package cmd

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// TODO: Go through all functions and methods called here and ensure global Viper
// usage is removed.
//
// REF: https://github.com/cosmos/cosmos-sdk/issues/6571

var (
	rootCmd = &cobra.Command{
		Use:   "simd",
		Short: "simulation app",
	}

	encodingConfig = simapp.MakeEncodingConfig()
	initClientCtx  = client.Context{}.
			WithJSONMarshaler(encodingConfig.Marshaler).
			WithTxGenerator(encodingConfig.TxGenerator).
			WithCodec(encodingConfig.Amino).
			WithInput(os.Stdin).
			WithAccountRetriever(types.NewAccountRetriever(encodingConfig.Marshaler)).
			WithBroadcastMode(flags.BroadcastBlock)
)

// Execute executes the root command.
func Execute() error {
	// Create and set a client.Context on the command's Context. During the pre-run
	// of the root command, a default initialized client.Context is provided to
	// seed child command execution with values such as AccountRetriver, Keyring,
	// and a Tendermint RPC. This requires the use of a pointer reference when
	// getting and setting the client.Context. Ideally, we utilize
	// https://github.com/spf13/cobra/pull/1118.
	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &client.Context{})

	return rootCmd.ExecuteContext(ctx)
}

func init() {
	authclient.Codec = encodingConfig.Marshaler

	// add application daemon commands
	cdc := encodingConfig.Amino
	ctx := server.NewDefaultContext()
	// rootCmd.PersistentPreRunE = server.PersistentPreRunEFn(ctx)
	rootCmd.PersistentFlags().String(flags.FlagHome, simapp.DefaultNodeHome, "The application's root directory")

	rootCmd.AddCommand(
		genutilcli.InitCmd(ctx, cdc, simapp.ModuleBasics, simapp.DefaultNodeHome),
		genutilcli.CollectGenTxsCmd(ctx, cdc, banktypes.GenesisBalancesIterator{}, simapp.DefaultNodeHome),
		genutilcli.MigrateGenesisCmd(ctx, cdc),
		genutilcli.GenTxCmd(
			ctx, cdc, simapp.ModuleBasics, staking.AppModuleBasic{},
			banktypes.GenesisBalancesIterator{}, simapp.DefaultNodeHome, simapp.DefaultNodeHome,
		),
		genutilcli.ValidateGenesisCmd(ctx, cdc, simapp.ModuleBasics),
		AddGenesisAccountCmd(ctx, cdc, encodingConfig.Marshaler, simapp.DefaultNodeHome, simapp.DefaultCLIHome),
		flags.NewCompletionCmd(rootCmd, true),
		testnetCmd(ctx, cdc, simapp.ModuleBasics, banktypes.GenesisBalancesIterator{}),
		debug.Cmd(cdc),
	)

	server.AddCommands(ctx, cdc, rootCmd, newApp, exportAppStateAndTMValidators)

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		queryCommand(),
		txCommand(),
		keys.Commands(),
	)
}

func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return client.SetCmdClientContextHandler(initClientCtx, cmd)
		},
		RunE: client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetAccountCmd(encodingConfig.Amino),
		rpc.ValidatorCommand(encodingConfig.Amino),
		rpc.BlockCommand(),
		authcmd.QueryTxsByEventsCmd(encodingConfig.Amino),
		authcmd.QueryTxCmd(encodingConfig.Amino),
	)

	simapp.ModuleBasics.AddQueryCommands(cmd, initClientCtx)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "network chain ID")

	return cmd
}

func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return client.SetCmdClientContextHandler(initClientCtx, cmd)
		},
		RunE: client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(initClientCtx),
		authcmd.GetSignBatchCommand(encodingConfig.Amino),
		authcmd.GetMultiSignCommand(initClientCtx),
		authcmd.GetValidateSignaturesCommand(initClientCtx),
		flags.LineBreak,
		authcmd.GetBroadcastCommand(initClientCtx),
		authcmd.GetEncodeCommand(initClientCtx),
		authcmd.GetDecodeCommand(initClientCtx),
		flags.LineBreak,
	)

	simapp.ModuleBasics.AddTxCommands(cmd, initClientCtx)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "network chain ID")

	return cmd
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) server.Application {
	var cache sdk.MultiStorePersistentCache

	if viper.GetBool(server.FlagInterBlockCache) {
		cache = store.NewCommitKVStoreCacheManager()
	}

	skipUpgradeHeights := make(map[int64]bool)
	for _, h := range viper.GetIntSlice(server.FlagUnsafeSkipUpgrades) {
		skipUpgradeHeights[int64(h)] = true
	}

	pruningOpts, err := server.GetPruningOptionsFromFlags()
	if err != nil {
		panic(err)
	}

	return simapp.NewSimApp(
		logger, db, traceStore, true, skipUpgradeHeights,
		viper.GetString(flags.FlagHome), viper.GetUint(server.FlagInvCheckPeriod),
		baseapp.SetPruning(pruningOpts),
		baseapp.SetMinGasPrices(viper.GetString(server.FlagMinGasPrices)),
		baseapp.SetHaltHeight(viper.GetUint64(server.FlagHaltHeight)),
		baseapp.SetHaltTime(viper.GetUint64(server.FlagHaltTime)),
		baseapp.SetInterBlockCache(cache),
	)
}

func exportAppStateAndTMValidators(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string,
) (json.RawMessage, []tmtypes.GenesisValidator, *abci.ConsensusParams, error) {

	var simApp *simapp.SimApp
	if height != -1 {
		simApp = simapp.NewSimApp(logger, db, traceStore, false, map[int64]bool{}, "", uint(1))

		err := simApp.LoadHeight(height)
		if err != nil {
			return nil, nil, nil, err
		}
	} else {
		simApp = simapp.NewSimApp(logger, db, traceStore, true, map[int64]bool{}, "", uint(1))
	}

	return simApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
}
