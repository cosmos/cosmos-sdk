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

	txCmd.AddCommand(
		//Add auth and bank commands
		client.PostCommands(
			bankcmd.SendTxCmd(cdc),
			bankcmd.GetBroadcastCommand(cdc),
			authcmd.GetSignCommand(cdc, authcmd.GetAccountDecoder(cdc)),
		)...)

	txCmd.AddCommand(
		client.LineBreak,
		stakecmd.GetTxCmd(storeStake, cdc),
		distrcmd.GetTxCmd("", cdc),
		govcmd.GetTxCmd("", cdc),
		slashingcmd.GetTxCmd("", cdc),
	)

	return txCmd
}
