package release

import "github.com/spf13/cobra"

func NewReleaseCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release",
		Short: "Manage releases for a deployed application",
	}

	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newListCmd(version))
	cmd.AddCommand(newActivateCmd(version))
	cmd.AddCommand(newRollbackCmd(version))

	return cmd
}
