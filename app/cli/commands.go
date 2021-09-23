package cli

import (
	"os"
	"path/filepath"
	"reflect"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/container"
)

type RootCommand *cobra.Command
type QueryCommand *cobra.Command
type TxCommand *cobra.Command

type inputs struct {
	RootCommands         []RootCommand
	QueryCommands        []QueryCommand
	TxCommands           []TxCommand
	ClientContextOptions []func(client.Context) client.Context
	DefaultHome          DefaultHome
}

var Provider = container.Options(
	container.AutoGroupTypes(reflect.TypeOf((*QueryCommand)(nil)).Elem()),
	container.AutoGroupTypes(reflect.TypeOf((*TxCommand)(nil)).Elem()),
	container.AutoGroupTypes(reflect.TypeOf((*RootCommand)(nil)).Elem()),
	container.AutoGroupTypes(reflect.TypeOf(func(client.Context) client.Context { return client.Context{} })),
	container.Provide(func(in inputs) *cobra.Command {
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}

		defaultHome := filepath.Join(userHomeDir, string(in.DefaultHome))

		ctx := client.Context{}.
			WithInput(os.Stdin).
			WithHomeDir(defaultHome)

		for _, opt := range in.ClientContextOptions {
			ctx = opt(ctx)
		}

		return &cobra.Command{}
	}),
)
