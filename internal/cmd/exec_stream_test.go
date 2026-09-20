package cmd

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adaouat/bifrost/internal/config"
	"github.com/adaouat/bifrost/internal/transport"
	forgeui "github.com/adaouat/forge/ui"
)

func TestDeployAgentCmd_WithoutInit(t *testing.T) {
	cmd := deployAgentCmd("/tmp/x/bifrost-agent", "/tmp/x/bifrost.yml", "/tmp/x/app.tar.gz", "r1", false)
	assert.NotContains(t, cmd, "--init")
}

func TestDeployAgentCmd_WithInit(t *testing.T) {
	cmd := deployAgentCmd("/tmp/x/bifrost-agent", "/tmp/x/bifrost.yml", "/tmp/x/app.tar.gz", "r1", true)
	assert.Contains(t, cmd, "--init", "--init must be forwarded to the remote agent so it can create releases_root/shared_root")
}

func TestExecAndForward_StreamErrorDrainsRemainingOutput(t *testing.T) {
	writeCompleted := make(chan struct{})
	execFn := func(w io.Writer) (*transport.ExecResult, error) {
		_, err := w.Write([]byte(`{"event":"error","step":"post_activate","message":"boom","exit_code":3}` + "\n"))
		require.NoError(t, err, "first write must succeed")

		// Simulate more output arriving after ForwardStream has already
		// returned on the error event above. This must still succeed —
		// otherwise the real SSH copy goroutine would stop draining the
		// channel and the remote agent process could hang indefinitely
		// waiting for a window update that never comes.
		_, err = w.Write([]byte("trailing hook output\n"))
		assert.NoError(t, err, "write after the stream error must not fail — draining must still be active")
		close(writeCompleted)

		return &transport.ExecResult{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}, ExitCode: 3}, nil
	}

	var out bytes.Buffer
	err := execAndForward(execFn, config.ResolvedServer{Name: "web-01"}, forgeui.Plain, &out)
	require.Error(t, err)

	select {
	case <-writeCompleted:
	case <-time.After(2 * time.Second):
		t.Fatal("second write on the pipe never completed — the reader was closed too early")
	}
}

func TestExecAndForward_StreamErrorTakesPrecedenceOverExecErr(t *testing.T) {
	execFn := func(w io.Writer) (*transport.ExecResult, error) {
		_, _ = w.Write([]byte(`{"event":"error","step":"post_activate","message":"boom","exit_code":3}` + "\n"))
		// Simulates the pipe-close race: the SSH session errors out because
		// the reader side was torn down, even though the agent already
		// reported a real, more specific failure.
		return nil, io.ErrClosedPipe
	}

	err := execAndForward(execFn, config.ResolvedServer{Name: "web-01"}, forgeui.Plain, &bytes.Buffer{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
	assert.NotErrorIs(t, err, io.ErrClosedPipe, "the agent's real error must not be masked by the pipe-close race")

	var exitErr *ExitError
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 3, exitErr.Code, "exit code must come from the agent's error event, not default to 1")
}

func TestExecAndForward_Success(t *testing.T) {
	execFn := func(w io.Writer) (*transport.ExecResult, error) {
		_, _ = w.Write([]byte(`{"event":"done","step":"deploy","release":"r1","duration_ms":1}` + "\n"))
		return &transport.ExecResult{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}, ExitCode: 0}, nil
	}

	err := execAndForward(execFn, config.ResolvedServer{Name: "web-01"}, forgeui.Plain, &bytes.Buffer{})
	require.NoError(t, err)
}

func TestExecAndForward_NonZeroExit_NoStreamError(t *testing.T) {
	execFn := func(w io.Writer) (*transport.ExecResult, error) {
		_, _ = w.Write([]byte(`{"event":"done","step":"deploy","release":"r1","duration_ms":1}` + "\n"))
		return &transport.ExecResult{Stdout: &bytes.Buffer{}, Stderr: bytes.NewBufferString("agent stderr"), ExitCode: 9}, nil
	}

	err := execAndForward(execFn, config.ResolvedServer{Name: "web-01"}, forgeui.Plain, &bytes.Buffer{})
	require.Error(t, err)

	var exitErr *ExitError
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 9, exitErr.Code)
	assert.Contains(t, exitErr.Message, "agent stderr")
}
