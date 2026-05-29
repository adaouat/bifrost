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

func TestInitCmd_CreatesRoots(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	cfg, err := os.ReadFile("../../testdata/bifrost-deploy-int-test.yml")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, cfg, "/tmp/bifrost.yml", 0o644))

	result, err := c.RunBifrost(ctx,
		"release", "init",
		"--config", "/tmp/bifrost.yml",
		"--env", "test",
		"--app", "app",
	)
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode, "init output:\n%s", result.Output)

	res, err := c.Exec(ctx, []string{"test", "-d", "/var/releases"})
	require.NoError(t, err)
	assert.Equal(t, 0, res.ExitCode, "releases_root should exist after init")

	res, err = c.Exec(ctx, []string{"test", "-d", "/var/shared"})
	require.NoError(t, err)
	assert.Equal(t, 0, res.ExitCode, "shared_root should exist after init")
}

func TestInitCmd_Idempotent(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	cfg, err := os.ReadFile("../../testdata/bifrost-deploy-int-test.yml")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, cfg, "/tmp/bifrost.yml", 0o644))

	for i := range 2 {
		result, err := c.RunBifrost(ctx,
			"release", "init",
			"--config", "/tmp/bifrost.yml",
			"--env", "test",
			"--app", "app",
		)
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode, "init run %d output:\n%s", i+1, result.Output)
	}
}
