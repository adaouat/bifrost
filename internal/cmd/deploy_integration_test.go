//go:build integration

package cmd_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/bifrost/internal/testutil"
)

var bifrostBin string

func TestMain(m *testing.M) {
	path, err := testutil.BuildBifrostBinary()
	if err != nil {
		panic("BuildBifrostBinary: " + err.Error())
	}
	bifrostBin = path
	os.Exit(m.Run())
}

func TestDeployCmd_E2E(t *testing.T) {
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
		"--release-name", "test-r1",
		"--init",
	)
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode, "deploy output:\n%s", result.Output)

	// Extracted file is present in the release directory.
	res, err := c.Exec(ctx, []string{"test", "-f", "/var/releases/test-r1/public/index.html"})
	require.NoError(t, err)
	assert.Equal(t, 0, res.ExitCode, "public/index.html should exist in release")

	// current symlink points to the new release.
	res, err = c.Exec(ctx, []string{"readlink", "/var/releases/current"})
	require.NoError(t, err)
	assert.Equal(t, "/var/releases/test-r1", res.Output)

	// Shared directory is symlinked.
	res, err = c.Exec(ctx, []string{"readlink", "/var/releases/test-r1/var/log"})
	require.NoError(t, err)
	assert.Equal(t, "/var/shared/var/log", res.Output)

	// Shared file is symlinked.
	res, err = c.Exec(ctx, []string{"readlink", "/var/releases/test-r1/.env"})
	require.NoError(t, err)
	assert.Equal(t, "/var/shared/.env", res.Output)
}

func TestDeployCmd_InteractiveHook_SkippedOnNonTTY(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	cfg, err := os.ReadFile("../../testdata/bifrost-interactive-hook-test.yml")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, cfg, "/tmp/bifrost-interactive.yml", 0o644))

	artifact, err := os.ReadFile("../../testdata/release.tar.gz")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, artifact, "/tmp/release.tar.gz", 0o644))

	result, err := c.RunBifrost(ctx,
		"deploy",
		"--config", "/tmp/bifrost-interactive.yml",
		"--env", "test",
		"--app", "app",
		"--artifact", "/tmp/release.tar.gz",
		"--release-name", "interactive-r1",
		"--init",
	)
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode, "deploy with interactive hook on non-TTY:\n%s\nstderr:\n%s", result.Output, result.Stderr)

	// Non-interactive hook must have run.
	res, err := c.Exec(ctx, []string{"test", "-f", "/tmp/post_extract_ran"})
	require.NoError(t, err)
	assert.Equal(t, 0, res.ExitCode, "post_extract non-interactive hook should have run")

	// Interactive hook must NOT have run (skipped on non-TTY).
	res, err = c.Exec(ctx, []string{"test", "-f", "/tmp/interactive_would_run"})
	require.NoError(t, err)
	assert.NotEqual(t, 0, res.ExitCode, "interactive hook should have been skipped on non-TTY")

	// Warning about skipped interactive hook should appear in stdout.
	assert.Contains(t, result.Output, "skipping interactive hook")
}

func TestDeployCmd_JSONOutput(t *testing.T) {
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
		"--output", "json",
		"--config", "/tmp/bifrost.yml",
		"--env", "test",
		"--app", "app",
		"--artifact", "/tmp/release.tar.gz",
		"--release-name", "json-r1",
		"--init",
	)
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode, "deploy output:\n%s\nstderr:\n%s", result.Output, result.Stderr)

	// Every line of stdout must be valid JSON.
	for _, line := range strings.Split(strings.TrimSpace(result.Output), "\n") {
		if line == "" {
			continue
		}
		var ev map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &ev), "invalid JSON line: %s", line)
	}

	// Must include a deploy-done event with the release name.
	assert.Contains(t, result.Output, `"event":"done"`)
	assert.Contains(t, result.Output, `"step":"deploy"`)
	assert.Contains(t, result.Output, `"release":"json-r1"`)

	// Must have start/done events for link, current_symlink, and purge steps.
	assert.Contains(t, result.Output, `"step":"link"`)
	assert.Contains(t, result.Output, `"step":"current_symlink"`)
	assert.Contains(t, result.Output, `"step":"purge"`)
}

func TestDeployCmd_JSONOutput_ErrorEvent(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	cfg, err := os.ReadFile("../../testdata/bifrost-deploy-int-test.yml")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, cfg, "/tmp/bifrost.yml", 0o644))

	// Copy garbage bytes as the artifact to trigger an extraction failure.
	require.NoError(t, c.CopyFile(ctx, []byte("not a real archive"), "/tmp/bad.tar.gz", 0o644))

	// Create the release roots so the deploy gets past the roots check.
	for _, dir := range []string{"/var/releases", "/var/shared"} {
		res, err := c.Exec(ctx, []string{"mkdir", "-p", dir})
		require.NoError(t, err)
		require.Equal(t, 0, res.ExitCode)
	}

	result, err := c.RunBifrost(ctx,
		"deploy",
		"--output", "json",
		"--config", "/tmp/bifrost.yml",
		"--env", "test",
		"--app", "app",
		"--artifact", "/tmp/bad.tar.gz",
		"--release-name", "json-err-r1",
	)
	require.NoError(t, err)
	assert.NotEqual(t, 0, result.ExitCode, "deploy with bad artifact should fail")

	// stdout must contain a JSON error event for the extract step.
	assert.Contains(t, result.Output, `"event":"error"`)
	assert.Contains(t, result.Output, `"step":"extract"`)
}

func TestDeployCmd_DryRun(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	cfg, err := os.ReadFile("../../testdata/bifrost-deploy-int-test.yml")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, cfg, "/tmp/bifrost.yml", 0o644))

	artifact, err := os.ReadFile("../../testdata/release.tar.gz")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, artifact, "/tmp/release.tar.gz", 0o644))

	// Create the release roots manually (--init would normally create them, but we
	// also need them for dry-run to be able to read config correctly without creating
	// the release dir).
	for _, dir := range []string{"/var/releases", "/var/shared"} {
		res, err := c.Exec(ctx, []string{"mkdir", "-p", dir})
		require.NoError(t, err)
		require.Equal(t, 0, res.ExitCode)
	}

	result, err := c.RunBifrost(ctx,
		"deploy",
		"--dry-run",
		"--config", "/tmp/bifrost.yml",
		"--env", "test",
		"--app", "app",
		"--artifact", "/tmp/release.tar.gz",
		"--release-name", "dry-r1",
	)
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode, "dry-run deploy output:\n%s\nstderr:\n%s", result.Output, result.Stderr)
	assert.Contains(t, result.Output, "DRY RUN")

	// Release directory must NOT have been created.
	res, err := c.Exec(ctx, []string{"test", "-d", "/var/releases/dry-r1"})
	require.NoError(t, err)
	assert.NotEqual(t, 0, res.ExitCode, "release directory should not exist after dry run")
}

func TestDeployCmd_DryRun_WouldPurge(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	// releases_to_keep=2; deploy r1, r2 first, then dry-run r3 — r1 should appear in Would purge.
	cfg, err := os.ReadFile("../../testdata/bifrost-deploy-purge-test.yml")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, cfg, "/tmp/bifrost-purge.yml", 0o644))

	artifact, err := os.ReadFile("../../testdata/release.tar.gz")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, artifact, "/tmp/release.tar.gz", 0o644))

	for i, name := range []string{"dryp-r1", "dryp-r2"} {
		args := []string{
			"deploy",
			"--config", "/tmp/bifrost-purge.yml",
			"--env", "test",
			"--app", "app",
			"--artifact", "/tmp/release.tar.gz",
			"--release-name", name,
		}
		if i == 0 {
			args = append(args, "--init")
		}
		result, err := c.RunBifrost(ctx, args...)
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode, "deploy %s output:\n%s", name, result.Output)
	}

	result, err := c.RunBifrost(ctx,
		"deploy",
		"--dry-run",
		"--config", "/tmp/bifrost-purge.yml",
		"--env", "test",
		"--app", "app",
		"--artifact", "/tmp/release.tar.gz",
		"--release-name", "dryp-r3",
	)
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode, "dry-run output:\n%s\nstderr:\n%s", result.Output, result.Stderr)
	assert.Contains(t, result.Output, "Would purge")
	assert.Contains(t, result.Output, "dryp-r1")
}

func TestDeployCmd_DryRun_SudoHook(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	cfg, err := os.ReadFile("../../testdata/bifrost-deploy-sudo-test.yml")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, cfg, "/tmp/bifrost-sudo.yml", 0o644))

	artifact, err := os.ReadFile("../../testdata/release.tar.gz")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, artifact, "/tmp/release.tar.gz", 0o644))

	for _, dir := range []string{"/var/releases", "/var/shared"} {
		res, err := c.Exec(ctx, []string{"mkdir", "-p", dir})
		require.NoError(t, err)
		require.Equal(t, 0, res.ExitCode)
	}

	result, err := c.RunBifrost(ctx,
		"deploy",
		"--dry-run",
		"--config", "/tmp/bifrost-sudo.yml",
		"--env", "test",
		"--app", "app",
		"--artifact", "/tmp/release.tar.gz",
		"--release-name", "sudo-r1",
	)
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode, "dry-run output:\n%s\nstderr:\n%s", result.Output, result.Stderr)
	assert.Contains(t, result.Output, "systemctl reload nginx")
	assert.Contains(t, result.Output, "(sudo)")
}

func TestDeployCmd_Hooks(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	cfg, err := os.ReadFile("../../testdata/bifrost-deploy-hooks-test.yml")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, cfg, "/tmp/bifrost-hooks.yml", 0o644))

	artifact, err := os.ReadFile("../../testdata/release.tar.gz")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, artifact, "/tmp/release.tar.gz", 0o644))

	result, err := c.RunBifrost(ctx,
		"deploy",
		"--config", "/tmp/bifrost-hooks.yml",
		"--env", "test",
		"--app", "app",
		"--artifact", "/tmp/release.tar.gz",
		"--release-name", "hooks-r1",
		"--init",
	)
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode, "deploy output:\n%s", result.Output)

	for _, marker := range []string{
		"/tmp/post_extract_ran",
		"/tmp/pre_link_ran",
		"/tmp/pre_enable_ran",
		"/tmp/post_enable_ran",
	} {
		res, err := c.Exec(ctx, []string{"test", "-f", marker})
		require.NoError(t, err)
		assert.Equal(t, 0, res.ExitCode, "hook marker should exist: %s", marker)
	}
}

func TestDeployCmd_Purge(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	cfg, err := os.ReadFile("../../testdata/bifrost-deploy-purge-test.yml")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, cfg, "/tmp/bifrost-purge.yml", 0o644))

	artifact, err := os.ReadFile("../../testdata/release.tar.gz")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, artifact, "/tmp/release.tar.gz", 0o644))

	// releases_to_keep=2; deploy 3 releases — purge-r1 must be removed after purge-r3.
	for i, name := range []string{"purge-r1", "purge-r2", "purge-r3"} {
		initFlag := "--init"
		args := []string{
			"deploy",
			"--config", "/tmp/bifrost-purge.yml",
			"--env", "test",
			"--app", "app",
			"--artifact", "/tmp/release.tar.gz",
			"--release-name", name,
		}
		if i == 0 {
			args = append(args, initFlag)
		}
		result, err := c.RunBifrost(ctx, args...)
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode, "deploy %s:\n%s", name, result.Output)
	}

	// purge-r1 must have been removed (oldest, beyond the keep=2 window).
	res, err := c.Exec(ctx, []string{"test", "-d", "/var/releases/purge-r1"})
	require.NoError(t, err)
	assert.NotEqual(t, 0, res.ExitCode, "purge-r1 should have been purged")

	// purge-r2 and purge-r3 must still exist.
	for _, name := range []string{"purge-r2", "purge-r3"} {
		res, err = c.Exec(ctx, []string{"test", "-d", "/var/releases/" + name})
		require.NoError(t, err)
		assert.Equal(t, 0, res.ExitCode, "%s should still exist", name)
	}

	// current must point to the latest release.
	res, err = c.Exec(ctx, []string{"readlink", "/var/releases/current"})
	require.NoError(t, err)
	assert.Equal(t, "/var/releases/purge-r3", res.Output)
}
