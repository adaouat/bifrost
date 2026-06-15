//go:build integration

package cmd_test

import (
	"context"
	"testing"

	"github.com/adaouat/bifrost/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPlainOutput_NoColorCodes verifies that --output plain suppresses foreground/
// background ANSI color codes in all output, including fang error rendering.
// TTY_FORCE=1 is used so color detection is active without a real terminal.
func TestPlainOutput_NoColorCodes(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	// With forced TTY and no output override, errors should contain color codes.
	res, err := c.Exec(ctx, []string{
		"sh", "-c", "TTY_FORCE=1 TERM=xterm /usr/local/bin/bifrost deploy",
	})
	require.NoError(t, err)
	require.NotEmpty(t, res.Stderr, "deploy with no --artifact must error to stderr")
	assert.Contains(t, res.Stderr, "\x1b[", "errors should have ANSI codes in default mode with forced TTY")

	// --output plain must suppress color ANSI codes even with forced TTY.
	// Fang uses codes like \x1b[101m (bright-red bg) and combined \x1b[1;97;101m.
	res, err = c.Exec(ctx, []string{
		"sh", "-c", "TTY_FORCE=1 TERM=xterm /usr/local/bin/bifrost --output plain deploy",
	})
	require.NoError(t, err)
	require.NotEmpty(t, res.Stderr, "deploy with no --artifact must error to stderr")
	// \x1b[10x (background 100-107) and combined \x1b[1;... (multi-param SGR)
	// are the color patterns used by fang's error panel.
	assert.NotContains(t, res.Stderr, "\x1b[10", "plain mode must suppress 10x background color codes")
	assert.NotContains(t, res.Stderr, "\x1b[1;", "plain mode must suppress combined SGR with color params")
}

// TestNoColor_SuppressesColors verifies that the NO_COLOR env var disables ANSI color
// codes in bifrost output. TTY_FORCE=1 is used to enable color detection in the
// non-TTY container environment so we can observe the difference NO_COLOR makes.
//
// NO_COLOR spec: colors are disabled; text decoration (bold, italic) may remain.
// We check for the absence of foreground/background color SGR sequences (30–37, 90–97)
// rather than all ANSI codes, since bold (\x1b[1m) and resets (\x1b[m) are retained.
func TestNoColor_SuppressesColors(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, bifrostBin)

	// TTY_FORCE=1 + TERM=xterm: colorprofile treats the session as a TTY and
	// enables ANSI colors. Trigger an error (missing --env flag) so fang renders
	// a styled error message to stderr.
	res, err := c.Exec(ctx, []string{
		"sh", "-c", "TTY_FORCE=1 TERM=xterm /usr/local/bin/bifrost deploy",
	})
	require.NoError(t, err)
	require.NotEmpty(t, res.Stderr, "deploy with no --artifact must error to stderr")
	// Fang uses colored text for the error panel; expect at least one SGR color code.
	assert.Contains(t, res.Stderr, "\x1b[", "error output should contain ANSI codes with forced TTY")

	// NO_COLOR=1 must suppress foreground/background color codes even with forced TTY.
	// Bold (\x1b[1m) and reset (\x1b[m) may remain per the NO_COLOR spec.
	res, err = c.Exec(ctx, []string{
		"sh", "-c", "TTY_FORCE=1 TERM=xterm NO_COLOR=1 /usr/local/bin/bifrost deploy",
	})
	require.NoError(t, err)
	require.NotEmpty(t, res.Stderr, "deploy with no --artifact must error to stderr")
	// Fang uses \x1b[101m (bright-red bg) and combined \x1b[1;97;101m.
	// With NO_COLOR, those codes should be absent; only \x1b[m (reset) and
	// \x1b[1m (bold) remain per the NO_COLOR spec.
	assert.NotContains(t, res.Stderr, "\x1b[10", "should have no 10x background color codes with NO_COLOR=1")
	assert.NotContains(t, res.Stderr, "\x1b[1;", "should have no combined SGR with color params with NO_COLOR=1")
}
