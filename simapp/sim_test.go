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
	"time"

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storetypes "cosmossdk.io/store/types"
	authzkeeper "cosmossdk.io/x/authz/keeper"
	"cosmossdk.io/x/feegrant"
	slashingtypes "cosmossdk.io/x/slashing/types"
	stakingtypes "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/simsx"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
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
	simsx.Run(t, NewSimApp, setupStateFactory)
}

func setupStateFactory(app *SimApp) simsx.SimStateFactory {
	blockedAddre, _ := BlockedAddresses(app.interfaceRegistry.SigningContext().AddressCodec())
	return simsx.SimStateFactory{
		Codec:         app.AppCodec(),
		AppStateFn:    simtestutil.AppStateFn(app.AppCodec(), app.AuthKeeper.AddressCodec(), app.StakingKeeper.ValidatorAddressCodec(), app.SimulationManager().Modules, app.DefaultGenesis()),
		BlockedAddr:   blockedAddre,
		AccountSource: app.AuthKeeper,
		BalanceSource: app.BankKeeper,
	}
}

var (
	exportAllModules       []string
	exportWithValidatorSet []string
)

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
		{app.keys[authzkeeper.StoreKey], newApp.keys[authzkeeper.StoreKey], [][]byte{}},
	}

	for _, skp := range storeKeysPrefixes {
		storeA := ctxA.KVStore(skp.A)
		storeB := ctxB.KVStore(skp.B)

		failedKVAs, failedKVBs := sdk.DiffKVStores(storeA, storeB, skp.Prefixes)
		require.Equal(t, len(failedKVAs), len(failedKVBs), "unequal sets of key-values to compare")

		fmt.Printf("compared %d different key/value pairs between %s and %s\n", len(failedKVAs), skp.A, skp.B)
		require.Equal(t, len(failedKVAs), 0, GetSimulationLog(skp.A.Name(), app.SimulationManager().StoreDecoders, failedKVAs, failedKVBs))
	}
}

// Scenario:
//
//	Start a fresh node and run n blocks, export state
//	set up a new node instance, Init chain from exported genesis
//	run new instance for n blocks
func TestAppSimulationAfterImport(t *testing.T) {
	simsx.Run(t, NewSimApp, setupStateFactory, func(t testing.TB, ti simsx.TestInstance[*SimApp], accs []simtypes.Account) {
		app := ti.App
		t.Log("exporting genesis...\n")
		exported, err := app.ExportAppStateAndValidators(false, exportWithValidatorSet, exportAllModules)
		require.NoError(t, err)

		importGenesisStateFactory := func(app *SimApp) simsx.SimStateFactory {
			return simsx.SimStateFactory{
				Codec: app.AppCodec(),
				AppStateFn: func(r *rand.Rand, _ []simtypes.Account, config simtypes.Config) (json.RawMessage, []simtypes.Account, string, time.Time) {
					t.Log("importing genesis...\n")
					genesisTimestamp := time.Unix(config.GenesisTime, 0)

					_, err = app.InitChain(&abci.InitChainRequest{
						AppStateBytes: exported.AppState,
						ChainId:       simsx.SimAppChainID,
						InitialHeight: exported.Height,
						Time:          genesisTimestamp,
					})
					if IsEmptyValidatorSetErr(err) {
						t.Skip("Skipping simulation as all validators have been unbonded")
						return nil, nil, "", time.Time{}
					}
					require.NoError(t, err)
					// use accounts from initial run
					return exported.AppState, accs, config.ChainID, genesisTimestamp
				},
				BlockedAddr:   must(BlockedAddresses(app.AuthKeeper.AddressCodec())),
				AccountSource: app.AuthKeeper,
				BalanceSource: app.BankKeeper,
			}
		}
		ti.Cfg.InitialBlockHeight = int(exported.Height)
		simsx.RunWithSeed(t, ti.Cfg, NewSimApp, importGenesisStateFactory, ti.Cfg.Seed, ti.Cfg.FuzzSeed)
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
	interBlockCachingAppFactory := func(logger log.Logger, db corestore.KVStoreWithBatch, traceStore io.Writer, loadLatest bool, appOpts servertypes.AppOptions, baseAppOptions ...func(*baseapp.BaseApp)) *SimApp {
		if FlagEnableStreamingValue {
			m := map[string]any{
				"streaming.abci.keys":             []string{"*"},
				"streaming.abci.plugin":           "abci_v1",
				"streaming.abci.stop-node-on-err": true,
			}
			others := appOpts
			appOpts = simsx.AppOptionsFn(func(k string) any {
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
	captureAndCheckHash := func(t testing.TB, ti simsx.TestInstance[*SimApp], _ []simtypes.Account) {
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
	simsx.RunWithSeeds(t, interBlockCachingAppFactory, setupStateFactory, seeds, []byte{}, captureAndCheckHash)
}

type ComparableStoreApp interface {
	LastBlockHeight() int64
	NewContextLegacy(isCheckTx bool, header cmtproto.Header) sdk.Context
	GetKey(storeKey string) *storetypes.KVStoreKey
	GetStoreKeys() []storetypes.StoreKey
}

func AssertEqualStores(t testing.TB, app, newApp ComparableStoreApp, storeDecoders simtypes.StoreDecoderRegistry, skipPrefixes map[string][][]byte) {
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
		simsx.RunWithSeeds(
			t,
			NewSimApp,
			setupStateFactory,
			[]int64{int64(binary.BigEndian.Uint64(rawSeed))},
			rawSeed[8:],
		)
	})
}

func must[T any](r T, err error) T {
	if err != nil {
		panic(err)
	}
	return r
}
