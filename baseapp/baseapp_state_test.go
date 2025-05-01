package baseapp

import (
	"context"
	"cosmossdk.io/log"
	pruningtypes "cosmossdk.io/store/pruning/types"
	storetypes "cosmossdk.io/store/types"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

var (
	capKey1 = storetypes.NewKVStoreKey("key1")
	capKey2 = storetypes.NewKVStoreKey("key2")

	ParamStoreKey = []byte("paramstore")
)

type (
	BaseAppSuite struct {
		baseApp  *BaseApp
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

func NewBaseAppSuite(t *testing.T, opts ...func(*BaseApp)) *BaseAppSuite {
	t.Helper()

	cdc := codectestutil.CodecOptions{}.NewCodec()
	baseapptestutil.RegisterInterfaces(cdc.InterfaceRegistry())

	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)
	db := dbm.NewMemDB()
	logger := log.NewLogger(os.Stdout, log.ColorOption(false))

	app := NewBaseApp(t.Name(), logger, db, txConfig.TxDecoder(), opts...)
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

func getQueryBaseapp(t *testing.T) *BaseApp {
	t.Helper()

	db := dbm.NewMemDB()
	name := t.Name()
	app := NewBaseApp(name, log.NewTestLogger(t), db, nil)

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

func TestABCI_CreateQueryContext_Before_Set_CheckState(t *testing.T) {
	t.Parallel()

	db := dbm.NewMemDB()
	name := t.Name()
	var height int64 = 2
	var headerHeight int64 = 1

	t.Run("valid height with different initial height", func(t *testing.T) {
		app := NewBaseApp(name, log.NewTestLogger(t), db, nil)

		_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
		require.NoError(t, err)
		_, err = app.Commit()
		require.NoError(t, err)

		_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 2})
		require.NoError(t, err)

		var queryCtx *sdk.Context
		var queryCtxErr error
		app.SetStreamingManager(storetypes.StreamingManager{
			ABCIListeners: []storetypes.ABCIListener{
				&mockABCIListener{
					ListenCommitFn: func(context.Context, abci.ResponseCommit, []*storetypes.StoreKVPair) error {
						qCtx, qErr := app.CreateQueryContext(0, true)
						queryCtx = &qCtx
						queryCtxErr = qErr
						return nil
					},
				},
			},
		})
		_, err = app.Commit()
		require.NoError(t, err)
		require.NoError(t, queryCtxErr)
		require.Equal(t, height, queryCtx.BlockHeight())
		_, err = app.InitChain(&abci.RequestInitChain{
			InitialHeight: headerHeight,
		})
		require.NoError(t, err)
	})
}

func TestSetMinGasPrices(t *testing.T) {
	minGasPrices := sdk.DecCoins{sdk.NewInt64DecCoin("stake", 5000)}
	suite := NewBaseAppSuite(t, SetMinGasPrices(minGasPrices.String()))

	ctx := suite.baseApp.stateManager.GetState(execModeCheck).Context()
	require.Equal(t, minGasPrices, ctx.MinGasPrices())
}

type ctxType string

const (
	QueryCtx   ctxType = "query"
	CheckTxCtx ctxType = "checkTx"
)

var ctxTypes = []ctxType{QueryCtx, CheckTxCtx}

func (c ctxType) GetCtx(t *testing.T, bapp *BaseApp) sdk.Context {
	t.Helper()
	switch c {
	case QueryCtx:
		ctx, err := bapp.CreateQueryContext(1, false)
		require.NoError(t, err)
		return ctx
	case CheckTxCtx:
		return bapp.stateManager.GetState(execModeCheck).Context()
	}
	// TODO: Not supported yet
	return bapp.stateManager.GetState(execModeFinalize).Context()
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
				SetQueryGasLimit(tc.queryGasLimit)(app)
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

	require.NoError(t, suite.baseApp.StoreConsensusParams(ctx, cmtproto.ConsensusParams{Block: &cmtproto.BlockParams{MaxGas: 0}}))
	require.Equal(t, uint64(0), suite.baseApp.GetMaximumBlockGas(ctx))

	require.NoError(t, suite.baseApp.StoreConsensusParams(ctx, cmtproto.ConsensusParams{Block: &cmtproto.BlockParams{MaxGas: -1}}))
	require.Equal(t, uint64(0), suite.baseApp.GetMaximumBlockGas(ctx))

	require.NoError(t, suite.baseApp.StoreConsensusParams(ctx, cmtproto.ConsensusParams{Block: &cmtproto.BlockParams{MaxGas: 5000000}}))
	require.Equal(t, uint64(5000000), suite.baseApp.GetMaximumBlockGas(ctx))

	require.NoError(t, suite.baseApp.StoreConsensusParams(ctx, cmtproto.ConsensusParams{Block: &cmtproto.BlockParams{MaxGas: -5000000}}))
	require.Panics(t, func() { suite.baseApp.GetMaximumBlockGas(ctx) })
}

func TestGetEmptyConsensusParams(t *testing.T) {
	suite := NewBaseAppSuite(t)
	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{})
	require.NoError(t, err)
	ctx := suite.baseApp.NewContext(true)

	cp := suite.baseApp.GetConsensusParams(ctx)
	require.Equal(t, cmtproto.ConsensusParams{}, cp)
	require.Equal(t, uint64(0), suite.baseApp.GetMaximumBlockGas(ctx))
}

func TestLoadVersionPruning(t *testing.T) {
	logger := log.NewNopLogger()
	pruningOptions := pruningtypes.NewCustomPruningOptions(10, 15)
	pruningOpt := SetPruning(pruningOptions)
	db := dbm.NewMemDB()
	name := t.Name()
	app := NewBaseApp(name, logger, db, nil, pruningOpt)

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
	app = NewBaseApp(name, logger, db, nil, pruningOpt)
	app.MountStores(capKey)

	err = app.LoadLatestVersion()
	require.Nil(t, err)
	testLoadVersionHelper(t, app, int64(7), lastCommitID)
}

func testLoadVersionHelper(t *testing.T, app *BaseApp, expectedHeight int64, expectedID storetypes.CommitID) {
	t.Helper()

	lastHeight := app.LastBlockHeight()
	lastID := app.LastCommitID()
	require.Equal(t, expectedHeight, lastHeight)
	require.Equal(t, expectedID, lastID)
}

type paramStore struct {
	db *dbm.MemDB
}

var _ ParamStore = (*paramStore)(nil)

func (ps paramStore) Set(_ context.Context, value cmtproto.ConsensusParams) error {
	bz, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return ps.db.Set(ParamStoreKey, bz)
}

func (ps paramStore) Has(_ context.Context) (bool, error) {
	return ps.db.Has(ParamStoreKey)
}

func (ps paramStore) Get(_ context.Context) (cmtproto.ConsensusParams, error) {
	bz, err := ps.db.Get(ParamStoreKey)
	if err != nil {
		return cmtproto.ConsensusParams{}, err
	}

	if len(bz) == 0 {
		return cmtproto.ConsensusParams{}, errors.New("params not found")
	}

	var params cmtproto.ConsensusParams
	if err := json.Unmarshal(bz, &params); err != nil {
		return cmtproto.ConsensusParams{}, err
	}

	return params, nil
}
