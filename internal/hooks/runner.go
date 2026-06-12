package hooks

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"text/template"

	"github.com/adaouat/bifrost/internal/config"
	forgeexec "github.com/adaouat/forge/exec"
)

// HookData is the template context available in hook cmd fields.
type HookData struct {
	Settings    config.Settings
	Variables   map[string]string
	Directories Directories
	Env         map[string]string
}

// Directories holds deployment path helpers available in hook templates.
type Directories struct {
	Working  string // Release directory being activated
	Current  string // Path to the `current` symlink
	Releases string // Releases root directory
	Shared   string // Shared root directory
}

// HookEventFn is called after each hook executes with its lifecycle name, zero-based index,
// rendered command string, and exit code (0 on success).
type HookEventFn func(lifecycle string, index int, cmd string, exitCode int)

// Run executes hooks in order, skipping interactive hooks with a warning.
// Hooks must already be sorted by priority (config.Merge guarantees this).
// workingDir is the default working directory. Hook output is written to out.
func Run(runner forgeexec.Runner, hooks []config.HookEntry, data HookData, workingDir string, out io.Writer) error {
	return RunInteractive(runner, hooks, data, workingDir, out, nil)
}

// RunInteractive is like Run but calls confirmFn before each interactive hook.
// If confirmFn is nil or returns false, the hook is skipped with a warning.
func RunInteractive(runner forgeexec.Runner, hooks []config.HookEntry, data HookData, workingDir string, out io.Writer, confirmFn func(cmd string) bool) error {
	return RunWithEvents(runner, hooks, data, workingDir, out, confirmFn, "", nil)
}

// RunWithEvents is like RunInteractive but calls eventFn after each hook executes.
// lifecycle is the name of the hook group (e.g. "pre_activate"). eventFn may be nil.
func RunWithEvents(runner forgeexec.Runner, hooks []config.HookEntry, data HookData, workingDir string, out io.Writer, confirmFn func(cmd string) bool, lifecycle string, eventFn HookEventFn) error {
	for i, h := range hooks {
		exitCode, err := runOne(runner, h, data, workingDir, out, confirmFn)
		if eventFn != nil {
			rendered, _ := renderCmd(h.Cmd, data)
			eventFn(lifecycle, i, rendered, exitCode)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// runOne executes a single hook and returns its exit code and any error.
// Exit code 0 means success; -1 means the hook was skipped.
func runOne(runner forgeexec.Runner, h config.HookEntry, data HookData, workingDir string, out io.Writer, confirmFn func(cmd string) bool) (exitCode int, err error) {
	if h.Interactive {
		if confirmFn == nil || !confirmFn(h.Cmd) {
			_, _ = fmt.Fprintf(out, "warning: skipping interactive hook %q\n", h.Cmd)
			return -1, nil
		}
	}

	rendered, err := renderCmd(h.Cmd, data)
	if err != nil {
		return -1, fmt.Errorf("hook template: %w", err)
	}

	dir := workingDir
	if h.CmdDir != "" {
		dir = h.CmdDir
	}

	name := "sh"
	args := []string{"-c", rendered}
	if h.Sudo {
		name = "sudo"
		args = []string{"sh", "-c", rendered}
	}

	stdout, stderr, runErr := runner.RunDir(dir, nil, name, args...)

	if stdout != "" {
		_, _ = fmt.Fprint(out, stdout)
	}

	if runErr != nil {
		if stderr != "" {
			_, _ = fmt.Fprint(out, stderr)
		}
		code := 1
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			code = exitErr.ExitCode()
		}
		if h.AllowFail {
			// Report the exit status, not runErr: forge embeds the captured
			// stderr in its error, which was already streamed to out above.
			_, _ = fmt.Fprintf(out, "warning: hook %q failed (allow_fail): exit status %d\n", h.Cmd, code)
			return code, nil
		}
		return code, fmt.Errorf("hook %q: %w", h.Cmd, runErr)
	}

	return 0, nil
}

func renderCmd(tmpl string, data HookData) (string, error) {
	t, err := template.New("hook").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("parsing %q: %w", tmpl, err)
	}
	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing %q: %w", tmpl, err)
	}
	return buf.String(), nil
}
