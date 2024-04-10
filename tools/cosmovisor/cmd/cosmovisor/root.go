package main

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "cosmovisor",
		Short: "A process manager for Cosmos SDK application binaries.",
		Long:  GetHelpText(),
	}

	rootCmd.AddCommand(
		NewIntCmd(),
		runCmd,
		configCmd,
		NewVersionCmd(),
		NewAddUpgradeCmd(),
	)

	return rootCmd
}

func mustNoError(err error) {
	if err != nil {
		panic(err)
	}
}
