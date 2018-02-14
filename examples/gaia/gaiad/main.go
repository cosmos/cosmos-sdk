package main

import (
	"errors"
	"os"

	"github.com/spf13/cobra"

	"github.com/tendermint/tmlibs/cli"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/version"
)

const (
	flagTo     = "to"
	flagAmount = "amount"
	flagFee    = "fee"
)

// gaiadCmd is the entry point for this binary
var (
	gaiadCmd = &cobra.Command{
		Use:   "gaiad",
		Short: "Gaia Daemon (server)",
	}
)

func todoNotImplemented(_ *cobra.Command, _ []string) error {
	return errors.New("TODO: Command not yet implemented")
}

func main() {
	// TODO: set this to something real
	var node baseapp.BaseApp

	AddNodeCommands(gaiadCmd, node)
	gaiadCmd.AddCommand(
		version.VersionCmd,
	)

	// prepare and add flags
	executor := cli.PrepareBaseCmd(gaiadCmd, "GA", os.ExpandEnv("$HOME/.gaiad"))
	executor.Execute()
}
