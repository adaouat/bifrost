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
// deploy path (shell-quoted args, SFTP chmod) and manages its own connection and
// staging cleanup. BinaryPath must be set (the package TestMain builds it).
func DeployOverSSH(ctx context.Context, t *testing.T, srv config.ResolvedServer, cfg *config.Config, env, app, artifactPath, releaseName string, mode forgeui.Mode) string {
	t.Helper()

	client, err := transport.Connect(srv)
	require.NoError(t, err, "ssh connect")
	defer func() { _ = client.Close() }()

	staging, err := transport.NewStaging(client, srv.StagingDir)
	require.NoError(t, err, "staging")
	defer func() { _ = staging.Cleanup() }()

	remoteAgent, err := staging.Upload(BinaryPath, "bifrost-agent")
	require.NoError(t, err, "upload agent")

	var flat bytes.Buffer
	require.NoError(t, config.GenerateFlatConfig(cfg, env, app, &flat), "generate flat config")
	remoteConfig, err := staging.UploadBytes(flat.Bytes(), "bifrost.yml")
	require.NoError(t, err, "upload config")

	remoteArtifact, err := staging.Upload(artifactPath, filepath.Base(artifactPath))
	require.NoError(t, err, "upload artifact")

	require.NoError(t, staging.Chmod(remoteAgent, 0o755), "chmod agent")

	agentCmd := fmt.Sprintf("%s deploy --output json --config %s --artifact %s --release-name %s",
		transport.ShellQuote(remoteAgent), transport.ShellQuote(remoteConfig),
		transport.ShellQuote(remoteArtifact), transport.ShellQuote(releaseName))

	pr, pw := io.Pipe()
	var execResult *transport.ExecResult
	var execErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer func() { _ = pw.Close(); wg.Done() }()
		execResult, execErr = client.ExecStream(agentCmd, pw)
	}()

	var out bytes.Buffer
	_, streamErr := tui.ForwardStream(pr, srv.Name, mode, &out)
	_ = pr.Close()
	wg.Wait()

	require.NoError(t, execErr, "agent exec")
	require.Equalf(t, 0, execResult.ExitCode, "agent exited non-zero: %s", execResult.Stderr.String())
	require.NoError(t, streamErr, "event stream")
	return out.String()
}
