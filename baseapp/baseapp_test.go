package baseapp_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/store/metrics"
	pruningtypes "cosmossdk.io/store/pruning/types"
	"cosmossdk.io/store/rootmulti"
	"cosmossdk.io/store/snapshots"
	snapshottypes "cosmossdk.io/store/snapshots/types"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
)

var (
	capKey1 = storetypes.NewKVStoreKey("key1")
	capKey2 = storetypes.NewKVStoreKey("key2")

	// testTxPriority is the CheckTx priority that we set in the test
	// AnteHandler.
	testTxPriority = int64(42)
)

type (
	BaseAppSuite struct {
		baseApp  *baseapp.BaseApp
		cdc      *codec.ProtoCodec
		txConfig client.TxConfig
	}

	SnapshotsConfig struct {
		blocks             uint64
		blockTxs           int
		snapshotInterval   uint64
		snapshotKeepRecent uint32
		pruningOpts        pruningtypes.PruningOptions
	}
)

func NewBaseAppSuite(t *testing.T, opts ...func(*baseapp.BaseApp)) *BaseAppSuite {
	t.Helper()
	cdc := codectestutil.CodecOptions{}.NewCodec()
	baseapptestutil.RegisterInterfaces(cdc.InterfaceRegistry())

	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)
	db := dbm.NewMemDB()

	app := baseapp.NewBaseApp(t.Name(), log.NewTestLogger(t), db, txConfig.TxDecoder(), opts...)
	require.Equal(t, t.Name(), app.Name())

	app.SetInterfaceRegistry(cdc.InterfaceRegistry())
	app.MsgServiceRouter().SetInterfaceRegistry(cdc.InterfaceRegistry())
	app.MountStores(capKey1, capKey2)
	app.SetParamStore(paramStore{db: dbm.NewMemDB()})
	app.SetTxDecoder(txConfig.TxDecoder())
	app.SetTxEncoder(txConfig.TxEncoder())

	// mount stores and seal
	require.Nil(t, app.LoadLatestVersion())

	return &BaseAppSuite{
		baseApp:  app,
		cdc:      cdc,
		txConfig: txConfig,
	}
}

func getQueryBaseapp(t *testing.T) *baseapp.BaseApp {
	t.Helper()

	db := dbm.NewMemDB()
	name := t.Name()
	app := baseapp.NewBaseApp(name, log.NewTestLogger(t), db, nil)

	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 2})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)

	return app
}

func NewBaseAppSuiteWithSnapshots(t *testing.T, cfg SnapshotsConfig, opts ...func(*baseapp.BaseApp)) *BaseAppSuite {
	t.Helper()
	snapshotTimeout := 1 * time.Minute
	snapshotStore, err := snapshots.NewStore(dbm.NewMemDB(), testutil.GetTempDir(t))
	require.NoError(t, err)

	suite := NewBaseAppSuite(
		t,
		append(
			opts,
			baseapp.SetSnapshot(snapshotStore, snapshottypes.NewSnapshotOptions(cfg.snapshotInterval, cfg.snapshotKeepRecent)),
			baseapp.SetPruning(cfg.pruningOpts),
		)...,
	)

	baseapptestutil.RegisterKeyValueServer(suite.baseApp.MsgServiceRouter(), MsgKeyValueImpl{})

	_, err = suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})
	require.NoError(t, err)

	r := rand.New(rand.NewSource(3920758213583))
	keyCounter := 0

	for height := int64(1); height <= int64(cfg.blocks); height++ {

		_, _, addr := testdata.KeyTestPubAddr()
		txs := [][]byte{}
		for txNum := 0; txNum < cfg.blockTxs; txNum++ {
			msgs := []sdk.Msg{}
			for msgNum := 0; msgNum < 100; msgNum++ {
				key := []byte(fmt.Sprintf("%v", keyCounter))
				value := make([]byte, 10000)

				_, err := r.Read(value)
				require.NoError(t, err)

				msgs = append(msgs, &baseapptestutil.MsgKeyValue{Key: key, Value: value, Signer: addr.String()})
				keyCounter++
			}

			builder := suite.txConfig.NewTxBuilder()
			err := builder.SetMsgs(msgs...)
			require.NoError(t, err)
			setTxSignature(t, builder, 0)

			txBytes, err := suite.txConfig.TxEncoder()(builder.GetTx())
			require.NoError(t, err)

			txs = append(txs, txBytes)
		}

		_, err := suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: height,
			Txs:    txs,
		})
		require.NoError(t, err)

		_, err = suite.baseApp.Commit()
		require.NoError(t, err)

		// wait for snapshot to be taken, since it happens asynchronously
		if cfg.snapshotInterval > 0 && uint64(height)%cfg.snapshotInterval == 0 {
			start := time.Now()
			for {
				if time.Since(start) > snapshotTimeout {
					t.Errorf("timed out waiting for snapshot after %v", snapshotTimeout)
				}

				snapshot, err := snapshotStore.Get(uint64(height), snapshottypes.CurrentFormat)
				require.NoError(t, err)

				if snapshot != nil {
					break
				}

				time.Sleep(100 * time.Millisecond)
			}
		}
	}

	return suite
}

func TestLoadVersion(t *testing.T) {
	logger := log.NewTestLogger(t)
	pruningOpt := baseapp.SetPruning(pruningtypes.NewPruningOptions(pruningtypes.PruningNothing))
	db := dbm.NewMemDB()
	name := t.Name()
	app := baseapp.NewBaseApp(name, logger, db, nil, pruningOpt)

	// make a cap key and mount the store
	err := app.LoadLatestVersion() // needed to make stores non-nil
	require.Nil(t, err)

	emptyCommitID := storetypes.CommitID{}

	// fresh store has zero/empty last commit
	lastHeight := app.LastBlockHeight()
	lastID := app.LastCommitID()
	require.Equal(t, int64(0), lastHeight)
	require.Equal(t, emptyCommitID, lastID)

	// execute a block, collect commit ID
	res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.NoError(t, err)
	commitID1 := storetypes.CommitID{Version: 1, Hash: res.AppHash}
	_, err = app.Commit()
	require.NoError(t, err)

	// execute a block, collect commit ID
	res, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 2})
	require.NoError(t, err)
	commitID2 := storetypes.CommitID{Version: 2, Hash: res.AppHash}
	_, err = app.Commit()
	require.NoError(t, err)

	// reload with LoadLatestVersion
	app = baseapp.NewBaseApp(name, logger, db, nil, pruningOpt)
	app.MountStores()

	err = app.LoadLatestVersion()
	require.Nil(t, err)

	testLoadVersionHelper(t, app, int64(2), commitID2)

	// Reload with LoadVersion, see if you can commit the same block and get
	// the same result.
	app = baseapp.NewBaseApp(name, logger, db, nil, pruningOpt)
	err = app.LoadVersion(1)
	require.Nil(t, err)

	testLoadVersionHelper(t, app, int64(1), commitID1)

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 2})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)

	testLoadVersionHelper(t, app, int64(2), commitID2)
}

func TestSetLoader(t *testing.T) {
	useDefaultLoader := func(app *baseapp.BaseApp) {
		app.SetStoreLoader(baseapp.DefaultStoreLoader)
	}

	initStore := func(t *testing.T, db dbm.DB, storeKey string, k, v []byte) {
		t.Helper()
		rs := rootmulti.NewStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
		rs.SetPruning(pruningtypes.NewPruningOptions(pruningtypes.PruningNothing))

		key := storetypes.NewKVStoreKey(storeKey)
		rs.MountStoreWithDB(key, storetypes.StoreTypeIAVL, nil)

		err := rs.LoadLatestVersion()
		require.Nil(t, err)
		require.Equal(t, int64(0), rs.LastCommitID().Version)

		// write some data in substore
		kv, _ := rs.GetStore(key).(storetypes.KVStore)
		require.NotNil(t, kv)
		kv.Set(k, v)

		commitID := rs.Commit()
		require.Equal(t, int64(1), commitID.Version)
	}

	checkStore := func(t *testing.T, db dbm.DB, ver int64, storeKey string, k, v []byte) {
		t.Helper()
		rs := rootmulti.NewStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
		rs.SetPruning(pruningtypes.NewPruningOptions(pruningtypes.PruningDefault))

		key := storetypes.NewKVStoreKey(storeKey)
		rs.MountStoreWithDB(key, storetypes.StoreTypeIAVL, nil)

		err := rs.LoadLatestVersion()
		require.Nil(t, err)
		require.Equal(t, ver, rs.LastCommitID().Version)

		// query data in substore
		kv, _ := rs.GetStore(key).(storetypes.KVStore)
		require.NotNil(t, kv)
		require.Equal(t, v, kv.Get(k))
	}

	testCases := map[string]struct {
		setLoader    func(*baseapp.BaseApp)
		origStoreKey string
		loadStoreKey string
	}{
		"don't set loader": {
			origStoreKey: "foo",
			loadStoreKey: "foo",
		},
		"default loader": {
			setLoader:    useDefaultLoader,
			origStoreKey: "foo",
			loadStoreKey: "foo",
		},
	}

	k := []byte("key")
	v := []byte("value")

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// prepare a db with some data
			db := dbm.NewMemDB()
			initStore(t, db, tc.origStoreKey, k, v)

			// load the app with the existing db
			opts := []func(*baseapp.BaseApp){baseapp.SetPruning(pruningtypes.NewPruningOptions(pruningtypes.PruningNothing))}
			if tc.setLoader != nil {
				opts = append(opts, tc.setLoader)
			}
			app := baseapp.NewBaseApp(t.Name(), log.NewTestLogger(t), db, nil, opts...)
			app.MountStores(storetypes.NewKVStoreKey(tc.loadStoreKey))
			err := app.LoadLatestVersion()
			require.Nil(t, err)

			// "execute" one block
			res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 2})
			require.NoError(t, err)
			require.NotNil(t, res.AppHash)
			_, err = app.Commit()
			require.NoError(t, err)

			// check db is properly updated
			checkStore(t, db, 2, tc.loadStoreKey, k, v)
			checkStore(t, db, 2, tc.loadStoreKey, []byte("foo"), nil)
		})
	}
}

func TestVersionSetterGetter(t *testing.T) {
	pruningOpt := baseapp.SetPruning(pruningtypes.NewPruningOptions(pruningtypes.PruningDefault))
	db := dbm.NewMemDB()
	name := t.Name()
	app := baseapp.NewBaseApp(name, log.NewTestLogger(t), db, nil, pruningOpt)

	require.Equal(t, "", app.Version())
	res, err := app.Query(context.TODO(), &abci.RequestQuery{Path: "app/version"})
	require.NoError(t, err)
	require.True(t, res.IsOK())
	require.Equal(t, "", string(res.Value))

	versionString := "1.0.0"
	app.SetVersion(versionString)
	require.Equal(t, versionString, app.Version())

	res, err = app.Query(context.TODO(), &abci.RequestQuery{Path: "app/version"})
	require.NoError(t, err)
	require.True(t, res.IsOK())
	require.Equal(t, versionString, string(res.Value))
}

func TestLoadVersionInvalid(t *testing.T) {
	logger := log.NewNopLogger()
	pruningOpt := baseapp.SetPruning(pruningtypes.NewPruningOptions(pruningtypes.PruningNothing))
	db := dbm.NewMemDB()
	name := t.Name()
	app := baseapp.NewBaseApp(name, logger, db, nil, pruningOpt)

	err := app.LoadLatestVersion()
	require.Nil(t, err)

	// require error when loading an invalid version
	err = app.LoadVersion(-1)
	require.Error(t, err)

	res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.NoError(t, err)
	commitID1 := storetypes.CommitID{Version: 1, Hash: res.AppHash}
	_, err = app.Commit()
	require.NoError(t, err)

	// create a new app with the stores mounted under the same cap key
	app = baseapp.NewBaseApp(name, logger, db, nil, pruningOpt)

	// require we can load the latest version
	err = app.LoadVersion(1)
	require.Nil(t, err)
	testLoadVersionHelper(t, app, int64(1), commitID1)

	// require error when loading an invalid version
	err = app.LoadVersion(2)
	require.Error(t, err)
}

func TestOptionFunction(t *testing.T) {
	testChangeNameHelper := func(name string) func(*baseapp.BaseApp) {
		return func(bap *baseapp.BaseApp) {
			bap.SetName(name)
		}
	}

	db := dbm.NewMemDB()
	bap := baseapp.NewBaseApp("starting name", log.NewTestLogger(t), db, nil, testChangeNameHelper("new name"))
	require.Equal(t, bap.Name(), "new name", "BaseApp should have had name changed via option function")
}

func TestBaseAppOptionSeal(t *testing.T) {
	suite := NewBaseAppSuite(t)

	require.Panics(t, func() {
		suite.baseApp.SetName("")
	})
	require.Panics(t, func() {
		suite.baseApp.SetVersion("")
	})
	require.Panics(t, func() {
		suite.baseApp.SetDB(nil)
	})
	require.Panics(t, func() {
		suite.baseApp.SetCMS(nil)
	})
	require.Panics(t, func() {
		suite.baseApp.SetInitChainer(nil)
	})
	require.Panics(t, func() {
		suite.baseApp.SetPreBlocker(nil)
	})
	require.Panics(t, func() {
		suite.baseApp.SetBeginBlocker(nil)
	})
	require.Panics(t, func() {
		suite.baseApp.SetEndBlocker(nil)
	})
	require.Panics(t, func() {
		suite.baseApp.SetPrepareCheckStater(nil)
	})
	require.Panics(t, func() {
		suite.baseApp.SetPrecommiter(nil)
	})
	require.Panics(t, func() {
		suite.baseApp.SetAnteHandler(nil)
	})
	require.Panics(t, func() {
		suite.baseApp.SetAddrPeerFilter(nil)
	})
	require.Panics(t, func() {
		suite.baseApp.SetIDPeerFilter(nil)
	})
	require.Panics(t, func() {
		suite.baseApp.SetFauxMerkleMode()
	})
}

func TestTxDecoder(t *testing.T) {
	cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
	baseapptestutil.RegisterInterfaces(cdc.InterfaceRegistry())

	// patch in TxConfig instead of using an output from x/auth/tx
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)

	tx := newTxCounter(t, txConfig, 1, 0)
	txBytes, err := txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	dTx, err := txConfig.TxDecoder()(txBytes)
	require.NoError(t, err)

	counter, _ := parseTxMemo(t, tx)
	dTxCounter, _ := parseTxMemo(t, dTx)
	require.Equal(t, counter, dTxCounter)
}

func TestCustomRunTxPanicHandler(t *testing.T) {
	customPanicMsg := "test panic"
	anteErr := errorsmod.Register("fakeModule", 100500, "fakeError")
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			panic(errorsmod.Wrap(anteErr, "anteHandler"))
		})
	}

	suite := NewBaseAppSuite(t, anteOpt)
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), NoopCounterServerImpl{})

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})
	require.NoError(t, err)

	suite.baseApp.AddRunTxRecoveryHandler(func(recoveryObj interface{}) error {
		err, ok := recoveryObj.(error)
		if !ok {
			return nil
		}

		if anteErr.Is(err) {
			panic(customPanicMsg)
		} else {
			return nil
		}
	})

	// transaction should panic with custom handler above
	{
		tx := newTxCounter(t, suite.txConfig, 0, 0)

		require.PanicsWithValue(t, customPanicMsg, func() {
			bz, err := suite.txConfig.TxEncoder()(tx)
			require.NoError(t, err)
			_, err = suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1, Txs: [][]byte{bz}})
			require.NoError(t, err)
		})
	}
}

func TestBaseAppAnteHandler(t *testing.T) {
	anteKey := []byte("ante-key")
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, anteKey))
	}
	suite := NewBaseAppSuite(t, anteOpt)

	deliverKey := []byte("deliver-key")
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), CounterServerImpl{t, capKey1, deliverKey})

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})
	require.NoError(t, err)

	// execute a tx that will fail ante handler execution
	//
	// NOTE: State should not be mutated here. This will be implicitly checked by
	// the next txs ante handler execution (anteHandlerTxTest).
	tx := newTxCounter(t, suite.txConfig, 0, 0)
	tx = setFailOnAnte(t, suite.txConfig, tx, true)

	txBytes, err := suite.txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	res, err := suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1, Txs: [][]byte{txBytes}})
	require.NoError(t, err)
	require.Empty(t, res.Events)
	require.False(t, res.TxResults[0].IsOK(), fmt.Sprintf("%v", res))

	ctx := getFinalizeBlockStateCtx(suite.baseApp)
	store := ctx.KVStore(capKey1)
	require.Equal(t, int64(0), getIntFromStore(t, store, anteKey))

	// execute at tx that will pass the ante handler (the checkTx state should
	// mutate) but will fail the message handler
	tx = newTxCounter(t, suite.txConfig, 0, 0)
	tx = setFailOnHandler(t, suite.txConfig, tx, true)

	txBytes, err = suite.txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	res, err = suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1, Txs: [][]byte{txBytes}})
	require.NoError(t, err)
	require.Empty(t, res.Events)
	require.False(t, res.TxResults[0].IsOK(), fmt.Sprintf("%v", res))

	ctx = getFinalizeBlockStateCtx(suite.baseApp)
	store = ctx.KVStore(capKey1)
	require.Equal(t, int64(1), getIntFromStore(t, store, anteKey))
	require.Equal(t, int64(0), getIntFromStore(t, store, deliverKey))

	// Execute a successful ante handler and message execution where state is
	// implicitly checked by previous tx executions.
	tx = newTxCounter(t, suite.txConfig, 1, 0)

	txBytes, err = suite.txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	res, err = suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1, Txs: [][]byte{txBytes}})
	require.NoError(t, err)
	require.NotEmpty(t, res.TxResults[0].Events)
	require.True(t, res.TxResults[0].IsOK(), fmt.Sprintf("%v", res))

	ctx = getFinalizeBlockStateCtx(suite.baseApp)
	store = ctx.KVStore(capKey1)
	require.Equal(t, int64(2), getIntFromStore(t, store, anteKey))
	require.Equal(t, int64(1), getIntFromStore(t, store, deliverKey))

	_, err = suite.baseApp.Commit()
	require.NoError(t, err)
}

// Test and ensure that invalid block heights always cause errors.
// See issues:
// - https://github.com/cosmos/cosmos-sdk/issues/11220
// - https://github.com/cosmos/cosmos-sdk/issues/7662
func TestABCI_CreateQueryContext(t *testing.T) {
	t.Parallel()

	db := dbm.NewMemDB()
	name := t.Name()
	app := baseapp.NewBaseApp(name, log.NewTestLogger(t), db, nil)

	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 2})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)
	testCases := []struct {
		name   string
		height int64
		prove  bool
		expErr bool
	}{
		{"valid height", 2, true, false},
		{"future height", 10, true, true},
		{"negative height, prove=true", -1, true, true},
		{"negative height, prove=false", -1, false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := app.CreateQueryContext(tc.height, tc.prove)
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSetMinGasPrices(t *testing.T) {
	minGasPrices := sdk.DecCoins{sdk.NewInt64DecCoin("stake", 5000)}
	suite := NewBaseAppSuite(t, baseapp.SetMinGasPrices(minGasPrices.String()))

	ctx := getCheckStateCtx(suite.baseApp)
	require.Equal(t, minGasPrices, ctx.MinGasPrices())
}

type ctxType string

const (
	QueryCtx   ctxType = "query"
	CheckTxCtx ctxType = "checkTx"
)

var ctxTypes = []ctxType{QueryCtx, CheckTxCtx}

func (c ctxType) GetCtx(t *testing.T, bapp *baseapp.BaseApp) sdk.Context {
	t.Helper()
	if c == QueryCtx {
		ctx, err := bapp.CreateQueryContext(1, false)
		require.NoError(t, err)
		return ctx
	} else if c == CheckTxCtx {
		return getCheckStateCtx(bapp)
	}
	// TODO: Not supported yet
	return getFinalizeBlockStateCtx(bapp)
}

func TestQueryGasLimit(t *testing.T) {
	testCases := []struct {
		queryGasLimit   uint64
		gasActuallyUsed uint64
		shouldQueryErr  bool
	}{
		{queryGasLimit: 100, gasActuallyUsed: 50, shouldQueryErr: false},  // Valid case
		{queryGasLimit: 100, gasActuallyUsed: 150, shouldQueryErr: true},  // gasActuallyUsed > queryGasLimit
		{queryGasLimit: 0, gasActuallyUsed: 50, shouldQueryErr: false},    // fuzzing with queryGasLimit = 0
		{queryGasLimit: 0, gasActuallyUsed: 0, shouldQueryErr: false},     // both queryGasLimit and gasActuallyUsed are 0
		{queryGasLimit: 200, gasActuallyUsed: 200, shouldQueryErr: false}, // gasActuallyUsed == queryGasLimit
		{queryGasLimit: 100, gasActuallyUsed: 1000, shouldQueryErr: true}, // gasActuallyUsed > queryGasLimit
	}

	for _, tc := range testCases {
		for _, ctxType := range ctxTypes {
			t.Run(fmt.Sprintf("%s: %d - %d", ctxType, tc.queryGasLimit, tc.gasActuallyUsed), func(t *testing.T) {
				app := getQueryBaseapp(t)
				baseapp.SetQueryGasLimit(tc.queryGasLimit)(app)
				ctx := ctxType.GetCtx(t, app)

				// query gas limit should have no effect when CtxType != QueryCtx
				if tc.shouldQueryErr && ctxType == QueryCtx {
					require.Panics(t, func() { ctx.GasMeter().ConsumeGas(tc.gasActuallyUsed, "test") })
				} else {
					require.NotPanics(t, func() { ctx.GasMeter().ConsumeGas(tc.gasActuallyUsed, "test") })
				}
			})
		}
	}
}

func TestGetMaximumBlockGas(t *testing.T) {
	suite := NewBaseAppSuite(t)
	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{})
	require.NoError(t, err)
	ctx := suite.baseApp.NewContext(true)

	err = suite.baseApp.StoreConsensusParams(ctx, cmtproto.ConsensusParams{Block: &cmtproto.BlockParams{MaxGas: 0}})
	require.NoError(t, err)
	require.Equal(t, uint64(0), suite.baseApp.GetMaximumBlockGas(ctx))

	err = suite.baseApp.StoreConsensusParams(ctx, cmtproto.ConsensusParams{Block: &cmtproto.BlockParams{MaxGas: -1}})
	require.NoError(t, err)
	require.Equal(t, uint64(0), suite.baseApp.GetMaximumBlockGas(ctx))

	err = suite.baseApp.StoreConsensusParams(ctx, cmtproto.ConsensusParams{Block: &cmtproto.BlockParams{MaxGas: 5000000}})
	require.NoError(t, err)
	require.Equal(t, uint64(5000000), suite.baseApp.GetMaximumBlockGas(ctx))

	err = suite.baseApp.StoreConsensusParams(ctx, cmtproto.ConsensusParams{Block: &cmtproto.BlockParams{MaxGas: -5000000}})
	require.NoError(t, err)
	require.Panics(t, func() { suite.baseApp.GetMaximumBlockGas(ctx) })
}

func TestLoadVersionPruning(t *testing.T) {
	logger := log.NewNopLogger()
	pruningOptions := pruningtypes.NewCustomPruningOptions(10, 15)
	pruningOpt := baseapp.SetPruning(pruningOptions)
	db := dbm.NewMemDB()
	name := t.Name()
	app := baseapp.NewBaseApp(name, logger, db, nil, pruningOpt)

	// make a cap key and mount the store
	capKey := storetypes.NewKVStoreKey("key1")
	app.MountStores(capKey)

	err := app.LoadLatestVersion() // needed to make stores non-nil
	require.Nil(t, err)

	emptyCommitID := storetypes.CommitID{}

	// fresh store has zero/empty last commit
	lastHeight := app.LastBlockHeight()
	lastID := app.LastCommitID()
	require.Equal(t, int64(0), lastHeight)
	require.Equal(t, emptyCommitID, lastID)

	var lastCommitID storetypes.CommitID

	// Commit seven blocks, of which 7 (latest) is kept in addition to 6, 5
	// (keep recent) and 3 (keep every).
	for i := int64(1); i <= 7; i++ {
		res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: i})
		require.NoError(t, err)
		_, err = app.Commit()
		require.NoError(t, err)
		lastCommitID = storetypes.CommitID{Version: i, Hash: res.AppHash}
	}

	for _, v := range []int64{1, 2, 4} {
		_, err = app.CommitMultiStore().CacheMultiStoreWithVersion(v)
		require.NoError(t, err)
	}

	for _, v := range []int64{3, 5, 6, 7} {
		_, err = app.CommitMultiStore().CacheMultiStoreWithVersion(v)
		require.NoError(t, err)
	}

	// reload with LoadLatestVersion, check it loads last version
	app = baseapp.NewBaseApp(name, logger, db, nil, pruningOpt)
	app.MountStores(capKey)

	err = app.LoadLatestVersion()
	require.Nil(t, err)
	testLoadVersionHelper(t, app, int64(7), lastCommitID)
}
