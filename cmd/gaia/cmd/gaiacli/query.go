package main

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	govcmd "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	slashingcmd "github.com/cosmos/cosmos-sdk/x/slashing/client/cli"
	stakecmd "github.com/cosmos/cosmos-sdk/x/stake/client/cli"
	amino "github.com/tendermint/go-amino"
)

func queryCmd(cdc *amino.Codec) *cobra.Command {
	//Add query commands
	queryCmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "Querying subcommands",
	}

	// Group staking queries under a subcommand
	stakeQueryCmd := &cobra.Command{
		Use:   "stake",
		Short: "Querying commands for the staking module",
	}

	stakeQueryCmd.AddCommand(client.GetCommands(
		stakecmd.GetCmdQueryDelegation(storeStake, cdc),
		stakecmd.GetCmdQueryDelegations(storeStake, cdc),
		stakecmd.GetCmdQueryUnbondingDelegation(storeStake, cdc),
		stakecmd.GetCmdQueryUnbondingDelegations(storeStake, cdc),
		stakecmd.GetCmdQueryRedelegation(storeStake, cdc),
		stakecmd.GetCmdQueryRedelegations(storeStake, cdc),
		stakecmd.GetCmdQueryValidator(storeStake, cdc),
		stakecmd.GetCmdQueryValidators(storeStake, cdc),
		stakecmd.GetCmdQueryValidatorDelegations(storeStake, cdc),
		stakecmd.GetCmdQueryValidatorUnbondingDelegations(queryRouteStake, cdc),
		stakecmd.GetCmdQueryValidatorRedelegations(queryRouteStake, cdc),
		stakecmd.GetCmdQueryParams(storeStake, cdc),
		stakecmd.GetCmdQueryPool(storeStake, cdc))...)

	// Group gov queries under a subcommand
	govQueryCmd := &cobra.Command{
		Use:   "gov",
		Short: "Querying commands for the governance module",
	}

	govQueryCmd.AddCommand(client.GetCommands(
		govcmd.GetCmdQueryProposal(storeGov, cdc),
		govcmd.GetCmdQueryProposals(storeGov, cdc),
		govcmd.GetCmdQueryVote(storeGov, cdc),
		govcmd.GetCmdQueryVotes(storeGov, cdc),
		govcmd.GetCmdQueryDeposit(storeGov, cdc),
		govcmd.GetCmdQueryDeposits(storeGov, cdc))...)

	// Group slashing queries under a subcommand
	slashingQueryCmd := &cobra.Command{
		Use:   "slashing",
		Short: "Querying commands for the slashing module",
	}

	slashingQueryCmd.AddCommand(client.GetCommands(
		slashingcmd.GetCmdQuerySigningInfo(storeSlashing, cdc))...)

	// Query commcmmand structure
	queryCmd.AddCommand(
		rpc.BlockCommand(),
		rpc.ValidatorCommand(),
		tx.SearchTxCmd(cdc),
		tx.QueryTxCmd(cdc),
		client.LineBreak,
		client.GetCommands(authcmd.GetAccountCmd(storeAcc, cdc, authcmd.GetAccountDecoder(cdc)))[0],
		stakeQueryCmd,
		govQueryCmd,
		slashingQueryCmd,
	)

	return queryCmd
}
