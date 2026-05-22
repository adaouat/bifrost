package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	var env, app string

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Display and validate the effective configuration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return fmt.Errorf("not yet implemented")
		},
	}

	f := cmd.Flags()
	f.StringVar(&env, "environment", "", "target environment key")
	f.StringVar(&env, "env", "", "alias for --environment")
	f.StringVar(&app, "application", "", "target application key")
	f.StringVar(&app, "app", "", "alias for --application")

	return cmd
}
