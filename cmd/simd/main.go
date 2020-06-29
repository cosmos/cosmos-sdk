package main

import (
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func main() {

	cobra.EnableCommandSorting = true

	// Read in the configuration file for the sdk
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(sdk.Bech32PrefixAccAddr, sdk.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(sdk.Bech32PrefixValAddr, sdk.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
	config.Seal()

	rootComd := &cobra.Command{
		Use:   "simd",
		Short: "simulation app",
	}

	// Construct Root Command
	addClientCommands(rootComd)
	addDaemonCommands(rootComd)

	executor := cli.PrepareMainCmd(rootComd, "GA", simapp.DefaultNodeHome)
	err := executor.Execute()
	if err != nil {
		panic(err)
	}

}
