package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var statFile = os.Stat

// resolveConfigPath returns the config file to load.
// Explicit --config always wins. Otherwise tries .config/bifrost.yml, then .bifrost.yml.
func resolveConfigPath(root *cobra.Command) string {
	if root.PersistentFlags().Changed("config") {
		path, _ := root.PersistentFlags().GetString("config")
		return path
	}
	if _, err := statFile(".config/bifrost.yml"); err == nil {
		return ".config/bifrost.yml"
	}
	return ".bifrost.yml"
}
