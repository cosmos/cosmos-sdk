//go:build sims

package simapp

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"io"
	"math/rand"
	"strings"
	"sync"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storetypes "cosmossdk.io/store/types"
	authzkeeper "cosmossdk.io/x/authz/keeper"
	"cosmossdk.io/x/feegrant"
	slashingtypes "cosmossdk.io/x/slashing/types"
	stakingtypes "cosmossdk.io/x/staking/types"
	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/testutils/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SimAppChainID hardcoded chainID for simulation

var FlagEnableStreamingValue bool

// Get flags every time the simulator is run
func init() {
	simcli.GetSimulatorFlags()
	flag.BoolVar(&FlagEnableStreamingValue, "EnableStreaming", false, "Enable streaming service")
}

// interBlockCacheOpt returns a BaseApp option function that sets the persistent
// inter-block write-through cache.
func interBlockCacheOpt() func(*baseapp.BaseApp) {
	return baseapp.SetInterBlockCache(store.NewCommitKVStoreCacheManager())
}

func TestFullAppSimulation(t *testing.T) {
	sims.Run(t, NewSimApp, setupStateFactory)
}

func setupStateFactory(app *SimApp) sims.SimStateFactory {
	return sims.SimStateFactory{
		Codec:       app.AppCodec(),
		AppStateFn:  simtestutil.AppStateFn(app.AppCodec(), app.AuthKeeper.AddressCodec(), app.StakingKeeper.ValidatorAddressCodec(), app.SimulationManager(), app.DefaultGenesis()),
		BlockedAddr: BlockedAddresses(),
	}
}

var (
	exportAllModules       = []string{}
	exportWithValidatorSet = []string{}
)

func TestAppImportExport(t *testing.T) {
	sims.Run(t, NewSimApp, setupStateFactory, func(t *testing.T, ti sims.TestInstance[*SimApp]) {
		app := ti.App
		t.Log("exporting genesis...\n")
		exported, err := app.ExportAppStateAndValidators(false, exportWithValidatorSet, exportAllModules)
		require.NoError(t, err)

		t.Log("importing genesis...\n")
		newTestInstance := sims.NewSimulationAppInstance(t, ti.Cfg, NewSimApp)
		newApp := newTestInstance.App
		var genesisState GenesisState
		require.NoError(t, json.Unmarshal(exported.AppState, &genesisState))
		ctxB := newApp.NewContextLegacy(true, cmtproto.Header{Height: app.LastBlockHeight()})
		_, err = newApp.ModuleManager.InitGenesis(ctxB, genesisState)
		if IsEmptyValidatorSetErr(err) {
			t.Skip("Skipping simulation as all validators have been unbonded")
			return
		}
		require.NoError(t, err)
		err = newApp.StoreConsensusParams(ctxB, exported.ConsensusParams)
		require.NoError(t, err)

		t.Log("comparing stores...")
		// skip certain prefixes
		skipPrefixes := map[string][][]byte{
			stakingtypes.StoreKey: {
				stakingtypes.UnbondingQueueKey, stakingtypes.RedelegationQueueKey, stakingtypes.ValidatorQueueKey,
				stakingtypes.HistoricalInfoKey, stakingtypes.UnbondingIDKey, stakingtypes.UnbondingIndexKey,
				stakingtypes.UnbondingTypeKey,
			},
			authzkeeper.StoreKey:   {authzkeeper.GrantQueuePrefix},
			feegrant.StoreKey:      {feegrant.FeeAllowanceQueueKeyPrefix},
			slashingtypes.StoreKey: {slashingtypes.ValidatorMissedBlockBitmapKeyPrefix},
		}
		AssertEqualStores(t, app, newApp, app.SimulationManager().StoreDecoders, skipPrefixes)
	})
}

// Scenario:
//
//	Start a fresh node and run n blocks, export state
//	set up a new node instance, Init chain from exported genesis
//	run new instance for n blocks
func TestAppSimulationAfterImport(t *testing.T) {
	sims.Run(t, NewSimApp, setupStateFactory, func(t *testing.T, ti sims.TestInstance[*SimApp]) {
		app := ti.App
		t.Log("exporting genesis...\n")
		exported, err := app.ExportAppStateAndValidators(false, exportWithValidatorSet, exportAllModules)
		require.NoError(t, err)

		t.Log("importing genesis...\n")
		newTestInstance := sims.NewSimulationAppInstance(t, ti.Cfg, NewSimApp)
		newApp := newTestInstance.App
		_, err = newApp.InitChain(&abci.InitChainRequest{
			AppStateBytes: exported.AppState,
			ChainId:       sims.SimAppChainID,
		})
		if IsEmptyValidatorSetErr(err) {
			t.Skip("Skipping simulation as all validators have been unbonded")
			return
		}
		require.NoError(t, err)
		newStateFactory := setupStateFactory(newApp)
		_, err = simulation.SimulateFromSeedX(
			t,
			newTestInstance.AppLogger,
			sims.WriteToDebugLog(newTestInstance.AppLogger),
			newApp.BaseApp,
			newStateFactory.AppStateFn,
			simtypes.RandomAccounts,
			simtestutil.SimulationOperations(newApp, newApp.AppCodec(), newTestInstance.Cfg, newApp.TxConfig()),
			newStateFactory.BlockedAddr,
			newTestInstance.Cfg,
			newStateFactory.Codec,
			newApp.TxConfig().SigningContext().AddressCodec(),
			ti.ExecLogWriter,
		)
		require.NoError(t, err)
	})
}

func IsEmptyValidatorSetErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "validator set is empty after InitGenesis")
}

func TestAppStateDeterminism(t *testing.T) {
	const numTimesToRunPerSeed = 3
	var seeds []int64
	if s := simcli.NewConfigFromFlags().Seed; s != simcli.DefaultSeedValue {
		// We will be overriding the random seed and just run a single simulation on the provided seed value
		for j := 0; j < numTimesToRunPerSeed; j++ { // multiple rounds
			seeds = append(seeds, s)
		}
	} else {
		// setup with 3 random seeds
		for i := 0; i < 3; i++ {
			seed := rand.Int63()
			for j := 0; j < numTimesToRunPerSeed; j++ { // multiple rounds
				seeds = append(seeds, seed)
			}
		}
	}
	// overwrite default app config
	interBlockCachingAppFactory := func(logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool, appOpts servertypes.AppOptions, baseAppOptions ...func(*baseapp.BaseApp)) *SimApp {
		if FlagEnableStreamingValue {
			m := map[string]any{
				"streaming.abci.keys":             []string{"*"},
				"streaming.abci.plugin":           "abci_v1",
				"streaming.abci.stop-node-on-err": true,
			}
			others := appOpts
			appOpts = sims.AppOptionsFn(func(k string) any {
				if v, ok := m[k]; ok {
					return v
				}
				return others.Get(k)
			})
		}
		return NewSimApp(logger, db, nil, true, appOpts, append(baseAppOptions, interBlockCacheOpt())...)
	}
	var mx sync.Mutex
	appHashResults := make(map[int64][][]byte)
	appSimLogger := make(map[int64][]simulation.LogWriter)
	captureAndCheckHash := func(t *testing.T, ti sims.TestInstance[*SimApp]) {
		seed, appHash := ti.Cfg.Seed, ti.App.LastCommitID().Hash
		mx.Lock()
		otherHashes, execWriters := appHashResults[seed], appSimLogger[seed]
		if len(otherHashes) < numTimesToRunPerSeed-1 {
			appHashResults[seed], appSimLogger[seed] = append(otherHashes, appHash), append(execWriters, ti.ExecLogWriter)
		} else { // cleanup
			delete(appHashResults, seed)
			delete(appSimLogger, seed)
		}
		mx.Unlock()

		var failNow bool
		// and check that all app hashes per seed are equal for each iteration
		for i := 0; i < len(otherHashes); i++ {
			if !assert.Equal(t, otherHashes[i], appHash) {
				execWriters[i].PrintLogs()
				failNow = true
			}
		}
		if failNow {
			ti.ExecLogWriter.PrintLogs()
			t.Fatalf("non-determinism in seed %d", seed)
		}
	}
	// run simulations
	sims.RunWithSeeds(t, interBlockCachingAppFactory, setupStateFactory, seeds, []byte{}, captureAndCheckHash)
}

type ComparableStoreApp interface {
	LastBlockHeight() int64
	NewContextLegacy(isCheckTx bool, header cmtproto.Header) sdk.Context
	GetKey(storeKey string) *storetypes.KVStoreKey
	GetStoreKeys() []storetypes.StoreKey
}

func AssertEqualStores(t *testing.T, app ComparableStoreApp, newApp ComparableStoreApp, storeDecoders simtypes.StoreDecoderRegistry, skipPrefixes map[string][][]byte) {
	ctxA := app.NewContextLegacy(true, cmtproto.Header{Height: app.LastBlockHeight()})
	ctxB := newApp.NewContextLegacy(true, cmtproto.Header{Height: app.LastBlockHeight()})

	storeKeys := app.GetStoreKeys()
	require.NotEmpty(t, storeKeys)

	for _, appKeyA := range storeKeys {
		// only compare kvstores
		if _, ok := appKeyA.(*storetypes.KVStoreKey); !ok {
			continue
		}

		keyName := appKeyA.Name()
		appKeyB := newApp.GetKey(keyName)

		storeA := ctxA.KVStore(appKeyA)
		storeB := ctxB.KVStore(appKeyB)

		failedKVAs, failedKVBs := simtestutil.DiffKVStores(storeA, storeB, skipPrefixes[keyName])
		require.Equal(t, len(failedKVAs), len(failedKVBs), "unequal sets of key-values to compare %s, key stores %s and %s", keyName, appKeyA, appKeyB)

		t.Logf("compared %d different key/value pairs between %s and %s\n", len(failedKVAs), appKeyA, appKeyB)
		if !assert.Equal(t, 0, len(failedKVAs), simtestutil.GetSimulationLog(keyName, storeDecoders, failedKVAs, failedKVBs)) {
			for _, v := range failedKVAs {
				t.Logf("store mismatch: %q\n", v)
			}
			t.FailNow()
		}
	}
}

func FuzzFullAppSimulation(f *testing.F) {
	f.Fuzz(func(t *testing.T, rawSeed []byte) {
		if len(rawSeed) < 8 {
			t.Skip()
			return
		}
		sims.RunWithSeeds(
			t,
			NewSimApp,
			setupStateFactory,
			[]int64{int64(binary.BigEndian.Uint64(rawSeed))},
			rawSeed[8:],
		)
	})
}
