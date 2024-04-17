package baseapp_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	"cosmossdk.io/store/metrics"
	pruningtypes "cosmossdk.io/store/pruning/types"
	"cosmossdk.io/store/rootmulti"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
		baseApp   *baseapp.BaseApp
		cdc       *codec.ProtoCodec
		txConfig  client.TxConfig
		logBuffer *bytes.Buffer
	}

	SnapshotsConfig struct {
		blocks             uint64
		blockTxs           int
		snapshotInterval   uint64
		snapshotKeepRecent uint32
		pruningOpts        pruningtypes.PruningOptions
	}
)

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

func TestLoadVersion(t *testing.T) {
	logger := log.NewTestLogger(t)
	pruningOpt := baseapp.SetPruning(pruningtypes.NewPruningOptions(pruningtypes.PruningNothing))
	db := dbm.NewMemDB()
	name := t.Name()
	app := baseapp.NewBaseApp(name, logger, db, nil, pruningOpt)

	// make a cap key and mount the store
	err := app.LoadLatestVersion() // needed to make stores non-nil
	require.Nil(t, err)

	emptyHash := sha256.Sum256([]byte{})
	appHash := emptyHash[:]
	emptyCommitID := storetypes.CommitID{Hash: appHash}

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

	emptyHash := sha256.Sum256([]byte{})
	emptyCommitID := storetypes.CommitID{
		Hash: emptyHash[:],
	}

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
