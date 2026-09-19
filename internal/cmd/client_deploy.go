package cmd

import (
	"fmt"
	"io"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"

	"github.com/adaouat/bifrost/internal/cmd/cmdutil"
	"github.com/adaouat/bifrost/internal/config"
	"github.com/adaouat/bifrost/internal/transport"
	"github.com/adaouat/bifrost/internal/tui"
	forgeui "github.com/adaouat/forge/ui"
)

// deployToServerFn is the per-server deploy step, replaced in tests to verify
// the loop's sequencing and failure handling without a real SSH connection.
var deployToServerFn = deployToServer

// runClientDeploy executes a deploy in client mode: for each resolved server,
// it uploads the agent binary, flat config, and artifact via SFTP, execs the
// agent with --output json, streams the event output back, then cleans up.
// Servers are processed sequentially; a failure on any server skips the
// remaining ones and the failure is returned as-is, preserving the agent's exit code.
func runClientDeploy(cmd *cobra.Command, version string, merged *config.MergedConfig, cfg *config.Config, env, app, artifact, releaseName, agentBinary string) error {
	out := cmd.OutOrStdout()
	mode := forgeui.ParseMode(outputMode(cmd))

	if dryRun, _ := cmd.Root().PersistentFlags().GetBool("dry-run"); dryRun {
		_, _ = fmt.Fprintln(out, "DRY RUN — no changes will be made")
		_, _ = fmt.Fprintln(out)
		for _, srv := range merged.Servers {
			_, _ = fmt.Fprintf(out, "  Would deploy   %s  →  %s (%s)  as release %s\n", artifact, srv.Name, srv.Host, releaseName)
		}
		return nil
	}

	if mode != forgeui.JSON {
		_, _ = fmt.Fprint(out, tui.DeployHeader(mode, env, app, releaseName))
	}

	for _, srv := range merged.Servers {
		if err := deployToServerFn(version, merged, cfg, env, app, artifact, releaseName, agentBinary, srv, mode, out); err != nil {
			return err
		}
	}
	return nil
}

func deployToServer(version string, _ *config.MergedConfig, cfg *config.Config, env, app, artifact, releaseName, agentBinary string, srv config.ResolvedServer, mode forgeui.Mode, out io.Writer) error {
	if mode != forgeui.JSON {
		_, _ = fmt.Fprint(out, tui.ServerHeader(mode, srv.Name, srv.Host))
	}

	client, err := transport.Connect(srv)
	if err != nil {
		return &ExitError{Code: Runtime, Message: fmt.Sprintf("connecting to %s: %v", srv.Name, err)}
	}
	defer func() { _ = client.Close() }()

	localAgentPath := agentBinary
	if localAgentPath == "" {
		p, err := transport.DetectPlatform(client)
		if err != nil {
			return &ExitError{Code: Runtime, Message: fmt.Sprintf("detecting platform on %s: %v", srv.Name, err)}
		}
		localAgentPath, err = transport.ResolveAgentBinary(version, p)
		if err != nil {
			return &ExitError{Code: Runtime, Message: fmt.Sprintf("resolving agent binary for %s: %v", srv.Name, err)}
		}
	}

	staging, err := transport.NewStaging(client, srv.StagingDir)
	if err != nil {
		return &ExitError{Code: Runtime, Message: fmt.Sprintf("creating staging dir on %s: %v", srv.Name, err)}
	}
	defer func() {
		if cleanErr := staging.Cleanup(); cleanErr != nil {
			_, _ = fmt.Fprintf(out, "WARNING: cleanup staging on %s: %v\n", srv.Name, cleanErr)
		}
	}()

	remoteAgent, err := staging.Upload(localAgentPath, "bifrost-agent")
	if err != nil {
		return fmt.Errorf("uploading agent to %s: %w", srv.Name, err)
	}

	flatConfigPath, cleanConfig, err := cmdutil.WriteTempFlatConfig(cfg, env, app)
	if err != nil {
		return fmt.Errorf("generating flat config for %s: %w", srv.Name, err)
	}
	defer cleanConfig()

	remoteConfig, err := staging.Upload(flatConfigPath, "bifrost.yml")
	if err != nil {
		return fmt.Errorf("uploading config to %s: %w", srv.Name, err)
	}

	remoteArtifact, err := staging.Upload(artifact, filepath.Base(artifact))
	if err != nil {
		return fmt.Errorf("uploading artifact to %s: %w", srv.Name, err)
	}

	if err := staging.Chmod(remoteAgent, 0o755); err != nil {
		return fmt.Errorf("chmod agent on %s: %w", srv.Name, err)
	}

	agentCmd := fmt.Sprintf("%s deploy --output json --config %s --artifact %s --release-name %s",
		transport.ShellQuote(remoteAgent), transport.ShellQuote(remoteConfig),
		transport.ShellQuote(remoteArtifact), transport.ShellQuote(releaseName))

	return execAndForward(func(w io.Writer) (*transport.ExecResult, error) {
		return client.ExecStream(agentCmd, w)
	}, srv, mode, out)
}

// execAndForward runs execStreamFn on its own goroutine, streaming its output
// through a pipe into tui.ForwardStream, and reconciles the two independently
// completing results. It is split out from deployToServer so this
// orchestration — including the failure-mode edge cases below — is
// unit-testable without a real SSH connection.
func execAndForward(execStreamFn func(io.Writer) (*transport.ExecResult, error), srv config.ResolvedServer, mode forgeui.Mode, out io.Writer) error {
	pr, pw := io.Pipe()
	var execResult *transport.ExecResult
	var execErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer func() {
			_ = pw.Close()
			wg.Done()
		}()
		execResult, execErr = execStreamFn(pw)
	}()

	_, streamErr := tui.ForwardStream(pr, srv.Name, mode, out)
	if streamErr != nil {
		// ForwardStream stopped reading before EOF (e.g. a mid-stream agent
		// error). Keep draining pr instead of closing it: closing would make
		// the writer goroutine's next write fail, which stops it reading from
		// the underlying SSH channel — the channel then stops acknowledging
		// received data, and the remote agent can block indefinitely trying
		// to flush its own stdout, so ExecStream would never return.
		go func() { _, _ = io.Copy(io.Discard, pr) }()
	}
	wg.Wait()
	_ = pr.Close()

	if streamErr != nil {
		return streamErr
	}
	if execErr != nil {
		return fmt.Errorf("agent exec on %s: %w", srv.Name, execErr)
	}
	if execResult.ExitCode != 0 {
		msg := execResult.Stderr.String()
		if msg == "" {
			msg = fmt.Sprintf("agent exited with code %d on %s", execResult.ExitCode, srv.Name)
		}
		return &ExitError{Code: execResult.ExitCode, Message: msg}
	}
	return nil
}
