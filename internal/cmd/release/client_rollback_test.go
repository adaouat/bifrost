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

func TestParseRollbackEvent_SingleLine(t *testing.T) {
	release, err := parseRollbackEvent([]byte(`{"event":"rollback","release":"r1","status":"done"}` + "\n"))
	require.NoError(t, err)
	assert.Equal(t, "r1", release)
}

func TestParseRollbackEvent_HookOutputLinesAheadOfRollbackEvent(t *testing.T) {
	stdout := []byte(`{"event":"hook_output","output":"migrating\n"}` + "\n" +
		`{"event":"hook_output","output":"done\n"}` + "\n" +
		`{"event":"rollback","release":"r1","status":"done"}` + "\n")

	release, err := parseRollbackEvent(stdout)
	require.NoError(t, err)
	assert.Equal(t, "r1", release, "must find the rollback event even with hook_output lines ahead of it")
}

func TestParseRollbackEvent_NoRollbackEvent(t *testing.T) {
	_, err := parseRollbackEvent([]byte(`{"event":"hook_output","output":"oops\n"}` + "\n"))
	require.Error(t, err)
}

func TestRunClientReleaseRollback_DryRun_NeverCallsServer(t *testing.T) {
	orig := releaseRollbackToServerFn
	defer func() { releaseRollbackToServerFn = orig }()

	releaseRollbackToServerFn = func(_ string, _ *config.Config, _, _, _ string, _ config.ResolvedServer, _ forgeui.Mode, _ io.Writer) error {
		t.Fatal("releaseRollbackToServerFn must not be called during a dry run")
		return nil
	}

	merged := &config.MergedConfig{
		Servers: []config.ResolvedServer{{Name: "web-01"}, {Name: "web-02"}},
	}

	root := &cobra.Command{Use: "root"}
	root.PersistentFlags().Bool("dry-run", true, "")
	var out bytes.Buffer
	root.SetOut(&out)

	err := runClientReleaseRollback(root, "dev", merged, &config.Config{}, "prod", "web", "")
	require.NoError(t, err)
	assert.Contains(t, out.String(), "DRY RUN")
	assert.Contains(t, out.String(), "web-01")
	assert.Contains(t, out.String(), "web-02")
}
