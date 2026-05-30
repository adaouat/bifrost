package config

import (
	"os"

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

	return cmd
}

var statFile = os.Stat

// configPath resolves the config file path from the root command's --config flag,
// falling back to .config/bifrost.yml then .bifrost.yml.
func configPath(cmd *cobra.Command) string {
	root := cmd.Root()
	if root.PersistentFlags().Changed("config") {
		path, _ := root.PersistentFlags().GetString("config")
		return path
	}
	if _, err := statFile(".config/bifrost.yml"); err == nil {
		return ".config/bifrost.yml"
	}
	return ".bifrost.yml"
}
