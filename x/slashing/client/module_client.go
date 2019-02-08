package client

import (
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/slashing/client/cli"
)

// ModuleClient exports all client functionality from this module
type ModuleClient struct {
	storeKey string
	cdc      *amino.Codec
}

func NewModuleClient(storeKey string, cdc *amino.Codec) ModuleClient {
	return ModuleClient{storeKey, cdc}
}

// GetQueryCmd returns the cli query commands for this module
func (mc ModuleClient) GetQueryCmd() *cobra.Command {
	// Group slashing queries under a subcommand
	slashingQueryCmd := &cobra.Command{
		Use:   slashing.ModuleName,
		Short: "Querying commands for the slashing module",
	}

	slashingQueryCmd.AddCommand(
		client.GetCommands(
			cli.GetCmdQuerySigningInfo(mc.storeKey, mc.cdc),
			cli.GetCmdQueryParams(mc.cdc),
		)...,
	)

	return slashingQueryCmd

}

// GetTxCmd returns the transaction commands for this module
func (mc ModuleClient) GetTxCmd() *cobra.Command {
	slashingTxCmd := &cobra.Command{
		Use:   slashing.ModuleName,
		Short: "Slashing transactions subcommands",
	}

	slashingTxCmd.AddCommand(client.PostCommands(
		cli.GetCmdUnjail(mc.cdc),
	)...)

	return slashingTxCmd
}
