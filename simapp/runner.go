package simapp

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/simsx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/cosmos/cosmos-sdk/client"

	"cosmossdk.io/log"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	"github.com/stretchr/testify/require"
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

func init() {
	// cli.GetSimulatorFlags()
}

type SimStateFactory struct {
	Codec       codec.Codec
	AppStateFn  simtypes.AppStateFn
	BlockedAddr map[string]bool
}
type SimulationApp interface {
	runtime.AppSimI
	SetNotSigverifyTx()
	GetBaseApp() *baseapp.BaseApp
	TxConfig() client.TxConfig
}

func Run[T SimulationApp](
	t *testing.T,
	appFactory func(
		logger log.Logger,
		db dbm.DB,
		traceStore io.Writer,
		loadLatest bool,
		appOpts servertypes.AppOptions,
		baseAppOptions ...func(*baseapp.BaseApp),
	) T,
	setupStateFactory func(app T) SimStateFactory,
	postRunActions ...func(t *testing.T, app TestInstance[T]),
) {
	RunWithSeeds(t, appFactory, setupStateFactory, defaultSeeds, nil, postRunActions...)
}

func RunWithSeeds[T SimulationApp](
	t *testing.T,
	appFactory func(
		logger log.Logger,
		db dbm.DB,
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
	cfg := cli.NewConfigFromFlags()
	cfg.ChainID = SimAppChainID
	for i := range seeds {
		seed := seeds[i]
		t.Run(fmt.Sprintf("seed: %d", seed), func(t *testing.T) {
			t.Parallel()
			// setup environment
			tCfg := cfg.Clone()
			tCfg.Seed = seed
			tCfg.FuzzSeed = fuzzSeed
			tCfg.T = t
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
			ops, reporter := prepareWeightedOps(app.SimulationManager(), stateFactory.Codec, tCfg, testInstance.App.TxConfig())
			simParams, err := simulation.SimulateFromSeed(
				t,
				runLogger,
				WriteToDebugLog(runLogger),
				app.GetBaseApp(),
				stateFactory.AppStateFn,
				simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
				ops,
				stateFactory.BlockedAddr,
				tCfg,
				stateFactory.Codec,
				app.TxConfig().SigningContext().AddressCodec(),
			)
			require.NoError(t, err)
			err = simtestutil.CheckExportSimulation(app, tCfg, simParams)
			require.NoError(t, err)
			if tCfg.Commit {
				simtestutil.PrintStats(testInstance.DB)
			}
			t.Log("+++ DONE: \n" + reporter.Summary().String())
			for _, step := range postRunActions {
				step(t, testInstance)
			}
		})
	}
}

// fetch weighted operation for all registered modules. included to avoid cyclic dependency in testutils/sims
func prepareWeightedOps(sm *module.SimulationManager, cdc codec.Codec, config simtypes.Config, txConfig client.TxConfig) (simulation.WeightedOperations, *simsx.BasicSimulationReporter) {
	signingCtx := cdc.InterfaceRegistry().SigningContext()
	simState := module.SimulationState{
		AppParams:      make(simtypes.AppParams),
		Cdc:            cdc,
		AddressCodec:   signingCtx.AddressCodec(),
		ValidatorCodec: signingCtx.ValidatorAddressCodec(),
		TxConfig:       txConfig,
		BondDenom:      sdk.DefaultBondDenom,
	}

	if config.ParamsFile != "" {
		bz, err := os.ReadFile(config.ParamsFile)
		if err != nil {
			panic(err)
		}

		err = json.Unmarshal(bz, &simState.AppParams)
		if err != nil {
			panic(err)
		}
	}

	simState.LegacyProposalContents = sm.GetProposalContents(simState) //nolint:staticcheck // we're testing the old way here
	simState.ProposalMsgs = sm.GetProposalMsgs(simState)

	wOps := make([]simtypes.WeightedOperation, 0, len(sm.Modules))
	weights := simsx.ParamWeightSource(simState.AppParams)
	reporter := simsx.NewBasicSimulationReporter()
	reg := simsx.NewSimsRegistryAdapter(reporter, sm.AK, sm.BK, txConfig)

	type weightedOperationsX interface {
		WeightedOperationsX(weight simsx.WeightSource, reg simsx.Registry)
	}

	for _, m := range sm.Modules {
		if xm, ok := m.(weightedOperationsX); ok {
			xm.WeightedOperationsX(weights, reg)
		} else {
			// support legacy entry factory method
			wOps = append(wOps, m.WeightedOperations(simState)...)
		}
	}
	return append(wOps, reg.ToLegacyWeightedOperations()...), reporter
}

type TestInstance[T SimulationApp] struct {
	App     T
	DB      dbm.DB
	WorkDir string
	Cfg     simtypes.Config
	Logger  log.Logger
}

func NewSimulationAppInstance[T SimulationApp](
	t *testing.T,
	tCfg simtypes.Config,
	appFactory func(logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool, appOpts servertypes.AppOptions, baseAppOptions ...func(*baseapp.BaseApp)) T,
) TestInstance[T] {
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
	app := appFactory(logger, db, nil, true, appOptions, FauxMerkleModeOpt, baseapp.SetChainID(SimAppChainID))
	if !cli.FlagSigverifyTxValue {
		app.SetNotSigverifyTx()
	}
	return TestInstance[T]{
		App:     app,
		DB:      db,
		WorkDir: workDir,
		Cfg:     tCfg,
		Logger:  logger,
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
