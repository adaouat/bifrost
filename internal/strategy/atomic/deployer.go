package atomic

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adaouat/bifrost/internal/hooks"
	"github.com/adaouat/bifrost/internal/strategy"
	"github.com/adaouat/bifrost/internal/tui"
	forgeexec "github.com/adaouat/forge/exec"
	forgeui "github.com/adaouat/forge/ui"
)

// Deployer implements strategy.Deployer for the atomic deployment strategy.
type Deployer struct {
	out       io.Writer
	mode      forgeui.Mode
	confirmFn func(string) bool
	runner    forgeexec.Runner
}

var _ strategy.Deployer = (*Deployer)(nil)

// New creates a new atomic Deployer.
func New(out io.Writer, mode forgeui.Mode, confirmFn func(string) bool) *Deployer {
	return &Deployer{out: out, mode: mode, confirmFn: confirmFn, runner: forgeexec.New(false, false)}
}

func (d *Deployer) jsonMode() bool { return d.mode == forgeui.JSON }

// Deploy executes a full atomic deployment.
func (d *Deployer) Deploy(ctx context.Context, opts strategy.DeployOptions) error {
	merged := opts.Config

	releaseDir, err := CreateReleaseDir(merged.ReleasesRoot, opts.ReleaseName)
	if err != nil {
		return err
	}

	var emit *tui.JSONEmitter
	if d.jsonMode() {
		emit = tui.NewJSONEmitter(d.out)
	}

	emitError := func(step string, err error) error {
		if emit != nil {
			emit.Emit(map[string]any{"event": "error", "step": step, "message": err.Error(), "exit_code": 1})
		}
		return err
	}

	runStep := func(step string, fn func() (map[string]any, error)) error {
		if d.jsonMode() {
			emit.Emit(map[string]any{"event": "start", "step": step})
		}
		start := time.Now()
		extras, err := fn()
		if err != nil {
			return emitError(step, err)
		}
		if d.jsonMode() {
			ev := map[string]any{"event": "done", "step": step, "duration_ms": time.Since(start).Milliseconds()}
			for k, v := range extras {
				ev[k] = v
			}
			emit.Emit(ev)
		}
		return nil
	}

	if !d.jsonMode() {
		_, _ = fmt.Fprint(d.out, tui.DeployHeader(d.mode, opts.Env, opts.App, filepath.Base(releaseDir)))
		tui.PrintStep(d.out, "Config loaded and validated", "")
		tui.PrintStep(d.out, "Release directory created", "")
		tui.PrintDetail(d.mode, d.out, "Root path: "+merged.ReleasesRoot)
		tui.PrintDetail(d.mode, d.out, "Shared path: "+merged.SharedRoot)
	}

	deployStart := time.Now()

	// Extract artifact.
	info, err := os.Stat(opts.Artifact)
	if err != nil {
		return fmt.Errorf("stat artifact: %w", err)
	}

	if d.jsonMode() {
		emit.Emit(map[string]any{"event": "start", "step": "extract", "artifact": opts.Artifact})
	}
	extractStart := time.Now()
	updateProgress, doneProgress := tui.NewProgressBar(d.mode, info.Size(), "Extracting artifact", d.out)
	var jsonProgressFn func(n int64)
	if d.jsonMode() {
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
	if err := Extract(ctx, opts.Artifact, releaseDir, progressFn); err != nil {
		return emitError("extract", fmt.Errorf("extracting artifact: %w", err))
	}
	doneProgress()
	if d.jsonMode() {
		emit.Emit(map[string]any{"event": "done", "step": "extract", "duration_ms": time.Since(extractStart).Milliseconds()})
	}
	if !d.jsonMode() {
		tui.PrintStep(d.out, "Artifact extracted", fmt.Sprintf("(%.1fs)", time.Since(extractStart).Seconds()))
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
		Env: envMap(),
	}
	hookEventFn := hookEmitter(emit)

	if err := hooks.RunWithEvents(d.runner, merged.Hooks.PostExtract, hookData, releaseDir, d.out, d.confirmFn, "post_extract", hookEventFn); err != nil {
		return emitError("post_extract", fmt.Errorf("post_extract hooks: %w", err))
	}
	if !d.jsonMode() && len(merged.Hooks.PostExtract) > 0 {
		n := len(merged.Hooks.PostExtract)
		tui.PrintStep(d.out, "post_extract hooks", fmt.Sprintf("(%d/%d)", n, n))
	}

	if err := hooks.RunWithEvents(d.runner, merged.Hooks.PreLink, hookData, releaseDir, d.out, d.confirmFn, "pre_link", hookEventFn); err != nil {
		return emitError("pre_link", fmt.Errorf("pre_link hooks: %w", err))
	}
	if !d.jsonMode() && len(merged.Hooks.PreLink) > 0 {
		n := len(merged.Hooks.PreLink)
		tui.PrintStep(d.out, "pre_link hooks", fmt.Sprintf("(%d/%d)", n, n))
	}

	if err := runStep("link", func() (map[string]any, error) {
		err := LinkShared(merged.SharedDirs, merged.SharedFiles, releaseDir, merged.SharedRoot)
		return map[string]any{"dirs": merged.SharedDirs, "files": merged.SharedFiles}, err
	}); err != nil {
		return fmt.Errorf("linking shared resources: %w", err)
	}
	if !d.jsonMode() {
		tui.PrintStep(d.out, "Shared directories linked", fmt.Sprintf("(%d)", len(merged.SharedDirs)))
		for _, dir := range merged.SharedDirs {
			tui.PrintDetail(d.mode, d.out, dir)
		}
		tui.PrintStep(d.out, "Shared files linked", fmt.Sprintf("(%d)", len(merged.SharedFiles)))
		for _, f := range merged.SharedFiles {
			tui.PrintDetail(d.mode, d.out, f)
		}
	}

	if err := hooks.RunWithEvents(d.runner, merged.Hooks.PreEnableRelease, hookData, releaseDir, d.out, d.confirmFn, "pre_enable_release", hookEventFn); err != nil {
		return emitError("pre_enable_release", fmt.Errorf("pre_enable_release hooks: %w", err))
	}
	if !d.jsonMode() && len(merged.Hooks.PreEnableRelease) > 0 {
		n := len(merged.Hooks.PreEnableRelease)
		tui.PrintStep(d.out, "pre_enable_release hooks", fmt.Sprintf("(%d/%d)", n, n))
	}

	if err := runStep("current_symlink", func() (map[string]any, error) {
		err := SetCurrent(merged.ReleasesRoot, releaseDir)
		return map[string]any{"path": releaseDir}, err
	}); err != nil {
		return fmt.Errorf("updating current symlink: %w", err)
	}
	if !d.jsonMode() {
		tui.PrintStep(d.out, "current symlink updated", "")
		tui.PrintDetail(d.mode, d.out, releaseDir)
	}

	if err := hooks.RunWithEvents(d.runner, merged.Hooks.PostEnableRelease, hookData, releaseDir, d.out, d.confirmFn, "post_enable_release", hookEventFn); err != nil {
		return emitError("post_enable_release", fmt.Errorf("post_enable_release hooks: %w", err))
	}
	if !d.jsonMode() && len(merged.Hooks.PostEnableRelease) > 0 {
		n := len(merged.Hooks.PostEnableRelease)
		tui.PrintStep(d.out, "post_enable_release hooks", fmt.Sprintf("(%d/%d)", n, n))
	}

	// PurgePlan failure is non-fatal: Purge carries its own error path.
	purgePlan, purgePlanErr := PurgePlan(merged.ReleasesRoot, filepath.Base(releaseDir), merged.Settings.ReleasesToKeep)
	if purgePlanErr != nil {
		purgePlan = nil
	}
	if err := runStep("purge", func() (map[string]any, error) {
		extras := map[string]any{"purged": purgePlan, "kept": merged.Settings.ReleasesToKeep}
		purge := func() error { return Purge(merged.ReleasesRoot, merged.Settings.ReleasesToKeep) }
		if d.jsonMode() {
			return extras, purge()
		}
		// Spin while purging (human + TTY) then resolve to "✓ Old releases purged — (kept N)".
		return extras, forgeui.NewSpinner(d.out, d.mode).Run("Old releases purged", func() (forgeui.Result, error) {
			return forgeui.Result{Detail: fmt.Sprintf("(kept %d)", merged.Settings.ReleasesToKeep)}, purge()
		})
	}); err != nil {
		return fmt.Errorf("purging old releases: %w", err)
	}
	if !d.jsonMode() {
		for _, r := range purgePlan {
			tui.PrintDetail(d.mode, d.out, r+" deleted")
		}
	}

	if d.jsonMode() {
		emit.Emit(map[string]any{
			"event":       "done",
			"step":        "deploy",
			"release":     filepath.Base(releaseDir),
			"duration_ms": time.Since(deployStart).Milliseconds(),
		})
	} else {
		tui.PrintSummary(d.mode, d.out, time.Since(deployStart), filepath.Base(releaseDir))
	}

	return nil
}

// hookEmitter returns a hook event callback that emits JSON events via emit.
func hookEmitter(emit *tui.JSONEmitter) hooks.HookEventFn {
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

// envMap returns the current process environment as a key→value map.
func envMap() map[string]string {
	env := os.Environ()
	m := make(map[string]string, len(env))
	for _, e := range env {
		k, v, _ := strings.Cut(e, "=")
		m[k] = v
	}
	return m
}
