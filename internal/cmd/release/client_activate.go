package release

import (
	"fmt"
	"io"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"

	"github.com/adaouat/bifrost/internal/cmderr"
	"github.com/adaouat/bifrost/internal/config"
	"github.com/adaouat/bifrost/internal/tui"
	forgeui "github.com/adaouat/forge/ui"
)

// releaseActivateToServerFn is the per-server activation step, replaced in
// tests to verify the loop's sequencing and failure handling without a real SSH connection.
var releaseActivateToServerFn = releaseActivateOnServer

// runClientReleaseActivate runs `release activate` against each resolved
// server over SSH. When releaseName is given, it is used for every server
// (non-interactive path). Otherwise the client queries each server's release
// list and presents a local huh form with one select per server (spec 08).
// Servers are processed sequentially; a failure on any server skips the
// remaining ones and the failure is returned as-is.
func runClientReleaseActivate(cmd *cobra.Command, version string, merged *config.MergedConfig, cfg *config.Config, env, app, releaseName, agentBinary string) error {
	out := cmd.OutOrStdout()
	modeFlag, _ := cmd.Root().PersistentFlags().GetString("output")
	mode := forgeui.ParseMode(modeFlag)

	var selections map[string]string
	if releaseName == "" {
		if !tui.IsTTY() {
			return &cmderr.ExitError{Code: cmderr.Usage, Message: "--release is required in non-interactive mode"}
		}
		var err error
		selections, err = selectReleasesPerServer(version, cfg, env, app, agentBinary, merged.Servers)
		if err != nil {
			return err
		}
	}

	if mode != forgeui.JSON {
		tui.PrintActionHeader(mode, out, "Activating", env, app)
	}

	for _, srv := range merged.Servers {
		name := releaseName
		if name == "" {
			name = selections[srv.Name]
		}
		if err := releaseActivateToServerFn(version, cfg, env, app, agentBinary, name, srv, mode, out); err != nil {
			return err
		}
	}
	return nil
}

func releaseActivateOnServer(version string, cfg *config.Config, env, app, agentBinary, releaseName string, srv config.ResolvedServer, mode forgeui.Mode, out io.Writer) error {
	client, remoteAgent, remoteConfig, cleanup, err := stageAgentSession(version, cfg, env, app, agentBinary, srv)
	if err != nil {
		return err
	}
	defer cleanup()

	agentCmd := fmt.Sprintf("%s release activate --release %s --no-confirm --config %s --env %s --app %s",
		remoteAgent, releaseName, remoteConfig, env, app)

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

	if mode == forgeui.JSON {
		emit := tui.NewJSONEmitter(out)
		emit.Emit(map[string]any{"event": "activate", "server": srv.Name, "release": releaseName, "status": "done"})
		return nil
	}

	tui.PrintServerResult(mode, out, srv.Name, fmt.Sprintf("Activated %s", releaseName))
	return nil
}

// selectReleasesPerServer queries every server's release list sequentially,
// then presents a local huh form with one Select field per server (spec 08).
// Returns the chosen release name keyed by server name.
func selectReleasesPerServer(version string, cfg *config.Config, env, app, agentBinary string, servers []config.ResolvedServer) (map[string]string, error) {
	type serverReleases struct {
		srv      config.ResolvedServer
		releases []string
		active   string
	}

	all := make([]serverReleases, 0, len(servers))
	for _, srv := range servers {
		entries, err := fetchReleaseEntries(version, cfg, env, app, agentBinary, srv)
		if err != nil {
			return nil, err
		}
		if len(entries) == 0 {
			return nil, fmt.Errorf("no releases found on %s", srv.Name)
		}

		releases := make([]string, len(entries))
		active := ""
		for i, e := range entries {
			releases[i] = e.Name
			if e.Active {
				active = e.Name
			}
		}
		all = append(all, serverReleases{srv: srv, releases: releases, active: active})
	}

	selections := make([]string, len(all))
	fields := make([]huh.Field, len(all))
	for i, sr := range all {
		options := make([]huh.Option[string], len(sr.releases))
		for j, r := range sr.releases {
			label := r
			if r == sr.active {
				label += " (current)"
			}
			options[j] = huh.NewOption(label, r)
		}
		fields[i] = huh.NewSelect[string]().
			Title(fmt.Sprintf("%s (%s)", sr.srv.Name, sr.srv.Host)).
			Options(options...).
			Value(&selections[i])
	}

	if err := huh.NewForm(huh.NewGroup(fields...)).
		WithTheme(tui.HuhTheme()).
		Run(); err != nil {
		return nil, fmt.Errorf("release selection: %w", err)
	}

	result := make(map[string]string, len(all))
	for i, sr := range all {
		result[sr.srv.Name] = selections[i]
	}
	return result, nil
}
