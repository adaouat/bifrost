package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/adaouat/bifrost/internal/cmd/config"
	"github.com/adaouat/bifrost/internal/cmd/release"
	"github.com/adaouat/bifrost/internal/tui"
	"github.com/adaouat/forge/updatecheck"
)

// ValidateOutputMode returns an error if mode is not one of: human, json, plain.
func ValidateOutputMode(mode string) error {
	switch mode {
	case "human", "json", "plain":
		return nil
	default:
		return fmt.Errorf("invalid output mode %q: must be human, json, or plain", mode)
	}
}

// NewRootCmd constructs the root bifrost command. version is the build version,
// used by --version and the update-check hint.
func NewRootCmd(version string) *cobra.Command {
	var (
		cfgFile string
		output  string
		dryRun  bool
		verbose bool
	)

	root := &cobra.Command{
		Use:   "bifrost",
		Short: "Atomic deployment CLI",
		Long:  tui.HelpLong(),
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := ValidateOutputMode(output); err != nil {
				return err
			}
			// plain and json modes have no colored output; set NO_COLOR so charm
			// libraries (fang, huh, bubbles) also suppress ANSI color sequences.
			if output == "plain" || output == "json" {
				_ = os.Setenv("NO_COLOR", "1")
			}
			return nil
		},
		// After each command, print a one-line update hint if a newer release
		// exists (cached 24h, errors swallowed). Skipped for dev builds, the
		// opt-out env var, and non-human output modes.
		PersistentPostRunE: func(c *cobra.Command, _ []string) error {
			if version == "dev" || output != "human" || os.Getenv("BIFROST_CHECK_UPDATE") == "false" {
				return nil
			}
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()
			cacheFile := ""
			if dir, err := os.UserCacheDir(); err == nil {
				cacheFile = filepath.Join(dir, "bifrost", "update-check.json")
			}
			updatecheck.Hinter{
				Repo:      "adaouat/bifrost",
				Bin:       "bifrost",
				Module:    "github.com/adaouat/bifrost/cmd/bifrost",
				Current:   version,
				CacheFile: cacheFile,
			}.Print(ctx, c.ErrOrStderr())
			return nil
		},
	}

	root.SetVersionTemplate(tui.VersionTemplate())

	root.AddCommand(config.NewConfigCmd())
	root.AddCommand(newDeployCmd(version))
	root.AddCommand(release.NewReleaseCmd(version))

	f := root.PersistentFlags()
	f.StringVar(&cfgFile, "config", "", "config file (default: .config/bifrost.yml, then .bifrost.yml)")
	f.StringVar(&output, "output", "human", "output mode (human, json, plain)")
	f.BoolVar(&dryRun, "dry-run", false, "simulate actions without applying them")
	f.BoolVar(&verbose, "verbose", false, "enable verbose logging")

	return root
}
