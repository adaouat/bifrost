package transport

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Platform identifies a remote OS/arch using goreleaser naming.
type Platform struct {
	OS   string
	Arch string
}

// DetectPlatform runs uname over SSH and maps the result to a Platform.
func DetectPlatform(client *Client) (Platform, error) {
	res, err := client.Exec("uname -s && uname -m")
	if err != nil {
		return Platform{}, err
	}
	if res.ExitCode != 0 {
		return Platform{}, fmt.Errorf("detecting remote platform: uname exited %d: %s", res.ExitCode, res.Stderr.String())
	}
	fields := strings.Fields(res.Stdout.String())
	if len(fields) < 2 {
		return Platform{}, fmt.Errorf("detecting remote platform: unexpected uname output %q", res.Stdout.String())
	}
	return mapPlatform(fields[0], fields[1])
}

// ResolveAgentBinary returns a local path to the agent binary for p at the given
// version, downloading and caching it from GitHub Releases on a cache miss.
func ResolveAgentBinary(version string, p Platform) (string, error) {
	dest, err := cachePath(version, p)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(dest); err == nil {
		return dest, nil
	}

	data, err := downloadAgent(version, p)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", fmt.Errorf("creating agent cache dir: %w", err)
	}
	if err := os.WriteFile(dest, data, 0o755); err != nil { //nolint:gosec // agent binary must be executable
		return "", fmt.Errorf("writing agent cache %s: %w", dest, err)
	}
	return dest, nil
}

// mapPlatform maps uname -s / uname -m output to a Platform.
func mapPlatform(unameS, unameM string) (Platform, error) {
	type key struct{ s, m string }
	table := map[key]Platform{
		{"Linux", "x86_64"}:  {OS: "linux", Arch: "amd64"},
		{"Linux", "aarch64"}: {OS: "linux", Arch: "arm64"},
		{"Darwin", "x86_64"}: {OS: "darwin", Arch: "amd64"},
		{"Darwin", "arm64"}:  {OS: "darwin", Arch: "arm64"},
	}
	p, ok := table[key{unameS, unameM}]
	if !ok {
		return Platform{}, fmt.Errorf("unsupported remote platform %s/%s: use --agent-binary to override", unameS, unameM)
	}
	return p, nil
}

// cacheKey is the cache-relative path for a version/platform binary.
func cacheKey(version string, p Platform) string {
	return filepath.Join("bifrost", "agents", version, p.OS+"_"+p.Arch, "bifrost")
}

// cachePath is the absolute local cache path for a version/platform binary.
func cachePath(version string, p Platform) (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("resolving user cache dir: %w", err)
	}
	return filepath.Join(base, cacheKey(version, p)), nil
}

func downloadURL(version string, p Platform) string {
	return fmt.Sprintf(
		"https://github.com/adaouat/bifrost/releases/download/v%s/bifrost_%s_%s_%s.tar.gz",
		version, version, p.OS, p.Arch,
	)
}

func downloadAgent(version string, p Platform) ([]byte, error) {
	url := downloadURL(version, p)
	resp, err := http.Get(url) //nolint:gosec // release URL is built from the trusted version + platform
	if err != nil {
		return nil, fmt.Errorf("downloading agent: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("downloading agent from %s: HTTP %d (no release for this version? use --agent-binary)", url, resp.StatusCode)
	}
	return extractBinary(resp.Body)
}

// extractBinary pulls the bifrost binary out of a goreleaser .tar.gz stream.
func extractBinary(r io.Reader) ([]byte, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("reading agent archive: %w", err)
	}
	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading agent archive: %w", err)
		}
		if hdr.Typeflag == tar.TypeReg && path.Base(hdr.Name) == "bifrost" {
			data, err := io.ReadAll(tr) //nolint:gosec // trusted release artifact
			if err != nil {
				return nil, fmt.Errorf("extracting agent binary: %w", err)
			}
			return data, nil
		}
	}
	return nil, fmt.Errorf("agent binary not found in archive")
}
