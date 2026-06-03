package cmd

import (
	"fmt"
	"os"

	clog "charm.land/log/v2"
	"github.com/adaouat/bifrost/internal/cmd/config"
	"github.com/adaouat/bifrost/internal/cmd/release"
	"github.com/adaouat/bifrost/internal/tui"
	"github.com/spf13/cobra"
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

func NewRootCmd() *cobra.Command {
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
			if verbose {
				clog.SetLevel(clog.DebugLevel)
			}
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
	}

	root.SetVersionTemplate(tui.VersionTemplate())

	root.AddCommand(config.NewConfigCmd())
	root.AddCommand(newDeployCmd())
	root.AddCommand(release.NewReleaseCmd())

	f := root.PersistentFlags()
	f.StringVar(&cfgFile, "config", "", "config file (default: .config/bifrost.yml, then .bifrost.yml)")
	f.StringVar(&output, "output", "human", "output mode (human, json, plain)")
	f.BoolVar(&dryRun, "dry-run", false, "simulate actions without applying them")
	f.BoolVar(&verbose, "verbose", false, "enable verbose logging")

	return root
}
