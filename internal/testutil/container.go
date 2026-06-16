//go:build integration

package testutil

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/pkg/stdcopy"
	moby "github.com/moby/moby/client"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const containerImage = "debian:bookworm-slim"

// Container wraps a testcontainers container with helpers for integration tests.
type Container struct {
	inner testcontainers.Container
}

// NewContainer starts a Linux container with the bifrost binary available at /usr/local/bin/bifrost.
// The container is automatically terminated when the test ends.
func NewContainer(ctx context.Context, t testing.TB, binaryPath string) *Container {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image: containerImage,
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      binaryPath,
				ContainerFilePath: "/usr/local/bin/bifrost",
				FileMode:          0o755,
			},
		},
		Cmd:        []string{"sleep", "infinity"},
		WaitingFor: wait.ForExec([]string{"true"}),
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("NewContainer: %v", err)
	}

	t.Cleanup(func() {
		_ = c.Terminate(context.Background())
	})

	return &Container{inner: c}
}

// ExecResult holds the output of a container command execution.
type ExecResult struct {
	ExitCode int
	Output   string
	Stderr   string
}

// Exec runs a command inside the container and returns the result.
// stdout and stderr are demultiplexed; Output contains stdout, Stderr contains stderr.
func (c *Container) Exec(ctx context.Context, cmd []string) (ExecResult, error) {
	code, reader, err := c.inner.Exec(ctx, cmd)
	if err != nil {
		return ExecResult{}, err
	}

	var stdout, stderr bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdout, &stderr, reader); err != nil {
		return ExecResult{}, err
	}

	return ExecResult{
		ExitCode: code,
		Output:   strings.TrimRight(stdout.String(), "\n"),
		Stderr:   strings.TrimRight(stderr.String(), "\n"),
	}, nil
}

// RunBifrost runs the bifrost binary with the given arguments.
func (c *Container) RunBifrost(ctx context.Context, args ...string) (ExecResult, error) {
	return c.Exec(ctx, append([]string{"/usr/local/bin/bifrost"}, args...))
}

// CopyFile copies content into the container at containerPath with the given file mode.
func (c *Container) CopyFile(ctx context.Context, content []byte, containerPath string, mode int64) error {
	return c.inner.CopyToContainer(ctx, content, containerPath, mode)
}

// ExecWithTTY runs cmd in the container with a pseudo-terminal allocated and
// feeds input as stdin. Use this for commands that require IsTTY()=true (e.g.
// `bifrost deploy --interactive`). stdout and stderr are merged in the raw TTY
// stream and returned as Output; Stderr is always empty.
//
// inputDelay sleeps before writing any stdin. Use it when the target process
// calls term.MakeRaw before reading (e.g. bubbletea/huh): without a delay
// the PTY's canonical ICRNL converts CR→LF, and LF (Ctrl+J) is not KeyEnter.
//
// inputInterval writes input bytes one at a time with a sleep between each.
// Use it when the target runs multiple independent bubbletea programs in
// sequence (e.g. deploy --interactive): bubbletea's read goroutine calls
// Read(4096) and may consume all buffered bytes in one call, discarding any
// beyond the first KeyEnter. Writing one byte per interval ensures each
// program reads exactly the one CR it needs before stdin is empty.
func (c *Container) ExecWithTTY(ctx context.Context, cmd []string, input []byte, inputDelay, inputInterval time.Duration) (ExecResult, error) {
	cli, err := moby.New(moby.FromEnv)
	if err != nil {
		return ExecResult{}, err
	}
	defer func() { _ = cli.Close() }()

	exec, err := cli.ExecCreate(ctx, c.inner.GetContainerID(), moby.ExecCreateOptions{
		Cmd:          cmd,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		TTY:          true,
	})
	if err != nil {
		return ExecResult{}, err
	}

	attach, err := cli.ExecAttach(ctx, exec.ID, moby.ExecAttachOptions{TTY: true})
	if err != nil {
		return ExecResult{}, err
	}

	// Drain output into a buffer. With TTY=true the stream is raw (no Docker
	// multiplexing headers). Some Docker runtimes (e.g. OrbStack) do not close
	// the hijacked connection when the exec process exits, so we poll ExecInspect
	// instead of relying on EOF, then close the connection to unblock the reader.
	var outBuf bytes.Buffer
	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		_, _ = io.Copy(&outBuf, attach.Reader)
	}()

	// Write stdin in a separate goroutine so it doesn't block the poll loop.
	// Do NOT close the connection after writing — closing the write side of a
	// hijacked TTY sends SIGHUP/EIO to the container process prematurely.
	go func() {
		if inputDelay > 0 {
			time.Sleep(inputDelay)
		}
		if inputInterval > 0 {
			for _, b := range input {
				_, _ = attach.Conn.Write([]byte{b})
				time.Sleep(inputInterval)
			}
		} else {
			_, _ = attach.Conn.Write(input)
		}
	}()

	// Poll ExecInspect until the process exits, then close the connection so the
	// read goroutine can drain and return.
	var exitCode int
	for {
		if ctx.Err() != nil {
			attach.Close()
			<-readDone
			return ExecResult{}, ctx.Err()
		}
		info, inspectErr := cli.ExecInspect(ctx, exec.ID, moby.ExecInspectOptions{})
		if inspectErr != nil {
			attach.Close()
			<-readDone
			return ExecResult{}, inspectErr
		}
		if !info.Running {
			exitCode = info.ExitCode
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	attach.Close()
	<-readDone

	return ExecResult{
		ExitCode: exitCode,
		Output:   strings.TrimRight(outBuf.String(), "\n"),
	}, nil
}
