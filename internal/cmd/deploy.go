package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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

			if isDryRun(cmd) {
				return deployDryRun(cmd, merged, artifact, releaseName)
			}

			releaseDir, err := atomic.CreateReleaseDir(merged.ReleasesRoot, releaseName)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			jsonMode := outputMode(cmd) == "json"
			var emit *tui.JSONEmitter
			if jsonMode {
				emit = tui.NewJSONEmitter(out)
			}

			emitError := func(step string, err error) error {
				if emit != nil {
					emit.Emit(map[string]any{"event": "error", "step": step, "message": err.Error(), "exit_code": 1})
				}
				return err
			}

			if !jsonMode {
				_, _ = fmt.Fprint(out, tui.DeployHeader(env, app, filepath.Base(releaseDir)))
				tui.PrintStep(out, "Config loaded and validated", "")
				tui.PrintStep(out, "Release directory created", "")
				tui.PrintDetail(out, "Root path: "+merged.ReleasesRoot)
				tui.PrintDetail(out, "Shared path: "+merged.SharedRoot)
			}

			deployStart := time.Now()

			// Extract artifact.
			if jsonMode {
				emit.Emit(map[string]any{"event": "start", "step": "extract", "artifact": artifact})
			}
			extractStart := time.Now()
			updateProgress, doneProgress := tui.NewProgressBar(info.Size(), "Extracting artifact", out)
			var jsonProgressFn func(n int64)
			if jsonMode {
				var written int64
				total := info.Size()
				jsonProgressFn = func(n int64) {
					written += n
					emit.Emit(map[string]any{"event": "progress", "step": "extract", "bytes": written, "total": total})
				}
			}
			progressFn := func(n int64) {
				updateProgress(n)
				if jsonProgressFn != nil {
					jsonProgressFn(n)
				}
			}
			if err := atomic.Extract(context.Background(), artifact, releaseDir, progressFn); err != nil {
				return emitError("extract", fmt.Errorf("extracting artifact: %w", err))
			}
			doneProgress()
			if jsonMode {
				emit.Emit(map[string]any{"event": "done", "step": "extract", "duration_ms": time.Since(extractStart).Milliseconds()})
			}
			if !jsonMode {
				tui.PrintStep(out, "Artifact extracted", fmt.Sprintf("(%.1fs)", time.Since(extractStart).Seconds()))
			}

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

			hookEventFn := deployHookEventFn(emit)

			if err := hooks.RunWithEvents(merged.Hooks.PostExtract, hookData, releaseDir, out, confirmFn, "post_extract", hookEventFn); err != nil {
				return emitError("post_extract", fmt.Errorf("post_extract hooks: %w", err))
			}
			if !jsonMode && len(merged.Hooks.PostExtract) > 0 {
				n := len(merged.Hooks.PostExtract)
				tui.PrintStep(out, "post_extract hooks", fmt.Sprintf("(%d/%d)", n, n))
			}

			if err := hooks.RunWithEvents(merged.Hooks.PreLink, hookData, releaseDir, out, confirmFn, "pre_link", hookEventFn); err != nil {
				return emitError("pre_link", fmt.Errorf("pre_link hooks: %w", err))
			}
			if !jsonMode && len(merged.Hooks.PreLink) > 0 {
				n := len(merged.Hooks.PreLink)
				tui.PrintStep(out, "pre_link hooks", fmt.Sprintf("(%d/%d)", n, n))
			}

			if jsonMode {
				emit.Emit(map[string]any{"event": "start", "step": "link"})
			}
			linkStart := time.Now()
			if err := atomic.LinkShared(merged.SharedDirs, merged.SharedFiles, releaseDir, merged.SharedRoot); err != nil {
				return emitError("link", fmt.Errorf("linking shared resources: %w", err))
			}
			if jsonMode {
				emit.Emit(map[string]any{
					"event":       "done",
					"step":        "link",
					"duration_ms": time.Since(linkStart).Milliseconds(),
					"dirs":        merged.SharedDirs,
					"files":       merged.SharedFiles,
				})
			}
			if !jsonMode {
				tui.PrintStep(out, "Shared directories linked", fmt.Sprintf("(%d)", len(merged.SharedDirs)))
				for _, d := range merged.SharedDirs {
					tui.PrintDetail(out, d)
				}
				tui.PrintStep(out, "Shared files linked", fmt.Sprintf("(%d)", len(merged.SharedFiles)))
				for _, f := range merged.SharedFiles {
					tui.PrintDetail(out, f)
				}
			}

			if err := hooks.RunWithEvents(merged.Hooks.PreEnableRelease, hookData, releaseDir, out, confirmFn, "pre_enable_release", hookEventFn); err != nil {
				return emitError("pre_enable_release", fmt.Errorf("pre_enable_release hooks: %w", err))
			}
			if !jsonMode && len(merged.Hooks.PreEnableRelease) > 0 {
				n := len(merged.Hooks.PreEnableRelease)
				tui.PrintStep(out, "pre_enable_release hooks", fmt.Sprintf("(%d/%d)", n, n))
			}

			if jsonMode {
				emit.Emit(map[string]any{"event": "start", "step": "current_symlink"})
			}
			currentStart := time.Now()
			if err := atomic.SetCurrent(merged.ReleasesRoot, releaseDir); err != nil {
				return emitError("current_symlink", fmt.Errorf("updating current symlink: %w", err))
			}
			currentTarget := filepath.Join(merged.ReleasesRoot, filepath.Base(releaseDir))
			if jsonMode {
				emit.Emit(map[string]any{
					"event":       "done",
					"step":        "current_symlink",
					"duration_ms": time.Since(currentStart).Milliseconds(),
					"path":        currentTarget,
				})
			}
			if !jsonMode {
				tui.PrintStep(out, "current symlink updated", "")
				tui.PrintDetail(out, currentTarget)
			}

			if err := hooks.RunWithEvents(merged.Hooks.PostEnableRelease, hookData, releaseDir, out, confirmFn, "post_enable_release", hookEventFn); err != nil {
				return emitError("post_enable_release", fmt.Errorf("post_enable_release hooks: %w", err))
			}
			if !jsonMode && len(merged.Hooks.PostEnableRelease) > 0 {
				n := len(merged.Hooks.PostEnableRelease)
				tui.PrintStep(out, "post_enable_release hooks", fmt.Sprintf("(%d/%d)", n, n))
			}

			// PurgePlan failure is non-fatal: Purge carries its own error path.
			// If planning fails, purgePlan is nil and the spinner / detail lines are skipped.
			purgePlan, purgePlanErr := atomic.PurgePlan(merged.ReleasesRoot, filepath.Base(releaseDir), merged.Settings.ReleasesToKeep)
			if purgePlanErr != nil {
				purgePlan = nil
			}
			if jsonMode {
				emit.Emit(map[string]any{"event": "start", "step": "purge"})
			}
			purgeStart := time.Now()
			var purgeErr error
			if !jsonMode && len(purgePlan) > 0 {
				noun := "releases"
				if len(purgePlan) == 1 {
					noun = "release"
				}
				label := fmt.Sprintf("Purging %d old %s...", len(purgePlan), noun)
				purgeErr = tui.RunWithSpinner(context.Background(), label, func(ctx context.Context) error {
					return atomic.Purge(merged.ReleasesRoot, merged.Settings.ReleasesToKeep)
				})
			} else {
				purgeErr = atomic.Purge(merged.ReleasesRoot, merged.Settings.ReleasesToKeep)
			}
			if purgeErr != nil {
				return emitError("purge", fmt.Errorf("purging old releases: %w", purgeErr))
			}
			if jsonMode {
				emit.Emit(map[string]any{
					"event":       "done",
					"step":        "purge",
					"duration_ms": time.Since(purgeStart).Milliseconds(),
					"purged":      purgePlan,
					"kept":        merged.Settings.ReleasesToKeep,
				})
			}
			if !jsonMode {
				tui.PrintStep(out, "Old releases purged", fmt.Sprintf("(kept %d)", merged.Settings.ReleasesToKeep))
				for _, r := range purgePlan {
					tui.PrintDetail(out, r+" deleted")
				}
			}

			if jsonMode {
				emit.Emit(map[string]any{
					"event":       "done",
					"step":        "deploy",
					"release":     filepath.Base(releaseDir),
					"duration_ms": time.Since(deployStart).Milliseconds(),
				})
			} else {
				tui.PrintSummary(out, time.Since(deployStart), filepath.Base(releaseDir))
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

// deployHookEventFn returns a hook event callback that emits JSON events via emit.
// Returns nil when emit is nil (non-JSON mode).
func deployHookEventFn(emit *tui.JSONEmitter) hooks.HookEventFn {
	if emit == nil {
		return nil
	}
	return func(lifecycle string, index int, cmd string, exitCode int) {
		emit.Emit(map[string]any{
			"event":     "hook",
			"lifecycle": lifecycle,
			"index":     index,
			"cmd":       cmd,
			"exit_code": exitCode,
		})
	}
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
