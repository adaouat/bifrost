package release

import "github.com/spf13/cobra"

func NewReleaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release",
		Short: "Manage releases for a deployed application",
	}

	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newActivateCmd())
	cmd.AddCommand(newRollbackCmd())

	return cmd
}
