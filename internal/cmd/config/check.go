package config

import (
	"fmt"
	"strings"

	"github.com/adaouat/bifrost/internal/cmderr"
	"github.com/adaouat/bifrost/internal/config"
	"github.com/spf13/cobra"
)

func newCheckCmd() *cobra.Command {
	var env, app string

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Validate required fields for a deployment target",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if env == "" {
				return fmt.Errorf("--env (or --environment) is required")
			}
			if app == "" {
				return fmt.Errorf("--app (or --application) is required")
			}

			cfg, err := config.Load(configPath(cmd))
			if err != nil {
				return err
			}
			merged, err := config.Merge(cfg, env, app)
			if err != nil {
				return err
			}
			if errs := config.Validate(merged); len(errs) > 0 {
				return &cmderr.ExitError{Code: 2, Message: strings.Join(errs, "\n")}
			}
			return nil
		},
	}

	f := cmd.Flags()
	f.StringVar(&env, "environment", "", "target environment key")
	f.StringVar(&env, "env", "", "alias for --environment")
	f.StringVar(&app, "application", "", "target application key")
	f.StringVar(&app, "app", "", "alias for --application")

	return cmd
}
