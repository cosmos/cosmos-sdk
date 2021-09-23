package cli

import (
	"reflect"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/container"
)

type RootCommand *cobra.Command
type QueryCommand *cobra.Command
type TxCommand *cobra.Command

type inputs struct {
	RootCommands  []RootCommand
	QueryCommands []QueryCommand
	TxCommands    []TxCommand
}

var Provider = container.Options(
	container.AutoGroupTypes(reflect.TypeOf((*QueryCommand)(nil))),
	container.AutoGroupTypes(reflect.TypeOf((*TxCommand)(nil))),
)
