package cmd

import "github.com/spf13/cobra"

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "bifrost",
		Short: "Atomic deployment CLI",
	}

	return root
}
