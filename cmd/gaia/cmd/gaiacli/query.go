package main

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	govClient "github.com/cosmos/cosmos-sdk/x/gov/client"
	slashingClient "github.com/cosmos/cosmos-sdk/x/slashing/client"
	stakeClient "github.com/cosmos/cosmos-sdk/x/stake/client"
	amino "github.com/tendermint/go-amino"
)

func queryCmd(cdc *amino.Codec) *cobra.Command {

	gmc := govClient.NewModuleClient()
	smc := stakeClient.NewModuleClient()
	slmc := slashingClient.NewModuleClient()

	//Add query commands
	queryCmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "Querying subcommands",
	}

	// Query commcmmand structure
	queryCmd.AddCommand(
		rpc.ValidatorCommand(),
		rpc.BlockCommand(),
		tx.SearchTxCmd(cdc),
		tx.QueryTxCmd(cdc),
		client.LineBreak,
		authcmd.GetAccountCmd(storeAcc, cdc),
		smc.GetQueryCmd(storeStake, cdc),
		gmc.GetQueryCmd(storeGov, cdc),
		slmc.GetQueryCmd(storeSlashing, cdc),
	)

	return queryCmd
}
