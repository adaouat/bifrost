package release

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/adaouat/bifrost/internal/cmd/cmdutil"
	"github.com/adaouat/bifrost/internal/cmderr"
	"github.com/adaouat/bifrost/internal/config"
)

// releaseEntry is the JSON shape of a single release in `release list --output json`.
type releaseEntry struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

func newListCmd(version string) *cobra.Command {
	var env, app, agentBinary string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all releases for an application",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load(releaseConfigPath(cmd))
			if err != nil {
				return err
			}

			merged, err := cmdutil.ResolveMergedConfig(cfg, env, app)
			if err != nil {
				return err
			}

			if len(merged.Servers) > 0 {
				return runClientReleaseList(cmd, version, merged, cfg, env, app, agentBinary)
			}

			releases, active, err := listReleases(merged.ReleasesRoot)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			mode, _ := cmd.Root().PersistentFlags().GetString("output")

			if mode == "json" {
				entries := make([]releaseEntry, len(releases))
				for i, r := range releases {
					entries[i] = releaseEntry{Name: r, Active: r == active}
				}
				data, _ := json.Marshal(entries)
				_, _ = fmt.Fprintf(out, "%s\n", data)
				return nil
			}

			renderReleaseList(out, env, app, releases, active)
			return nil
		},
	}

	f := cmd.Flags()
	f.StringVar(&env, "environment", "", "target environment key")
	f.StringVar(&env, "env", "", "alias for --environment")
	f.StringVar(&app, "application", "", "target application key")
	f.StringVar(&app, "app", "", "alias for --application")
	f.StringVar(&agentBinary, "agent-binary", "", "path to a prebuilt agent binary (skips download in client mode)")

	return cmd
}

// listReleases returns release directory names sorted newest-first (by
// modification time — --release-name accepts arbitrary values, so a
// lexicographic sort would not reliably reflect creation order) and the name
// of the currently active release (empty if no current symlink exists).
func listReleases(releasesRoot string) (releases []string, active string, err error) {
	entries, err := os.ReadDir(releasesRoot)
	if err != nil {
		return nil, "", cmderr.Wrap(cmderr.Runtime, fmt.Errorf("reading releases dir: %w", err))
	}

	type releaseDir struct {
		name    string
		modTime time.Time
	}
	var dirs []releaseDir
	for _, e := range entries {
		if e.Name() == "current" || !e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			return nil, "", cmderr.Wrap(cmderr.Runtime, fmt.Errorf("stat release %s: %w", e.Name(), err))
		}
		dirs = append(dirs, releaseDir{name: e.Name(), modTime: info.ModTime()})
	}

	sort.Slice(dirs, func(i, j int) bool { return dirs[i].modTime.After(dirs[j].modTime) })

	releases = make([]string, len(dirs))
	for i, d := range dirs {
		releases[i] = d.name
	}

	currentLink := filepath.Join(releasesRoot, "current")
	if target, err := os.Readlink(currentLink); err == nil {
		active = filepath.Base(target)
	}

	return releases, active, nil
}

// renderReleaseList writes the human-mode release list to out.
func renderReleaseList(out io.Writer, env, app string, releases []string, active string) {
	_, _ = fmt.Fprintf(out, "\n  Releases for %s › %s  (%d total)\n\n", env, app, len(releases))
	for _, r := range releases {
		if r == active {
			_, _ = fmt.Fprintf(out, "    %s  ← current\n", r)
		} else {
			_, _ = fmt.Fprintf(out, "    %s\n", r)
		}
	}
	_, _ = fmt.Fprintln(out)
}
