package cli

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/app/internal"

	"github.com/cosmos/cosmos-sdk/app/query"

	"github.com/cosmos/cosmos-sdk/container"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"go.uber.org/dig"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/server"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/cosmos/cosmos-sdk/x/crisis"
)

type Options struct {
	AppName          string
	Description      string
	DefaultAppConfig *app.Config
	DefaultHome      string
	EnvPrefix        string
}

func Run(options Options) {
	err := container.Compose(
		func(in inputs) {
			rootCmd := makeRootCmd(in, options)
			if err := svrcmd.Execute(rootCmd, string(in.DefaultHome)); err != nil {
				switch e := err.(type) {
				case server.ErrorCode:
					os.Exit(e.Code)

				default:
					os.Exit(1)
				}
			}
		},

		// Provide default home
		container.Provide(func() client.DefaultHome { return client.DefaultHome(options.DefaultHome) }),
		// Provide codec
		app.CodecProvider,
		query.Module,
		internal.AppConfigProvider(options.DefaultAppConfig),
	)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(-1)
	}
}

type inputs struct {
	dig.In

	DefaultHome       client.DefaultHome
	RootCommands      []*cobra.Command             `group:"root"`
	Codec             codec.JSONCodec              `optional:"true"`
	InterfaceRegistry codectypes.InterfaceRegistry `optional:"true"`
	Amino             *codec.LegacyAmino           `optional:"true"`
	TxConfig          client.TxConfig              `optional:"true"`
	AccountRetriever  client.AccountRetriever      `optional:"true"`
}

func makeRootCmd(in inputs, options Options) *cobra.Command {
	initClientCtx := client.Context{}.
		WithInput(os.Stdin).
		WithHomeDir(options.DefaultHome).
		WithViper(options.EnvPrefix)
	if in.Codec != nil {
		initClientCtx = initClientCtx.WithJSONCodec(in.Codec)
	}
	if in.InterfaceRegistry != nil {
		initClientCtx = initClientCtx.WithInterfaceRegistry(in.InterfaceRegistry)
	}
	if in.TxConfig != nil {
		initClientCtx = initClientCtx.WithTxConfig(in.TxConfig)
	}
	if in.AccountRetriever != nil {
		initClientCtx = initClientCtx.WithAccountRetriever(in.AccountRetriever)
	}
	if in.Amino != nil {
		initClientCtx = initClientCtx.WithLegacyAmino(in.Amino)
	}

	rootCmd := &cobra.Command{
		Use:   options.AppName,
		Short: options.Description,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// set the default command outputs
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			initClientCtx = client.ReadHomeFlag(initClientCtx, cmd)

			initClientCtx, err := config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			return server.InterceptConfigsPreRunHandler(cmd)
		},
	}

	cfg := sdk.GetConfig()
	cfg.Seal()

	rootCmd.AddCommand(in.RootCommands...)

	rootCmd.AddCommand(
		//TODO: AddGenesisAccountCmd(options.DefaultHome),
		tmcli.NewCompletionCmd(rootCmd, true),
		//TODO: testnetCmd(simapp.ModuleBasics, banktypes.GenesisBalancesIterator{}),
		debug.Cmd(),
		config.Cmd(),
	)

	// TODO:
	//server.AddCommands(rootCmd, options.DefaultHome, appProvider.AppCreator, appProvider.AppExportor, addModuleInitFlags)

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		keys.Commands(options.DefaultHome),
	)

	// TODO: rootCmd.AddCommand(server.RosettaCommand(clientCtx.InterfaceRegistry, clientCtx.JSONCodec))

	return rootCmd
}

func addModuleInitFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
}
