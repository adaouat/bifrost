package cmd_test

import (
	"bytes"
	"testing"

	"github.com/adaouat/bifrost/internal/cmd"
	"github.com/adaouat/bifrost/internal/cmderr"
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

func TestDeployCmd_ClientMode_NotYetImplemented(t *testing.T) {
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
	assert.Equal(t, 1, exitErr.Code)
	assert.Contains(t, exitErr.Message, "client mode")
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
