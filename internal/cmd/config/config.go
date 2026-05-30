package config

import (
	"github.com/adaouat/bifrost/internal/cmd/cmdutil"
	"github.com/spf13/cobra"
)

// NewConfigCmd returns the config parent command with its subcommands.
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage and inspect the effective configuration",
	}

	cmd.AddCommand(newShowCmd())
	cmd.AddCommand(newCheckCmd())
	cmd.AddCommand(newInitCmd())

	return cmd
}

func configPath(cmd *cobra.Command) string {
	explicit, _ := cmd.Root().PersistentFlags().GetString("config")
	return cmdutil.ResolvePath(explicit)
}
