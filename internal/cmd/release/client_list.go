package release

import (
	"encoding/json"
	"fmt"
	"io"

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

	entries, err := fetchReleaseEntries(version, cfg, env, app, agentBinary, srv)
	if err != nil {
		return err
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

// fetchReleaseEntries stages the agent on srv, runs `release list --output json`,
// and returns the parsed release entries.
func fetchReleaseEntries(version string, cfg *config.Config, env, app, agentBinary string, srv config.ResolvedServer) ([]releaseEntry, error) {
	session, err := transport.StageAgentSession(version, cfg, env, app, agentBinary, srv)
	if err != nil {
		return nil, cmderr.Wrap(cmderr.Runtime, err)
	}
	defer session.Close()

	agentCmd := fmt.Sprintf("%s release list --output json --config %s --env %s --app %s",
		transport.ShellQuote(session.RemoteAgent), transport.ShellQuote(session.RemoteConfig),
		transport.ShellQuote(env), transport.ShellQuote(app))

	res, err := session.Client.Exec(agentCmd)
	if err != nil {
		return nil, fmt.Errorf("agent exec on %s: %w", srv.Name, err)
	}
	if res.ExitCode != 0 {
		msg := res.Stderr.String()
		if msg == "" {
			msg = fmt.Sprintf("agent exited with code %d on %s", res.ExitCode, srv.Name)
		}
		return nil, &cmderr.ExitError{Code: res.ExitCode, Message: msg}
	}

	var entries []releaseEntry
	if err := json.Unmarshal(res.Stdout.Bytes(), &entries); err != nil {
		return nil, fmt.Errorf("parsing release list from %s: %w", srv.Name, err)
	}
	return entries, nil
}
