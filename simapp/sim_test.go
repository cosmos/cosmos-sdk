package simapp

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	FlagEnableStreamingValue bool
)

// Get flags every time the simulator is run
func init() {
	GetSimulatorFlags()
	flag.BoolVar(&FlagEnableStreamingValue, "EnableStreaming", false, "Enable streaming service")
}

type StoreKeysPrefixes struct {
	A        storetypes.StoreKey
	B        storetypes.StoreKey
	Prefixes [][]byte
}

// fauxMerkleModeOpt returns a BaseApp option to use a dbStoreAdapter instead of
// an IAVLStore for faster simulation speed.
func fauxMerkleModeOpt(bapp *baseapp.BaseApp) {
	bapp.SetFauxMerkleMode()
}

// interBlockCacheOpt returns a BaseApp option function that sets the persistent
// inter-block write-through cache.
func interBlockCacheOpt() func(*baseapp.BaseApp) {
	return baseapp.SetInterBlockCache(store.NewCommitKVStoreCacheManager())
}

func TestFullAppSimulation(t *testing.T) {
	config, db, dir, logger, skip, err := SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application simulation")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		require.NoError(t, db.Close())
		require.NoError(t, os.RemoveAll(dir))
	}()

	app := NewSimApp(logger, db, nil, true, map[int64]bool{}, DefaultNodeHome, FlagPeriodValue, MakeTestEncodingConfig(), EmptyAppOptions{}, fauxMerkleModeOpt)
	require.Equal(t, "SimApp", app.Name())

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		AppStateFn(app.AppCodec(), app.SimulationManager()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		SimulationOperations(app, app.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		PrintStats(db)
	}
}

func TestAppImportExport(t *testing.T) {
	config, db, dir, logger, skip, err := SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application import/export simulation")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		require.NoError(t, db.Close())
		require.NoError(t, os.RemoveAll(dir))
	}()

	app := NewSimApp(logger, db, nil, true, map[int64]bool{}, DefaultNodeHome, FlagPeriodValue, MakeTestEncodingConfig(), EmptyAppOptions{}, fauxMerkleModeOpt)
	require.Equal(t, "SimApp", app.Name())

	// Run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		AppStateFn(app.AppCodec(), app.SimulationManager()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		SimulationOperations(app, app.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		PrintStats(db)
	}

	fmt.Printf("exporting genesis...\n")

	exported, err := app.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err)

	fmt.Printf("importing genesis...\n")

	_, newDB, newDir, _, _, err := SetupSimulation("leveldb-app-sim-2", "Simulation-2")
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		require.NoError(t, newDB.Close())
		require.NoError(t, os.RemoveAll(newDir))
	}()

	newApp := NewSimApp(log.NewNopLogger(), newDB, nil, true, map[int64]bool{}, DefaultNodeHome, FlagPeriodValue, MakeTestEncodingConfig(), EmptyAppOptions{}, fauxMerkleModeOpt)
	require.Equal(t, "SimApp", newApp.Name())

	var genesisState GenesisState
	err = json.Unmarshal(exported.AppState, &genesisState)
	require.NoError(t, err)

	defer func() {
		if r := recover(); r != nil {
			err := fmt.Sprintf("%v", r)
			if !strings.Contains(err, "validator set is empty after InitGenesis") {
				panic(r)
			}
			logger.Info("Skipping simulation as all validators have been unbonded")
			logger.Info("err", err, "stacktrace", string(debug.Stack()))
		}
	}()

	ctxA := app.NewContext(true, tmproto.Header{Height: app.LastBlockHeight()})
	ctxB := newApp.NewContext(true, tmproto.Header{Height: app.LastBlockHeight()})
	newApp.mm.InitGenesis(ctxB, app.AppCodec(), genesisState)
	newApp.StoreConsensusParams(ctxB, exported.ConsensusParams)

	fmt.Printf("comparing stores...\n")

	storeKeysPrefixes := []StoreKeysPrefixes{
		{app.keys[authtypes.StoreKey], newApp.keys[authtypes.StoreKey], [][]byte{}},
		{
			app.keys[stakingtypes.StoreKey], newApp.keys[stakingtypes.StoreKey],
			[][]byte{
				stakingtypes.UnbondingQueueKey, stakingtypes.RedelegationQueueKey, stakingtypes.ValidatorQueueKey,
				stakingtypes.HistoricalInfoKey,
			},
		}, // ordering may change but it doesn't matter
		{app.keys[slashingtypes.StoreKey], newApp.keys[slashingtypes.StoreKey], [][]byte{}},
		{app.keys[minttypes.StoreKey], newApp.keys[minttypes.StoreKey], [][]byte{}},
		{app.keys[distrtypes.StoreKey], newApp.keys[distrtypes.StoreKey], [][]byte{}},
		{app.keys[banktypes.StoreKey], newApp.keys[banktypes.StoreKey], [][]byte{banktypes.BalancesPrefix}},
		{app.keys[paramtypes.StoreKey], newApp.keys[paramtypes.StoreKey], [][]byte{}},
		{app.keys[govtypes.StoreKey], newApp.keys[govtypes.StoreKey], [][]byte{}},
		{app.keys[evidencetypes.StoreKey], newApp.keys[evidencetypes.StoreKey], [][]byte{}},
		{app.keys[capabilitytypes.StoreKey], newApp.keys[capabilitytypes.StoreKey], [][]byte{}},
		{app.keys[authzkeeper.StoreKey], newApp.keys[authzkeeper.StoreKey], [][]byte{authzkeeper.GrantKey, authzkeeper.GrantQueuePrefix}},
		{app.keys[quarantine.StoreKey], newApp.keys[quarantine.StoreKey], [][]byte{}},
		{app.keys[sanction.StoreKey], newApp.keys[sanction.StoreKey], [][]byte{}},
	}

	for _, skp := range storeKeysPrefixes {
		storeA := ctxA.KVStore(skp.A)
		storeB := ctxB.KVStore(skp.B)

		failedKVAs, failedKVBs := sdk.DiffKVStores(storeA, storeB, skp.Prefixes)
		require.Equal(t, len(failedKVAs), len(failedKVBs), "unequal sets of key-values to compare")

		fmt.Printf("compared %d different key/value pairs between %s and %s\n", len(failedKVAs), skp.A, skp.B)
		require.Equal(t, 0, len(failedKVAs), GetSimulationLog(skp.A.Name(), app.SimulationManager().StoreDecoders, failedKVAs, failedKVBs))
	}
}

func TestAppSimulationAfterImport(t *testing.T) {
	config, db, dir, logger, skip, err := SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application simulation after import")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		require.NoError(t, db.Close())
		require.NoError(t, os.RemoveAll(dir))
	}()

	app := NewSimApp(logger, db, nil, true, map[int64]bool{}, DefaultNodeHome, FlagPeriodValue, MakeTestEncodingConfig(), EmptyAppOptions{}, fauxMerkleModeOpt)
	require.Equal(t, "SimApp", app.Name())

	// Run randomized simulation
	stopEarly, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		AppStateFn(app.AppCodec(), app.SimulationManager()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		SimulationOperations(app, app.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		PrintStats(db)
	}

	fmt.Printf("exporting genesis...\n")

	exported, err := app.ExportAppStateAndValidators(true, []string{})
	require.NoError(t, err)

	if stopEarly || len(exported.Validators) == 0 {
		fmt.Println("can't import a zero-validator genesis, exiting test...")
		return
	}

	fmt.Printf("importing genesis...\n")

	_, newDB, newDir, _, _, err := SetupSimulation("leveldb-app-sim-2", "Simulation-2")
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		require.NoError(t, newDB.Close())
		require.NoError(t, os.RemoveAll(newDir))
	}()

	newApp := NewSimApp(log.NewNopLogger(), newDB, nil, true, map[int64]bool{}, DefaultNodeHome, FlagPeriodValue, MakeTestEncodingConfig(), EmptyAppOptions{}, fauxMerkleModeOpt)
	require.Equal(t, "SimApp", newApp.Name())

	newApp.InitChain(abci.RequestInitChain{
		AppStateBytes: exported.AppState,
	})

	_, _, err = simulation.SimulateFromSeed(
		t,
		os.Stdout,
		newApp.BaseApp,
		AppStateFn(app.AppCodec(), app.SimulationManager()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		SimulationOperations(newApp, newApp.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)
	require.NoError(t, err)
}

// TODO: Make another test for the fuzzer itself, which just has noOp txs
// and doesn't depend on the application.
func TestAppStateDeterminism(t *testing.T) {
	if !FlagEnabledValue {
		t.Skip("skipping application simulation")
	}

	vpr := viper.New()
	if FlagEnableStreamingValue {
		m := make(map[string]interface{})
		m["streaming.abci.keys"] = []string{"*"}
		m["streaming.abci.plugin"] = "abci_v1"
		m["streaming.abci.stop-node-on-err"] = true
		for key, value := range m {
			vpr.SetDefault(key, value)
		}
	}

	config := NewConfigFromFlags()
	config.InitialBlockHeight = 1
	config.ExportParamsPath = ""
	config.OnOperation = false
	config.AllInvariants = false
	config.ChainID = helpers.SimAppChainID

	numSeeds := 3
	numTimesToRunPerSeed := 5
	appHashList := make([]json.RawMessage, numTimesToRunPerSeed)

	for i := 0; i < numSeeds; i++ {
		config.Seed = rand.Int63()

		for j := 0; j < numTimesToRunPerSeed; j++ {
			var logger log.Logger
			if FlagVerboseValue {
				logger = log.TestingLogger()
			} else {
				logger = log.NewNopLogger()
			}

			db := dbm.NewMemDB()
			app := NewSimApp(logger, db, nil, true, map[int64]bool{}, DefaultNodeHome, FlagPeriodValue, MakeTestEncodingConfig(), vpr, interBlockCacheOpt())

			fmt.Printf(
				"running non-determinism simulation; seed %d: %d/%d, attempt: %d/%d\n",
				config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
			)

			_, _, err := simulation.SimulateFromSeed(
				t,
				os.Stdout,
				app.BaseApp,
				AppStateFn(app.AppCodec(), app.SimulationManager()),
				simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
				SimulationOperations(app, app.AppCodec(), config),
				app.ModuleAccountAddrs(),
				config,
				app.AppCodec(),
			)
			require.NoError(t, err)

			if config.Commit {
				PrintStats(db)
			}

			appHash := app.LastCommitID().Hash
			appHashList[j] = appHash

			if j != 0 {
				require.Equal(
					t, string(appHashList[0]), string(appHashList[j]),
					"non-determinism in seed %d: %d/%d, attempt: %d/%d\n", config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
				)
			}
		}
	}
}
