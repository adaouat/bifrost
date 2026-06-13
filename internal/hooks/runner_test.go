package hooks

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/adaouat/bifrost/internal/config"
	"github.com/adaouat/forge/exec/exectest"
	"github.com/adaouat/forge/exitcode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// okRunner returns a MockRunner with n queued success responses.
func okRunner(n int) *exectest.MockRunner {
	mr := exectest.NewMockRunner()
	for range n {
		mr.QueueResponse("", "", nil)
	}
	return mr
}

func prio(n int) *int { return &n }

func TestRun_Empty(t *testing.T) {
	var out bytes.Buffer
	require.NoError(t, Run(exectest.NewMockRunner(), nil, HookData{}, "/tmp", &out))
}

func TestRun_BasicShExecution(t *testing.T) {
	mr := okRunner(1)
	hooks := []config.HookEntry{{Cmd: "echo hello", Priority: prio(99999)}}
	var out bytes.Buffer
	require.NoError(t, Run(mr, hooks, HookData{}, "/tmp", &out))
	require.Len(t, mr.Calls, 1)
	assert.Equal(t, "sh", mr.Calls[0].Name)
	assert.Equal(t, []string{"-c", "echo hello"}, mr.Calls[0].Args)
}

func TestRun_SudoWrapping(t *testing.T) {
	mr := okRunner(1)
	hooks := []config.HookEntry{{Cmd: "systemctl reload nginx", Sudo: true, Priority: prio(99999)}}
	var out bytes.Buffer
	require.NoError(t, Run(mr, hooks, HookData{}, "/tmp", &out))
	require.Len(t, mr.Calls, 1)
	assert.Equal(t, "sudo", mr.Calls[0].Name)
	assert.Equal(t, []string{"sh", "-c", "systemctl reload nginx"}, mr.Calls[0].Args)
}

func TestRun_TemplateRendering(t *testing.T) {
	mr := okRunner(1)
	hooks := []config.HookEntry{{
		Cmd:      "echo {{ .Directories.Working }} {{ .Variables.env }}",
		Priority: prio(99999),
	}}
	data := HookData{
		Directories: Directories{Working: "/releases/20260524"},
		Variables:   map[string]string{"env": "prod"},
	}
	var out bytes.Buffer
	require.NoError(t, Run(mr, hooks, data, "/tmp", &out))
	require.Len(t, mr.Calls, 1)
	assert.Equal(t, []string{"-c", "echo /releases/20260524 prod"}, mr.Calls[0].Args)
}

func TestRun_InvalidTemplateReturnsError(t *testing.T) {
	mr := exectest.NewMockRunner()
	hooks := []config.HookEntry{{Cmd: "{{ .Unclosed", Priority: prio(99999)}}
	var out bytes.Buffer
	require.Error(t, Run(mr, hooks, HookData{}, "/tmp", &out))
	assert.Empty(t, mr.Calls, "command must not run when the template is invalid")
}

func TestRun_NonZeroExitReturnsError(t *testing.T) {
	mr := exectest.NewMockRunner()
	mr.QueueResponse("", "", errors.New("exit status 1"))
	hooks := []config.HookEntry{{Cmd: "exit 1", Priority: prio(99999)}}
	var out bytes.Buffer
	require.Error(t, Run(mr, hooks, HookData{}, "/tmp", &out))
}

func TestRun_HookFailureCarriesRuntimeExitCode(t *testing.T) {
	mr := exectest.NewMockRunner()
	mr.QueueResponse("", "", errors.New("exit status 1"))
	hooks := []config.HookEntry{{Cmd: "false", Priority: prio(99999)}}
	var out bytes.Buffer
	err := Run(mr, hooks, HookData{}, "/tmp", &out)
	require.Error(t, err)
	assert.Equal(t, exitcode.Runtime, exitcode.Resolve(err), "a failing hook must resolve to Runtime (3)")
}

func TestRun_TemplateErrorCarriesRuntimeExitCode(t *testing.T) {
	mr := exectest.NewMockRunner()
	hooks := []config.HookEntry{{Cmd: "{{ .Unclosed", Priority: prio(99999)}}
	var out bytes.Buffer
	err := Run(mr, hooks, HookData{}, "/tmp", &out)
	require.Error(t, err)
	assert.Equal(t, exitcode.Runtime, exitcode.Resolve(err), "a template error must resolve to Runtime (3)")
}

func TestRun_CmdDirOverridesWorkingDir(t *testing.T) {
	mr := okRunner(1)
	// workingDir is "/" and CmdDir is "/tmp" — verifies CmdDir wins.
	hooks := []config.HookEntry{{
		Cmd:      "echo dir",
		CmdDir:   "/tmp",
		Priority: prio(99999),
	}}
	var out bytes.Buffer
	require.NoError(t, Run(mr, hooks, HookData{}, "/", &out))
	require.Len(t, mr.Calls, 1)
	assert.Equal(t, "/tmp", mr.Calls[0].Dir)
}

func TestRun_AllowFailContinuesOnNonZeroExit(t *testing.T) {
	mr := exectest.NewMockRunner()
	mr.QueueResponse("", "", errors.New("exit status 1"))
	mr.QueueResponse("", "", nil)

	hooks := []config.HookEntry{
		{Cmd: "fail", AllowFail: true, Priority: prio(10)},
		{Cmd: "succeed", Priority: prio(20)},
	}
	var out bytes.Buffer
	require.NoError(t, Run(mr, hooks, HookData{}, "/tmp", &out))
	assert.Len(t, mr.Calls, 2)
}

func TestRun_AllowFailWritesWarning(t *testing.T) {
	mr := exectest.NewMockRunner()
	mr.QueueResponse("", "", errors.New("exit status 1"))

	hooks := []config.HookEntry{{Cmd: "fail", AllowFail: true, Priority: prio(99999)}}
	var out bytes.Buffer
	require.NoError(t, Run(mr, hooks, HookData{}, "/tmp", &out))
	assert.Contains(t, out.String(), "warning")
}

func TestRun_AllowFailWarningDoesNotDuplicateStderr(t *testing.T) {
	mr := exectest.NewMockRunner()
	// forge's CmdRunner folds stderr into the returned error; the warning must
	// not echo that stderr a second time after it was already streamed to out.
	mr.QueueResponse("", "boom\n", errors.New("sh: exit status 1: boom"))

	hooks := []config.HookEntry{{Cmd: "fail", AllowFail: true, Priority: prio(99999)}}
	var out bytes.Buffer
	require.NoError(t, Run(mr, hooks, HookData{}, "/tmp", &out))

	assert.Equal(t, 1, strings.Count(out.String(), "boom"), "stderr must appear once, not again in the warning")
	assert.Contains(t, out.String(), `warning: hook "fail" failed (allow_fail): exit status 1`)
}

func TestRun_InteractiveSkippedWhenNotEnabled(t *testing.T) {
	mr := exectest.NewMockRunner()
	hooks := []config.HookEntry{{Cmd: "echo hi", Interactive: true, Priority: prio(99999)}}
	var out bytes.Buffer
	// interactive=false means prompts are disabled — hook is skipped with a warning.
	require.NoError(t, Run(mr, hooks, HookData{}, "/tmp", &out))
	assert.Empty(t, mr.Calls, "interactive hook must not execute when prompts disabled")
	assert.Contains(t, out.String(), "warning")
}

func TestRun_InteractiveExecutesWhenConfirmed(t *testing.T) {
	mr := okRunner(1)
	hooks := []config.HookEntry{{Cmd: "echo hi", Interactive: true, Priority: prio(99999)}}
	var out bytes.Buffer
	// confirmFn returns true — hook proceeds.
	require.NoError(t, RunInteractive(mr, hooks, HookData{}, "/tmp", &out, func(cmd string) bool { return true }))
	assert.Len(t, mr.Calls, 1)
}

func TestRun_InteractiveSkippedWhenDeclined(t *testing.T) {
	mr := exectest.NewMockRunner()
	hooks := []config.HookEntry{{Cmd: "echo hi", Interactive: true, Priority: prio(99999)}}
	var out bytes.Buffer
	// confirmFn returns false — hook is skipped with a warning.
	require.NoError(t, RunInteractive(mr, hooks, HookData{}, "/tmp", &out, func(cmd string) bool { return false }))
	assert.Empty(t, mr.Calls)
	assert.Contains(t, out.String(), "warning")
}

func TestRun_TemplateRendering_AllDirectories(t *testing.T) {
	mr := okRunner(1)
	hooks := []config.HookEntry{{
		Cmd:      "echo {{ .Directories.Working }} {{ .Directories.Current }} {{ .Directories.Releases }} {{ .Directories.Shared }}",
		Priority: prio(99999),
	}}
	data := HookData{
		Directories: Directories{
			Working:  "/releases/20260524",
			Current:  "/releases/current",
			Releases: "/releases",
			Shared:   "/shared",
		},
	}
	var out bytes.Buffer
	require.NoError(t, Run(mr, hooks, data, "/tmp", &out))
	require.Len(t, mr.Calls, 1)
	assert.Equal(t, []string{"-c", "echo /releases/20260524 /releases/current /releases /shared"}, mr.Calls[0].Args)
}

func TestRun_TemplateRendering_Settings(t *testing.T) {
	mr := okRunner(1)
	hooks := []config.HookEntry{{
		Cmd:      "echo keep={{ .Settings.ReleasesToKeep }}",
		Priority: prio(99999),
	}}
	data := HookData{
		Settings: config.Settings{ReleasesToKeep: 7},
	}
	var out bytes.Buffer
	require.NoError(t, Run(mr, hooks, data, "/tmp", &out))
	require.Len(t, mr.Calls, 1)
	assert.Equal(t, []string{"-c", "echo keep=7"}, mr.Calls[0].Args)
}

func TestRun_TemplateRendering_Env(t *testing.T) {
	mr := okRunner(1)
	hooks := []config.HookEntry{{
		Cmd:      "echo db={{ .Env.DB_HOST }}",
		Priority: prio(99999),
	}}
	data := HookData{
		Env: map[string]string{"DB_HOST": "db.prod.internal"},
	}
	var out bytes.Buffer
	require.NoError(t, Run(mr, hooks, data, "/tmp", &out))
	require.Len(t, mr.Calls, 1)
	assert.Equal(t, []string{"-c", "echo db=db.prod.internal"}, mr.Calls[0].Args)
}

func TestRun_ExecutesInOrder(t *testing.T) {
	mr := okRunner(3)
	hooks := []config.HookEntry{
		{Cmd: "first", Priority: prio(10)},
		{Cmd: "second", Priority: prio(20)},
		{Cmd: "third", Priority: prio(30)},
	}
	var out bytes.Buffer
	require.NoError(t, Run(mr, hooks, HookData{}, "/tmp", &out))
	require.Len(t, mr.Calls, 3)
	assert.Equal(t, "first", mr.Calls[0].Args[1])
	assert.Equal(t, "second", mr.Calls[1].Args[1])
	assert.Equal(t, "third", mr.Calls[2].Args[1])
}
