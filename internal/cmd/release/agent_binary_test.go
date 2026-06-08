package release

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestSSHCapableCommands_HaveAgentBinaryFlag(t *testing.T) {
	cmds := map[string]*cobra.Command{
		"list":     newListCmd("dev"),
		"activate": newActivateCmd(),
		"rollback": newRollbackCmd(),
	}
	for name, c := range cmds {
		assert.NotNilf(t, c.Flags().Lookup("agent-binary"), "%s release command missing --agent-binary flag", name)
	}
}

func TestInitCommand_HasNoAgentBinaryFlag(t *testing.T) {
	// release init is local-only — it does not run over SSH and takes no agent.
	assert.Nil(t, newInitCmd().Flags().Lookup("agent-binary"))
}
