package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/adaouat/bifrost/internal/config"
)

func newInitCmd() *cobra.Command {
	var env, app string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create releases_root and shared_root for a deployment target",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if env == "" {
				return fmt.Errorf("--env (or --environment) is required")
			}
			if app == "" {
				return fmt.Errorf("--app (or --application) is required")
			}

			cfg, err := config.Load(resolveConfigPath(cmd.Root()))
			if err != nil {
				return err
			}
			merged, err := config.Merge(cfg, env, app)
			if err != nil {
				return err
			}
			if errs := config.Validate(merged); len(errs) > 0 {
				return &ExitError{Code: 2, Message: strings.Join(errs, "\n")}
			}

			return ensureRoots(merged.ReleasesRoot, merged.SharedRoot, true)
		},
	}

	f := cmd.Flags()
	f.StringVar(&env, "environment", "", "target environment key")
	f.StringVar(&env, "env", "", "alias for --environment")
	f.StringVar(&app, "application", "", "target application key")
	f.StringVar(&app, "app", "", "alias for --application")

	return cmd
}
