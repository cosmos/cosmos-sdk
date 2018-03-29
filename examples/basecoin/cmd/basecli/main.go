package main

import (
	"errors"
	"os"

	"github.com/spf13/cobra"

	"github.com/tendermint/tmlibs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"

	coolcmd "github.com/cosmos/cosmos-sdk/examples/basecoin/x/cool/commands"
	"github.com/cosmos/cosmos-sdk/version"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/commands"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/commands"
	ibccmd "github.com/cosmos/cosmos-sdk/x/ibc/commands"
	simplestakingcmd "github.com/cosmos/cosmos-sdk/x/simplestake/commands"

	"github.com/cosmos/cosmos-sdk/examples/basecoin/app"
	"github.com/cosmos/cosmos-sdk/examples/basecoin/types"
)

// gaiacliCmd is the entry point for this binary
var (
	basecliCmd = &cobra.Command{
		Use:   "basecli",
		Short: "Basecoin light-client",
	}
)

func todoNotImplemented(_ *cobra.Command, _ []string) error {
	return errors.New("TODO: Command not yet implemented")
}

func main() {
	// disable sorting
	cobra.EnableCommandSorting = false

	// get the codec
	cdc := app.MakeCodec()

	// TODO: setup keybase, viper object, etc. to be passed into
	// the below functions and eliminate global vars, like we do
	// with the cdc

	// add standard rpc, and tx commands
	rpc.AddCommands(basecliCmd)
	basecliCmd.AddCommand(client.LineBreak)
	tx.AddCommands(basecliCmd, cdc)
	basecliCmd.AddCommand(client.LineBreak)

	// add query/post commands (custom to binary)
	basecliCmd.AddCommand(
		client.GetCommands(
			authcmd.GetAccountCmd("main", cdc, types.GetAccountDecoder(cdc)),
		)...)
	basecliCmd.AddCommand(
		client.PostCommands(
			bankcmd.SendTxCmd(cdc),
		)...)
	basecliCmd.AddCommand(
		client.PostCommands(
			coolcmd.QuizTxCmd(cdc),
		)...)
	basecliCmd.AddCommand(
		client.PostCommands(
			coolcmd.SetTrendTxCmd(cdc),
		)...)
	basecliCmd.AddCommand(
		client.PostCommands(
			ibccmd.IBCTransferCmd(cdc),
		)...)
	basecliCmd.AddCommand(
		client.PostCommands(
			ibccmd.IBCRelayCmd(cdc),
			simplestakingcmd.BondTxCmd(cdc),
		)...)
	basecliCmd.AddCommand(
		client.PostCommands(
			simplestakingcmd.UnbondTxCmd(cdc),
		)...)

	// add proxy, version and key info
	basecliCmd.AddCommand(
		client.LineBreak,
		lcd.ServeCommand(cdc),
		keys.Commands(),
		client.LineBreak,
		version.VersionCmd,
	)

	// prepare and add flags
	executor := cli.PrepareMainCmd(basecliCmd, "BC", os.ExpandEnv("$HOME/.basecli"))
	executor.Execute()
}
