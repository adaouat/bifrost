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
