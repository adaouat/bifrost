package cmd

import (
	"fmt"
	"io"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"

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
func runClientDeploy(cmd *cobra.Command, version string, merged *config.MergedConfig, cfg *config.Config, env, app, artifact, releaseName, agentBinary string, init_ bool) error {
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
		if err := deployToServerFn(version, merged, cfg, env, app, artifact, releaseName, agentBinary, init_, srv, mode, out); err != nil {
			return err
		}
	}
	return nil
}

func deployToServer(version string, _ *config.MergedConfig, cfg *config.Config, env, app, artifact, releaseName, agentBinary string, init_ bool, srv config.ResolvedServer, mode forgeui.Mode, out io.Writer) error {
	if mode != forgeui.JSON {
		_, _ = fmt.Fprint(out, tui.ServerHeader(mode, srv.Name, srv.Host))
	}

	session, err := transport.StageAgentSession(version, cfg, env, app, agentBinary, srv)
	if err != nil {
		return Wrap(Runtime, err)
	}
	defer session.Close()

	remoteArtifact, err := session.Staging.Upload(artifact, filepath.Base(artifact))
	if err != nil {
		return fmt.Errorf("uploading artifact to %s: %w", srv.Name, err)
	}

	agentCmd := deployAgentCmd(session.RemoteAgent, session.RemoteConfig, remoteArtifact, releaseName, init_)

	return execAndForward(func(w io.Writer) (*transport.ExecResult, error) {
		return session.Client.ExecStream(agentCmd, w)
	}, srv, mode, out)
}

// deployAgentCmd builds the remote agent's deploy command. init_ forwards
// --init so the agent can create releases_root/shared_root on the target,
// matching the local (non-SSH) deploy path.
func deployAgentCmd(remoteAgent, remoteConfig, remoteArtifact, releaseName string, init_ bool) string {
	cmd := fmt.Sprintf("%s deploy --output json --config %s --artifact %s --release-name %s",
		transport.ShellQuote(remoteAgent), transport.ShellQuote(remoteConfig),
		transport.ShellQuote(remoteArtifact), transport.ShellQuote(releaseName))
	if init_ {
		cmd += " --init"
	}
	return cmd
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
