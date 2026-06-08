package release

import (
	"bytes"
	"io"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/bifrost/internal/cmderr"
	"github.com/adaouat/bifrost/internal/config"
	forgeui "github.com/adaouat/forge/ui"
)

func TestRunClientReleaseRollback_SequentialLoop(t *testing.T) {
	orig := releaseRollbackToServerFn
	defer func() { releaseRollbackToServerFn = orig }()

	var called []string
	releaseRollbackToServerFn = func(_ string, _ *config.Config, _, _, _ string, srv config.ResolvedServer, _ forgeui.Mode, _ io.Writer) error {
		called = append(called, srv.Name)
		return nil
	}

	merged := &config.MergedConfig{
		Servers: []config.ResolvedServer{{Name: "web-01"}, {Name: "web-02"}, {Name: "web-03"}},
	}

	root := &cobra.Command{Use: "root"}
	var out bytes.Buffer
	root.SetOut(&out)

	err := runClientReleaseRollback(root, "dev", merged, &config.Config{}, "prod", "web", "")
	require.NoError(t, err)
	assert.Equal(t, []string{"web-01", "web-02", "web-03"}, called, "servers must be rolled back in resolved order")
}

func TestRunClientReleaseRollback_FailureSkipsRemainingServers(t *testing.T) {
	orig := releaseRollbackToServerFn
	defer func() { releaseRollbackToServerFn = orig }()

	failure := &cmderr.ExitError{Code: 7, Message: "agent failed"}

	var called []string
	releaseRollbackToServerFn = func(_ string, _ *config.Config, _, _, _ string, srv config.ResolvedServer, _ forgeui.Mode, _ io.Writer) error {
		called = append(called, srv.Name)
		if srv.Name == "web-02" {
			return failure
		}
		return nil
	}

	merged := &config.MergedConfig{
		Servers: []config.ResolvedServer{{Name: "web-01"}, {Name: "web-02"}, {Name: "web-03"}},
	}

	root := &cobra.Command{Use: "root"}
	var out bytes.Buffer
	root.SetOut(&out)

	err := runClientReleaseRollback(root, "dev", merged, &config.Config{}, "prod", "web", "")
	require.Error(t, err)
	assert.Equal(t, []string{"web-01", "web-02"}, called, "remaining servers must be skipped after a failure")

	var exitErr *cmderr.ExitError
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 7, exitErr.Code, "exit code must come from the failing server's agent")
}
