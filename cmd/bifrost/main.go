package main

import (
	"context"
	"os"

	"charm.land/fang/v2"
	"github.com/adaouat/bifrost/internal/cmd"
)

func main() {
	if err := fang.Execute(context.Background(), cmd.NewRootCmd()); err != nil {
		os.Exit(1)
	}
}
