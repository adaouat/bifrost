package atomic_test

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/bifrost/internal/strategy/atomic"
)

// writeTarGzSymlink writes a .tar.gz at archivePath containing a single symlink
// entry named linkName that points at target.
func writeTarGzSymlink(t *testing.T, archivePath, linkName, target string) {
	t.Helper()
	f, err := os.Create(archivePath)
	require.NoError(t, err)
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name:     linkName,
		Linkname: target,
		Typeflag: tar.TypeSymlink,
		Mode:     0o777,
	}))
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	require.NoError(t, f.Close())
}

// TestExtract_RejectsAbsoluteSymlinkTarget proves an archive symlink pointing at
// an absolute path is rejected: a later entry written under it would escape the
// release directory.
func TestExtract_RejectsAbsoluteSymlinkTarget(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "evil.tar.gz")
	writeTarGzSymlink(t, archivePath, "link", "/etc/passwd")

	dest := filepath.Join(dir, "release")
	err := atomic.Extract(context.Background(), archivePath, dest, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsafe symlink")

	_, statErr := os.Lstat(filepath.Join(dest, "link"))
	assert.True(t, os.IsNotExist(statErr), "the unsafe symlink must not be created")
}

// TestExtract_RejectsEscapingSymlinkTarget proves a relative symlink target that
// climbs out of the release directory via .. is rejected.
func TestExtract_RejectsEscapingSymlinkTarget(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "evil.tar.gz")
	writeTarGzSymlink(t, archivePath, "link", "../../../../etc/passwd")

	dest := filepath.Join(dir, "release")
	err := atomic.Extract(context.Background(), archivePath, dest, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsafe symlink")
}

// tarEntry describes one entry to write via writeTarGzEntries.
type tarEntry struct {
	name     string
	linkname string
	content  string
	typeflag byte
}

// writeTarGzEntries writes a .tar.gz at archivePath containing entries, in order.
func writeTarGzEntries(t *testing.T, archivePath string, entries []tarEntry) {
	t.Helper()
	f, err := os.Create(archivePath)
	require.NoError(t, err)
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	for _, e := range entries {
		hdr := &tar.Header{
			Name:     e.name,
			Typeflag: e.typeflag,
			Mode:     0o777,
		}
		if e.typeflag == tar.TypeSymlink {
			hdr.Linkname = e.linkname
		} else {
			hdr.Size = int64(len(e.content))
		}
		require.NoError(t, tw.WriteHeader(hdr))
		if e.typeflag == tar.TypeReg {
			_, err := tw.Write([]byte(e.content))
			require.NoError(t, err)
		}
	}
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	require.NoError(t, f.Close())
}

// TestExtract_RejectsSymlinkChainEscape proves a two-link chain that defeats a
// purely lexical per-entry check is still rejected: symlink "a" -> "." looks
// contained in isolation, and so does symlink "a/b" -> ".." (Join(dest/a, "..")
// lexically resolves back to dest) — but the OS follows "a" first, so "b" is
// actually created one level above dest. A later entry under "b" would then
// write outside the release directory.
func TestExtract_RejectsSymlinkChainEscape(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "evil-chain.tar.gz")
	writeTarGzEntries(t, archivePath, []tarEntry{
		{name: "a", typeflag: tar.TypeSymlink, linkname: "."},
		{name: "a/b", typeflag: tar.TypeSymlink, linkname: ".."},
		{name: "b/evil.sh", typeflag: tar.TypeReg, content: "evil"},
	})

	dest := filepath.Join(dir, "release")
	require.NoError(t, os.MkdirAll(dest, 0o755))

	err := atomic.Extract(context.Background(), archivePath, dest, nil)
	require.Error(t, err)

	// Without the ancestor-symlink guard, the OS resolves "a" (a symlink to
	// ".") while creating "a/b", so the second symlink actually lands at
	// dest/b, not dest/a/b — and a later entry under "b" then escapes through
	// it to dir/evil.sh, one level above the release directory.
	_, statErr := os.Lstat(filepath.Join(dir, "evil.sh"))
	assert.True(t, os.IsNotExist(statErr), "no file must have escaped the release directory")
}

// TestExtract_AllowsNestedDirectoriesWithoutSymlinkAncestors proves the
// ancestor-symlink guard doesn't over-restrict ordinary nested paths that
// contain no symlinks at all.
func TestExtract_AllowsNestedDirectoriesWithoutSymlinkAncestors(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "nested.tar.gz")
	writeTarGzEntries(t, archivePath, []tarEntry{
		{name: "foo/bar/baz.txt", typeflag: tar.TypeReg, content: "hello"},
	})

	dest := filepath.Join(dir, "release")
	err := atomic.Extract(context.Background(), archivePath, dest, nil)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dest, "foo", "bar", "baz.txt"))
	require.NoError(t, err)
	assert.Equal(t, "hello", string(data))
}

// TestExtract_AllowsSafeRelativeSymlink proves a relative symlink that resolves
// inside the release directory is still extracted normally (no over-restriction).
func TestExtract_AllowsSafeRelativeSymlink(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "ok.tar.gz")
	writeTarGzSymlink(t, archivePath, "link", "shared/config")

	dest := filepath.Join(dir, "release")
	require.NoError(t, os.MkdirAll(dest, 0o755))

	err := atomic.Extract(context.Background(), archivePath, dest, nil)
	require.NoError(t, err)

	target, err := os.Readlink(filepath.Join(dest, "link"))
	require.NoError(t, err)
	assert.Equal(t, "shared/config", target)
}
