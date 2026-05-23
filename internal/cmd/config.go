package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/adaouat/bifrost/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	var env, app string

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Display and validate the effective configuration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfgPath, err := cmd.Root().PersistentFlags().GetString("config")
			if err != nil {
				return fmt.Errorf("reading --config flag: %w", err)
			}

			cfg, err := config.Load(cfgPath)
			if err != nil {
				return err
			}

			if env == "" && app == "" {
				return printJSON(cmd, cfg)
			}

			merged, err := config.Merge(cfg, env, app)
			if err != nil {
				return err
			}
			if errs := config.Validate(merged); len(errs) > 0 {
				return &ExitError{
					Code:    2,
					Message: strings.Join(errs, "\n"),
				}
			}
			return printJSON(cmd, merged)
		},
	}

	f := cmd.Flags()
	f.StringVar(&env, "environment", "", "target environment key")
	f.StringVar(&env, "env", "", "alias for --environment")
	f.StringVar(&app, "application", "", "target application key")
	f.StringVar(&app, "app", "", "alias for --application")

	return cmd
}

func printJSON(cmd *cobra.Command, v any) error {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("serializing to JSON: %w", err)
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), string(out))
	return err
}
