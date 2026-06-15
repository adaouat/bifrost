package atomic

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/adaouat/bifrost/internal/cmderr"
	"github.com/adaouat/bifrost/internal/config"
	"github.com/adaouat/bifrost/internal/hooks"
	"github.com/adaouat/bifrost/internal/strategy"
	"github.com/adaouat/bifrost/internal/tui"
	forgeexec "github.com/adaouat/forge/exec"
	forgeui "github.com/adaouat/forge/ui"
)

// deployStepTotal counts the numbered step lines a human-mode deploy renders:
// seven always-present steps plus one per configured hook group.
func deployStepTotal(cfg *config.MergedConfig) int {
	total := 7 // config, release dir, extract, shared dirs, shared files, symlink, purge
	for _, h := range [][]config.HookEntry{
		cfg.Hooks.PreExtract, cfg.Hooks.PostExtract,
		cfg.Hooks.PreLink, cfg.Hooks.PostLink,
		cfg.Hooks.PreActivate, cfg.Hooks.PostActivate,
		cfg.Hooks.PrePurge, cfg.Hooks.PostPurge,
	} {
		if len(h) > 0 {
			total++
		}
	}
	return total
}

// Deployer implements strategy.Deployer for the atomic deployment strategy.
type Deployer struct {
	out       io.Writer
	mode      forgeui.Mode
	confirmFn func(string) bool
	runner    forgeexec.Runner
	logger    *slog.Logger
}

var _ strategy.Deployer = (*Deployer)(nil)

// New creates a new atomic Deployer.
func New(out io.Writer, mode forgeui.Mode, confirmFn func(string) bool) *Deployer {
	return &Deployer{out: out, mode: mode, confirmFn: confirmFn, runner: forgeexec.New(false, false)}
}

// WithLogger sets the operator-debug diagnostic logger and returns d for chaining.
// When unset, diagnostics are discarded. See forge ADR-0011.
func (d *Deployer) WithLogger(l *slog.Logger) *Deployer {
	d.logger = l
	return d
}

// WithRunner replaces the command runner used for hook execution and returns d
// for chaining. Tests inject a fake runner to avoid real shell execution.
func (d *Deployer) WithRunner(r forgeexec.Runner) *Deployer {
	d.runner = r
	return d
}

// debug emits an operator-debug log line when a logger is set.
func (d *Deployer) debug(msg string, args ...any) {
	if d.logger != nil {
		d.logger.Debug(msg, args...)
	}
}

func (d *Deployer) jsonMode() bool { return d.mode == forgeui.JSON }

// Deploy executes a full atomic deployment.
func (d *Deployer) Deploy(ctx context.Context, opts strategy.DeployOptions) error {
	merged := opts.Config

	releaseDir, err := CreateReleaseDir(merged.ReleasesRoot, opts.ReleaseName)
	if err != nil {
		return cmderr.Wrap(cmderr.Runtime, err)
	}
	d.debug("release dir created", "path", releaseDir, "env", opts.Env, "app", opts.App)

	// One spinner numbers every human-mode step line [N/total]; JSON mode emits
	// events instead and never touches it.
	sp := forgeui.NewSpinner(d.out, d.mode).Total(deployStepTotal(merged))

	var emit *tui.JSONEmitter
	if d.jsonMode() {
		emit = tui.NewJSONEmitter(d.out)
	}

	emitError := func(step string, err error) error {
		if emit != nil {
			emit.Emit(map[string]any{"event": "error", "step": step, "message": err.Error(), "exit_code": cmderr.Resolve(err)})
		}
		return err
	}

	// checkCancel is a no-op until extraction succeeds, then warns once if ctx
	// was cancelled (e.g. Ctrl+C) — the remaining steps run to completion
	// regardless, so the user should know cancellation is being ignored.
	checkCancel := func() {}

	runStep := func(step string, fn func() (map[string]any, error)) error {
		checkCancel()
		if d.jsonMode() {
			emit.Emit(map[string]any{"event": "start", "step": step})
		}
		d.debug("step started", "step", step)
		start := time.Now()
		extras, err := fn()
		if err != nil {
			return emitError(step, cmderr.Wrap(cmderr.Runtime, err))
		}
		if d.jsonMode() {
			ev := map[string]any{"event": "done", "step": step, "duration_ms": time.Since(start).Milliseconds()}
			for k, v := range extras {
				ev[k] = v
			}
			emit.Emit(ev)
		}
		args := []any{"step", step, "duration_ms", time.Since(start).Milliseconds()}
		for k, v := range extras {
			args = append(args, k, v)
		}
		d.debug("step completed", args...)
		return nil
	}

	if !d.jsonMode() {
		_, _ = fmt.Fprint(d.out, tui.DeployHeader(d.mode, opts.Env, opts.App, filepath.Base(releaseDir)))
		sp.Step("Config loaded and validated", "")
		sp.Step("Release directory created", "")
		tui.PrintDetail(d.mode, d.out, "Root path: "+merged.ReleasesRoot)
		tui.PrintDetail(d.mode, d.out, "Shared path: "+merged.SharedRoot)
	}

	deployStart := time.Now()

	hookData := hooks.HookData{
		Settings:  merged.Settings,
		Variables: merged.Variables,
		Directories: hooks.Directories{
			Working:  releaseDir,
			Current:  filepath.Join(merged.ReleasesRoot, "current"),
			Releases: merged.ReleasesRoot,
			Shared:   merged.SharedRoot,
		},
		Env: hooks.OSEnv(),
	}
	hookEventFn := hookEmitter(emit)

	// In JSON mode hook stdout/stderr must become events, not raw writes to the
	// shared stream; otherwise they corrupt the NDJSON the client parses.
	hookOut := d.out
	if emit != nil {
		hookOut = tui.NewHookOutputWriter(emit)
	}

	// runHookStage runs one lifecycle's hooks and renders its step line. Failures
	// carry the Runtime exit code (set by the hooks runner) through emitError.
	runHookStage := func(name string, entries []config.HookEntry) error {
		checkCancel()
		if err := hooks.RunWithEvents(d.runner, entries, hookData, releaseDir, hookOut, d.confirmFn, name, hookEventFn); err != nil {
			return emitError(name, fmt.Errorf("%s hooks: %w", name, err))
		}
		if !d.jsonMode() && len(entries) > 0 {
			n := len(entries)
			sp.Step(name+" hooks", fmt.Sprintf("(%d/%d)", n, n))
		}
		return nil
	}

	if err := runHookStage("pre_extract", merged.Hooks.PreExtract); err != nil {
		return err
	}

	// Extract artifact.
	info, err := os.Stat(opts.Artifact)
	if err != nil {
		return cmderr.Wrap(cmderr.Runtime, fmt.Errorf("stat artifact: %w", err))
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
		return emitError("extract", cmderr.Wrap(cmderr.Runtime, fmt.Errorf("extracting artifact: %w", err)))
	}
	doneProgress()
	d.debug("artifact extracted", "artifact", opts.Artifact, "bytes", info.Size())
	if d.jsonMode() {
		emit.Emit(map[string]any{"event": "done", "step": "extract", "duration_ms": time.Since(extractStart).Milliseconds()})
	}
	if !d.jsonMode() {
		sp.Step("Artifact extracted", fmt.Sprintf("(%.1fs)", time.Since(extractStart).Seconds()))
	}

	// From here on the deploy can no longer be safely cancelled, so a Ctrl+C
	// gets a one-time warning instead of being silently ignored.
	cancelWarned := false
	checkCancel = func() {
		if cancelWarned || ctx.Err() == nil {
			return
		}
		cancelWarned = true
		const msg = "Deploy in progress — cannot be cancelled, continuing..."
		if d.jsonMode() {
			emit.Emit(map[string]any{"event": "warning", "message": msg})
		} else {
			tui.PrintWarning(d.mode, d.out, msg)
		}
	}

	if err := runHookStage("post_extract", merged.Hooks.PostExtract); err != nil {
		return err
	}

	if err := runHookStage("pre_link", merged.Hooks.PreLink); err != nil {
		return err
	}

	if err := runStep("link", func() (map[string]any, error) {
		err := LinkShared(merged.SharedDirs, merged.SharedFiles, releaseDir, merged.SharedRoot)
		return map[string]any{"dirs": merged.SharedDirs, "files": merged.SharedFiles}, err
	}); err != nil {
		return fmt.Errorf("linking shared resources: %w", err)
	}
	if !d.jsonMode() {
		sp.Step("Shared directories linked", fmt.Sprintf("(%d)", len(merged.SharedDirs)))
		for _, dir := range merged.SharedDirs {
			tui.PrintDetail(d.mode, d.out, dir)
		}
		sp.Step("Shared files linked", fmt.Sprintf("(%d)", len(merged.SharedFiles)))
		for _, f := range merged.SharedFiles {
			tui.PrintDetail(d.mode, d.out, f)
		}
	}

	if err := runHookStage("post_link", merged.Hooks.PostLink); err != nil {
		return err
	}

	if err := runHookStage("pre_activate", merged.Hooks.PreActivate); err != nil {
		return err
	}

	if err := runStep("current_symlink", func() (map[string]any, error) {
		err := SetCurrent(merged.ReleasesRoot, releaseDir)
		return map[string]any{"path": releaseDir}, err
	}); err != nil {
		return fmt.Errorf("updating current symlink: %w", err)
	}
	if !d.jsonMode() {
		sp.Step("current symlink updated", "")
		tui.PrintDetail(d.mode, d.out, releaseDir)
	}

	if err := runHookStage("post_activate", merged.Hooks.PostActivate); err != nil {
		return err
	}

	if err := runHookStage("pre_purge", merged.Hooks.PrePurge); err != nil {
		return err
	}

	// PurgePlan failure is non-fatal: Purge carries its own error path.
	purgePlan, purgePlanErr := PurgePlan(merged.ReleasesRoot, filepath.Base(releaseDir), merged.Settings.ReleasesToKeep)
	if purgePlanErr != nil {
		d.debug("purge plan failed; the purge step still runs", "error", purgePlanErr)
		purgePlan = nil
	}
	if err := runStep("purge", func() (map[string]any, error) {
		extras := map[string]any{"purged": purgePlan, "kept": merged.Settings.ReleasesToKeep}
		purge := func() error { return Purge(merged.ReleasesRoot, merged.Settings.ReleasesToKeep) }
		if d.jsonMode() {
			return extras, purge()
		}
		// Spin while purging (human + TTY) then resolve to "✓ Old releases purged — (kept N)".
		return extras, sp.Run("Old releases purged", func() (forgeui.Result, error) {
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

	if err := runHookStage("post_purge", merged.Hooks.PostPurge); err != nil {
		return err
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
