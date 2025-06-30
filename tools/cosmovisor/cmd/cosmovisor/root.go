package main

import (
	"github.com/spf13/cobra"

	"cosmossdk.io/tools/cosmovisor/v2"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "cosmovisor",
		Short: "A process manager for Cosmos SDK application binaries.",
		Long:  GetHelpText(),
	}

	rootCmd.AddCommand(
		NewInitCmd(),
		runCmd,
		configCmd,
		NewVersionCmd(),
		NewAddUpgradeCmd(),
		NewShowManualUpgradesCmd(),
		NewBatchAddUpgradeCmd(),
		NewPrepareUpgradeCmd(),
	)

	rootCmd.PersistentFlags().StringP(cosmovisor.FlagCosmovisorConfig, "c", "", "path to cosmovisor config file")
	return rootCmd
}
