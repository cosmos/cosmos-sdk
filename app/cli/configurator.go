package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/app"
)

type Configurator interface {
	SetGenesisHandler(handler app.GenesisBasicHandler)
	SetTxCommand(command *cobra.Command)
	SetQueryCommand(command *cobra.Command)

	RootCommand() *cobra.Command
	RootTxCommand() *cobra.Command
	RootQueryCommand() *cobra.Command
}
