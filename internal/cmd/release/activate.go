package release

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"

	"github.com/adaouat/bifrost/internal/cmderr"
	"github.com/adaouat/bifrost/internal/config"
	"github.com/adaouat/bifrost/internal/hooks"
	"github.com/adaouat/bifrost/internal/strategy/atomic"
	"github.com/adaouat/bifrost/internal/tui"
)

func newActivateCmd() *cobra.Command {
	var env, app, releaseName string
	var noConfirm bool

	cmd := &cobra.Command{
		Use:   "activate",
		Short: "Activate a previously deployed release",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if env == "" {
				return fmt.Errorf("--env (or --environment) is required")
			}
			if app == "" {
				return fmt.Errorf("--app (or --application) is required")
			}
			if releaseName == "" {
				return fmt.Errorf("--release is required")
			}

			cfg, err := config.Load(releaseConfigPath(cmd))
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

			releaseDir := filepath.Join(merged.ReleasesRoot, releaseName)
			if _, err := os.Stat(releaseDir); err != nil {
				return &cmderr.ExitError{Code: 3, Message: fmt.Sprintf("release not found: %s", releaseName)}
			}

			currentLink := filepath.Join(merged.ReleasesRoot, "current")
			if target, err := os.Readlink(currentLink); err == nil && target == releaseDir {
				if noConfirm || !tui.IsTTY() {
					return &cmderr.ExitError{Code: 3, Message: fmt.Sprintf("release %q is already current", releaseName)}
				}
				var ok bool
				if err := huh.NewConfirm().
					Title(fmt.Sprintf("Release %q is already current. Re-activate?", releaseName)).
					Value(&ok).
					Run(); err != nil || !ok {
					return fmt.Errorf("re-activation cancelled")
				}
			}

			if err := atomic.LinkShared(merged.SharedDirs, merged.SharedFiles, releaseDir, merged.SharedRoot); err != nil {
				return fmt.Errorf("linking shared resources: %w", err)
			}

			hookData := hooks.HookData{
				Settings:  merged.Settings,
				Variables: merged.Variables,
				Directories: hooks.Directories{
					Working:  releaseDir,
					Current:  currentLink,
					Releases: merged.ReleasesRoot,
					Shared:   merged.SharedRoot,
				},
				Env: releaseOsEnv(),
			}
			confirmFn := releaseInteractiveConfirm()

			if err := hooks.RunInteractive(merged.Hooks.PreEnableRelease, hookData, releaseDir, cmd.OutOrStdout(), confirmFn); err != nil {
				return fmt.Errorf("pre_enable_release hooks: %w", err)
			}

			if err := atomic.SetCurrent(merged.ReleasesRoot, releaseDir); err != nil {
				return fmt.Errorf("updating current symlink: %w", err)
			}

			if err := hooks.RunInteractive(merged.Hooks.PostEnableRelease, hookData, releaseDir, cmd.OutOrStdout(), confirmFn); err != nil {
				return fmt.Errorf("post_enable_release hooks: %w", err)
			}

			return nil
		},
	}

	f := cmd.Flags()
	f.StringVar(&env, "environment", "", "target environment key")
	f.StringVar(&env, "env", "", "alias for --environment")
	f.StringVar(&app, "application", "", "target application key")
	f.StringVar(&app, "app", "", "alias for --application")
	f.StringVar(&releaseName, "release", "", "release name to activate (interactive selector if omitted)")
	f.BoolVar(&noConfirm, "no-confirm", false, "exit 3 if release is already current instead of prompting")

	return cmd
}

// releaseConfigPath resolves the config file path from the root command's --config flag,
// falling back to .config/bifrost.yml then .bifrost.yml.
func releaseConfigPath(cmd *cobra.Command) string {
	root := cmd.Root()
	if root.PersistentFlags().Changed("config") {
		path, _ := root.PersistentFlags().GetString("config")
		return path
	}
	if _, err := os.Stat(".config/bifrost.yml"); err == nil {
		return ".config/bifrost.yml"
	}
	return ".bifrost.yml"
}

// releaseInteractiveConfirm returns a hook confirm function that shows a huh prompt on TTY.
func releaseInteractiveConfirm() func(cmd string) bool {
	if !tui.IsTTY() {
		return nil
	}
	return func(hookCmd string) bool {
		var ok bool
		if err := huh.NewConfirm().
			Title("Run interactive hook?").
			Description(hookCmd).
			Value(&ok).
			Run(); err != nil {
			return false
		}
		return ok
	}
}

// releaseOsEnv returns the current process environment as a key→value map.
func releaseOsEnv() map[string]string {
	env := os.Environ()
	m := make(map[string]string, len(env))
	for _, e := range env {
		k, v, _ := strings.Cut(e, "=")
		m[k] = v
	}
	return m
}
