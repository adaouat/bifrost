package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"

	"github.com/adaouat/bifrost/internal/config"
	"github.com/adaouat/bifrost/internal/hooks"
	"github.com/adaouat/bifrost/internal/strategy/atomic"
	"github.com/adaouat/bifrost/internal/tui"
)

func newDeployCmd() *cobra.Command {
	var env, app, artifact, releaseName string
	var init_ bool

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy an application from an artifact archive",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if env == "" {
				return fmt.Errorf("--env (or --environment) is required")
			}
			if app == "" {
				return fmt.Errorf("--app (or --application) is required")
			}
			if artifact == "" {
				return fmt.Errorf("--artifact is required")
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

			if err := ensureRoots(merged.ReleasesRoot, merged.SharedRoot, init_); err != nil {
				return err
			}

			info, err := os.Stat(artifact)
			if err != nil {
				return &ExitError{Code: 3, Message: fmt.Sprintf("artifact not found: %s", artifact)}
			}

			releaseDir, err := atomic.CreateReleaseDir(merged.ReleasesRoot, releaseName)
			if err != nil {
				return err
			}

			updateProgress, doneProgress := tui.NewProgressBar(info.Size(), cmd.OutOrStdout())
			if err := atomic.Extract(context.Background(), artifact, releaseDir, updateProgress); err != nil {
				return fmt.Errorf("extracting artifact: %w", err)
			}
			doneProgress()

			hookData := hooks.HookData{
				Settings:  merged.Settings,
				Variables: merged.Variables,
				Directories: hooks.Directories{
					Working:  releaseDir,
					Current:  filepath.Join(merged.ReleasesRoot, "current"),
					Releases: merged.ReleasesRoot,
					Shared:   merged.SharedRoot,
				},
				Env: osEnv(),
			}
			confirmFn := interactiveConfirm()

			if err := hooks.RunInteractive(merged.Hooks.PostExtract, hookData, releaseDir, cmd.OutOrStdout(), confirmFn); err != nil {
				return fmt.Errorf("post_extract hooks: %w", err)
			}

			if err := hooks.RunInteractive(merged.Hooks.PreLink, hookData, releaseDir, cmd.OutOrStdout(), confirmFn); err != nil {
				return fmt.Errorf("pre_link hooks: %w", err)
			}

			if err := atomic.LinkShared(merged.SharedDirs, merged.SharedFiles, releaseDir, merged.SharedRoot); err != nil {
				return fmt.Errorf("linking shared resources: %w", err)
			}

			if err := hooks.RunInteractive(merged.Hooks.PreEnableRelease, hookData, releaseDir, cmd.OutOrStdout(), confirmFn); err != nil {
				return fmt.Errorf("pre_enable_release hooks: %w", err)
			}

			if err := atomic.SetCurrent(merged.ReleasesRoot, releaseDir); err != nil {
				return fmt.Errorf("updating current symlink: %w", err)
			}

			if err := hooks.RunInteractive(merged.Hooks.PostEnableRelease, hookData, releaseDir, cmd.OutOrStdout(), confirmFn); err != nil {
				return fmt.Errorf("post_enable_release hooks: %w", err)
			}

			if err := atomic.Purge(merged.ReleasesRoot, merged.Settings.ReleasesToKeep); err != nil {
				return fmt.Errorf("purging old releases: %w", err)
			}

			return nil
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

// interactiveConfirm returns a confirmFn that shows a huh prompt on TTY,
// or nil (skip with warning) when stdout is not a terminal.
func interactiveConfirm() func(cmd string) bool {
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

// osEnv returns the current process environment as a key→value map.
func osEnv() map[string]string {
	env := os.Environ()
	m := make(map[string]string, len(env))
	for _, e := range env {
		k, v, _ := strings.Cut(e, "=")
		m[k] = v
	}
	return m
}

func ensureRoots(releasesRoot, sharedRoot string, create bool) error {
	for _, dir := range []string{releasesRoot, sharedRoot} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if !create {
				return &ExitError{Code: 3, Message: fmt.Sprintf("directory does not exist: %s (use --init to create)", dir)}
			}
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("creating directory %s: %w", dir, err)
			}
		} else if err != nil {
			return fmt.Errorf("checking directory %s: %w", dir, err)
		}
	}
	return nil
}
