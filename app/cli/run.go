package cli

import (
	"os"

	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"go.uber.org/dig"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/cosmos/cosmos-sdk/app/internal"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/server"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/spf13/cobra"
)

type Options struct {
	AppName          string
	Description      string
	DefaultAppConfig *app.Config
	DefaultHome      string
	EnvPrefix        string
}

func Run(options Options) {
	rootCmd := newRootCmd(options)

	if err := svrcmd.Execute(rootCmd, options.DefaultHome); err != nil {
		switch e := err.(type) {
		case server.ErrorCode:
			os.Exit(e.Code)

		default:
			os.Exit(1)
		}
	}
}

type Inputs struct {
	dig.In

	RootCommands []*cobra.Command `group:"cli.root"`
}

func newRootCmd(options Options) *cobra.Command {
	a, err := internal.NewAppProvider(options.DefaultAppConfig)
	if err != nil {
		panic(err)
	}

	err = a.Provide(func() string { return options.DefaultHome }, dig.Name("cli.default-home"))
	if err != nil {
		panic(err)
	}

	initClientCtx := client.Context{}.
		WithInput(os.Stdin).
		WithHomeDir(options.DefaultHome).
		WithViper(options.EnvPrefix)

	err = a.Invoke(func(
		codec codec.JSONCodec,
		registry codectypes.InterfaceRegistry,
		txConfig client.TxConfig,
		amino *codec.LegacyAmino,
		accountRetriever client.AccountRetriever,
	) {
		initClientCtx = initClientCtx.
			WithJSONCodec(codec).
			WithInterfaceRegistry(registry).
			WithTxConfig(txConfig).
			WithLegacyAmino(amino).
			WithAccountRetriever(accountRetriever)
	})
	if err != nil {
		panic(err)
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

	err = a.Invoke(func(inputs Inputs) {
		rootCmd.AddCommand(inputs.RootCommands...)
	})
	if err != nil {
		panic(err)
	}

	rootCmd.AddCommand(
		//TODO: AddGenesisAccountCmd(options.DefaultHome),
		tmcli.NewCompletionCmd(rootCmd, true),
		//TODO: testnetCmd(simapp.ModuleBasics, banktypes.GenesisBalancesIterator{}),
		debug.Cmd(),
		config.Cmd(),
	)

	server.AddCommands(rootCmd, options.DefaultHome, a.AppCreator, a.AppExportor, addModuleInitFlags)

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
