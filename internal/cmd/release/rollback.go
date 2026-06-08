package release

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	forgeexec "github.com/adaouat/forge/exec"

	"github.com/adaouat/bifrost/internal/cmderr"
	"github.com/adaouat/bifrost/internal/config"
	"github.com/adaouat/bifrost/internal/hooks"
	"github.com/adaouat/bifrost/internal/strategy/atomic"
	"github.com/adaouat/bifrost/internal/tui"
)

func newRollbackCmd(version string) *cobra.Command {
	var env, app, agentBinary string

	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "Activate the release preceding the current one",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if env == "" {
				return fmt.Errorf("--env (or --environment) is required")
			}
			if app == "" {
				return fmt.Errorf("--app (or --application) is required")
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
				return &cmderr.ExitError{Code: cmderr.Config, Message: errs.Error()}
			}

			if len(merged.Servers) > 0 {
				return runClientReleaseRollback(cmd, version, merged, cfg, env, app, agentBinary)
			}

			releases, active, err := listReleases(merged.ReleasesRoot)
			if err != nil {
				return err
			}

			target, err := findRollbackTarget(releases, active)
			if err != nil {
				return err
			}

			releaseDir := filepath.Join(merged.ReleasesRoot, target)
			currentLink := filepath.Join(merged.ReleasesRoot, "current")

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
			hookRunner := forgeexec.New(false, false)

			if err := hooks.RunInteractive(hookRunner, merged.Hooks.PreEnableRelease, hookData, releaseDir, cmd.OutOrStdout(), confirmFn); err != nil {
				return fmt.Errorf("pre_enable_release hooks: %w", err)
			}

			if err := atomic.SetCurrent(merged.ReleasesRoot, releaseDir); err != nil {
				return fmt.Errorf("updating current symlink: %w", err)
			}

			if err := hooks.RunInteractive(hookRunner, merged.Hooks.PostEnableRelease, hookData, releaseDir, cmd.OutOrStdout(), confirmFn); err != nil {
				return fmt.Errorf("post_enable_release hooks: %w", err)
			}

			if mode, _ := cmd.Root().PersistentFlags().GetString("output"); mode == "json" {
				emit := tui.NewJSONEmitter(cmd.OutOrStdout())
				emit.Emit(map[string]any{"event": "rollback", "release": target, "status": "done"})
			}

			return nil
		},
	}

	f := cmd.Flags()
	f.StringVar(&env, "environment", "", "target environment key")
	f.StringVar(&env, "env", "", "alias for --environment")
	f.StringVar(&app, "application", "", "target application key")
	f.StringVar(&app, "app", "", "alias for --application")
	f.StringVar(&agentBinary, "agent-binary", "", "path to a prebuilt agent binary (skips download in client mode)")

	return cmd
}

// findRollbackTarget returns the release immediately preceding active in the
// newest-first sorted list. Returns ExitError{3} if rollback is not possible.
func findRollbackTarget(releases []string, active string) (string, error) {
	if active == "" {
		return "", &cmderr.ExitError{Code: cmderr.Runtime, Message: "no active release — nothing to roll back from"}
	}

	for i, r := range releases {
		if r == active {
			if i+1 >= len(releases) {
				return "", &cmderr.ExitError{Code: cmderr.Runtime, Message: "no previous release to roll back to"}
			}
			return releases[i+1], nil
		}
	}

	return "", &cmderr.ExitError{Code: cmderr.Runtime, Message: fmt.Sprintf("active release %q not found in releases list", active)}
}
