package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDeployCmd() *cobra.Command {
	var env, app, artifact, releaseName string
	var init_ bool

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy an application from an artifact archive",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return fmt.Errorf("not yet implemented")
		},
	}

	f := cmd.Flags()
	f.StringVar(&env, "environment", "", "target environment key")
	f.StringVar(&env, "env", "", "alias for --environment")
	f.StringVar(&app, "application", "", "target application key")
	f.StringVar(&app, "app", "", "alias for --application")
	f.StringVar(&artifact, "artifact", "", "path to artifact file (.tar.gz, .zip)")
	f.StringVar(&releaseName, "release-name", "", "override auto-generated release name")
	f.BoolVar(&init_, "init", false, "create releases_root and shared_root if missing")

	return cmd
}
