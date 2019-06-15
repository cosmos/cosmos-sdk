package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	agentcmd "github.com/cosmos/cosmos-sdk/x/group/client/cli"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"
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
	agentQueryCmd := &cobra.Command{
		Use:   "group",
		Short: "Querying commands for the group module",
	}

	agentQueryCmd.AddCommand(client.GetCommands(
		agentcmd.GetCmdGetGroup(mc.storeKey, mc.cdc),
	)...)

	return agentQueryCmd
}

// GetTxCmd returns the transaction commands for this module
func (mc ModuleClient) GetTxCmd() *cobra.Command {
	agentTxCmd := &cobra.Command{
		Use:   "group",
		Short: "Agent transactions subcommands",
	}

	agentTxCmd.AddCommand(client.PostCommands(
		agentcmd.GetCmdCreateGroup(mc.cdc),
	)...)

	return agentTxCmd
}
