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

// Run executes hooks in order. Hooks must already be sorted by priority
// (config.Merge guarantees this). workingDir is the default working directory
// for each hook. Hook stdout and stderr are written to out.
func Run(hooks []config.HookEntry, data HookData, workingDir string, out io.Writer) error {
	for _, h := range hooks {
		if err := runOne(h, data, workingDir, out); err != nil {
			return err
		}
	}
	return nil
}

func runOne(h config.HookEntry, data HookData, workingDir string, out io.Writer) error {
	rendered, err := renderCmd(h.Cmd, data)
	if err != nil {
		return fmt.Errorf("hook template: %w", err)
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
		return fmt.Errorf("hook %q: %w", h.Cmd, runErr)
	}

	return nil
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
