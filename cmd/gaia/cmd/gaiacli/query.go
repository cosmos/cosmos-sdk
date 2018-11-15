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

	// Query commcmmand structure
	queryCmd.AddCommand(
		rpc.BlockCommand(),
		rpc.ValidatorCommand(),
		tx.SearchTxCmd(cdc),
		tx.QueryTxCmd(cdc),
		client.LineBreak,
		client.GetCommands(authcmd.GetAccountCmd(storeAcc, cdc, authcmd.GetAccountDecoder(cdc)))[0],
		stakecmd.GetQueryCmd(storeStake, cdc),
		govcmd.GetQueryCmd(storeGov, cdc),
		slashingcmd.GetQueryCmd(storeSlashing, cdc),
	)

	return queryCmd
}
