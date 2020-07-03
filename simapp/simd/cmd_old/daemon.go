package main

import (
	"encoding/json"
	"io"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

const flagInvCheckPeriod = "inv-check-period"

var invCheckPeriod uint

func addDaemonCommands(rootDaemonCmd *cobra.Command) {
	appCodec, cdc := simapp.MakeCodecs()

	ctx := server.NewDefaultContext()
	rootDaemonCmd.PersistentPreRunE = server.PersistentPreRunEFn(ctx)

	rootDaemonCmd.AddCommand(
		genutilcli.InitCmd(ctx, cdc, simapp.ModuleBasics, simapp.DefaultNodeHome),
		genutilcli.CollectGenTxsCmd(ctx, cdc, banktypes.GenesisBalancesIterator{}, simapp.DefaultNodeHome),
		genutilcli.MigrateGenesisCmd(ctx, cdc),
		genutilcli.GenTxCmd(
			ctx, cdc, simapp.ModuleBasics, staking.AppModuleBasic{},
			banktypes.GenesisBalancesIterator{}, simapp.DefaultNodeHome, simapp.DefaultNodeHome,
		),
		genutilcli.ValidateGenesisCmd(ctx, cdc, simapp.ModuleBasics),
		AddGenesisAccountCmd(ctx, cdc, appCodec, simapp.DefaultNodeHome, simapp.DefaultCLIHome),
		flags.NewCompletionCmd(rootDaemonCmd, true),
		testnetCmd(ctx, cdc, simapp.ModuleBasics, banktypes.GenesisBalancesIterator{}),
		debug.Cmd(cdc))

	server.AddCommands(ctx, cdc, rootDaemonCmd, newApp, exportAppStateAndTMValidators)

	// prepare and add flags
	rootDaemonCmd.PersistentFlags().UintVar(&invCheckPeriod, flagInvCheckPeriod,
		0, "Assert registered invariants every N blocks")

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

	// TODO: Make sure custom pruning works.
	return simapp.NewSimApp(
		logger, db, traceStore, true, skipUpgradeHeights,
		viper.GetString(flags.FlagHome), invCheckPeriod,
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
