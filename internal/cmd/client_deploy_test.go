package cmd

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/bifrost/internal/config"
	forgeui "github.com/adaouat/forge/ui"
)

func TestRunClientDeploy_SequentialLoop(t *testing.T) {
	orig := deployToServerFn
	defer func() { deployToServerFn = orig }()

	var called []string
	deployToServerFn = func(_ string, _ *config.MergedConfig, _ *config.Config, _, _, _, _, _ string, srv config.ResolvedServer, _ forgeui.Mode, _ io.Writer) error {
		called = append(called, srv.Name)
		return nil
	}

	merged := &config.MergedConfig{
		Servers: []config.ResolvedServer{{Name: "web-01"}, {Name: "web-02"}, {Name: "web-03"}},
	}

	root := &cobra.Command{Use: "root"}
	var out bytes.Buffer
	root.SetOut(&out)

	err := runClientDeploy(root, "dev", merged, &config.Config{}, "prod", "web", "artifact.tar.gz", "20260601-120000", "")
	require.NoError(t, err)
	assert.Equal(t, []string{"web-01", "web-02", "web-03"}, called, "servers must be deployed in resolved order")
	assert.Equal(t, 1, strings.Count(out.String(), "Environment"), "deploy header must be printed once, not per server")
}

func TestRunClientDeploy_FailureSkipsRemainingServers(t *testing.T) {
	orig := deployToServerFn
	defer func() { deployToServerFn = orig }()

	failure := &ExitError{Code: 7, Message: "agent failed"}

	var called []string
	deployToServerFn = func(_ string, _ *config.MergedConfig, _ *config.Config, _, _, _, _, _ string, srv config.ResolvedServer, _ forgeui.Mode, _ io.Writer) error {
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

	err := runClientDeploy(root, "dev", merged, &config.Config{}, "prod", "web", "artifact.tar.gz", "20260601-120000", "")
	require.Error(t, err)
	assert.Equal(t, []string{"web-01", "web-02"}, called, "remaining servers must be skipped after a failure")

	var exitErr *ExitError
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 7, exitErr.Code, "exit code must come from the failing server's agent")
}
