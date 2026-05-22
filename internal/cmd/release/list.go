package release

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var env, app string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all releases for an application",
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
