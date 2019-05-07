package client

import (
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/cli"
)

// PrepareMainCmd is meant for client side libs that want some more flags
func PrepareMainCmd(cmd *cobra.Command, envPrefix, defaultHome string) Executor {

	// Add --chain-id to persistent flags and mark it required
	rootCmd.PersistentFlags().String(FlagChainID, "", "Chain ID of tendermint node")
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return InitConfig(rootCmd)
	}

	// add tendermint setup
	return cli.PrepareMainCmd(cmd, envPrefix, defaultHome)
}
