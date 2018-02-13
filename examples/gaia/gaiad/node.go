package main

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/baseapp"
)

const (
	flagWithTendermint = "with-tendermint"
)

var (
	initNodeCmd = &cobra.Command{
		Use:   "init <flags???>",
		Short: "Initialize full node",
		RunE:  todoNotImplemented,
	}

	resetNodeCmd = &cobra.Command{
		Use:   "unsafe_reset_all",
		Short: "Reset full node data (danger, must resync)",
		RunE:  todoNotImplemented,
	}
)

// AddNodeCommands registers all commands to interact
// with a local full-node as subcommands of the argument.
//
// Accept an application it should start
func AddNodeCommands(cmd *cobra.Command, node baseapp.BaseApp) {
	cmd.AddCommand(
		initNodeCmd,
		startNodeCmd(node),
		resetNodeCmd,
	)
}

func startNodeCmd(node baseapp.BaseApp) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Run the full node",
		RunE:  todoNotImplemented,
	}
	cmd.Flags().Bool(flagWithTendermint, true, "run abci app embedded in-process with tendermint")
	return cmd
}
