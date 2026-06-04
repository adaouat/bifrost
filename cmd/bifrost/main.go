package main

import (
	"context"
	"os"

	"charm.land/fang/v2"

	"github.com/adaouat/bifrost/internal/cmd"
	forgeexit "github.com/adaouat/forge/exitcode"
)

// Version is injected at build time via -ldflags "-X main.Version={{.Tag}}".
var Version = "dev"

func main() {
	err := fang.Execute(context.Background(), cmd.NewRootCmd(Version),
		fang.WithVersion(Version), fang.WithColorSchemeFunc(colorScheme))
	os.Exit(forgeexit.Resolve(err))
}
