package config

import (
	"encoding/json"
	"fmt"

	"github.com/adaouat/bifrost/internal/config"
	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	var env, app string

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Display the effective configuration as JSON",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if (env == "") != (app == "") {
				return fmt.Errorf("--env and --app must be used together")
			}

			cfg, err := config.Load(configPath(cmd))
			if err != nil {
				return err
			}

			if env == "" {
				return printJSON(cmd, cfg)
			}

			merged, err := config.Merge(cfg, env, app)
			if err != nil {
				return err
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
