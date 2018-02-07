package main

import (
	"errors"
	"os"

	"github.com/spf13/cobra"

	"github.com/tendermint/tmlibs/cli"

	"github.com/cosmos/cosmos-sdk/baseapp"
)

const (
	flagTo     = "to"
	flagAmount = "amount"
	flagFee    = "fee"
)

// chubCmd is the entry point for this binary
var (
	chubCmd = &cobra.Command{
		Use:   "chub",
		Short: "Cosmos Hub command-line tool",
	}

	lineBreak = &cobra.Command{Run: func(*cobra.Command, []string) {}}

	getAccountCmd = &cobra.Command{
		Use:   "account <address>",
		Short: "Query account balance",
		RunE:  todoNotImplemented,
	}
)

func todoNotImplemented(_ *cobra.Command, _ []string) error {
	return errors.New("TODO: Command not yet implemented")
}

func postSendCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Create and sign a send tx",
		RunE:  todoNotImplemented,
	}
	cmd.Flags().String(flagTo, "", "Address to send coins")
	cmd.Flags().String(flagAmount, "", "Amount of coins to send")
	cmd.Flags().String(flagFee, "", "Fee to pay along with transaction")
	return cmd
}

func main() {
	// disable sorting
	cobra.EnableCommandSorting = false

	// TODO: set this to something real
	var node baseapp.BaseApp

	// add commands
	AddGetCommand(getAccountCmd)
	AddPostCommand(postSendCommand())

	chubCmd.AddCommand(
		NodeCommands(node),
		KeyCommands(),
		ClientCommands(),

		lineBreak,
		VersionCmd,
	)

	// prepare and add flags
	// executor := cli.PrepareMainCmd(chubCmd, "CH", os.ExpandEnv("$HOME/.cosmos-chub"))
	executor := cli.PrepareBaseCmd(chubCmd, "CH", os.ExpandEnv("$HOME/.cosmos-chub"))
	executor.Execute()
}
