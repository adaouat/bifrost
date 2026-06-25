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
