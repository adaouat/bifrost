package main

import (
	"context"
	"os"

	"github.com/adaouat/bifrost/internal/cmd"
	"github.com/adaouat/bifrost/internal/tui"
	"github.com/adaouat/forge/cli"
	forgeexit "github.com/adaouat/forge/exitcode"
)

// Version is injected at build time via -ldflags "-X main.Version={{.Tag}}".
var Version = "dev"

func main() {
	err := cli.Run(context.Background(), cmd.NewRootCmd(Version), Version, tui.Accent())
	os.Exit(forgeexit.Resolve(err))
}
