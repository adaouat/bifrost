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

// releaseRollbackToServerFn is the per-server rollback step, replaced in tests
// to verify the loop's sequencing and failure handling without a real SSH connection.
var releaseRollbackToServerFn = releaseRollbackOnServer

// runClientReleaseRollback runs `release rollback` against each resolved
// server over SSH. There is no selection TUI: each server independently rolls
// back to the release immediately preceding its own current one (spec 08).
// Servers are processed sequentially; a failure on any server skips the
// remaining ones and the failure is returned as-is.
func runClientReleaseRollback(cmd *cobra.Command, version string, merged *config.MergedConfig, cfg *config.Config, env, app, agentBinary string) error {
	out := cmd.OutOrStdout()
	modeFlag, _ := cmd.Root().PersistentFlags().GetString("output")
	mode := forgeui.ParseMode(modeFlag)

	if mode != forgeui.JSON {
		tui.PrintActionHeader(mode, out, "Rolling back", env, app)
	}

	for _, srv := range merged.Servers {
		if err := releaseRollbackToServerFn(version, cfg, env, app, agentBinary, srv, mode, out); err != nil {
			return err
		}
	}
	return nil
}

func releaseRollbackOnServer(version string, cfg *config.Config, env, app, agentBinary string, srv config.ResolvedServer, mode forgeui.Mode, out io.Writer) error {
	client, remoteAgent, remoteConfig, cleanup, err := stageAgentSession(version, cfg, env, app, agentBinary, srv)
	if err != nil {
		return err
	}
	defer cleanup()

	agentCmd := fmt.Sprintf("%s release rollback --output json --config %s --env %s --app %s",
		transport.ShellQuote(remoteAgent), transport.ShellQuote(remoteConfig),
		transport.ShellQuote(env), transport.ShellQuote(app))

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

	var ev struct {
		Release string `json:"release"`
	}
	if err := json.Unmarshal(res.Stdout.Bytes(), &ev); err != nil {
		return fmt.Errorf("parsing rollback result from %s: %w", srv.Name, err)
	}

	if mode == forgeui.JSON {
		emit := tui.NewJSONEmitter(out)
		emit.Emit(map[string]any{"event": "rollback", "server": srv.Name, "release": ev.Release, "status": "done"})
		return nil
	}

	tui.PrintServerResult(mode, out, srv.Name, fmt.Sprintf("Rolled back to %s", ev.Release))
	return nil
}
