//go:build integration

package cmd_test

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/bifrost/internal/testutil"
)

// TestExecWithTTY_BasicEcho validates that ExecWithTTY can run a simple command
// and that the connection is properly closed after the process exits.
func TestExecWithTTY_BasicEcho(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	res, err := c.ExecWithTTY(ctx, []string{"echo", "hello"}, nil, 0, 0)
	require.NoError(t, err)
	t.Logf("ExecWithTTY echo: output=%q exitCode=%d", res.Output, res.ExitCode)
	assert.Equal(t, 0, res.ExitCode)
	assert.Contains(t, res.Output, "hello")
}

// TestExecWithTTY_StdinInput validates that ExecWithTTY properly feeds stdin
// to the container process via the PTY.
func TestExecWithTTY_StdinInput(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	input := []byte("hello\r")
	res, err := c.ExecWithTTY(ctx, []string{"sh", "-c", "read -r line; echo got:$line"}, input, 0, 0)
	require.NoError(t, err)
	t.Logf("ExecWithTTY stdin: output=%q exitCode=%d", res.Output, res.ExitCode)
	assert.Equal(t, 0, res.ExitCode)
	// The PTY echoes input, so stdout contains the echoed chars plus output
	assert.Contains(t, res.Output, "hello")
}

// TestDeployCmd_Interactive_ScriptedConfirm proves that `bifrost deploy
// --interactive` completes a full deploy normally when the operator confirms
// every step. A Docker exec TTY is used so the binary sees IsTTY()=true;
// repeated CR bytes drive each huh confirm prompt to accept its default (yes).
func TestDeployCmd_Interactive_ScriptedConfirm(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	cfg, err := os.ReadFile("../../testdata/bifrost-deploy-int-test.yml")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, cfg, "/tmp/bifrost-int.yml", 0o644))

	artifact, err := os.ReadFile("../../testdata/release.tar.gz")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, artifact, "/tmp/release.tar.gz", 0o644))

	// Feed one CR per confirm step (7 base + buffer). Each CR is written with a
	// 300ms interval via ExecWithTTY so bubbletea's read goroutine receives
	// exactly one CR per huh.Run() call — sending all at once would let the first
	// program consume every CR in a single Read(4096), leaving subsequent prompts
	// with empty stdin.
	input := bytes.Repeat([]byte{'\r'}, 20)

	res, err := c.ExecWithTTY(ctx, []string{
		"/usr/local/bin/bifrost", "deploy",
		"--interactive",
		"--config", "/tmp/bifrost-int.yml",
		"--env", "test",
		"--app", "app",
		"--artifact", "/tmp/release.tar.gz",
		"--release-name", "interactive-r1",
		"--init",
	}, input, 1000*time.Millisecond, 300*time.Millisecond)
	require.NoError(t, err)
	t.Logf("deploy output: %q", res.Output)
	assert.Equal(t, 0, res.ExitCode,
		"--interactive deploy should complete when all steps confirmed:\n%s", res.Output)
	assert.True(t, strings.Contains(res.Output, "Deployed in") || strings.Contains(res.Output, "interactive-r1"),
		"expected deploy completion in output:\n%s", res.Output)

	// current symlink must point to the new release.
	link, err := c.Exec(ctx, []string{"readlink", "/var/releases/current"})
	require.NoError(t, err)
	assert.Equal(t, "/var/releases/interactive-r1", link.Output)
}
