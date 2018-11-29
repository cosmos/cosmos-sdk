package main

import (
	"github.com/spf13/cobra"

	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/cosmos/cosmos-sdk/version"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	auth "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	bank "github.com/cosmos/cosmos-sdk/x/bank/client/rest"

	"github.com/cosmos/cosmos-sdk/docs/examples/democoin/app"
	coolcmd "github.com/cosmos/cosmos-sdk/docs/examples/democoin/x/cool/client/cli"
	powcmd "github.com/cosmos/cosmos-sdk/docs/examples/democoin/x/pow/client/cli"
	simplestakingcmd "github.com/cosmos/cosmos-sdk/docs/examples/democoin/x/simplestake/client/cli"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// rootCmd is the entry point for this binary
var (
	rootCmd = &cobra.Command{
		Use:   "democli",
		Short: "Democoin light-client",
	}
	storeAcc = "acc"
)

func main() {
	// disable sorting
	cobra.EnableCommandSorting = false

	// get the codec
	cdc := app.MakeCodec()

	// Setup certain SDK config
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("demoacc", "demopub")
	config.SetBech32PrefixForValidator("demoval", "demovalpub")
	config.SetBech32PrefixForConsensusNode("democons", "democonspub")
	config.Seal()

	// TODO: setup keybase, viper object, etc. to be passed into
	// the below functions and eliminate global vars, like we do
	// with the cdc

	// add standard rpc, and tx commands

	rootCmd.AddCommand(
		rpc.InitClientCommand(),
		rpc.StatusCommand(),
		client.LineBreak,
		tx.SearchTxCmd(cdc),
		tx.QueryTxCmd(cdc),
		client.LineBreak,
	)

	// add query/post commands (custom to binary)
	// start with commands common to basecoin
	rootCmd.AddCommand(
		authcmd.GetAccountCmd(storeAcc, cdc),
	)
	rootCmd.AddCommand(
		bankcmd.SendTxCmd(cdc),
	)
	rootCmd.AddCommand(
		client.PostCommands(
			simplestakingcmd.BondTxCmd(cdc),
		)...)
	rootCmd.AddCommand(
		client.PostCommands(
			simplestakingcmd.UnbondTxCmd(cdc),
		)...)
	// and now democoin specific commands
	rootCmd.AddCommand(
		client.PostCommands(
			coolcmd.QuizTxCmd(cdc),
			coolcmd.SetTrendTxCmd(cdc),
			powcmd.MineCmd(cdc),
		)...)

	// add proxy, version and key info
	rootCmd.AddCommand(
		client.LineBreak,
		lcd.ServeCommand(cdc, registerRoutes),
		keys.Commands(),
		client.LineBreak,
		version.VersionCmd,
	)

	// prepare and add flags
	executor := cli.PrepareMainCmd(rootCmd, "BC", app.DefaultCLIHome)
	err := executor.Execute()
	if err != nil {
		// handle with #870
		panic(err)
	}
}

func registerRoutes(rs *lcd.RestServer) {
	keys.RegisterRoutes(rs.Mux, rs.CliCtx.Indent)
	rpc.RegisterRoutes(rs.CliCtx, rs.Mux)
	tx.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc)
	auth.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, storeAcc)
	bank.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)
}
