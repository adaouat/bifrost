package cmd

import (
	"github.com/adaouat/bifrost/internal/cmd/cmdutil"
	"github.com/spf13/cobra"
)

func resolveConfigPath(root *cobra.Command) string {
	explicit, _ := root.PersistentFlags().GetString("config")
	return cmdutil.ResolvePath(explicit)
}

// outputMode returns the --output flag value from the root command.
func outputMode(cmd *cobra.Command) string {
	mode, _ := cmd.Root().PersistentFlags().GetString("output")
	return mode
}

// isDryRun returns true if the --dry-run flag is set on the root command.
func isDryRun(cmd *cobra.Command) bool {
	v, _ := cmd.Root().PersistentFlags().GetBool("dry-run")
	return v
}
