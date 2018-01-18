package main

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/app"
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

// NodeCommands registers a sub-tree of commands to interact with
// a local full-node.
//
// Accept an application it should start
func NodeCommands(node app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "Run the full node",
	}
	cmd.AddCommand(
		initNodeCmd,
		startNodeCmd(node),
		resetNodeCmd,
	)
	return cmd
}

func startNodeCmd(node app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Run the full node",
		RunE:  todoNotImplemented,
	}
	cmd.Flags().Bool(flagWithTendermint, true, "run abci app embedded in-process with tendermint")
	return cmd
}
