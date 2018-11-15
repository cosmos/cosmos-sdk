package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/slashing/client/cli"
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"
)

// ModuleClient exports all client functionality from this module
type ModuleClient struct{}

func NewModuleClient() ModuleClient {
	return ModuleClient{}
}

// GetQueryCmd returns the cli query commands for this module
func (mc ModuleClient) GetQueryCmd(storeKey string, cdc *amino.Codec) *cobra.Command {
	// Group slashing queries under a subcommand
	slashingQueryCmd := &cobra.Command{
		Use:   "slashing",
		Short: "Querying commands for the slashing module",
	}

	slashingQueryCmd.AddCommand(client.GetCommands(
		cli.GetCmdQuerySigningInfo(storeKey, cdc))...)

	return slashingQueryCmd

}

// GetTxCmd returns the transaction commands for this module
func (mc ModuleClient) GetTxCmd(storeKey string, cdc *amino.Codec) *cobra.Command {
	slashingTxCmd := &cobra.Command{
		Use:   "slashing",
		Short: "Slashing transactions subcommands",
	}

	slashingTxCmd.AddCommand(client.PostCommands(
		cli.GetCmdUnjail(cdc),
	)...)

	return slashingTxCmd
}
