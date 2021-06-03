package cli

import (
	"os"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/cosmos/cosmos-sdk/app/internal"

	"github.com/spf13/cobra"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/server"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingcli "github.com/cosmos/cosmos-sdk/x/auth/vesting/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
)

type Options struct {
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

func newRootCmd(options Options) *cobra.Command {
	a, err := internal.NewAppProvider(options.DefaultAppConfig)
	if err != nil {
		panic(err)
	}

	initClientCtx := client.Context{}.
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithHomeDir(options.DefaultHome).
		WithViper("") // In simapp, we don't use any prefix for env variables.

	err = a.Invoke(func(
		codec codec.JSONCodec,
		registry codectypes.InterfaceRegistry,
		txConfig client.TxConfig,
		amino *codec.LegacyAmino,
	) {
		initClientCtx = initClientCtx.
			WithJSONCodec(codec).
			WithInterfaceRegistry(registry).
			WithTxConfig(txConfig).
			WithLegacyAmino(amino)
	})
	if err != nil {
		panic(err)
	}

	rootCmd := &cobra.Command{
		Use:   "simd",
		Short: "simulation app",
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

	initRootCmd(options, rootCmd, a, initClientCtx)

	return rootCmd
}

func initRootCmd(options Options, rootCmd *cobra.Command, a *internal.AppProvider, clientCtx client.Context) {
	cfg := sdk.GetConfig()
	cfg.Seal()

	rootCmd.AddCommand(
		genutilcli.InitCmd(simapp.ModuleBasics, options.DefaultHome),
		genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, options.DefaultHome),
		genutilcli.MigrateGenesisCmd(),
		genutilcli.GenTxCmd(simapp.ModuleBasics, clientCtx.TxConfig, banktypes.GenesisBalancesIterator{}, options.DefaultHome),
		genutilcli.ValidateGenesisCmd(simapp.ModuleBasics),
		//AddGenesisAccountCmd(options.DefaultHome),
		tmcli.NewCompletionCmd(rootCmd, true),
		//TODO: testnetCmd(simapp.ModuleBasics, banktypes.GenesisBalancesIterator{}),
		debug.Cmd(),
		config.Cmd(),
	)

	server.AddCommands(rootCmd, options.DefaultHome, a.AppCreator, a.AppExportor, addModuleInitFlags)

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		queryCommand(),
		txCommand(),
		keys.Commands(options.DefaultHome),
	)

	// add rosetta
	rootCmd.AddCommand(server.RosettaCommand(clientCtx.InterfaceRegistry, clientCtx.JSONCodec))
}

func addModuleInitFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
}

func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		// NOTE: moved these to the auth module:
		//authcmd.GetAccountCmd(),
		rpc.ValidatorCommand(),
		rpc.BlockCommand(),
		//authcmd.QueryTxsByEventsCmd(),
		//authcmd.QueryTxCmd(),
	)

	simapp.ModuleBasics.AddQueryCommands(cmd)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		// NOTE: moved these to the auth module:
		//authcmd.GetSignCommand(),
		//authcmd.GetSignBatchCommand(),
		//authcmd.GetMultiSignCommand(),
		//authcmd.GetMultiSignBatchCmd(),
		//authcmd.GetValidateSignaturesCommand(),
		//flags.LineBreak,
		//authcmd.GetBroadcastCommand(),
		//authcmd.GetEncodeCommand(),
		//authcmd.GetDecodeCommand(),
		//flags.LineBreak,
		vestingcli.GetTxCmd(),
	)

	simapp.ModuleBasics.AddTxCommands(cmd)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}
