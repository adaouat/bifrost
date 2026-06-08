package release

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/adaouat/bifrost/internal/cmderr"
	"github.com/adaouat/bifrost/internal/config"
	"github.com/adaouat/bifrost/internal/transport"
	"github.com/adaouat/bifrost/internal/tui"
	forgeui "github.com/adaouat/forge/ui"
)

// releaseListToServerFn is the per-server release list step, replaced in tests
// to verify the loop's sequencing and failure handling without a real SSH connection.
var releaseListToServerFn = releaseListOnServer

// runClientReleaseList runs `release list` against each resolved server over
// SSH. Servers are queried sequentially; a failure on any server skips the
// remaining ones and the failure is returned as-is, preserving the agent's exit code.
func runClientReleaseList(cmd *cobra.Command, version string, merged *config.MergedConfig, cfg *config.Config, env, app, agentBinary string) error {
	out := cmd.OutOrStdout()
	modeFlag, _ := cmd.Root().PersistentFlags().GetString("output")
	mode := forgeui.ParseMode(modeFlag)

	for _, srv := range merged.Servers {
		if err := releaseListToServerFn(version, cfg, env, app, agentBinary, srv, mode, out); err != nil {
			return err
		}
	}
	return nil
}

func releaseListOnServer(version string, cfg *config.Config, env, app, agentBinary string, srv config.ResolvedServer, mode forgeui.Mode, out io.Writer) error {
	if mode != forgeui.JSON {
		_, _ = fmt.Fprint(out, tui.ServerHeader(mode, srv.Name, srv.Host))
	}

	client, remoteAgent, remoteConfig, cleanup, err := stageAgentSession(version, cfg, env, app, agentBinary, srv)
	if err != nil {
		return err
	}
	defer cleanup()

	agentCmd := fmt.Sprintf("%s release list --output json --config %s --env %s --app %s",
		remoteAgent, remoteConfig, env, app)

	res, err := client.Exec(agentCmd)
	if err != nil {
		return fmt.Errorf("agent exec on %s: %w", srv.Name, err)
	}
	if res.ExitCode != 0 {
		msg := res.Stderr.String()
		if msg == "" {
			msg = fmt.Sprintf("agent exited with code %d on %s", res.ExitCode, srv.Name)
		}
		return &cmderr.ExitError{Code: res.ExitCode, Message: msg}
	}

	var entries []releaseEntry
	if err := json.Unmarshal(res.Stdout.Bytes(), &entries); err != nil {
		return fmt.Errorf("parsing release list from %s: %w", srv.Name, err)
	}

	if mode == forgeui.JSON {
		emit := tui.NewJSONEmitter(out)
		emit.Emit(map[string]any{"event": "list", "server": srv.Name, "releases": entries})
		return nil
	}

	releases := make([]string, len(entries))
	active := ""
	for i, e := range entries {
		releases[i] = e.Name
		if e.Active {
			active = e.Name
		}
	}
	renderReleaseList(out, env, app, releases, active)
	return nil
}

// stageAgentSession connects to srv, resolves and uploads the agent binary and
// a flat config to its staging directory, and makes the agent executable. The
// returned cleanup func closes the connection, removes the staging dir, and
// deletes the temporary local config file.
func stageAgentSession(version string, cfg *config.Config, env, app, agentBinary string, srv config.ResolvedServer) (client *transport.Client, remoteAgent, remoteConfig string, cleanup func(), err error) {
	client, err = transport.Connect(srv)
	if err != nil {
		return nil, "", "", nil, &cmderr.ExitError{Code: cmderr.Runtime, Message: fmt.Sprintf("connecting to %s: %v", srv.Name, err)}
	}

	localAgentPath := agentBinary
	if localAgentPath == "" {
		p, err := transport.DetectPlatform(client)
		if err != nil {
			_ = client.Close()
			return nil, "", "", nil, &cmderr.ExitError{Code: cmderr.Runtime, Message: fmt.Sprintf("detecting platform on %s: %v", srv.Name, err)}
		}
		localAgentPath, err = transport.ResolveAgentBinary(version, p)
		if err != nil {
			_ = client.Close()
			return nil, "", "", nil, &cmderr.ExitError{Code: cmderr.Runtime, Message: fmt.Sprintf("resolving agent binary for %s: %v", srv.Name, err)}
		}
	}

	staging, err := transport.NewStaging(client, srv.StagingDir)
	if err != nil {
		_ = client.Close()
		return nil, "", "", nil, &cmderr.ExitError{Code: cmderr.Runtime, Message: fmt.Sprintf("creating staging dir on %s: %v", srv.Name, err)}
	}

	remoteAgent, err = staging.Upload(localAgentPath, "bifrost-agent")
	if err != nil {
		_ = staging.Cleanup()
		_ = client.Close()
		return nil, "", "", nil, fmt.Errorf("uploading agent to %s: %w", srv.Name, err)
	}

	flatConfigPath, cleanConfig, err := writeTempFlatConfig(cfg, env, app)
	if err != nil {
		_ = staging.Cleanup()
		_ = client.Close()
		return nil, "", "", nil, fmt.Errorf("generating flat config for %s: %w", srv.Name, err)
	}

	remoteConfig, err = staging.Upload(flatConfigPath, "bifrost.yml")
	if err != nil {
		cleanConfig()
		_ = staging.Cleanup()
		_ = client.Close()
		return nil, "", "", nil, fmt.Errorf("uploading config to %s: %w", srv.Name, err)
	}

	chmodRes, err := client.Exec("chmod +x " + remoteAgent)
	if err != nil {
		cleanConfig()
		_ = staging.Cleanup()
		_ = client.Close()
		return nil, "", "", nil, fmt.Errorf("chmod agent on %s: %w", srv.Name, err)
	}
	if chmodRes.ExitCode != 0 {
		cleanConfig()
		_ = staging.Cleanup()
		_ = client.Close()
		return nil, "", "", nil, fmt.Errorf("chmod agent on %s: exit %d: %s", srv.Name, chmodRes.ExitCode, chmodRes.Stderr.String())
	}

	cleanup = func() {
		cleanConfig()
		if cleanErr := staging.Cleanup(); cleanErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "WARNING: cleanup staging on %s: %v\n", srv.Name, cleanErr)
		}
		_ = client.Close()
	}
	return client, remoteAgent, remoteConfig, cleanup, nil
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
