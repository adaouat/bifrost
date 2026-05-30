package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/cobra"
)

const scaffold = `# bifrost configuration
# Reference: https://github.com/adaouat/bifrost/blob/main/docs/specs/02-configuration.md

# ── Strategy ───────────────────────────────────────────────────────────────────
strategy: atomic # v0 valid values: atomic

# ── Global defaults ────────────────────────────────────────────────────────────
paths:
  releases_root: /var/www/releases # REQUIRED after merge
  shared_root: /var/www/shared # REQUIRED after merge
  shared:
    directories: [] # Relative paths symlinked into each release
    files: [] # Relative paths symlinked into each release

settings:
  releases_to_keep: 10 # Number of old releases to retain

variables: {} # String key-value pairs for hook templates: {{ .Variables.key }}

hooks:
  post_extract: [] # After extraction, before shared linking
  pre_link: [] # Before shared resource linking
  pre_enable_release: [] # Before current symlink update
  post_enable_release: [] # After current symlink update

# ── Environments ───────────────────────────────────────────────────────────────
environments:
  production:
    name: "Production"
    applications:
      web:
        name: "Web Application"
        paths:
          shared:
            directories:
              - var/log
              - var/cache
            files:
              - .env
        hooks:
          post_enable_release:
            - cmd: "systemctl reload php-fpm"
              sudo: true
              priority: 10
`

var (
	initWrite = func(path string, data []byte, perm uint32) error {
		return os.WriteFile(path, data, fs.FileMode(perm))
	}
	initStat = func(path string) error {
		_, err := os.Stat(path)
		return err
	}
)

// SetInitWrite replaces the write function used by config init (for testing).
// Pass nil to restore the default.
// Tests that call this must not run in parallel — these are package-level globals.
func SetInitWrite(fn func(string, []byte, uint32) error) {
	if fn == nil {
		initWrite = func(path string, data []byte, perm uint32) error {
			return os.WriteFile(path, data, fs.FileMode(perm))
		}
	} else {
		initWrite = fn
	}
}

// SetInitStat replaces the stat function used by config init (for testing).
// Pass nil to restore the default.
// Tests that call this must not run in parallel — these are package-level globals.
func SetInitStat(fn func(string) error) {
	if fn == nil {
		initStat = func(path string) error {
			_, err := os.Stat(path)
			return err
		}
	} else {
		initStat = fn
	}
}

func newInitCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Scaffold a default .bifrost.yml configuration file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			path := initPath(cmd)

			if !force {
				if err := initStat(path); err == nil {
					return fmt.Errorf("%s already exists — use --force to overwrite", path)
				} else if !errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("checking %s: %w", path, err)
				}
			}

			if err := initWrite(path, []byte(scaffold), 0o644); err != nil {
				return fmt.Errorf("writing %s: %w", path, err)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "created %s\n", path)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing file")

	return cmd
}

// initPath returns the output path for config init.
// Uses --config if explicitly set, otherwise defaults to .bifrost.yml.
func initPath(cmd *cobra.Command) string {
	root := cmd.Root()
	if root.PersistentFlags().Changed("config") {
		path, _ := root.PersistentFlags().GetString("config")
		return path
	}
	return ".bifrost.yml"
}
