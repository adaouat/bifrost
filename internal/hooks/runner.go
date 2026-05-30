package hooks

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"text/template"

	"github.com/adaouat/bifrost/internal/config"
)

// execCommand is the function used to create exec.Cmd instances.
// Replaced in tests to avoid shell execution on the local host.
var execCommand = exec.Command

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
func Run(hooks []config.HookEntry, data HookData, workingDir string, out io.Writer) error {
	return RunInteractive(hooks, data, workingDir, out, nil)
}

// RunInteractive is like Run but calls confirmFn before each interactive hook.
// If confirmFn is nil or returns false, the hook is skipped with a warning.
func RunInteractive(hooks []config.HookEntry, data HookData, workingDir string, out io.Writer, confirmFn func(cmd string) bool) error {
	return RunWithEvents(hooks, data, workingDir, out, confirmFn, "", nil)
}

// RunWithEvents is like RunInteractive but calls eventFn after each hook executes.
// lifecycle is the name of the hook group (e.g. "pre_enable_release"). eventFn may be nil.
func RunWithEvents(hooks []config.HookEntry, data HookData, workingDir string, out io.Writer, confirmFn func(cmd string) bool, lifecycle string, eventFn HookEventFn) error {
	for i, h := range hooks {
		exitCode, err := runOne(h, data, workingDir, out, confirmFn)
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
func runOne(h config.HookEntry, data HookData, workingDir string, out io.Writer, confirmFn func(cmd string) bool) (exitCode int, err error) {
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

	var cmd *exec.Cmd
	if h.Sudo {
		cmd = execCommand("sudo", "sh", "-c", rendered)
	} else {
		cmd = execCommand("sh", "-c", rendered)
	}

	dir := workingDir
	if h.CmdDir != "" {
		dir = h.CmdDir
	}
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()

	if stdout.Len() > 0 {
		_, _ = fmt.Fprint(out, stdout.String())
	}

	if runErr != nil {
		if stderr.Len() > 0 {
			_, _ = fmt.Fprint(out, stderr.String())
		}
		code := 1
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		}
		if h.AllowFail {
			_, _ = fmt.Fprintf(out, "warning: hook %q failed (allow_fail): %v\n", h.Cmd, runErr)
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
