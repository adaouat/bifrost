//go:build integration

package testutil

import (
	"context"
	"io"
	"strings"
	"testing"

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
}

// Exec runs a command inside the container and returns the result.
func (c *Container) Exec(ctx context.Context, cmd []string) (ExecResult, error) {
	code, reader, err := c.inner.Exec(ctx, cmd)
	if err != nil {
		return ExecResult{}, err
	}

	out, err := io.ReadAll(reader)
	if err != nil {
		return ExecResult{}, err
	}

	return ExecResult{
		ExitCode: code,
		Output:   strings.TrimRight(string(out), "\n"),
	}, nil
}

// RunBifrost runs the bifrost binary with the given arguments.
func (c *Container) RunBifrost(ctx context.Context, args ...string) (ExecResult, error) {
	return c.Exec(ctx, append([]string{"/usr/local/bin/bifrost"}, args...))
}
