//go:build integration

package cmd_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/bifrost/internal/testutil"
)

// TestDeployCmd_SIGINTAfterExtract_WarnsAndCompletes proves that a Ctrl+C
// arriving after extraction does not abort the deploy: it runs to
// completion and prints a warning that cancellation is being ignored.
func TestDeployCmd_SIGINTAfterExtract_WarnsAndCompletes(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	cfg, err := os.ReadFile("../../testdata/bifrost-deploy-cancel-test.yml")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, cfg, "/tmp/bifrost-cancel.yml", 0o644))

	artifact, err := os.ReadFile("../../testdata/release.tar.gz")
	require.NoError(t, err)
	require.NoError(t, c.CopyFile(ctx, artifact, "/tmp/release.tar.gz", 0o644))

	// Run the deploy in the background, redirected to a file. The
	// post_extract hook sleeps for 2s, giving us a window to deliver SIGINT
	// after extraction has completed.
	startRes, err := c.Exec(ctx, []string{"sh", "-c",
		"/usr/local/bin/bifrost --output plain deploy " +
			"--config /tmp/bifrost-cancel.yml --env test --app app " +
			"--artifact /tmp/release.tar.gz --release-name cancel-r1 --init " +
			"</dev/null >/tmp/out.log 2>/tmp/err.log &"})
	require.NoError(t, err)
	require.Equal(t, 0, startRes.ExitCode, "starting background deploy:\n%s", startRes.Stderr)

	// Give the deploy time to pass extraction and enter the post_extract
	// hook's sleep, then send SIGINT directly to the bifrost process.
	time.Sleep(500 * time.Millisecond)
	sigRes, err := c.Exec(ctx, []string{"sh", "-c",
		`for p in /proc/[0-9]*; do ` +
			`[ "$(cat $p/comm 2>/dev/null)" = "bifrost" ] && kill -INT ${p#/proc/}; ` +
			`done; true`})
	require.NoError(t, err)
	require.Equal(t, 0, sigRes.ExitCode, "sending SIGINT:\n%s", sigRes.Stderr)

	// Wait for the deploy to finish.
	var out string
	require.Eventually(t, func() bool {
		res, err := c.Exec(ctx, []string{"cat", "/tmp/out.log"})
		if err != nil || res.ExitCode != 0 {
			return false
		}
		out = res.Output
		return strings.Contains(out, "Deployed in")
	}, 10*time.Second, 200*time.Millisecond, "deploy did not finish")

	assert.Contains(t, out, "cannot be cancelled", "expected a cancellation warning in the output")

	// current must point to the new release: the deploy ran to completion.
	res, err := c.Exec(ctx, []string{"readlink", "/var/releases/current"})
	require.NoError(t, err)
	assert.Equal(t, "/var/releases/cancel-r1", res.Output)
}
