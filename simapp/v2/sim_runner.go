package simapp

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"maps"
	"math/rand"
	"os"
	"slices"
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/appmodule"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/comet"
	corecontext "cosmossdk.io/core/context"
	"cosmossdk.io/core/server"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/runtime/v2"
	"cosmossdk.io/schema/appdata"
	"cosmossdk.io/server/v2/appmanager"
	"cosmossdk.io/server/v2/cometbft"
	"cosmossdk.io/server/v2/streaming"
	storev2 "cosmossdk.io/store/v2"
	consensustypes "cosmossdk.io/x/consensus/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simsx"
	simsxv2 "github.com/cosmos/cosmos-sdk/simsx/v2"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
)

type Tx = transaction.Tx
type (
	HasWeightedOperationsX              = simsx.HasWeightedOperationsX
	HasWeightedOperationsXWithProposals = simsx.HasWeightedOperationsXWithProposals
	HasProposalMsgsX                    = simsx.HasProposalMsgsX
	HasLegacyProposalMsgs               = simsx.HasLegacyProposalMsgs
)

const SimAppChainID = "simulation-app"

// DefaultSeeds list of seeds was imported from the original simulation runner: https://github.com/cosmos/tools/blob/v1.0.0/cmd/runsim/main.go#L32
var DefaultSeeds = []int64{
	1, 2, 4, 7,
	32, 123, 124, 582, 1893, 2989,
	3012, 4728, 37827, 981928, 87821, 891823782,
	989182, 89182391, 11, 22, 44, 77, 99, 2020,
	3232, 123123, 124124, 582582, 18931893,
	29892989, 30123012, 47284728, 7601778, 8090485,
	977367484, 491163361, 424254581, 673398983,
}

const (
	maxTimePerBlock   = 10_000 * time.Second
	minTimePerBlock   = maxTimePerBlock / 2
	timeRangePerBlock = maxTimePerBlock - minTimePerBlock
)

type (
	AuthKeeper interface {
		simsx.ModuleAccountSource
		simsx.AccountSource
	}

	BankKeeper interface {
		simsx.BalanceSource
		GetBlockedAddresses() map[string]bool
	}

	StakingKeeper interface {
		UnbondingTime(ctx context.Context) (time.Duration, error)
	}

	ModuleManager interface {
		Modules() map[string]appmodulev2.AppModule
		StoreKeys() map[string]string
	}

	// SimulationApp abstract blockchain app
	SimulationApp[T Tx] interface {
		appmanager.TransactionFuzzer[T]
		InitGenesis(
			ctx context.Context,
			blockRequest *server.BlockRequest[T],
			initGenesisJSON []byte,
			txDecoder transaction.Codec[T],
		) (*server.BlockResponse, store.WriterMap, error)

		GetApp() *runtime.App[T]
		TxConfig() client.TxConfig
		AppCodec() codec.Codec
		DefaultGenesis() map[string]json.RawMessage
		Store() storev2.RootStore
		Close() error
	}

	// TestInstance system under test
	TestInstance[T Tx] struct {
		RandSource    simsxv2.RandSource
		App           SimulationApp[T]
		TxDecoder     transaction.Codec[T]
		BankKeeper    BankKeeper
		AuthKeeper    AuthKeeper
		StakingKeeper StakingKeeper
		TXBuilder     simsxv2.TXBuilder[T]
		AppManager    appmanager.AppManager[T]
		ModuleManager ModuleManager
		StreamManager streaming.Manager
		StreamHook    *appdata.Listener
	}

	AppFactory[T Tx, V SimulationApp[T]] func(config depinject.Config, outputs ...any) (V, error)
	AppConfigFactory                     func() depinject.Config
)

// SetupTestInstance initializes and returns the system under test.
func SetupTestInstance[T Tx, V SimulationApp[T]](
	tb testing.TB,
	appFactory AppFactory[T, V],
	appConfigFactory AppConfigFactory,
	randSource simsxv2.RandSource,
	dbBackend string,
) TestInstance[T] {
	tb.Helper()
	vp := viper.New()
	vp.Set("store.app-db-backend", dbBackend)
	vp.Set("home", tb.TempDir())

	depInjCfg := depinject.Configs(
		depinject.Supply(log.NewNopLogger(), runtime.GlobalConfig(vp.AllSettings())),
		appConfigFactory(),
	)
	var (
		bankKeeper BankKeeper
		authKeeper AuthKeeper
		stKeeper   StakingKeeper
	)

	err := depinject.Inject(depInjCfg,
		&authKeeper,
		&bankKeeper,
		&stKeeper,
	)
	require.NoError(tb, err)

	xapp, err := appFactory(depinject.Configs(depinject.Supply(log.NewNopLogger(), runtime.GlobalConfig(vp.AllSettings()))))
	require.NoError(tb, err)
	return TestInstance[T]{
		RandSource:    randSource,
		App:           xapp,
		BankKeeper:    bankKeeper,
		AuthKeeper:    authKeeper,
		StakingKeeper: stKeeper,
		AppManager:    xapp.GetApp(),
		ModuleManager: xapp.GetApp().ModuleManager(),
		TxDecoder:     simsxv2.NewGenericTxDecoder[T](xapp.TxConfig()),
		TXBuilder:     simsxv2.NewSDKTXBuilder[T](xapp.TxConfig(), simsxv2.DefaultGenTxGas),
	}
}

// InitializeChain sets up the blockchain with an initial state, validator set, and history using the provided genesis data.
func (ti TestInstance[T]) InitializeChain(
	tb testing.TB,
	ctx context.Context,
	chainID string,
	genesisTimestamp time.Time,
	initialHeight uint64,
	genesisAppState json.RawMessage,
) ChainState[T] {
	tb.Helper()
	initRsp, stateRoot := doChainInitWithGenesis(
		tb,
		ctx,
		chainID,
		genesisTimestamp,
		initialHeight,
		genesisAppState,
		ti,
	)
	activeValidatorSet := simsxv2.NewValSet().Update(initRsp.ValidatorUpdates)
	valsetHistory := simsxv2.NewValSetHistory(initialHeight)
	valsetHistory.Add(genesisTimestamp, activeValidatorSet)
	return ChainState[T]{
		ChainID:            chainID,
		BlockTime:          genesisTimestamp,
		BlockHeight:        initialHeight,
		ActiveValidatorSet: activeValidatorSet,
		ValsetHistory:      valsetHistory,
		AppHash:            stateRoot,
	}
}

// RunWithSeeds runs a series of subtests for each of the given set of seeds for deterministic simulation testing.
func RunWithSeeds[T Tx, V SimulationApp[T]](
	t *testing.T,
	appFactory AppFactory[T, V],
	appConfigFactory AppConfigFactory,
	seeds []int64,
	postRunActions ...func(t testing.TB, cs ChainState[T], app TestInstance[T], accs []simtypes.Account),
) {
	t.Helper()
	cfg := cli.NewConfigFromFlags()
	cfg.ChainID = SimAppChainID
	for _, seed := range seeds {
		t.Run(fmt.Sprintf("seed: %d", seed), func(t *testing.T) {
			t.Parallel()
			RunWithSeed(t, appFactory, appConfigFactory, cfg, seed, postRunActions...)
		})
	}
}

// RunWithSeed initializes and executes a simulation run with the given seed, generating blocks and transactions.
func RunWithSeed[T Tx, V SimulationApp[T]](
	tb testing.TB,
	appFactory AppFactory[T, V],
	appConfigFactory AppConfigFactory,
	tCfg simtypes.Config,
	seed int64,
	postRunActions ...func(t testing.TB, cs ChainState[T], app TestInstance[T], accs []simtypes.Account),
) {
	tb.Helper()
	RunWithRandSource(tb, appFactory, appConfigFactory, tCfg, simsxv2.NewSeededRandSource(seed), postRunActions...)
}

// RunWithRandSource initializes and executes a simulation run with the given rand source, generating blocks and transactions.
func RunWithRandSource[T Tx, V SimulationApp[T]](
	tb testing.TB,
	appFactory AppFactory[T, V],
	appConfigFactory AppConfigFactory,
	tCfg simtypes.Config,
	randSource simsxv2.RandSource,
	postRunActions ...func(t testing.TB, cs ChainState[T], app TestInstance[T], accs []simtypes.Account),
) {
	tb.Helper()
	initialBlockHeight := tCfg.InitialBlockHeight
	require.NotEmpty(tb, initialBlockHeight, "initial block height must not be 0")

	setupFn := func(ctx context.Context, r *rand.Rand) (TestInstance[T], ChainState[T], []simtypes.Account) {
		testInstance := SetupTestInstance[T, V](tb, appFactory, appConfigFactory, randSource, tCfg.DBBackend)
		accounts, genesisAppState, chainID, genesisTimestamp := prepareInitialGenesisState(
			testInstance.App,
			r,
			testInstance.BankKeeper,
			tCfg,
			testInstance.ModuleManager,
		)
		cs := testInstance.InitializeChain(
			tb,
			ctx,
			chainID,
			genesisTimestamp,
			initialBlockHeight,
			genesisAppState,
		)

		return testInstance, cs, accounts
	}
	RunWithRandSourceX(tb, tCfg, setupFn, randSource, postRunActions...)
}

// RunWithRandSourceX entrypoint for custom chain setups.
// The function runs the full simulation test circle for the specified random source and setup function, followed by optional post-run actions.
// when tb implements ResetTimer, the method is called after setup, before jumping into the main loop
func RunWithRandSourceX[T Tx](
	tb testing.TB,
	tCfg simtypes.Config,
	setupChainStateFn func(ctx context.Context, r *rand.Rand) (TestInstance[T], ChainState[T], []simtypes.Account),
	randSource rand.Source,
	postRunActions ...func(t testing.TB, cs ChainState[T], app TestInstance[T], accs []simtypes.Account),
) {
	tb.Helper()
	r := rand.New(randSource)
	rootCtx, done := context.WithCancel(context.Background())
	defer done()

	testInstance, chainState, accounts := setupChainStateFn(rootCtx, r)
	customFactoryParams := make(map[string]json.RawMessage)
	if tCfg.ParamsFile != "" {
		bz, err := os.ReadFile(tCfg.ParamsFile)
		require.NoError(tb, err)
		require.NoError(tb, json.Unmarshal(bz, &customFactoryParams))
	}

	modules := testInstance.ModuleManager.Modules()
	msgFactoriesFn := prepareSimsMsgFactories(tb, r, modules, simsx.ParamWeightSource(customFactoryParams))

	if b, ok := tb.(interface{ ResetTimer() }); ok {
		b.ResetTimer()
	}

	doMainLoop(
		tb,
		rootCtx,
		testInstance,
		&chainState,
		msgFactoriesFn,
		r,
		tCfg,
		accounts,
	)

	for _, step := range postRunActions {
		step(tb, chainState, testInstance, accounts)
	}
	require.NoError(tb, testInstance.App.Close(), "closing app")
}

// prepareInitialGenesisState initializes the genesis state for simulation by generating accounts, app state, chain ID, and timestamp.
// It uses a random seed, configuration parameters, and module manager to customize the state.
// Blocked accounts are removed from the simulation accounts list based on the bank keeper's configuration.
func prepareInitialGenesisState[T Tx](
	app SimulationApp[T],
	r *rand.Rand,
	bankKeeper BankKeeper,
	tCfg simtypes.Config,
	moduleManager ModuleManager,
) ([]simtypes.Account, json.RawMessage, string, time.Time) {
	txConfig := app.TxConfig()
	appStateFn := simtestutil.AppStateFn(
		app.AppCodec(),
		txConfig.SigningContext().AddressCodec(),
		txConfig.SigningContext().ValidatorAddressCodec(),
		toLegacySimsModule(moduleManager.Modules()),
		app.DefaultGenesis(),
	)
	params := simulation.RandomParams(r)
	accounts := slices.DeleteFunc(simtypes.RandomAccounts(r, params.NumKeys()),
		func(acc simtypes.Account) bool { // remove blocked accounts
			return bankKeeper.GetBlockedAddresses()[acc.AddressBech32]
		})

	appState, accounts, chainID, genesisTimestamp := appStateFn(r, accounts, tCfg)
	return accounts, appState, chainID, genesisTimestamp
}

// doChainInitWithGenesis initializes the blockchain state with the provided genesis data and returns the initial block response and state root.
func doChainInitWithGenesis[T Tx](
	tb testing.TB,
	ctx context.Context,
	chainID string,
	genesisTimestamp time.Time,
	initialHeight uint64,
	genesisAppState json.RawMessage,
	testInstance TestInstance[T],
) (*server.BlockResponse, store.Hash) {
	tb.Helper()
	app := testInstance.App
	txDecoder := testInstance.TxDecoder
	appStore := testInstance.App.Store()
	genesisReq := &server.BlockRequest[T]{
		Height:    initialHeight,
		Time:      genesisTimestamp,
		Hash:      make([]byte, 32),
		ChainId:   chainID,
		AppHash:   make([]byte, 32),
		IsGenesis: true,
	}

	initialConsensusParams := &consensustypes.MsgUpdateParams{
		Block: &cmtproto.BlockParams{
			MaxBytes: 200000,
			MaxGas:   100_000_000,
		},
		Evidence: &cmtproto.EvidenceParams{
			MaxAgeNumBlocks: 302400,
			MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
			MaxBytes:        10000,
		},
		Validator: &cmtproto.ValidatorParams{PubKeyTypes: []string{cmttypes.ABCIPubKeyTypeEd25519, cmttypes.ABCIPubKeyTypeSecp256k1}},
	}
	genesisCtx := context.WithValue(ctx, corecontext.CometParamsInitInfoKey, initialConsensusParams)
	initRsp, genesisStateChanges, err := app.InitGenesis(genesisCtx, genesisReq, genesisAppState, txDecoder)
	require.NoError(tb, err)

	require.NoError(tb, appStore.SetInitialVersion(initialHeight-1))
	changeSet, err := genesisStateChanges.GetStateChanges()
	require.NoError(tb, err)

	stateRoot, err := appStore.Commit(&store.Changeset{Changes: changeSet, Version: initialHeight - 1})
	require.NoError(tb, err)
	return initRsp, stateRoot
}

// ChainState represents the state of a blockchain during a simulation run.
type ChainState[T Tx] struct {
	ChainID            string
	BlockTime          time.Time
	BlockHeight        uint64
	ActiveValidatorSet simsxv2.WeightedValidators
	ValsetHistory      *simsxv2.ValSetHistory
	AppHash            store.Hash
}

// doMainLoop executes the main simulation loop after chain setup with genesis block.
// Based on the initial seed and configurations, a deterministic set of messages is generated
// and executed. Events like validators missing votes or double signing are included in this
// process. The runtime tracks the validator's state and history.
func doMainLoop[T Tx](
	tb testing.TB,
	rootCtx context.Context,
	testInstance TestInstance[T],
	cs *ChainState[T],
	nextMsgFactory func() simsx.SimMsgFactoryX,
	r *rand.Rand,
	tCfg simtypes.Config,
	accounts []simtypes.Account,
) {
	tb.Helper()
	if len(cs.ActiveValidatorSet) == 0 {
		tb.Fatal("no active validators in chain setup")
		return
	}

	numBlocks := tCfg.NumBlocks
	maxTXPerBlock := tCfg.BlockSize

	var (
		txSkippedCounter int
		txTotalCounter   int
	)
	rootReporter := simsx.NewBasicSimulationReporter()
	futureOpsReg := simsxv2.NewFutureOpsRegistry()

	for end := cs.BlockHeight + numBlocks; cs.BlockHeight < end; cs.BlockHeight++ {
		if len(cs.ActiveValidatorSet) == 0 {
			tb.Skipf("run out of validators in block: %d\n", cs.BlockHeight)
			return
		}
		cs.BlockTime = cs.BlockTime.Add(minTimePerBlock).
			Add(time.Duration(int64(r.Intn(int(timeRangePerBlock/time.Second)))) * time.Second)
		cs.ValsetHistory.Add(cs.BlockTime, cs.ActiveValidatorSet)
		blockReqN := &server.BlockRequest[T]{
			Height:  cs.BlockHeight,
			Time:    cs.BlockTime,
			Hash:    cs.AppHash,
			AppHash: cs.AppHash,
			ChainId: cs.ChainID,
		}

		cometInfo := comet.Info{
			ValidatorsHash: nil,
			Evidence:       cs.ValsetHistory.MissBehaviour(r),
			// pick one of top 10
			ProposerAddress: cs.ActiveValidatorSet[r.Intn(min(len(cs.ActiveValidatorSet), 10))].Address,
			LastCommit:      cs.ActiveValidatorSet.NewCommitInfo(r),
		}
		fOps, pos := futureOpsReg.PopScheduledFor(cs.BlockTime), 0
		addressCodec := testInstance.App.TxConfig().SigningContext().AddressCodec()
		simsCtx := context.WithValue(rootCtx, corecontext.CometInfoKey, cometInfo) // required for ContextAwareCometInfoService
		resultHandlers := make([]simsx.SimDeliveryResultHandler, 0, maxTXPerBlock)
		var txPerBlockCounter int
		blockRsp, updates, err := testInstance.App.DeliverSims(simsCtx, blockReqN, func(ctx context.Context) iter.Seq[T] {
			return func(yield func(T) bool) {
				unbondingTime, err := testInstance.StakingKeeper.UnbondingTime(ctx)
				require.NoError(tb, err)
				cs.ValsetHistory.SetMaxHistory(minBlocksInUnbondingPeriod(unbondingTime))
				testData := simsx.NewChainDataSource(ctx, r, testInstance.AuthKeeper, testInstance.BankKeeper, addressCodec, accounts...)

				for txPerBlockCounter < maxTXPerBlock {
					txPerBlockCounter++
					mergedMsgFactory := func() simsx.SimMsgFactoryX {
						if pos < len(fOps) {
							pos++
							return fOps[pos-1]
						}
						return nextMsgFactory()
					}()
					reporter := rootReporter.WithScope(mergedMsgFactory.MsgType())
					if fx, ok := mergedMsgFactory.(simsx.HasFutureOpsRegistry); ok {
						fx.SetFutureOpsRegistry(futureOpsReg)
						continue
					}

					// the stf context is required to access state via keepers
					signers, msg := mergedMsgFactory.Create()(ctx, testData, reporter)
					if reporter.IsSkipped() {
						txSkippedCounter++
						require.NoError(tb, reporter.Close())
						continue
					}
					resultHandlers = append(resultHandlers, mergedMsgFactory.DeliveryResultHandler())
					reporter.Success(msg)
					require.NoError(tb, reporter.Close())

					tx, err := testInstance.TXBuilder.Build(ctx, testInstance.AuthKeeper, signers, msg, r, cs.ChainID)
					require.NoError(tb, err)
					blockReqN.Txs = append(blockReqN.Txs, tx)
					if !yield(tx) {
						return
					}
				}
			}
		})
		require.NoError(tb, err, "%d, %s", blockReqN.Height, blockReqN.Time)
		changeSet, err := updates.GetStateChanges()
		require.NoError(tb, err)
		cs.AppHash, err = testInstance.App.Store().Commit(&store.Changeset{
			Version: blockReqN.Height,
			Changes: changeSet,
		})

		require.NoError(tb, err)
		require.Equal(tb, len(resultHandlers), len(blockRsp.TxResults), "txPerBlockCounter: %d, totalSkipped: %d", txPerBlockCounter, txSkippedCounter)
		for i, v := range blockRsp.TxResults {
			require.NoError(tb, resultHandlers[i](v.Error))
		}
		txTotalCounter += txPerBlockCounter
		cs.ActiveValidatorSet = cs.ActiveValidatorSet.Update(blockRsp.ValidatorUpdates)

		if len(testInstance.StreamManager.Listeners) == 0 && testInstance.StreamHook == nil {
			continue
		}
		// stream data
		strmCtx, cancel := context.WithTimeout(rootCtx, time.Second)
		rawTxs := simsx.Collect(blockReqN.Txs, func(a T) []byte { return a.Bytes() })
		require.NoError(tb, cometbft.StreamOut[T](
			strmCtx,
			int64(blockReqN.Height),
			rawTxs,
			blockReqN.Txs,
			*blockRsp,
			changeSet,
			testInstance.StreamManager,
			testInstance.StreamHook,
			true,
			tb.Logf,
		))
		cancel()
	}
	fmt.Println("+++ reporter:\n" + rootReporter.Summary().String())
	fmt.Printf("Tx total: %d skipped: %d\n", txTotalCounter, txSkippedCounter)
}

// prepareSimsMsgFactories constructs and returns a function to retrieve simulation message factories for all modules.
// It initializes proposal and factory registries, registers proposals and weighted operations, and sorts deterministically.
func prepareSimsMsgFactories(tb testing.TB, r *rand.Rand, modules map[string]appmodulev2.AppModule, weights simsx.WeightSource) func() simsx.SimMsgFactoryX {
	tb.Helper()
	moduleNames := slices.Collect(maps.Keys(modules))
	slices.Sort(moduleNames) // make deterministic

	// get all proposal types
	proposalRegistry := simsx.NewUniqueTypeRegistry()
	for _, n := range moduleNames {
		switch xm := modules[n].(type) {
		case HasProposalMsgsX:
			xm.ProposalMsgsX(weights, proposalRegistry)
		case HasLegacyProposalMsgs:
			tb.Logf("Ignoring legacy proposal messages for module: %s", n)
		}
	}
	// register all msg factories
	factoryRegistry := simsx.NewUnorderedRegistry()
	for _, n := range moduleNames {
		switch xm := modules[n].(type) {
		case HasWeightedOperationsX:
			xm.WeightedOperationsX(weights, factoryRegistry)
		case HasWeightedOperationsXWithProposals:
			xm.WeightedOperationsX(weights, factoryRegistry, proposalRegistry.Iterator(), nil)
		}
	}
	return simsxv2.NextFactoryFn(factoryRegistry.Elements(), r)
}

func toLegacySimsModule(modules map[string]appmodule.AppModule) []module.AppModuleSimulation {
	r := make([]module.AppModuleSimulation, 0, len(modules))
	names := slices.Collect(maps.Keys(modules))
	slices.Sort(names) // make deterministic
	for _, v := range names {
		if m, ok := modules[v].(module.AppModuleSimulation); ok {
			r = append(r, m)
		}
	}
	return r
}

func minBlocksInUnbondingPeriod(unbondingTime time.Duration) int {
	maxblocks := unbondingTime / maxTimePerBlock
	return max(int(maxblocks)-1, 1)
}
