package release

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newActivateCmd() *cobra.Command {
	var env, app, releaseName string

	cmd := &cobra.Command{
		Use:   "activate",
		Short: "Activate a previously deployed release",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return fmt.Errorf("not yet implemented")
		},
	}

	f := cmd.Flags()
	f.StringVar(&env, "environment", "", "target environment key")
	f.StringVar(&env, "env", "", "alias for --environment")
	f.StringVar(&app, "application", "", "target application key")
	f.StringVar(&app, "app", "", "alias for --application")
	f.StringVar(&releaseName, "release", "", "release name to activate (interactive selector if omitted)")

	return cmd
}
