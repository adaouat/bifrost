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

func TestRunClientReleaseActivate_NonInteractiveSequentialLoop(t *testing.T) {
	orig := releaseActivateToServerFn
	defer func() { releaseActivateToServerFn = orig }()

	type call struct {
		server  string
		release string
	}
	var calls []call
	releaseActivateToServerFn = func(_ string, _ *config.Config, _, _, _, releaseName string, srv config.ResolvedServer, _ forgeui.Mode, _ io.Writer) error {
		calls = append(calls, call{server: srv.Name, release: releaseName})
		return nil
	}

	merged := &config.MergedConfig{
		Servers: []config.ResolvedServer{{Name: "web-01"}, {Name: "web-02"}},
	}

	root := &cobra.Command{Use: "root"}
	var out bytes.Buffer
	root.SetOut(&out)

	err := runClientReleaseActivate(root, "dev", merged, &config.Config{}, "prod", "web", "20260601-120000", "")
	require.NoError(t, err)
	assert.Equal(t, []call{
		{server: "web-01", release: "20260601-120000"},
		{server: "web-02", release: "20260601-120000"},
	}, calls, "the given release name must be used for every server, in resolved order")
}

func TestRunClientReleaseActivate_FailureSkipsRemainingServers(t *testing.T) {
	orig := releaseActivateToServerFn
	defer func() { releaseActivateToServerFn = orig }()

	failure := &cmderr.ExitError{Code: 7, Message: "agent failed"}

	var called []string
	releaseActivateToServerFn = func(_ string, _ *config.Config, _, _, _, _ string, srv config.ResolvedServer, _ forgeui.Mode, _ io.Writer) error {
		called = append(called, srv.Name)
		if srv.Name == "web-01" {
			return failure
		}
		return nil
	}

	merged := &config.MergedConfig{
		Servers: []config.ResolvedServer{{Name: "web-01"}, {Name: "web-02"}},
	}

	root := &cobra.Command{Use: "root"}
	var out bytes.Buffer
	root.SetOut(&out)

	err := runClientReleaseActivate(root, "dev", merged, &config.Config{}, "prod", "web", "20260601-120000", "")
	require.Error(t, err)
	assert.Equal(t, []string{"web-01"}, called, "remaining servers must be skipped after a failure")

	var exitErr *cmderr.ExitError
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 7, exitErr.Code, "exit code must come from the failing server's agent")
}

func TestRunClientReleaseActivate_NonTTYWithoutRelease_ExitsCode1(t *testing.T) {
	merged := &config.MergedConfig{
		Servers: []config.ResolvedServer{{Name: "web-01"}},
	}

	root := &cobra.Command{Use: "root"}
	var out bytes.Buffer
	root.SetOut(&out)

	err := runClientReleaseActivate(root, "dev", merged, &config.Config{}, "prod", "web", "", "")
	require.Error(t, err)

	var exitErr *cmderr.ExitError
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 1, exitErr.Code, "non-TTY without --release must exit with code 1")
}
