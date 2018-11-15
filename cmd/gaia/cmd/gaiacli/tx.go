package main

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	distClient "github.com/cosmos/cosmos-sdk/x/distribution/client"
	govClient "github.com/cosmos/cosmos-sdk/x/gov/client"
	slashingClient "github.com/cosmos/cosmos-sdk/x/slashing/client"
	stakeClient "github.com/cosmos/cosmos-sdk/x/stake/client"
	amino "github.com/tendermint/go-amino"
)

func txCmd(cdc *amino.Codec) *cobra.Command {

	gmc := govClient.NewModuleClient()
	dmc := distClient.NewModuleClient()
	smc := stakeClient.NewModuleClient()
	slmc := slashingClient.NewModuleClient()

	//Add transaction generation commands
	txCmd := &cobra.Command{
		Use:   "tx",
		Short: "Transactions subcommands",
	}

	txCmd.AddCommand(
		bankcmd.SendTxCmd(cdc),
		client.LineBreak,
		authcmd.GetSignCommand(cdc),
		bankcmd.GetBroadcastCommand(cdc),
		client.LineBreak,
		smc.GetTxCmd(storeStake, cdc),
		dmc.GetTxCmd("", cdc),
		gmc.GetTxCmd("", cdc),
		slmc.GetTxCmd("", cdc),
	)

	return txCmd
}
