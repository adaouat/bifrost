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
	"github.com/adaouat/bifrost/internal/strategy"
	"github.com/adaouat/bifrost/internal/strategy/atomic"
	"github.com/adaouat/bifrost/internal/tui"
	forgeui "github.com/adaouat/forge/ui"
)

func newDeployCmd() *cobra.Command {
	var env, app, artifact, releaseName, agentBinary string
	var init_ bool

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy an application from an artifact archive",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if artifact == "" {
				return fmt.Errorf("--artifact is required")
			}

			cfg, err := config.Load(resolveConfigPath(cmd.Root()))
			if err != nil {
				return err
			}

			var merged *config.MergedConfig
			if config.IsFlat(cfg) {
				merged = config.MergeFlat(cfg)
			} else {
				if env == "" {
					return fmt.Errorf("--env (or --environment) is required")
				}
				if app == "" {
					return fmt.Errorf("--app (or --application) is required")
				}
				merged, err = config.Merge(cfg, env, app)
				if err != nil {
					return err
				}
			}

			if errs := config.Validate(merged); len(errs) > 0 {
				return &ExitError{Code: Config, Message: errs.Error()}
			}

			if len(merged.Servers) > 0 {
				return &ExitError{Code: Usage, Message: "client mode not yet implemented (servers configured)"}
			}

			if err := ensureRoots(merged.ReleasesRoot, merged.SharedRoot, init_); err != nil {
				return err
			}

			if _, err := os.Stat(artifact); err != nil {
				return &ExitError{Code: Runtime, Message: fmt.Sprintf("artifact not found: %s", artifact)}
			}

			if isDryRun(cmd) {
				return deployDryRun(cmd, merged, artifact, releaseName)
			}

			out := cmd.OutOrStdout()
			mode := forgeui.ParseMode(outputMode(cmd))
			confirmFn := interactiveConfirm()

			var deployer strategy.Deployer = atomic.New(out, mode, confirmFn)
			return deployer.Deploy(context.Background(), strategy.DeployOptions{
				Config:      merged,
				Artifact:    artifact,
				ReleaseName: releaseName,
				Env:         env,
				App:         app,
			})
		},
	}

	f := cmd.Flags()
	f.StringVar(&env, "environment", "", "target environment key")
	f.StringVar(&env, "env", "", "alias for --environment")
	f.StringVar(&app, "application", "", "target application key")
	f.StringVar(&app, "app", "", "alias for --application")
	f.StringVar(&artifact, "artifact", "", "path to artifact file (.tar.gz, .zip)")
	f.StringVar(&releaseName, "release-name", "", "override auto-generated release name")
	f.StringVar(&agentBinary, "agent-binary", "", "path to a prebuilt agent binary (skips download in client mode)")
	f.BoolVar(&init_, "init", false, "create releases_root and shared_root if missing")

	return cmd
}

// deployDryRun prints what a real deploy would do without executing any actions.
func deployDryRun(cmd *cobra.Command, merged *config.MergedConfig, artifact, releaseName string) error {
	releaseDir := atomic.PlanReleaseName(merged.ReleasesRoot, releaseName)
	relBase := filepath.Base(releaseDir)
	currentLink := filepath.Join(merged.ReleasesRoot, "current")
	out := cmd.OutOrStdout()

	hookLine := func(lifecycle string, h config.HookEntry) {
		suffix := ""
		if h.Sudo {
			suffix = "  (sudo)"
		}
		_, _ = fmt.Fprintf(out, "  Would run      [%s]  %s%s\n", lifecycle, h.Cmd, suffix)
	}

	_, _ = fmt.Fprintln(out, "DRY RUN — no changes will be made")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintf(out, "  Would create   %s\n", releaseDir)
	_, _ = fmt.Fprintf(out, "  Would extract  %s  →  %s\n", artifact, releaseDir)
	for _, h := range merged.Hooks.PostExtract {
		hookLine("post_extract", h)
	}
	for _, h := range merged.Hooks.PreLink {
		hookLine("pre_link", h)
	}
	for _, rel := range merged.SharedDirs {
		_, _ = fmt.Fprintf(out, "  Would link     %s  →  %s\n",
			filepath.Join(releaseDir, rel), filepath.Join(merged.SharedRoot, rel))
	}
	for _, rel := range merged.SharedFiles {
		_, _ = fmt.Fprintf(out, "  Would link     %s  →  %s\n",
			filepath.Join(releaseDir, rel), filepath.Join(merged.SharedRoot, rel))
	}
	for _, h := range merged.Hooks.PreEnableRelease {
		hookLine("pre_enable_release", h)
	}
	_, _ = fmt.Fprintf(out, "  Would update   %s  →  %s\n", currentLink, relBase)
	for _, h := range merged.Hooks.PostEnableRelease {
		hookLine("post_enable_release", h)
	}
	if candidates, err := atomic.PurgePlan(merged.ReleasesRoot, relBase, merged.Settings.ReleasesToKeep); err == nil && len(candidates) > 0 {
		_, _ = fmt.Fprintf(out, "  Would purge    %s  (keeping %d)\n", strings.Join(candidates, ", "), merged.Settings.ReleasesToKeep)
	}
	return nil
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

func ensureRoots(releasesRoot, sharedRoot string, create bool) error {
	for _, dir := range []string{releasesRoot, sharedRoot} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if !create {
				return &ExitError{Code: Runtime, Message: fmt.Sprintf("directory does not exist: %s (use --init to create)", dir)}
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
