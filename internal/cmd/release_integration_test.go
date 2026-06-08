//go:build integration

package cmd_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/adaouat/bifrost/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReleaseListCmd(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	cfg, err := os.ReadFile("../../testdata/bifrost-deploy-int-test.yml")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, cfg, "/tmp/bifrost.yml", 0o644))

	artifact, err := os.ReadFile("../../testdata/release.tar.gz")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, artifact, "/tmp/release.tar.gz", 0o644))

	for _, name := range []string{"r1", "r2"} {
		result, err := c.RunBifrost(ctx,
			"deploy",
			"--config", "/tmp/bifrost.yml",
			"--env", "test",
			"--app", "app",
			"--artifact", "/tmp/release.tar.gz",
			"--release-name", name,
			"--init",
		)
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode, "deploy %s:\n%s", name, result.Output)
	}

	result, err := c.RunBifrost(ctx,
		"release", "list",
		"--config", "/tmp/bifrost.yml",
		"--env", "test",
		"--app", "app",
	)
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode, "release list output:\n%s", result.Output)
	assert.Contains(t, result.Output, "r1")
	assert.Contains(t, result.Output, "r2")
	// r2 is the most recently deployed — should appear first (newest-first order)
	assert.Less(t, strings.Index(result.Output, "r2"), strings.Index(result.Output, "r1"))
	// M8 human-mode output format
	assert.Contains(t, result.Output, "Releases for test › app")
	assert.Contains(t, result.Output, "← current")
}

func TestReleaseListCmd_JSONOutput(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	cfg, err := os.ReadFile("../../testdata/bifrost-deploy-int-test.yml")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, cfg, "/tmp/bifrost.yml", 0o644))

	artifact, err := os.ReadFile("../../testdata/release.tar.gz")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, artifact, "/tmp/release.tar.gz", 0o644))

	for _, name := range []string{"r1", "r2"} {
		result, err := c.RunBifrost(ctx,
			"deploy",
			"--config", "/tmp/bifrost.yml",
			"--env", "test",
			"--app", "app",
			"--artifact", "/tmp/release.tar.gz",
			"--release-name", name,
			"--init",
		)
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode, "deploy %s:\n%s", name, result.Output)
	}

	result, err := c.RunBifrost(ctx,
		"release", "list",
		"--output", "json",
		"--config", "/tmp/bifrost.yml",
		"--env", "test",
		"--app", "app",
	)
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode, "release list --output json:\n%s", result.Output)

	var releases []map[string]any
	require.NoError(t, json.Unmarshal([]byte(result.Output), &releases), "output must be a JSON array")
	require.Len(t, releases, 2)
	assert.Equal(t, "r2", releases[0]["name"])
	assert.Equal(t, true, releases[0]["active"])
	assert.Equal(t, "r1", releases[1]["name"])
	assert.Equal(t, false, releases[1]["active"])
}

func TestReleaseRollbackCmd_SwitchesToPrevious(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	cfg, err := os.ReadFile("../../testdata/bifrost-deploy-int-test.yml")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, cfg, "/tmp/bifrost.yml", 0o644))

	artifact, err := os.ReadFile("../../testdata/release.tar.gz")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, artifact, "/tmp/release.tar.gz", 0o644))

	for _, name := range []string{"r1", "r2"} {
		result, err := c.RunBifrost(ctx,
			"deploy",
			"--config", "/tmp/bifrost.yml",
			"--env", "test",
			"--app", "app",
			"--artifact", "/tmp/release.tar.gz",
			"--release-name", name,
			"--init",
		)
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode, "deploy %s:\n%s", name, result.Output)
	}

	res, err := c.Exec(ctx, []string{"readlink", "/var/releases/current"})
	require.NoError(t, err)
	assert.Equal(t, "/var/releases/r2", res.Output, "current should point to r2 after two deploys")

	result, err := c.RunBifrost(ctx,
		"release", "rollback",
		"--config", "/tmp/bifrost.yml",
		"--env", "test",
		"--app", "app",
	)
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode, "release rollback output:\n%s", result.Output)

	res, err = c.Exec(ctx, []string{"readlink", "/var/releases/current"})
	require.NoError(t, err)
	assert.Equal(t, "/var/releases/r1", res.Output, "current should point to r1 after rollback")
}

func TestReleaseRollbackCmd_JSONOutput(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	cfg, err := os.ReadFile("../../testdata/bifrost-deploy-int-test.yml")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, cfg, "/tmp/bifrost.yml", 0o644))

	artifact, err := os.ReadFile("../../testdata/release.tar.gz")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, artifact, "/tmp/release.tar.gz", 0o644))

	for _, name := range []string{"r1", "r2"} {
		result, err := c.RunBifrost(ctx,
			"deploy",
			"--config", "/tmp/bifrost.yml",
			"--env", "test",
			"--app", "app",
			"--artifact", "/tmp/release.tar.gz",
			"--release-name", name,
			"--init",
		)
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode, "deploy %s:\n%s", name, result.Output)
	}

	result, err := c.RunBifrost(ctx,
		"release", "rollback",
		"--output", "json",
		"--config", "/tmp/bifrost.yml",
		"--env", "test",
		"--app", "app",
	)
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode, "release rollback --output json:\n%s", result.Output)

	var ev map[string]any
	require.NoError(t, json.Unmarshal([]byte(result.Output), &ev), "output must be a single JSON object")
	assert.Equal(t, "rollback", ev["event"])
	assert.Equal(t, "r1", ev["release"])
	assert.Equal(t, "done", ev["status"])
}

func TestReleaseRollbackCmd_NoPreviousRelease(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	cfg, err := os.ReadFile("../../testdata/bifrost-deploy-int-test.yml")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, cfg, "/tmp/bifrost.yml", 0o644))

	artifact, err := os.ReadFile("../../testdata/release.tar.gz")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, artifact, "/tmp/release.tar.gz", 0o644))

	result, err := c.RunBifrost(ctx,
		"deploy",
		"--config", "/tmp/bifrost.yml",
		"--env", "test",
		"--app", "app",
		"--artifact", "/tmp/release.tar.gz",
		"--release-name", "only-release",
		"--init",
	)
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode, "deploy output:\n%s", result.Output)

	result, err = c.RunBifrost(ctx,
		"release", "rollback",
		"--config", "/tmp/bifrost.yml",
		"--env", "test",
		"--app", "app",
	)
	require.NoError(t, err)
	assert.Equal(t, 3, result.ExitCode, "rollback with single release should exit 3:\n%s", result.Output)
}

func TestReleaseActivateCmd_AlreadyCurrent_NonInteractive(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	cfg, err := os.ReadFile("../../testdata/bifrost-deploy-int-test.yml")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, cfg, "/tmp/bifrost.yml", 0o644))

	artifact, err := os.ReadFile("../../testdata/release.tar.gz")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, artifact, "/tmp/release.tar.gz", 0o644))

	result, err := c.RunBifrost(ctx,
		"deploy",
		"--config", "/tmp/bifrost.yml",
		"--env", "test",
		"--app", "app",
		"--artifact", "/tmp/release.tar.gz",
		"--release-name", "r1",
		"--init",
	)
	require.NoError(t, err)
	require.Equal(t, 0, result.ExitCode, "deploy r1:\n%s", result.Output)

	// Activate r1 again — it is already current; non-TTY must not prompt.
	result, err = c.RunBifrost(ctx,
		"release", "activate",
		"--config", "/tmp/bifrost.yml",
		"--env", "test",
		"--app", "app",
		"--release", "r1",
	)
	require.NoError(t, err)
	assert.Equal(t, 3, result.ExitCode, "should exit 3 when release is already current on non-TTY")
}

func TestReleaseActivateCmd_DryRun(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	cfg, err := os.ReadFile("../../testdata/bifrost-deploy-int-test.yml")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, cfg, "/tmp/bifrost.yml", 0o644))

	artifact, err := os.ReadFile("../../testdata/release.tar.gz")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, artifact, "/tmp/release.tar.gz", 0o644))

	for _, name := range []string{"r1", "r2"} {
		result, err := c.RunBifrost(ctx,
			"deploy",
			"--config", "/tmp/bifrost.yml",
			"--env", "test",
			"--app", "app",
			"--artifact", "/tmp/release.tar.gz",
			"--release-name", name,
			"--init",
		)
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode, "deploy %s:\n%s", name, result.Output)
	}

	// r2 is current; dry-run activate r1 — current must not change.
	result, err := c.RunBifrost(ctx,
		"release", "activate",
		"--dry-run",
		"--config", "/tmp/bifrost.yml",
		"--env", "test",
		"--app", "app",
		"--release", "r1",
	)
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode, "dry-run activate:\n%s\nstderr:\n%s", result.Output, result.Stderr)
	assert.Contains(t, result.Output, "DRY RUN")

	res, err := c.Exec(ctx, []string{"readlink", "/var/releases/current"})
	require.NoError(t, err)
	assert.Equal(t, "/var/releases/r2", res.Output, "current should still point to r2 after dry-run")
}

func TestReleaseActivateCmd_NoRelease_NonInteractive(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	cfg, err := os.ReadFile("../../testdata/bifrost-deploy-int-test.yml")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, cfg, "/tmp/bifrost.yml", 0o644))

	artifact, err := os.ReadFile("../../testdata/release.tar.gz")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, artifact, "/tmp/release.tar.gz", 0o644))

	result, err := c.RunBifrost(ctx,
		"deploy",
		"--config", "/tmp/bifrost.yml",
		"--env", "test",
		"--app", "app",
		"--artifact", "/tmp/release.tar.gz",
		"--release-name", "r1",
		"--init",
	)
	require.NoError(t, err)
	require.Equal(t, 0, result.ExitCode, "deploy:\n%s", result.Output)

	result, err = c.RunBifrost(ctx,
		"release", "activate",
		"--config", "/tmp/bifrost.yml",
		"--env", "test",
		"--app", "app",
	)
	require.NoError(t, err)
	assert.NotEqual(t, 0, result.ExitCode, "should fail without --release in non-interactive mode")
	assert.Contains(t, result.Stderr, "non-interactive")
}

func TestReleaseActivateCmd_NonExistentRelease(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	cfg, err := os.ReadFile("../../testdata/bifrost-deploy-int-test.yml")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, cfg, "/tmp/bifrost.yml", 0o644))

	artifact, err := os.ReadFile("../../testdata/release.tar.gz")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, artifact, "/tmp/release.tar.gz", 0o644))

	result, err := c.RunBifrost(ctx,
		"deploy",
		"--config", "/tmp/bifrost.yml",
		"--env", "test",
		"--app", "app",
		"--artifact", "/tmp/release.tar.gz",
		"--release-name", "r1",
		"--init",
	)
	require.NoError(t, err)
	require.Equal(t, 0, result.ExitCode, "deploy:\n%s", result.Output)

	result, err = c.RunBifrost(ctx,
		"release", "activate",
		"--config", "/tmp/bifrost.yml",
		"--env", "test",
		"--app", "app",
		"--release", "does-not-exist",
	)
	require.NoError(t, err)
	assert.NotEqual(t, 0, result.ExitCode, "activating a non-existent release should fail")
}

func TestReleaseActivateCmd_SwitchesCurrent(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	cfg, err := os.ReadFile("../../testdata/bifrost-deploy-int-test.yml")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, cfg, "/tmp/bifrost.yml", 0o644))

	artifact, err := os.ReadFile("../../testdata/release.tar.gz")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, artifact, "/tmp/release.tar.gz", 0o644))

	for _, name := range []string{"r1", "r2"} {
		result, err := c.RunBifrost(ctx,
			"deploy",
			"--config", "/tmp/bifrost.yml",
			"--env", "test",
			"--app", "app",
			"--artifact", "/tmp/release.tar.gz",
			"--release-name", name,
			"--init",
		)
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode, "deploy %s:\n%s", name, result.Output)
	}

	res, err := c.Exec(ctx, []string{"readlink", "/var/releases/current"})
	require.NoError(t, err)
	assert.Equal(t, "/var/releases/r2", res.Output, "current should point to r2 after two deploys")

	result, err := c.RunBifrost(ctx,
		"release", "activate",
		"--config", "/tmp/bifrost.yml",
		"--env", "test",
		"--app", "app",
		"--release", "r1",
	)
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode, "release activate output:\n%s", result.Output)

	res, err = c.Exec(ctx, []string{"readlink", "/var/releases/current"})
	require.NoError(t, err)
	assert.Equal(t, "/var/releases/r1", res.Output, "current should point to r1 after activate")
}
