package main

import (
	"context"
	"errors"
	"os"

	"charm.land/fang/v2"
	"github.com/adaouat/bifrost/internal/cmd"
	"github.com/adaouat/bifrost/internal/cmderr"
)

func main() {
	err := fang.Execute(context.Background(), cmd.NewRootCmd())
	if err == nil {
		return
	}
	var exitErr *cmderr.ExitError
	if errors.As(err, &exitErr) {
		os.Exit(exitErr.Code)
	}
	os.Exit(1)
}
