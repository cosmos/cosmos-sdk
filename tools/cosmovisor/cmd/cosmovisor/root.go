package main

import (
	"github.com/spf13/cobra"

	"cosmossdk.io/tools/cosmovisor"
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
		NewShowUpgradeInfoCmd(),
	)

	rootCmd.PersistentFlags().StringP(cosmovisor.FlagCosmovisorConfig, "c", "", "path to cosmovisor config file")
	return rootCmd
}
