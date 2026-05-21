//go:build integration

package main_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/bchatard/bifrost/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var binaryPath string

func TestMain(m *testing.M) {
	path, err := testutil.BuildBifrostBinary()
	if err != nil {
		fmt.Fprintf(os.Stderr, "build bifrost: %v\n", err)
		os.Exit(1)
	}

	binaryPath = path
	os.Exit(m.Run())
}

func TestHelp(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, binaryPath)

	result, err := c.RunBifrost(ctx, "--help")
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Output, "bifrost")
}

func TestVersion(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewContainer(ctx, t, binaryPath)

	result, err := c.RunBifrost(ctx, "--version")
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
}
