package main

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/app"
)

var (
	initNodeCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize full node",
		RunE:  todoNotImplemented,
	}

	resetNodeCmd = &cobra.Command{
		Use:   "unsafe_reset_all",
		Short: "Reset full node data (danger, must resync)",
		RunE:  todoNotImplemented,
	}
)

func startNodeCmd(node app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Run the full node",
		RunE:  todoNotImplemented,
	}
	return cmd
}

func nodeCommand(node app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "Run the full node",
		RunE:  todoNotImplemented,
	}
	cmd.AddCommand(
		initNodeCmd,
		startNodeCmd(node),
		resetNodeCmd,
	)
	return cmd
}
