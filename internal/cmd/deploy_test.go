package cmd_test

import (
	"bytes"
	"testing"

	"github.com/adaouat/bifrost/internal/cmd"
	"github.com/adaouat/bifrost/internal/cmderr"
	"github.com/adaouat/bifrost/internal/tui"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeployCmd_HasAgentBinaryFlag(t *testing.T) {
	root := cmd.NewRootCmd("dev")
	var deploy *cobra.Command
	for _, c := range root.Commands() {
		if c.Name() == "deploy" {
			deploy = c
			break
		}
	}
	require.NotNil(t, deploy)
	assert.NotNil(t, deploy.Flags().Lookup("agent-binary"))
}

func TestDeployCmd_ClientMode_AttemptsSSHConnection(t *testing.T) {
	// Client mode is now implemented: it tries to SSH to the configured server.
	// With no real server at 192.168.1.10, it fails with a Runtime (exit code 3)
	// SSH connection error rather than the old Usage (exit code 1) stub error.
	root := cmd.NewRootCmd("dev")
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{
		"deploy",
		"--config", "../../testdata/bifrost-servers.yml",
		"--env", "prod",
		"--app", "web",
		"--artifact", "../../testdata/release.tar.gz",
	})

	err := root.Execute()
	require.Error(t, err)

	var exitErr *cmderr.ExitError
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 3, exitErr.Code)
}

func TestDeployCmd_HasInteractiveFlag(t *testing.T) {
	root := cmd.NewRootCmd("dev")
	var deploy *cobra.Command
	for _, c := range root.Commands() {
		if c.Name() == "deploy" {
			deploy = c
			break
		}
	}
	require.NotNil(t, deploy)
	assert.NotNil(t, deploy.Flags().Lookup("interactive"))
}

// TestDeployCmd_Interactive_RequiresTTY proves --interactive fails fast with a
// Usage error when stdout isn't a real terminal (as in this test), before any
// config loading or deploy steps run.
func TestDeployCmd_Interactive_RequiresTTY(t *testing.T) {
	root := cmd.NewRootCmd("dev")
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{
		"deploy",
		"--artifact", "../../testdata/release.tar.gz",
		"--interactive",
	})

	err := root.Execute()
	require.Error(t, err)

	var exitErr *cmderr.ExitError
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, cmderr.Usage, exitErr.Code)
}

// TestDeployCmd_Interactive_RejectsServerConfig proves --interactive is rejected
// with a Usage error for a server (remote) config, even on a TTY: the agent runs
// remotely in JSON mode, so per-step prompts can't work. Without the guard the
// command would silently fall through to a client-mode SSH deploy.
func TestDeployCmd_Interactive_RejectsServerConfig(t *testing.T) {
	tui.SetIsTTY(func() bool { return true })
	defer tui.SetIsTTY(nil)

	root := cmd.NewRootCmd("dev")
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{
		"deploy",
		"--config", "../../testdata/bifrost-servers.yml",
		"--env", "prod",
		"--app", "web",
		"--artifact", "../../testdata/release.tar.gz",
		"--interactive",
	})

	err := root.Execute()
	require.Error(t, err)

	var exitErr *cmderr.ExitError
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, cmderr.Usage, exitErr.Code)
	assert.Contains(t, exitErr.Error(), "remote")
}

func TestDeployCmd_FlatConfig_RequiresNoEnvApp(t *testing.T) {
	root := cmd.NewRootCmd("dev")
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	// flat config + no --env/--app + no real paths → will fail on ensureRoots, not on missing flags
	root.SetArgs([]string{
		"deploy",
		"--config", "../../testdata/bifrost-flat.yml",
		"--artifact", "../../testdata/release.tar.gz",
	})

	err := root.Execute()
	// Should not fail with "env is required" — may fail for other reasons (paths don't exist)
	if err != nil {
		assert.NotContains(t, err.Error(), "--env (or --environment) is required")
		assert.NotContains(t, err.Error(), "--app (or --application) is required")
	}
}
