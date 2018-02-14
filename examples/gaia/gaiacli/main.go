package main

import (
	"errors"
	"os"

	"github.com/spf13/cobra"

	"github.com/tendermint/tmlibs/cli"

	"github.com/cosmos/cosmos-sdk/version"
)

const (
	flagTo     = "to"
	flagAmount = "amount"
	flagFee    = "fee"
)

// gaiacliCmd is the entry point for this binary
var (
	gaiacliCmd = &cobra.Command{
		Use:   "gaiacli",
		Short: "Gaia light-client",
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

	// generic client commands
	AddClientCommands(gaiacliCmd)
	// query commands (custom to binary)
	gaiacliCmd.AddCommand(
		GetCommands(getAccountCmd)...)
	// post tx commands (custom to binary)
	gaiacliCmd.AddCommand(
		PostCommands(postSendCommand())...)

	// add proxy, version and key info
	gaiacliCmd.AddCommand(
		lineBreak,
		serveCommand(),
		KeyCommands(),
		lineBreak,
		version.VersionCmd,
	)

	// prepare and add flags
	executor := cli.PrepareBaseCmd(gaiacliCmd, "GA", os.ExpandEnv("$HOME/.gaiacli"))
	executor.Execute()
}
