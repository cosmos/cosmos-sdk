package client

import (
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/client"
	distCmds "github.com/cosmos/cosmos-sdk/x/distribution/client/cli"
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
	return &cobra.Command{Hidden: true}
}

// GetTxCmd returns the transaction commands for this module
func (mc ModuleClient) GetTxCmd() *cobra.Command {
	distTxCmd := &cobra.Command{
		Use:   "dist",
		Short: "Distribution transactions subcommands",
	}

	distTxCmd.AddCommand(client.PostCommands(
		distCmds.GetCmdWithdrawRewards(mc.cdc),
		distCmds.GetCmdSetWithdrawAddr(mc.cdc),
	)...)

	return distTxCmd
}
