package client

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
)

// PrepareMainCmd is meant for client side libs that want some more flags
//
// This adds FlagKeyringBackend with the default value.
func PrepareMainCmd(cmd *cobra.Command, envPrefix, defaultHome string) cli.Executor {
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test)")
	viper.BindPFlag(flags.FlagKeyringBackend, cmd.Flags().Lookup(flags.FlagKeyringBackend))

	return cli.PrepareMainCmd(cmd, envPrefix, defaultHome)
}
