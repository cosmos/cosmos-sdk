package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	distCmds "github.com/cosmos/cosmos-sdk/x/distribution/client/cli"
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
	// Return a hidden command to staisfy the interface, but not polute the cli
	return &cobra.Command{Hidden: true}
}

// GetTxCmd returns the transaction commands for this module
func (mc ModuleClient) GetTxCmd(storeKey string, cdc *amino.Codec) *cobra.Command {
	distTxCmd := &cobra.Command{
		Use:   "dist",
		Short: "Distribution transactions subcommands",
	}

	distTxCmd.AddCommand(client.PostCommands(
		distCmds.GetCmdWithdrawRewards(cdc),
		distCmds.GetCmdSetWithdrawAddr(cdc),
	)...)

	return distTxCmd
}
