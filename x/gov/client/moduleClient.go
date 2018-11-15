package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	govCli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
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
	// Group gov queries under a subcommand
	govQueryCmd := &cobra.Command{
		Use:   "gov",
		Short: "Querying commands for the governance module",
	}

	govQueryCmd.AddCommand(client.GetCommands(
		govCli.GetCmdQueryProposal(storeKey, cdc),
		govCli.GetCmdQueryProposals(storeKey, cdc),
		govCli.GetCmdQueryVote(storeKey, cdc),
		govCli.GetCmdQueryVotes(storeKey, cdc),
		govCli.GetCmdQueryParams(storeKey, cdc),
		govCli.GetCmdQueryDeposit(storeKey, cdc),
		govCli.GetCmdQueryDeposits(storeKey, cdc),
		govCli.GetCmdQueryTally(storeKey, cdc))...)

	return govQueryCmd
}

// GetTxCmd returns the transaction commands for this module
func (mc ModuleClient) GetTxCmd(storeKey string, cdc *amino.Codec) *cobra.Command {
	govTxCmd := &cobra.Command{
		Use:   "gov",
		Short: "Governance transactions subcommands",
	}

	govTxCmd.AddCommand(client.PostCommands(
		govCli.GetCmdDeposit(cdc),
		govCli.GetCmdVote(cdc),
		govCli.GetCmdSubmitProposal(cdc),
	)...)

	return govTxCmd
}
