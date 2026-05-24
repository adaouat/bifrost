package release

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/adaouat/bifrost/internal/cmderr"
	"github.com/adaouat/bifrost/internal/config"
)

func newListCmd() *cobra.Command {
	var env, app string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all releases for an application",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if env == "" {
				return fmt.Errorf("--env (or --environment) is required")
			}
			if app == "" {
				return fmt.Errorf("--app (or --application) is required")
			}

			cfg, err := config.Load(releaseConfigPath(cmd))
			if err != nil {
				return err
			}
			merged, err := config.Merge(cfg, env, app)
			if err != nil {
				return err
			}
			if errs := config.Validate(merged); len(errs) > 0 {
				return &cmderr.ExitError{Code: 2, Message: strings.Join(errs, "\n")}
			}

			releases, active, err := listReleases(merged.ReleasesRoot)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			for _, r := range releases {
				if r == active {
					_, _ = fmt.Fprintf(out, "* %s\n", r)
				} else {
					_, _ = fmt.Fprintf(out, "  %s\n", r)
				}
			}
			return nil
		},
	}

	f := cmd.Flags()
	f.StringVar(&env, "environment", "", "target environment key")
	f.StringVar(&env, "env", "", "alias for --environment")
	f.StringVar(&app, "application", "", "target application key")
	f.StringVar(&app, "app", "", "alias for --application")

	return cmd
}

// listReleases returns release directory names sorted newest-first and the
// name of the currently active release (empty if no current symlink exists).
func listReleases(releasesRoot string) (releases []string, active string, err error) {
	entries, err := os.ReadDir(releasesRoot)
	if err != nil {
		return nil, "", fmt.Errorf("reading releases dir: %w", err)
	}

	for _, e := range entries {
		if e.Name() == "current" {
			continue
		}
		if e.IsDir() {
			releases = append(releases, e.Name())
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(releases)))

	currentLink := filepath.Join(releasesRoot, "current")
	if target, err := os.Readlink(currentLink); err == nil {
		active = filepath.Base(target)
	}

	return releases, active, nil
}
