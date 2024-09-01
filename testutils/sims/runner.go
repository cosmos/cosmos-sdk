package sims

import (
	"fmt"
	"io"
	"path/filepath"
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
)

const SimAppChainID = "simulation-app"

// this list of seeds was imported from the original simulation runner: https://github.com/cosmos/tools/blob/v1.0.0/cmd/runsim/main.go#L32
var defaultSeeds = []int64{
	1, 2, 4, 7,
	32, 123, 124, 582, 1893, 2989,
	3012, 4728, 37827, 981928, 87821, 891823782,
	989182, 89182391, 11, 22, 44, 77, 99, 2020,
	3232, 123123, 124124, 582582, 18931893,
	29892989, 30123012, 47284728, 7601778, 8090485,
	977367484, 491163361, 424254581, 673398983,
}

type SimStateFactory struct {
	Codec       codec.Codec
	AppStateFn  simtypes.AppStateFn
	BlockedAddr map[string]bool
}

// SimulationApp abstract app that is used by sims
type SimulationApp interface {
	runtime.AppSimI
	SetNotSigverifyTx()
	GetBaseApp() *baseapp.BaseApp
	TxConfig() client.TxConfig
}

// Run is a helper function that runs a simulation test with the given parameters.
// It calls the RunWithSeeds function with the default seeds and parameters.
//
// This is the entrypoint to run simulation tests that used to run with the runsim binary.
func Run[T SimulationApp](
	t *testing.T,
	appFactory func(
		logger log.Logger,
		db corestore.KVStoreWithBatch,
		traceStore io.Writer,
		loadLatest bool,
		appOpts servertypes.AppOptions,
		baseAppOptions ...func(*baseapp.BaseApp),
	) T,
	setupStateFactory func(app T) SimStateFactory,
	postRunActions ...func(t *testing.T, app TestInstance[T]),
) {
	t.Helper()
	RunWithSeeds(t, appFactory, setupStateFactory, defaultSeeds, nil, postRunActions...)
}

// RunWithSeeds is a helper function that runs a simulation test with the given parameters.
// It iterates over the provided seeds and runs the simulation test for each seed in parallel.
//
// It sets up the environment, creates an instance of the simulation app,
// calls the simulation.SimulateFromSeed function to run the simulation, and performs post-run actions for each seed.
// The execution is deterministic and can be used for fuzz tests as well.
//
// The system under test is isolated for each run but unlike the old runsim command, there is no Process separation.
// This means, global caches may be reused for example. This implementation build upon the vanialla Go stdlib test framework.
func RunWithSeeds[T SimulationApp](
	t *testing.T,
	appFactory func(
		logger log.Logger,
		db corestore.KVStoreWithBatch,
		traceStore io.Writer,
		loadLatest bool,
		appOpts servertypes.AppOptions,
		baseAppOptions ...func(*baseapp.BaseApp),
	) T,
	setupStateFactory func(app T) SimStateFactory,
	seeds []int64,
	fuzzSeed []byte,
	postRunActions ...func(t *testing.T, app TestInstance[T]),
) {
	t.Helper()
	cfg := cli.NewConfigFromFlags()
	cfg.ChainID = SimAppChainID
	for i := range seeds {
		seed := seeds[i]
		t.Run(fmt.Sprintf("seed: %d", seed), func(t *testing.T) {
			t.Parallel()
			// setup environment
			tCfg := cfg.With(t, seed, fuzzSeed)
			testInstance := NewSimulationAppInstance(t, tCfg, appFactory)
			var runLogger log.Logger
			if cli.FlagVerboseValue {
				runLogger = log.NewTestLogger(t)
			} else {
				runLogger = log.NewTestLoggerInfo(t)
			}
			runLogger = runLogger.With("seed", tCfg.Seed)

			app := testInstance.App
			stateFactory := setupStateFactory(app)
			simParams, err := simulation.SimulateFromSeedX(
				t,
				runLogger,
				WriteToDebugLog(runLogger),
				app.GetBaseApp(),
				stateFactory.AppStateFn,
				simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
				simtestutil.SimulationOperations(app, stateFactory.Codec, tCfg, testInstance.App.TxConfig()),
				stateFactory.BlockedAddr,
				tCfg,
				stateFactory.Codec,
				app.TxConfig().SigningContext().AddressCodec(),
				testInstance.ExecLogWriter,
			)
			require.NoError(t, err)
			err = simtestutil.CheckExportSimulation(app, tCfg, simParams)
			require.NoError(t, err)
			if tCfg.Commit {
				simtestutil.PrintStats(testInstance.DB.(*dbm.GoLevelDB))
			}
			for _, step := range postRunActions {
				step(t, testInstance)
			}
		})
	}
}

// TestInstance is a generic type that represents an instance of a SimulationApp used for testing simulations.
// It contains the following fields:
//   - App: The instance of the SimulationApp under test.
//   - DB: The LevelDB database for the simulation app.
//   - WorkDir: The temporary working directory for the simulation app.
//   - Cfg: The configuration flags for the simulator.
//   - AppLogger: The logger used for logging in the app during the simulation, with seed value attached.
//   - ExecLogWriter: Captures block and operation data coming from the simulation
type TestInstance[T SimulationApp] struct {
	App           T
	DB            corestore.KVStoreWithBatch
	WorkDir       string
	Cfg           simtypes.Config
	AppLogger     log.Logger
	ExecLogWriter simulation.LogWriter
}

// NewSimulationAppInstance initializes and returns a TestInstance of a SimulationApp.
// The function takes a testing.T instance, a simtypes.Config instance, and an appFactory function as parameters.
// It creates a temporary working directory and a LevelDB database for the simulation app.
// The function then initializes a logger based on the verbosity flag and sets the logger's seed to the test configuration's seed.
// The database is closed and cleaned up on test completion.
func NewSimulationAppInstance[T SimulationApp](
	t *testing.T,
	tCfg simtypes.Config,
	appFactory func(logger log.Logger, db corestore.KVStoreWithBatch, traceStore io.Writer, loadLatest bool, appOpts servertypes.AppOptions, baseAppOptions ...func(*baseapp.BaseApp)) T,
) TestInstance[T] {
	t.Helper()
	workDir := t.TempDir()
	dbDir := filepath.Join(workDir, "leveldb-app-sim")
	var logger log.Logger
	if cli.FlagVerboseValue {
		logger = log.NewTestLogger(t)
	} else {
		logger = log.NewTestLoggerError(t)
	}
	logger = logger.With("seed", tCfg.Seed)

	db, err := dbm.NewDB("Simulation", dbm.BackendType(tCfg.DBBackend), dbDir)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})
	appOptions := make(simtestutil.AppOptionsMap)
	appOptions[flags.FlagHome] = workDir
	appOptions[server.FlagInvCheckPeriod] = cli.FlagPeriodValue

	app := appFactory(logger, db, nil, true, appOptions, baseapp.SetChainID(SimAppChainID))
	if !cli.FlagSigverifyTxValue {
		app.SetNotSigverifyTx()
	}
	return TestInstance[T]{
		App:           app,
		DB:            db,
		WorkDir:       workDir,
		Cfg:           tCfg,
		AppLogger:     logger,
		ExecLogWriter: &simulation.StandardLogWriter{Seed: tCfg.Seed},
	}
}

var _ io.Writer = writerFn(nil)

type writerFn func(p []byte) (n int, err error)

func (w writerFn) Write(p []byte) (n int, err error) {
	return w(p)
}

// WriteToDebugLog is an adapter to io.Writer interface
func WriteToDebugLog(logger log.Logger) io.Writer {
	return writerFn(func(p []byte) (n int, err error) {
		logger.Debug(string(p))
		return len(p), nil
	})
}

// AppOptionsFn is an adapter to the single method AppOptions interface
type AppOptionsFn func(string) any

func (f AppOptionsFn) Get(k string) any {
	return f(k)
}

// FauxMerkleModeOpt returns a BaseApp option to use a dbStoreAdapter instead of
// an IAVLStore for faster simulation speed.
func FauxMerkleModeOpt(bapp *baseapp.BaseApp) {
	bapp.SetFauxMerkleMode()
}
