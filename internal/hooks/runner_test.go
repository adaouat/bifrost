package hooks

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/adaouat/bifrost/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHelperProcess is a subprocess helper invoked by captureExec.
// It exits with code 0 or 1 depending on BIFROST_TEST_EXIT.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("BIFROST_TEST_HELPER") != "1" {
		return
	}
	if os.Getenv("BIFROST_TEST_EXIT") == "1" {
		fmt.Fprintln(os.Stderr, "hook stderr output")
		os.Exit(1)
	}
	os.Exit(0)
}

type cmdCall struct {
	name string
	args []string
}

// captureExec returns a mock execCommand that records calls and exits the helper subprocess with exitCode.
func captureExec(calls *[]cmdCall, exitCode int) func(string, ...string) *exec.Cmd {
	return func(name string, args ...string) *exec.Cmd {
		*calls = append(*calls, cmdCall{name, args})
		cmd := exec.Command(os.Args[0], "-test.run=^TestHelperProcess$")
		cmd.Env = append(os.Environ(), "BIFROST_TEST_HELPER=1")
		if exitCode != 0 {
			cmd.Env = append(cmd.Env, fmt.Sprintf("BIFROST_TEST_EXIT=%d", exitCode))
		}
		return cmd
	}
}

func prio(n int) *int { return &n }

func TestRun_Empty(t *testing.T) {
	var out bytes.Buffer
	require.NoError(t, Run(nil, HookData{}, "/tmp", &out))
}

func TestRun_BasicShExecution(t *testing.T) {
	var calls []cmdCall
	execCommand = captureExec(&calls, 0)
	t.Cleanup(func() { execCommand = exec.Command })

	hooks := []config.HookEntry{{Cmd: "echo hello", Priority: prio(99999)}}
	var out bytes.Buffer
	require.NoError(t, Run(hooks, HookData{}, "/tmp", &out))
	require.Len(t, calls, 1)
	assert.Equal(t, "sh", calls[0].name)
	assert.Equal(t, []string{"-c", "echo hello"}, calls[0].args)
}

func TestRun_SudoWrapping(t *testing.T) {
	var calls []cmdCall
	execCommand = captureExec(&calls, 0)
	t.Cleanup(func() { execCommand = exec.Command })

	hooks := []config.HookEntry{{Cmd: "systemctl reload nginx", Sudo: true, Priority: prio(99999)}}
	var out bytes.Buffer
	require.NoError(t, Run(hooks, HookData{}, "/tmp", &out))
	require.Len(t, calls, 1)
	assert.Equal(t, "sudo", calls[0].name)
	assert.Equal(t, []string{"sh", "-c", "systemctl reload nginx"}, calls[0].args)
}

func TestRun_TemplateRendering(t *testing.T) {
	var calls []cmdCall
	execCommand = captureExec(&calls, 0)
	t.Cleanup(func() { execCommand = exec.Command })

	hooks := []config.HookEntry{{
		Cmd:      "echo {{ .Directories.Working }} {{ .Variables.env }}",
		Priority: prio(99999),
	}}
	data := HookData{
		Directories: Directories{Working: "/releases/20260524"},
		Variables:   map[string]string{"env": "prod"},
	}
	var out bytes.Buffer
	require.NoError(t, Run(hooks, data, "/tmp", &out))
	require.Len(t, calls, 1)
	assert.Equal(t, []string{"-c", "echo /releases/20260524 prod"}, calls[0].args)
}

func TestRun_InvalidTemplateReturnsError(t *testing.T) {
	hooks := []config.HookEntry{{Cmd: "{{ .Unclosed", Priority: prio(99999)}}
	var out bytes.Buffer
	require.Error(t, Run(hooks, HookData{}, "/tmp", &out))
}

func TestRun_NonZeroExitReturnsError(t *testing.T) {
	var calls []cmdCall
	execCommand = captureExec(&calls, 1)
	t.Cleanup(func() { execCommand = exec.Command })

	hooks := []config.HookEntry{{Cmd: "exit 1", Priority: prio(99999)}}
	var out bytes.Buffer
	require.Error(t, Run(hooks, HookData{}, "/tmp", &out))
}

func TestRun_CmdDirOverridesWorkingDir(t *testing.T) {
	var captured *exec.Cmd
	execCommand = func(name string, args ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=^TestHelperProcess$")
		cmd.Env = append(os.Environ(), "BIFROST_TEST_HELPER=1")
		captured = cmd
		return cmd
	}
	t.Cleanup(func() { execCommand = exec.Command })

	// workingDir is "/" and CmdDir is "/tmp" — both exist, verifies CmdDir wins
	hooks := []config.HookEntry{{
		Cmd:      "echo dir",
		CmdDir:   "/tmp",
		Priority: prio(99999),
	}}
	var out bytes.Buffer
	require.NoError(t, Run(hooks, HookData{}, "/", &out))
	require.NotNil(t, captured)
	assert.Equal(t, "/tmp", captured.Dir)
}

func TestRun_AllowFailContinuesOnNonZeroExit(t *testing.T) {
	var calls []cmdCall
	callNum := 0
	execCommand = func(name string, args ...string) *exec.Cmd {
		calls = append(calls, cmdCall{name, args})
		callNum++
		cmd := exec.Command(os.Args[0], "-test.run=^TestHelperProcess$")
		env := append(os.Environ(), "BIFROST_TEST_HELPER=1")
		if callNum == 1 {
			env = append(env, "BIFROST_TEST_EXIT=1")
		}
		cmd.Env = env
		return cmd
	}
	t.Cleanup(func() { execCommand = exec.Command })

	hooks := []config.HookEntry{
		{Cmd: "fail", AllowFail: true, Priority: prio(10)},
		{Cmd: "succeed", Priority: prio(20)},
	}
	var out bytes.Buffer
	require.NoError(t, Run(hooks, HookData{}, "/tmp", &out))
	assert.Len(t, calls, 2)
}

func TestRun_AllowFailWritesWarning(t *testing.T) {
	execCommand = captureExec(new([]cmdCall), 1)
	t.Cleanup(func() { execCommand = exec.Command })

	hooks := []config.HookEntry{{Cmd: "fail", AllowFail: true, Priority: prio(99999)}}
	var out bytes.Buffer
	require.NoError(t, Run(hooks, HookData{}, "/tmp", &out))
	assert.Contains(t, out.String(), "warning")
}

func TestRun_ExecutesInOrder(t *testing.T) {
	var calls []cmdCall
	execCommand = captureExec(&calls, 0)
	t.Cleanup(func() { execCommand = exec.Command })

	hooks := []config.HookEntry{
		{Cmd: "first", Priority: prio(10)},
		{Cmd: "second", Priority: prio(20)},
		{Cmd: "third", Priority: prio(30)},
	}
	var out bytes.Buffer
	require.NoError(t, Run(hooks, HookData{}, "/tmp", &out))
	require.Len(t, calls, 3)
	assert.Equal(t, "first", calls[0].args[1])
	assert.Equal(t, "second", calls[1].args[1])
	assert.Equal(t, "third", calls[2].args[1])
}
