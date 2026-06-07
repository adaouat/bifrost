package cmd

import (
	"fmt"
	"io"
	"os"
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
func runClientDeploy(cmd *cobra.Command, version string, merged *config.MergedConfig, cfg *config.Config, env, app, artifact, releaseName, agentBinary string) error {
	out := cmd.OutOrStdout()
	mode := forgeui.ParseMode(outputMode(cmd))

	for _, srv := range merged.Servers {
		if err := deployToServerFn(version, merged, cfg, env, app, artifact, releaseName, agentBinary, srv, mode, out); err != nil {
			return err
		}
	}
	return nil
}

func deployToServer(version string, _ *config.MergedConfig, cfg *config.Config, env, app, artifact, releaseName, agentBinary string, srv config.ResolvedServer, mode forgeui.Mode, out io.Writer) error {
	if mode != forgeui.JSON {
		_, _ = fmt.Fprint(out, tui.DeployHeader(mode, env, app, releaseName))
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

	flatConfigPath, cleanConfig, err := writeTempFlatConfig(cfg, env, app)
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

	chmodRes, err := client.Exec("chmod +x " + remoteAgent)
	if err != nil {
		return fmt.Errorf("chmod agent on %s: %w", srv.Name, err)
	}
	if chmodRes.ExitCode != 0 {
		return fmt.Errorf("chmod agent on %s: exit %d: %s", srv.Name, chmodRes.ExitCode, chmodRes.Stderr.String())
	}

	agentCmd := fmt.Sprintf("%s deploy --output json --config %s --artifact %s --release-name %s",
		remoteAgent, remoteConfig, remoteArtifact, releaseName)

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
		execResult, execErr = client.ExecStream(agentCmd, pw)
	}()

	_, streamErr := tui.ForwardStream(pr, srv.Name, mode, out)
	_ = pr.Close()
	wg.Wait()

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
	return streamErr
}

// writeTempFlatConfig generates a flat config for the given env+app and writes
// it to a temp file. The caller must call the returned cleanup func when done.
func writeTempFlatConfig(cfg *config.Config, envName, appName string) (path string, cleanup func(), err error) {
	f, err := os.CreateTemp("", "bifrost-config-*.yml")
	if err != nil {
		return "", nil, fmt.Errorf("creating temp config: %w", err)
	}
	cleanFn := func() { _ = os.Remove(f.Name()) }

	if err := config.GenerateFlatConfig(cfg, envName, appName, f); err != nil {
		cleanFn()
		_ = f.Close()
		return "", nil, fmt.Errorf("writing flat config: %w", err)
	}
	if err := f.Close(); err != nil {
		cleanFn()
		return "", nil, fmt.Errorf("closing temp config: %w", err)
	}
	return f.Name(), cleanFn, nil
}
