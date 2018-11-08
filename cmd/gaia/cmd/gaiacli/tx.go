package main

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	distrcmd "github.com/cosmos/cosmos-sdk/x/distribution/client/cli"
	govcmd "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	slashingcmd "github.com/cosmos/cosmos-sdk/x/slashing/client/cli"
	stakecmd "github.com/cosmos/cosmos-sdk/x/stake/client/cli"
	amino "github.com/tendermint/go-amino"
)

func txCmd(cdc *amino.Codec) *cobra.Command {
	//Add transaction generation commands
	txCmd := &cobra.Command{
		Use:   "tx",
		Short: "Transactions subcommands",
	}

	stakeTxCmd := &cobra.Command{
		Use:   "stake",
		Short: "Staking transaction subcommands",
	}

	stakeTxCmd.AddCommand(client.PostCommands(
		stakecmd.GetCmdCreateValidator(cdc),
		stakecmd.GetCmdEditValidator(cdc),
		stakecmd.GetCmdDelegate(cdc),
		stakecmd.GetCmdRedelegate(storeStake, cdc),
		stakecmd.GetCmdUnbond(storeStake, cdc),
	)...)

	distTxCmd := &cobra.Command{
		Use:   "dist",
		Short: "Distribution transactions subcommands",
	}

	distTxCmd.AddCommand(client.PostCommands(
		distrcmd.GetCmdWithdrawRewards(cdc),
		distrcmd.GetCmdSetWithdrawAddr(cdc),
	)...)

	govTxCmd := &cobra.Command{
		Use:   "gov",
		Short: "Governance transactions subcommands",
	}

	govTxCmd.AddCommand(client.PostCommands(
		govcmd.GetCmdDeposit(cdc),
		govcmd.GetCmdVote(cdc),
		govcmd.GetCmdSubmitProposal(cdc),
	)...)

	slashingTxCmd := &cobra.Command{
		Use:   "slashing",
		Short: "Slashing transactions subcommands",
	}

	slashingTxCmd.AddCommand(client.PostCommands(
		slashingcmd.GetCmdUnjail(cdc),
	)...)

	txCmd.AddCommand(
		//Add auth and bank commands
		client.PostCommands(
			bankcmd.SendTxCmd(cdc),
			bankcmd.GetBroadcastCommand(cdc),
			authcmd.GetSignCommand(cdc, authcmd.GetAccountDecoder(cdc)),
		)...)

	txCmd.AddCommand(
		client.LineBreak,
		stakeTxCmd,
		distTxCmd,
		govTxCmd,
		slashingTxCmd,
	)

	return txCmd
}
