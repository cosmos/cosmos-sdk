package client

import (
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govCli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
)

// ModuleClient exports all client functionality from the governance module. The
// governance ModuleClient is slightly different from other ModuleClients in that
// it contains a slice of "proposal" child commands. These commands are respective
// to proposal type handlers that are implemented in other modules but are mounted
// under the governance CLI (eg. parameter change proposals).
type ModuleClient struct {
	storeKey string
	cdc      *amino.Codec
	pcmds    []*cobra.Command
}

func NewModuleClient(storeKey string, cdc *amino.Codec, pcmds ...*cobra.Command) ModuleClient {
	return ModuleClient{storeKey, cdc, pcmds}
}

// GetQueryCmd returns the cli query commands for this module
func (mc ModuleClient) GetQueryCmd() *cobra.Command {
	// Group gov queries under a subcommand
	govQueryCmd := &cobra.Command{
		Use:   gov.ModuleName,
		Short: "Querying commands for the governance module",
	}

	govQueryCmd.AddCommand(client.GetCommands(
		govCli.GetCmdQueryProposal(mc.storeKey, mc.cdc),
		govCli.GetCmdQueryProposals(mc.storeKey, mc.cdc),
		govCli.GetCmdQueryVote(mc.storeKey, mc.cdc),
		govCli.GetCmdQueryVotes(mc.storeKey, mc.cdc),
		govCli.GetCmdQueryParam(mc.storeKey, mc.cdc),
		govCli.GetCmdQueryParams(mc.storeKey, mc.cdc),
		govCli.GetCmdQueryProposer(mc.storeKey, mc.cdc),
		govCli.GetCmdQueryDeposit(mc.storeKey, mc.cdc),
		govCli.GetCmdQueryDeposits(mc.storeKey, mc.cdc),
		govCli.GetCmdQueryTally(mc.storeKey, mc.cdc))...)

	return govQueryCmd
}

// GetTxCmd returns the transaction commands for this module
func (mc ModuleClient) GetTxCmd() *cobra.Command {
	govTxCmd := &cobra.Command{
		Use:   gov.ModuleName,
		Short: "Governance transactions subcommands",
	}

	cmdSubmitProp := govCli.GetCmdSubmitProposal(mc.cdc)
	for _, pcmd := range mc.pcmds {
		cmdSubmitProp.AddCommand(client.PostCommands(pcmd)[0])
	}

	govTxCmd.AddCommand(client.PostCommands(
		govCli.GetCmdDeposit(mc.cdc),
		govCli.GetCmdVote(mc.cdc),
		cmdSubmitProp,
	)...)

	return govTxCmd
}
