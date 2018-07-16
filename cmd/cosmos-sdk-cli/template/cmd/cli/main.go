package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/cosmos/cosmos-sdk/version"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"

	"github.com/cosmos/cosmos-sdk/cmd/cosmos-sdk-cli/template/app"
	"github.com/cosmos/cosmos-sdk/cmd/cosmos-sdk-cli/template/types"
)

// rootCmd is the entry point for this binary
var (
	rootCmd = &cobra.Command{
		Use:   "myawesomeprojectcli",
		Short: "MyAwesomeProject light-client",
	}
)

func main() {
	// disable sorting
	cobra.EnableCommandSorting = false

	// get the codec
	cdc := app.MakeCodec()

	// TODO: setup keybase, viper object, etc. to be passed into
	// the below functions and eliminate global vars, like we do
	// with the cdc

	// add standard rpc, and tx commands
	rpc.AddCommands(rootCmd)
	rootCmd.AddCommand(client.LineBreak)
	tx.AddCommands(rootCmd, cdc)
	rootCmd.AddCommand(client.LineBreak)

	// add query/post commands (custom to binary)
	// start with commands common to basecoin
	rootCmd.AddCommand(
		client.GetCommands(
			authcmd.GetAccountCmd("acc", cdc, types.GetAccountDecoder(cdc)),
		)...)
	rootCmd.AddCommand(
		client.PostCommands(
			bankcmd.SendTxCmd(cdc),
		)...)
	// and now myawesomeproject specific commands here

	// add proxy, version and key info
	rootCmd.AddCommand(
		client.LineBreak,
		lcd.ServeCommand(cdc),
		keys.Commands(),
		client.LineBreak,
		version.VersionCmd,
	)

	// prepare and add flags
	executor := cli.PrepareMainCmd(rootCmd, "BC", os.ExpandEnv("$HOME/.myawesomeprojectcli"))
	err := executor.Execute()
	if err != nil {
		// handle with #870
		panic(err)
	}
}
