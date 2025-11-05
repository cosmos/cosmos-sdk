package simsx

import (
	"io"
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
)

func RunBlocks[T SimulationApp](
	tb testing.TB,
	appFactory func(
		logger log.Logger,
		db dbm.DB,
		traceStore io.Writer,
		loadLatest bool,
		appOpts servertypes.AppOptions,
		baseAppOptions ...func(*baseapp.BaseApp),
	) T,
	setupStateFactory func(app T) SimStateFactory,
	postRunActions ...func(t testing.TB, app TestInstance[T], accs []simtypes.Account),
) {
	tb.Helper()

	cfg := cli.NewConfigFromFlags()
	cfg.ChainID = SimAppChainID
	// setup environment
	testInstance := NewSimulationAppInstance(tb, cfg, appFactory)
	var runLogger log.Logger
	if cli.FlagVerboseValue {
		runLogger = log.NewTestLogger(tb)
	} else {
		runLogger = log.NewTestLoggerInfo(tb)
	}
	runLogger = runLogger.With("seed", cfg.Seed)

	app := testInstance.App
	stateFactory := setupStateFactory(app)
	ops, _ := prepareWeightedOps(app.SimulationManager(), stateFactory, cfg, testInstance.App.TxConfig(), runLogger)
	simParams, accs, err := simulation.SimulateFromSeedX(
		tb,
		runLogger,
		WriteToDebugLog(runLogger),
		app.GetBaseApp(),
		stateFactory.AppStateFn,
		simtypes.RandomAccounts,
		ops,
		stateFactory.BlockedAddr,
		cfg,
		stateFactory.Codec,
		testInstance.ExecLogWriter,
	)
	require.NoError(tb, err)
	err = simtestutil.CheckExportSimulation(app, cfg, simParams)
	require.NoError(tb, err)
	if cfg.Commit {
		simtestutil.PrintStats(testInstance.DB)
	}
	// not using tb.Log to always print the summary
	for _, step := range postRunActions {
		step(tb, testInstance, accs)
	}
	require.NoError(tb, app.Close())
}
