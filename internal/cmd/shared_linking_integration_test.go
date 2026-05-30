//go:build integration

package cmd_test

import (
	"context"
	"os"
	"testing"

	"github.com/adaouat/bifrost/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLinkShared_SeedsReleaseDirToShared verifies that when a shared path is
// absent from the shared root but present in the extracted release, the
// algorithm moves (seeds) it to the shared root before creating the symlink.
func TestLinkShared_SeedsReleaseDirToShared(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	cfg, err := os.ReadFile("../../testdata/bifrost-shared-seeding-test.yml")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, cfg, "/tmp/bifrost-seeding.yml", 0o644))

	artifact, err := os.ReadFile("../../testdata/release.tar.gz")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, artifact, "/tmp/release.tar.gz", 0o644))

	// Deploy without pre-creating /var/shared/config — the tarball contains
	// ./config/app.yml so the algorithm must seed it into shared.
	result, err := c.RunBifrost(ctx,
		"deploy",
		"--config", "/tmp/bifrost-seeding.yml",
		"--env", "test",
		"--app", "app",
		"--artifact", "/tmp/release.tar.gz",
		"--release-name", "seed-r1",
		"--init",
	)
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode, "deploy output:\n%s\nstderr:\n%s", result.Output, result.Stderr)

	// The seeded file must be present in the shared root.
	res, err := c.Exec(ctx, []string{"test", "-f", "/var/shared/config/app.yml"})
	require.NoError(t, err)
	assert.Equal(t, 0, res.ExitCode, "config/app.yml should have been seeded into shared root")

	// The release dir must have a symlink pointing to the shared root.
	res, err = c.Exec(ctx, []string{"readlink", "/var/releases/seed-r1/config"})
	require.NoError(t, err)
	assert.Equal(t, "/var/shared/config", res.Output, "release config dir should be symlinked to shared root")
}

// TestLinkShared_TargetExists_SymlinksWithoutSeeding verifies that when the
// shared target already exists, the algorithm replaces any existing link in
// the release dir with a fresh symlink to the shared root (no seeding).
func TestLinkShared_TargetExists_SymlinksWithoutSeeding(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	cfg, err := os.ReadFile("../../testdata/bifrost-deploy-int-test.yml")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, cfg, "/tmp/bifrost.yml", 0o644))

	artifact, err := os.ReadFile("../../testdata/release.tar.gz")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, artifact, "/tmp/release.tar.gz", 0o644))

	// First deploy creates the shared target.
	result, err := c.RunBifrost(ctx,
		"deploy",
		"--config", "/tmp/bifrost.yml",
		"--env", "test",
		"--app", "app",
		"--artifact", "/tmp/release.tar.gz",
		"--release-name", "exists-r1",
		"--init",
	)
	require.NoError(t, err)
	require.Equal(t, 0, result.ExitCode, "first deploy:\n%s", result.Output)

	// Second deploy: /var/shared/var/log already exists — exercise "target exists" branch.
	result, err = c.RunBifrost(ctx,
		"deploy",
		"--config", "/tmp/bifrost.yml",
		"--env", "test",
		"--app", "app",
		"--artifact", "/tmp/release.tar.gz",
		"--release-name", "exists-r2",
	)
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode, "second deploy:\n%s", result.Output)

	// Both releases must have symlinks to the same shared target.
	for _, release := range []string{"exists-r1", "exists-r2"} {
		res, err := c.Exec(ctx, []string{"readlink", "/var/releases/" + release + "/var/log"})
		require.NoError(t, err)
		assert.Equal(t, "/var/shared/var/log", res.Output, "%s/var/log should link to shared", release)
	}
}

// TestLinkShared_AbsentTarget_CreatesSharedDir verifies that when a shared
// directory path is absent from both the shared root and the release dir, the
// algorithm creates an empty directory in the shared root and symlinks it.
func TestLinkShared_AbsentTarget_CreatesSharedDir(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	cfg, err := os.ReadFile("../../testdata/bifrost-deploy-int-test.yml")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, cfg, "/tmp/bifrost.yml", 0o644))

	artifact, err := os.ReadFile("../../testdata/release.tar.gz")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, artifact, "/tmp/release.tar.gz", 0o644))

	// The tarball does not contain var/log; /var/shared/var/log does not exist.
	// The algorithm must create an empty shared dir and symlink it.
	result, err := c.RunBifrost(ctx,
		"deploy",
		"--config", "/tmp/bifrost.yml",
		"--env", "test",
		"--app", "app",
		"--artifact", "/tmp/release.tar.gz",
		"--release-name", "absdir-r1",
		"--init",
	)
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode, "deploy output:\n%s", result.Output)

	// Shared dir was created (it is a real directory, not a symlink).
	res, err := c.Exec(ctx, []string{"test", "-d", "/var/shared/var/log"})
	require.NoError(t, err)
	assert.Equal(t, 0, res.ExitCode, "/var/shared/var/log should be a directory")

	// Release dir entry is a symlink to the shared dir.
	res, err = c.Exec(ctx, []string{"readlink", "/var/releases/absdir-r1/var/log"})
	require.NoError(t, err)
	assert.Equal(t, "/var/shared/var/log", res.Output)
}

// TestLinkShared_AbsentTarget_CreatesSharedFile verifies that when a shared
// file path is absent from both the shared root and the release dir, the
// algorithm creates an empty file in the shared root and symlinks it.
func TestLinkShared_AbsentTarget_CreatesSharedFile(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	cfg, err := os.ReadFile("../../testdata/bifrost-deploy-int-test.yml")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, cfg, "/tmp/bifrost.yml", 0o644))

	artifact, err := os.ReadFile("../../testdata/release.tar.gz")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, artifact, "/tmp/release.tar.gz", 0o644))

	// The tarball does not contain .env; /var/shared/.env does not exist.
	// The algorithm must create an empty shared file and symlink it.
	result, err := c.RunBifrost(ctx,
		"deploy",
		"--config", "/tmp/bifrost.yml",
		"--env", "test",
		"--app", "app",
		"--artifact", "/tmp/release.tar.gz",
		"--release-name", "absfile-r1",
		"--init",
	)
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode, "deploy output:\n%s", result.Output)

	// Shared file was created (it is a regular file).
	res, err := c.Exec(ctx, []string{"test", "-f", "/var/shared/.env"})
	require.NoError(t, err)
	assert.Equal(t, 0, res.ExitCode, "/var/shared/.env should be a regular file")

	// Release dir entry is a symlink to the shared file.
	res, err = c.Exec(ctx, []string{"readlink", "/var/releases/absfile-r1/.env"})
	require.NoError(t, err)
	assert.Equal(t, "/var/shared/.env", res.Output)
}
