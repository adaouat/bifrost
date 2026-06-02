package main

import (
	"context"
	"os"

	"charm.land/fang/v2"

	"github.com/adaouat/bifrost/internal/cmd"
	forgeexit "github.com/adaouat/forge/exitcode"
)

func main() {
	err := fang.Execute(context.Background(), cmd.NewRootCmd())
	os.Exit(forgeexit.Resolve(err))
}
