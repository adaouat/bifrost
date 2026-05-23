package main

import (
	"context"
	"errors"
	"os"

	"charm.land/fang/v2"
	"github.com/adaouat/bifrost/internal/cmd"
)

func main() {
	err := fang.Execute(context.Background(), cmd.NewRootCmd())
	if err == nil {
		return
	}
	var exitErr *cmd.ExitError
	if errors.As(err, &exitErr) {
		os.Exit(exitErr.Code)
	}
	os.Exit(1)
}
