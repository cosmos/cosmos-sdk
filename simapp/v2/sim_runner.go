package simapp

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
	"iter"

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
	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/server/v2/appmanager"
	cometbfttypes "cosmossdk.io/server/v2/cometbft/types"
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

	SimulationApp[T Tx] interface {
		GetApp() *runtime.App[T]
		ModuleManager() *runtime.MM[T]
		TxConfig() client.TxConfig
		AppCodec() codec.Codec
		DefaultGenesis() map[string]json.RawMessage
		Store() storev2.RootStore
		Close() error
	}

	StakingKeeper interface {
		UnbondingTime(ctx context.Context) (time.Duration, error)
	}

	ModuleManager interface {
		Modules() map[string]appmodulev2.AppModule
	}

	TestInstance[T Tx] struct {
		App           SimulationApp[T]
		TxDecoder     transaction.Codec[T]
		BankKeeper    BankKeeper
		AuthKeeper    AuthKeeper
		StakingKeeper StakingKeeper
		TXBuilder     simsxv2.TXBuilder[T]
		AppManager    appmanager.AppManager[T]
		ModuleManager ModuleManager
	}

	AppFactory[T Tx, V SimulationApp[T]] func(config depinject.Config, outputs ...any) (V, error)
)

func SetupTestInstance[T Tx, V SimulationApp[T]](t *testing.T, factory AppFactory[T, V], appConfig depinject.Config) TestInstance[T] {
	nodeHome := t.TempDir()
	currentDir, err := os.Getwd()
	require.NoError(t, err)
	configPath := filepath.Join(currentDir, "testdata")
	v, err := serverv2.ReadConfig(configPath)
	require.NoError(t, err)
	v.Set("home", nodeHome)
	v.Set("store.app-db-backend", "memdb")

	depInjCfg := depinject.Configs(
		depinject.Supply(log.NewNopLogger(), runtime.GlobalConfig(v.AllSettings())),
		appConfig,
	)
	var (
		bankKeeper BankKeeper
		authKeeper AuthKeeper
		stKeeper   StakingKeeper
	)

	err = depinject.Inject(depInjCfg,
		&authKeeper,
		&bankKeeper,
		&stKeeper,
	)
	require.NoError(t, err)

	xapp, err := factory(depinject.Configs(depinject.Supply(log.NewNopLogger(), runtime.GlobalConfig(v.AllSettings()))))
	require.NoError(t, err)
	return TestInstance[T]{
		App:           xapp,
		BankKeeper:    bankKeeper,
		AuthKeeper:    authKeeper,
		StakingKeeper: stKeeper,
		AppManager:    xapp.GetApp().AppManager,
		ModuleManager: xapp.ModuleManager(),
		TxDecoder:     simsxv2.NewGenericTxDecoder[T](xapp.TxConfig()),
		TXBuilder:     simsxv2.NewSDKTXBuilder[T](xapp.TxConfig(), simsxv2.DefaultGenTxGas),
	}
}

func RunWithSeeds[T Tx](t *testing.T, seeds []int64) {
	t.Helper()
	cfg := cli.NewConfigFromFlags()
	cfg.ChainID = SimAppChainID
	for i := range seeds {
		seed := seeds[i]
		t.Run(fmt.Sprintf("seed: %d", seed), func(t *testing.T) {
			t.Parallel()
			RunWithSeed(t, NewSimApp[T], AppConfig(), cfg, seed)
		})
	}
}

func RunWithSeed[T Tx, V SimulationApp[T]](t *testing.T, appFactory AppFactory[T, V], appConfig depinject.Config, tCfg simtypes.Config, seed int64) {
	r := rand.New(rand.NewSource(seed))
	testInstance := SetupTestInstance[T, V](t, appFactory, appConfig)
	accounts, genesisAppState, chainID, genesisTimestamp := prepareInitialGenesisState(testInstance.App, r, testInstance.BankKeeper, tCfg, testInstance.ModuleManager)

	appManager := testInstance.AppManager
	appStore := testInstance.App.Store()
	txConfig := testInstance.App.TxConfig()
	rootCtx, done := context.WithCancel(context.Background())
	defer done()
	initRsp, stateRoot := doChainInitWithGenesis(t, rootCtx, chainID, genesisTimestamp, appManager, testInstance.TxDecoder, genesisAppState, appStore)
	activeValidatorSet := simsxv2.NewValSet().Update(initRsp.ValidatorUpdates)
	valsetHistory := simsxv2.NewValSetHistory(1)
	valsetHistory.Add(genesisTimestamp, activeValidatorSet)

	emptySimParams := make(map[string]json.RawMessage) // todo read sims params from disk as before

	modules := testInstance.ModuleManager.Modules()
	msgFactoriesFn := prepareSimsMsgFactories(r, modules, simsx.ParamWeightSource(emptySimParams))

	cs := chainState[T]{
		chainID:            chainID,
		blockTime:          genesisTimestamp,
		activeValidatorSet: activeValidatorSet,
		valsetHistory:      valsetHistory,
		stateRoot:          stateRoot,
		app:                appManager,
		appStore:           appStore,
		txConfig:           txConfig,
	}
	doMainLoop(
		t,
		rootCtx,
		cs,
		msgFactoriesFn,
		r,
		testInstance.AuthKeeper,
		testInstance.BankKeeper,
		accounts,
		testInstance.TXBuilder,
		testInstance.StakingKeeper,
	)
	require.NoError(t, testInstance.App.Close(), "closing app")
}

func prepareInitialGenesisState[T Tx](
	app SimulationApp[T],
	r *rand.Rand,
	bankKeeper BankKeeper,
	tCfg simtypes.Config,
	moduleManager ModuleManager,
) ([]simtypes.Account, json.RawMessage, string, time.Time) {
	txConfig := app.TxConfig()
	// todo: replace legacy testdata functions ?
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

func doChainInitWithGenesis[T Tx](
	t *testing.T,
	ctx context.Context,
	chainID string,
	genesisTimestamp time.Time,
	app appmanager.AppManager[T],
	txDecoder transaction.Codec[T],
	genesisAppState json.RawMessage,
	appStore cometbfttypes.Store,
) (*server.BlockResponse, store.Hash) {
	genesisReq := &server.BlockRequest[T]{
		Height:    0,
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
	require.NoError(t, err)

	require.NoError(t, appStore.SetInitialVersion(0))
	changeSet, err := genesisStateChanges.GetStateChanges()
	require.NoError(t, err)

	stateRoot, err := appStore.Commit(&store.Changeset{Changes: changeSet})
	require.NoError(t, err)
	return initRsp, stateRoot
}

type chainState[T Tx] struct {
	chainID            string
	blockTime          time.Time
	activeValidatorSet simsxv2.WeightedValidators
	valsetHistory      *simsxv2.ValSetHistory
	stateRoot          store.Hash
	app                appmanager.AppManager[T]
	appStore           storev2.RootStore
	txConfig           client.TxConfig
}

func doMainLoop[T Tx](
	t *testing.T,
	rootCtx context.Context,
	cs chainState[T],
	nextMsgFactory func() simsx.SimMsgFactoryX,
	r *rand.Rand,
	authKeeper AuthKeeper,
	bankKeeper simsx.BalanceSource,
	accounts []simtypes.Account,
	txBuilder simsxv2.TXBuilder[T],
	stakingKeeper StakingKeeper,
) {
	blockTime := cs.blockTime
	activeValidatorSet := cs.activeValidatorSet
	if len(activeValidatorSet) == 0 {
		t.Fatal("no active validators in chain setup")
		return
	}
	valsetHistory := cs.valsetHistory
	stateRoot := cs.stateRoot
	chainID := cs.chainID
	app := cs.app
	appStore := cs.appStore

	const ( // todo: read from CLI instead
		numBlocks     = 100 // 500 default
		maxTXPerBlock = 200 // 200 default
	)

	var (
		txSkippedCounter int
		txTotalCounter   int
	)
	rootReporter := simsx.NewBasicSimulationReporter()
	futureOpsReg := simsxv2.NewFutureOpsRegistry()

	for i := 0; i < numBlocks; i++ {
		if len(activeValidatorSet) == 0 {
			t.Skipf("run out of validators in block: %d\n", i+1)
			return
		}
		blockTime = blockTime.Add(minTimePerBlock)
		blockTime = blockTime.Add(time.Duration(int64(r.Intn(int(timeRangePerBlock/time.Second)))) * time.Second)
		valsetHistory.Add(blockTime, activeValidatorSet)
		blockReqN := &server.BlockRequest[T]{
			Height:  uint64(1 + i),
			Time:    blockTime,
			Hash:    stateRoot,
			AppHash: stateRoot,
			ChainId: chainID,
		}

		cometInfo := comet.Info{
			ValidatorsHash:  nil,
			Evidence:        valsetHistory.MissBehaviour(r),
			ProposerAddress: activeValidatorSet[0].Address,
			LastCommit:      activeValidatorSet.NewCommitInfo(r),
		}
		fOps, pos := futureOpsReg.PopScheduledFor(blockTime), 0
		addressCodec := cs.txConfig.SigningContext().AddressCodec()
		simsCtx := context.WithValue(rootCtx, corecontext.CometInfoKey, cometInfo) // required for ContextAwareCometInfoService
		resultHandlers := make([]simsx.SimDeliveryResultHandler, 0, maxTXPerBlock)
		var txPerBlockCounter int
		blockRsp, updates, err := app.DeliverSims(simsCtx, blockReqN, func(ctx context.Context) iter.Seq[T] {
			return func(yield func(T) bool) {
				unbondingTime, err := stakingKeeper.UnbondingTime(ctx)
				require.NoError(t, err)
				valsetHistory.SetMaxHistory(minBlocksInUnbondingPeriod(unbondingTime))
				testData := simsx.NewChainDataSource(ctx, r, authKeeper, bankKeeper, addressCodec, accounts...)

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
						require.NoError(t, reporter.Close())
						continue
					}
					resultHandlers = append(resultHandlers, mergedMsgFactory.DeliveryResultHandler())
					reporter.Success(msg)
					require.NoError(t, reporter.Close())

					tx, err := txBuilder.Build(ctx, authKeeper, signers, msg, r, chainID)
					require.NoError(t, err)
					if !yield(tx) {
						return
					}
				}
			}
		})
		require.NoError(t, err, "%d, %s", blockReqN.Height, blockReqN.Time)
		changeSet, err := updates.GetStateChanges()
		require.NoError(t, err)
		stateRoot, err = appStore.Commit(&store.Changeset{
			Version: blockReqN.Height,
			Changes: changeSet,
		})

		require.NoError(t, err)
		require.Equal(t, len(resultHandlers), len(blockRsp.TxResults), "txPerBlockCounter: %d, totalSkipped: %d", txPerBlockCounter, txSkippedCounter)
		for i, v := range blockRsp.TxResults {
			require.NoError(t, resultHandlers[i](v.Error))
		}
		txTotalCounter += txPerBlockCounter
		var removed int
		for _, v := range blockRsp.ValidatorUpdates {
			if v.Power == 0 {
				removed++
			}
		}
		activeValidatorSet = activeValidatorSet.Update(blockRsp.ValidatorUpdates)
		//fmt.Printf("active validator set after height %d: %d, %s, %X\n", blockReqN.Height, len(activeValidatorSet), blockReqN.Time, stateRoot)
	}
	fmt.Println("+++ reporter:\n" + rootReporter.Summary().String())
	fmt.Printf("Tx total: %d skipped: %d\n", txTotalCounter, txSkippedCounter)
}

func prepareSimsMsgFactories(
	r *rand.Rand,
	modules map[string]appmodulev2.AppModule,
	weights simsx.WeightSource,
) func() simsx.SimMsgFactoryX {
	moduleNames := maps.Keys(modules)
	slices.Sort(moduleNames) // make deterministic

	// get all proposal types
	proposalRegistry := simsx.NewUniqueTypeRegistry()
	for _, n := range moduleNames {
		switch xm := modules[n].(type) {
		case HasProposalMsgsX:
			xm.ProposalMsgsX(weights, proposalRegistry)
			// todo: register legacy and v1 msg proposals
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
	names := maps.Keys(modules)
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
