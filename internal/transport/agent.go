package transport

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
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
	if err := verifyChecksum(version, p, data); err != nil {
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

var downloadBaseURL = "https://github.com/adaouat/bifrost/releases/download"

// httpClient bounds agent and checksum downloads so a hung server can't stall a deploy.
var httpClient = &http.Client{Timeout: 120 * time.Second}

// downloadURL builds the raw-binary release asset URL. The leading "v" is
// stripped so the tag path (v{ver}) and asset name ({ver}) match goreleaser.
func downloadURL(version string, p Platform) string {
	return fmt.Sprintf("%s/v%s/%s", downloadBaseURL, strings.TrimPrefix(version, "v"), assetName(version, p))
}

// assetName is the goreleaser release asset filename for a platform.
func assetName(version string, p Platform) string {
	v := strings.TrimPrefix(version, "v")
	return fmt.Sprintf("bifrost_%s_%s_%s", v, p.OS, p.Arch)
}

func downloadAgent(version string, p Platform) ([]byte, error) {
	url := downloadURL(version, p)
	resp, err := httpClient.Get(url) //nolint:gosec // release URL is built from the trusted version + platform
	if err != nil {
		return nil, fmt.Errorf("downloading agent: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("downloading agent from %s: HTTP %d (no release for this version? use --agent-binary)", url, resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body) //nolint:gosec // trusted release artifact
	if err != nil {
		return nil, fmt.Errorf("reading agent download: %w", err)
	}
	return data, nil
}

// verifyChecksum downloads the release checksums.txt and verifies that data's
// SHA-256 matches the digest recorded for the platform's asset. The agent runs
// remotely (often via sudo), so an unverified binary must never be cached.
func verifyChecksum(version string, p Platform, data []byte) error {
	url := checksumURL(version)
	resp, err := httpClient.Get(url) //nolint:gosec // release URL is built from the trusted version
	if err != nil {
		return fmt.Errorf("downloading checksums: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading checksums from %s: HTTP %d", url, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading checksums: %w", err)
	}

	name := assetName(version, p)
	want, err := checksumFor(string(body), name)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(data)
	if got := hex.EncodeToString(sum[:]); got != want {
		return fmt.Errorf("agent checksum mismatch for %s: expected %s, got %s", name, want, got)
	}
	return nil
}

// checksumFor returns the hex SHA-256 recorded for filename in goreleaser's
// checksums.txt ("<sha256>  <filename>" per line).
func checksumFor(checksums, filename string) (string, error) {
	for line := range strings.SplitSeq(checksums, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == filename {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("no checksum for %s in checksums.txt", filename)
}

func checksumURL(version string) string {
	return fmt.Sprintf("%s/v%s/checksums.txt", downloadBaseURL, strings.TrimPrefix(version, "v"))
}
