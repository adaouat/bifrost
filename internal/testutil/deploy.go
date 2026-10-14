//go:build integration

package testutil

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/adaouat/bifrost/internal/config"
	"github.com/adaouat/bifrost/internal/transport"
	"github.com/adaouat/bifrost/internal/tui"
	forgeui "github.com/adaouat/forge/ui"
)

// DeployOverSSH stages the agent binary, the flat config for (env, app), and the
// artifact on the remote, runs `bifrost-agent deploy --output json`, forwards the
// event stream in mode, and returns the rendered output. It mirrors the client
// deploy path and manages its own connection and staging cleanup. BinaryPath
// must be set (the package TestMain builds it).
func DeployOverSSH(ctx context.Context, t *testing.T, srv config.ResolvedServer, cfg *config.Config, env, app, artifactPath, releaseName string, mode forgeui.Mode) string {
	t.Helper()

	session, err := transport.StageAgentSession("dev", cfg, env, app, BinaryPath, srv)
	require.NoError(t, err, "staging agent session")
	defer session.Close()

	remoteArtifact, err := session.Staging.Upload(artifactPath, filepath.Base(artifactPath))
	require.NoError(t, err, "upload artifact")

	agentCmd := fmt.Sprintf("%s deploy --output json --config %s --artifact %s --release-name %s",
		transport.ShellQuote(session.RemoteAgent), transport.ShellQuote(session.RemoteConfig),
		transport.ShellQuote(remoteArtifact), transport.ShellQuote(releaseName))

	pr, pw := io.Pipe()
	var execResult *transport.ExecResult
	var execErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer func() { _ = pw.Close(); wg.Done() }()
		execResult, execErr = session.Client.ExecStream(agentCmd, pw)
	}()

	var out bytes.Buffer
	_, streamErr := tui.ForwardStream(pr, srv.Name, mode, &out)
	if streamErr != nil {
		// Drain instead of closing early — see execAndForward (M26): closing
		// here would stop the writer goroutine from draining the SSH channel,
		// which can hang the remote agent trying to flush its own stdout.
		go func() { _, _ = io.Copy(io.Discard, pr) }()
	}
	wg.Wait()
	_ = pr.Close()

	require.NoError(t, execErr, "agent exec")
	require.Equalf(t, 0, execResult.ExitCode, "agent exited non-zero: %s", execResult.Stderr.String())
	require.NoError(t, streamErr, "event stream")
	return out.String()
}
