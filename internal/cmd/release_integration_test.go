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
