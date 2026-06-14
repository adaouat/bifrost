package release

import (
	"fmt"
	"os"
	"path/filepath"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"

	forgeexec "github.com/adaouat/forge/exec"

	"github.com/adaouat/bifrost/internal/cmd/cmdutil"
	"github.com/adaouat/bifrost/internal/cmderr"
	"github.com/adaouat/bifrost/internal/config"
	"github.com/adaouat/bifrost/internal/hooks"
	"github.com/adaouat/bifrost/internal/strategy/atomic"
	"github.com/adaouat/bifrost/internal/tui"
)

func newActivateCmd(version string) *cobra.Command {
	var env, app, releaseName, agentBinary string
	var noConfirm bool

	cmd := &cobra.Command{
		Use:   "activate",
		Short: "Activate a previously deployed release",
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
				return runClientReleaseActivate(cmd, version, merged, cfg, env, app, releaseName, agentBinary)
			}

			if releaseName == "" {
				if !tui.IsTTY() {
					return fmt.Errorf("--release is required in non-interactive mode")
				}
				selected, err := selectRelease(merged.ReleasesRoot)
				if err != nil {
					return err
				}
				releaseName = selected
			}

			dryRun, _ := cmd.Root().PersistentFlags().GetBool("dry-run")
			if dryRun {
				out := cmd.OutOrStdout()
				currentLink := filepath.Join(merged.ReleasesRoot, "current")
				_, _ = fmt.Fprintln(out, "DRY RUN — no changes will be made")
				_, _ = fmt.Fprintln(out)
				_, _ = fmt.Fprintf(out, "  Would update   %s  →  %s\n", currentLink, releaseName)
				return nil
			}

			releaseDir := filepath.Join(merged.ReleasesRoot, releaseName)
			if _, err := os.Stat(releaseDir); err != nil {
				return &cmderr.ExitError{Code: cmderr.Runtime, Message: fmt.Sprintf("release not found: %s", releaseName)}
			}

			currentLink := filepath.Join(merged.ReleasesRoot, "current")
			if target, err := os.Readlink(currentLink); err == nil && target == releaseDir {
				if noConfirm || !tui.IsTTY() {
					return &cmderr.ExitError{Code: cmderr.Runtime, Message: fmt.Sprintf("release %q is already current", releaseName)}
				}
				var ok bool
				if err := huh.NewConfirm().
					Title(fmt.Sprintf("Release %q is already current. Re-activate?", releaseName)).
					Value(&ok).
					WithTheme(tui.HuhTheme()).
					Run(); err != nil || !ok {
					return fmt.Errorf("re-activation cancelled")
				}
			}

			if err := atomic.LinkShared(merged.SharedDirs, merged.SharedFiles, releaseDir, merged.SharedRoot); err != nil {
				return fmt.Errorf("linking shared resources: %w", err)
			}

			hookData := hooks.HookData{
				Settings:  merged.Settings,
				Variables: merged.Variables,
				Directories: hooks.Directories{
					Working:  releaseDir,
					Current:  currentLink,
					Releases: merged.ReleasesRoot,
					Shared:   merged.SharedRoot,
				},
				Env: hooks.OSEnv(),
			}
			confirmFn := tui.InteractiveHookConfirm()
			hookRunner := forgeexec.New(false, false)

			if err := hooks.RunInteractive(hookRunner, merged.Hooks.PreActivate, hookData, releaseDir, cmd.OutOrStdout(), confirmFn); err != nil {
				return fmt.Errorf("pre_activate hooks: %w", err)
			}

			if err := atomic.SetCurrent(merged.ReleasesRoot, releaseDir); err != nil {
				return fmt.Errorf("updating current symlink: %w", err)
			}

			if err := hooks.RunInteractive(hookRunner, merged.Hooks.PostActivate, hookData, releaseDir, cmd.OutOrStdout(), confirmFn); err != nil {
				return fmt.Errorf("post_activate hooks: %w", err)
			}

			return nil
		},
	}

	f := cmd.Flags()
	f.StringVar(&env, "environment", "", "target environment key")
	f.StringVar(&env, "env", "", "alias for --environment")
	f.StringVar(&app, "application", "", "target application key")
	f.StringVar(&app, "app", "", "alias for --application")
	f.StringVar(&releaseName, "release", "", "release name to activate (interactive selector if omitted)")
	f.StringVar(&agentBinary, "agent-binary", "", "path to a prebuilt agent binary (skips download in client mode)")
	f.BoolVar(&noConfirm, "no-confirm", false, "exit 3 if release is already current instead of prompting")

	return cmd
}

func releaseConfigPath(cmd *cobra.Command) string {
	explicit, _ := cmd.Root().PersistentFlags().GetString("config")
	return cmdutil.ResolvePath(explicit)
}

// selectRelease shows a huh.Select prompt listing available releases. The current release
// is annotated with "(current)" in the label. Returns the selected release name.
func selectRelease(releasesRoot string) (string, error) {
	releases, active, err := listReleases(releasesRoot)
	if err != nil {
		return "", err
	}
	if len(releases) == 0 {
		return "", fmt.Errorf("no releases found in %s", releasesRoot)
	}

	options := make([]huh.Option[string], len(releases))
	for i, r := range releases {
		label := r
		if r == active {
			label += " (current)"
		}
		options[i] = huh.NewOption(label, r)
	}

	var selected string
	if err := huh.NewSelect[string]().
		Title("Which release would you like to activate?").
		Options(options...).
		Value(&selected).
		WithTheme(tui.HuhTheme()).
		Run(); err != nil {
		return "", fmt.Errorf("release selection: %w", err)
	}
	return selected, nil
}
