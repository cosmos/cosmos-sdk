package baseapp_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/testutil/testdata_pulsar"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/db/memdb"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/snapshots"
	snapshottypes "github.com/cosmos/cosmos-sdk/snapshots/types"
	stypes "github.com/cosmos/cosmos-sdk/store/v2"
	"github.com/cosmos/cosmos-sdk/store/v2/multi"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/middleware"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

var (
	capKey1 = sdk.NewKVStoreKey("key1")
	capKey2 = sdk.NewKVStoreKey("key2")

	encCfg = simapp.MakeTestEncodingConfig()
)

func init() {
	registerTestCodec(encCfg.Amino)
}

type paramStore struct {
	db *memdb.MemDB
	rw dbm.DBReadWriter
}

func newParamStore(db *memdb.MemDB) *paramStore {
	return &paramStore{db: db, rw: db.ReadWriter()}
}

func (ps *paramStore) Set(_ sdk.Context, key []byte, value interface{}) {
	bz, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}

	ps.rw.Set(key, bz)
}

func (ps *paramStore) Has(_ sdk.Context, key []byte) bool {
	ok, err := ps.rw.Has(key)
	if err != nil {
		panic(err)
	}

	return ok
}

func (ps *paramStore) Get(_ sdk.Context, key []byte, ptr interface{}) {
	bz, err := ps.rw.Get(key)
	if err != nil {
		panic(err)
	}

	if len(bz) == 0 {
		return
	}

	if err := json.Unmarshal(bz, ptr); err != nil {
		panic(err)
	}
}

func defaultLogger() log.Logger {
	logger, _ := log.NewDefaultLogger("plain", "info", false)
	return logger.With("module", "sdk/app")
}

func newBaseApp(name string, options ...baseapp.AppOption) *baseapp.BaseApp {
	logger := defaultLogger()
	db := memdb.NewDB()
	codec := codec.NewLegacyAmino()
	registerTestCodec(codec)
	return baseapp.NewBaseApp(name, logger, db, options...)
}

func registerTestCodec(cdc *codec.LegacyAmino) {
	// register test types
	cdc.RegisterConcrete(&txTest{}, "cosmos-sdk/baseapp/txTest", nil)
	cdc.RegisterConcrete(&msgCounter{}, "cosmos-sdk/baseapp/msgCounter", nil)
	cdc.RegisterConcrete(&msgCounter2{}, "cosmos-sdk/baseapp/msgCounter2", nil)
	cdc.RegisterConcrete(&msgKeyValue{}, "cosmos-sdk/baseapp/msgKeyValue", nil)
	cdc.RegisterConcrete(&msgNoRoute{}, "cosmos-sdk/baseapp/msgNoRoute", nil)
}

// aminoTxEncoder creates a amino TxEncoder for testing purposes.
func aminoTxEncoder(cdc *codec.LegacyAmino) sdk.TxEncoder {
	return legacytx.StdTxConfig{Cdc: cdc}.TxEncoder()
}

// simple one store baseapp
func setupBaseApp(t *testing.T, options ...baseapp.AppOption) *baseapp.BaseApp {
	options = append(options, baseapp.SetSubstores(capKey1, capKey2))
	app := newBaseApp(t.Name(), options...)
	require.Equal(t, t.Name(), app.Name())

	app.SetParamStore(newParamStore(memdb.NewDB()))

	// stores are mounted
	err := app.Init()
	require.Nil(t, err)
	return app
}

// testTxHandler is a tx.Handler used for the mock app, it does not
// contain any signature verification logic.
func testTxHandler(options middleware.TxHandlerOptions, customTxHandlerMiddleware handlerFun) tx.Handler {
	return middleware.ComposeMiddlewares(
		middleware.NewRunMsgsTxHandler(options.MsgServiceRouter, options.LegacyRouter),
		middleware.NewTxDecoderMiddleware(options.TxDecoder),
		middleware.GasTxMiddleware,
		middleware.RecoveryTxMiddleware,
		middleware.NewIndexEventsTxMiddleware(options.IndexEvents),
		middleware.ValidateBasicMiddleware,
		middleware.ConsumeBlockGasMiddleware,
		CustomTxHandlerMiddleware(customTxHandlerMiddleware),
	)
}

// simple one store baseapp with data and snapshots. Each tx is 1 MB in size (uncompressed).
func setupBaseAppWithSnapshots(t *testing.T, blocks uint, blockTxs int, options ...baseapp.AppOption) (*baseapp.BaseApp, func()) {
	codec := codec.NewLegacyAmino()
	registerTestCodec(codec)
	routerOpt := baseapp.AppOptionFunc(func(bapp *baseapp.BaseApp) {
		legacyRouter := middleware.NewLegacyRouter()
		legacyRouter.AddRoute(sdk.NewRoute(routeMsgKeyValue, func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
			kv := msg.(*msgKeyValue)
			bapp.Store().GetKVStore(capKey2).Set(kv.Key, kv.Value)
			any, err := codectypes.NewAnyWithValue(msg)
			if err != nil {
				return nil, err
			}

			return &sdk.Result{
				MsgResponses: []*codectypes.Any{any},
			}, nil
		}))
		txHandler := testTxHandler(
			middleware.TxHandlerOptions{
				LegacyRouter:     legacyRouter,
				MsgServiceRouter: middleware.NewMsgServiceRouter(encCfg.InterfaceRegistry),
				TxDecoder:        testTxDecoder(encCfg.Amino),
			},
			func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) { return ctx, nil },
		)
		bapp.SetTxHandler(txHandler)
	})

	snapshotInterval := uint64(2)
	snapshotTimeout := 1 * time.Minute
	snapshotDir, err := os.MkdirTemp("", "baseapp")
	require.NoError(t, err)
	snapshotStore, err := snapshots.NewStore(memdb.NewDB(), snapshotDir)
	require.NoError(t, err)
	teardown := func() {
		os.RemoveAll(snapshotDir)
	}

	app := setupBaseApp(t, append(options,
		baseapp.SetSnapshotStore(snapshotStore),
		baseapp.SetSnapshotInterval(snapshotInterval),
		routerOpt)...)

	app.InitChain(abci.RequestInitChain{})

	r := rand.New(rand.NewSource(3920758213583))
	keyCounter := 0
	for height := int64(1); height <= int64(blocks); height++ {
		app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: height}})
		for txNum := 0; txNum < blockTxs; txNum++ {
			tx := txTest{Msgs: []sdk.Msg{}}
			for msgNum := 0; msgNum < 100; msgNum++ {
				key := []byte(fmt.Sprintf("%v", keyCounter))
				value := make([]byte, 10000)
				_, err := r.Read(value)
				require.NoError(t, err)
				tx.Msgs = append(tx.Msgs, msgKeyValue{Key: key, Value: value})
				keyCounter++
			}
			txBytes, err := encCfg.Amino.Marshal(tx)
			require.NoError(t, err)
			resp := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
			require.True(t, resp.IsOK(), "%v", resp.String())
		}
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()

		// Wait for snapshot to be taken, since it happens asynchronously.
		if uint64(height)%snapshotInterval == 0 {
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

	return app, teardown
}

func TestMountStores(t *testing.T) {
	app := setupBaseApp(t)

	// check both stores
	store1 := app.Store().GetKVStore(capKey1)
	require.NotNil(t, store1)
	store2 := app.Store().GetKVStore(capKey2)
	require.NotNil(t, store2)
}

type MockTxHandler struct {
	T *testing.T
}

func (th MockTxHandler) CheckTx(goCtx context.Context, req tx.Request, checkReq tx.RequestCheckTx) (tx.Response, tx.ResponseCheckTx, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	require.NotNil(th.T, ctx.ConsensusParams())
	return tx.Response{}, tx.ResponseCheckTx{}, nil
}

func (th MockTxHandler) DeliverTx(goCtx context.Context, req tx.Request) (tx.Response, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	require.NotNil(th.T, ctx.ConsensusParams())
	return tx.Response{}, nil
}

func (th MockTxHandler) SimulateTx(goCtx context.Context, req tx.Request) (tx.Response, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	require.NotNil(th.T, ctx.ConsensusParams())
	return tx.Response{}, nil
}

func TestConsensusParamsNotNil(t *testing.T) {
	app := setupBaseApp(t, baseapp.AppOptionFunc(func(app *baseapp.BaseApp) {
		app.SetBeginBlocker(func(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
			require.NotNil(t, ctx.ConsensusParams())
			return abci.ResponseBeginBlock{}
		})
	}), baseapp.AppOptionFunc(func(app *baseapp.BaseApp) {
		app.SetEndBlocker(func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
			require.NotNil(t, ctx.ConsensusParams())
			return abci.ResponseEndBlock{}
		})
	}), baseapp.AppOptionFunc(func(app *baseapp.BaseApp) {
		app.SetTxHandler(MockTxHandler{T: t})
	}))

	header := tmproto.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	app.EndBlock(abci.RequestEndBlock{Height: header.Height})
	app.CheckTx(abci.RequestCheckTx{})
	app.DeliverTx(abci.RequestDeliverTx{})
	app.Simulate([]byte{})
}

// Test that we can make commits and then reload old versions.
// Test that LoadLatestVersion actually does.
func TestLoadVersion(t *testing.T) {
	logger := defaultLogger()
	pruningOpt := baseapp.SetPruning(stypes.PruneNothing)
	db := memdb.NewDB()
	name := t.Name()
	app := baseapp.NewBaseApp(name, logger, db, pruningOpt)

	// make a cap key and mount the store
	err := app.Init() // needed to make stores non-nil
	require.Nil(t, err)

	emptyCommitID := stypes.CommitID{}

	// fresh store has zero/empty last commit
	lastHeight := app.LastBlockHeight()
	lastID := app.LastCommitID()
	require.Equal(t, int64(0), lastHeight)
	require.Equal(t, emptyCommitID, lastID)

	// execute a block, collect commit ID
	header := tmproto.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	_ = app.Commit()

	// execute a block, collect commit ID
	header = tmproto.Header{Height: 2}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	res := app.Commit()
	commitID2 := stypes.CommitID{Version: 2, Hash: res.Data}
	app.CloseStore()

	// reload with latest version
	app = baseapp.NewBaseApp(name, logger, db, pruningOpt)
	err = app.Init()
	require.Nil(t, err)
	testLoadVersionHelper(t, app, int64(2), commitID2)
}

func initStore(t *testing.T, db dbm.DBConnection, storeKey string, k, v []byte) {
	key := sdk.NewKVStoreKey(storeKey)
	opts := multi.DefaultStoreParams()
	opts.Pruning = stypes.PruneNothing
	require.NoError(t, opts.RegisterSubstore(key, stypes.StoreTypePersistent))
	rs, err := multi.NewV1MultiStoreAsV2(db, opts)
	require.NoError(t, err)
	require.Equal(t, int64(0), rs.LastCommitID().Version)

	// write some data in substore
	kv := rs.GetKVStore(key)
	require.NotNil(t, kv)
	kv.Set(k, v)
	commitID := rs.Commit()
	require.Equal(t, int64(1), commitID.Version)
	require.NoError(t, rs.Close())
}

func checkStore(t *testing.T, db dbm.DBConnection, ver int64, storeKey string, k, v []byte) {
	opts := multi.DefaultStoreParams()
	opts.Pruning = stypes.PruneNothing
	key := sdk.NewKVStoreKey(storeKey)
	require.NoError(t, opts.RegisterSubstore(key, stypes.StoreTypePersistent))
	rs, err := multi.NewV1MultiStoreAsV2(db, opts)
	require.NoError(t, err)
	require.Equal(t, ver, rs.LastCommitID().Version)

	// query data in substore
	kv := rs.GetKVStore(key)
	require.NotNil(t, kv)
	require.Equal(t, v, kv.Get(k))
	require.NoError(t, rs.Close())
}

func TestVersionSetterGetter(t *testing.T) {
	logger := defaultLogger()
	pruningOpt := baseapp.SetPruning(stypes.PruneDefault)
	db := memdb.NewDB()
	name := t.Name()
	app := baseapp.NewBaseApp(name, logger, db, pruningOpt)
	require.Equal(t, "", app.Version())
	res := app.Query(abci.RequestQuery{Path: "app/version"})
	require.True(t, res.IsOK())
	require.Equal(t, "", string(res.Value))

	versionString := "1.0.0"
	app.SetVersion(versionString)
	require.Equal(t, versionString, app.Version())
	res = app.Query(abci.RequestQuery{Path: "app/version"})
	require.True(t, res.IsOK())
	require.Equal(t, versionString, string(res.Value))
}

func TestLoadVersionPruning(t *testing.T) {
	logger := log.NewNopLogger()
	pruningOptions := stypes.PruningOptions{
		KeepRecent: 2,
		Interval:   1,
	}
	pruningOpt := baseapp.SetPruning(pruningOptions)
	schemaOpt := baseapp.SetSubstores(sdk.NewKVStoreKey("key1"))
	db := memdb.NewDB()
	name := t.Name()
	app := baseapp.NewBaseApp(name, logger, db, pruningOpt, schemaOpt)

	err := app.Init() // needed to make stores non-nil
	require.Nil(t, err)

	emptyCommitID := stypes.CommitID{}

	// fresh store has zero/empty last commit
	lastHeight := app.LastBlockHeight()
	lastID := app.LastCommitID()
	require.Equal(t, int64(0), lastHeight)
	require.Equal(t, emptyCommitID, lastID)

	var lastCommitID stypes.CommitID

	// Commit seven blocks, of which 7 (latest) is kept in addition to 6, 5
	// (keep recent) and 3 (keep every).
	for i := int64(1); i <= 7; i++ {
		app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: i}})
		res := app.Commit()
		lastCommitID = stypes.CommitID{Version: i, Hash: res.Data}
	}

	// TODO: behavior change -
	// CacheMultiStoreWithVersion returned no error on missing version (?)
	for _, v := range []int64{1, 2, 4} {
		_, err = app.Store().GetVersion(v)
		require.NoError(t, err, "version=%v", v)
	}

	for _, v := range []int64{3, 5, 6, 7} {
		_, err = app.Store().GetVersion(v)
		require.NoError(t, err, "version=%v", v)
	}
	require.NoError(t, app.CloseStore())

	// reload app, check it loads last version
	app = baseapp.NewBaseApp(name, logger, db, pruningOpt, schemaOpt)

	err = app.Init()
	require.Nil(t, err)
	testLoadVersionHelper(t, app, int64(7), lastCommitID)
}

func testLoadVersionHelper(t *testing.T, app *baseapp.BaseApp, expectedHeight int64, expectedID stypes.CommitID) {
	lastHeight := app.LastBlockHeight()
	lastID := app.LastCommitID()
	require.Equal(t, expectedHeight, lastHeight)
	require.Equal(t, expectedID, lastID)
}

func TestOptionFunction(t *testing.T) {
	logger := defaultLogger()
	db := memdb.NewDB()
	bap := baseapp.NewBaseApp("starting name", logger, db, testChangeNameHelper("new name"))
	require.Equal(t, bap.Name(), "new name", "BaseApp should have had name changed via option function")
}

func testChangeNameHelper(name string) baseapp.AppOptionFunc {
	return func(bap *baseapp.BaseApp) { bap.SetName(name) }
}

// Test that Info returns the latest committed state.
func TestInfo(t *testing.T) {
	app := newBaseApp(t.Name())

	// ----- test an empty response -------
	reqInfo := abci.RequestInfo{}
	res := app.Info(reqInfo)

	// should be empty
	assert.Equal(t, "", res.Version)
	assert.Equal(t, t.Name(), res.GetData())
	assert.Equal(t, int64(0), res.LastBlockHeight)
	require.Equal(t, []uint8(nil), res.LastBlockAppHash)
	require.Equal(t, app.AppVersion(), res.AppVersion)
	// ----- test a proper response -------
	// TODO
}

func TestBaseAppOptionSeal(t *testing.T) {
	app := setupBaseApp(t)

	require.Panics(t, func() {
		app.SetName("")
	})
	require.Panics(t, func() {
		app.SetVersion("")
	})
	require.Panics(t, func() {
		app.SetInitChainer(nil)
	})
	require.Panics(t, func() {
		app.SetBeginBlocker(nil)
	})
	require.Panics(t, func() {
		app.SetEndBlocker(nil)
	})
	require.Panics(t, func() {
		app.SetTxHandler(nil)
	})
	require.Panics(t, func() {
		app.SetAddrPeerFilter(nil)
	})
	require.Panics(t, func() {
		app.SetIDPeerFilter(nil)
	})
	require.Panics(t, func() {
		app.SetFauxMerkleMode()
	})
}

func TestSetMinGasPrices(t *testing.T) {
	minGasPrices := sdk.DecCoins{sdk.NewInt64DecCoin("stake", 5000)}
	app := newBaseApp(t.Name(), baseapp.SetMinGasPrices(minGasPrices.String()))
	require.Equal(t, minGasPrices, app.MinGasPrices())
}

func TestInitChainer(t *testing.T) {
	name := t.Name()
	// keep the db and logger ourselves so
	// we can reload the same  app later
	db := memdb.NewDB()
	logger := defaultLogger()
	capKey := sdk.NewKVStoreKey("main")
	capKey2 := sdk.NewKVStoreKey("key2")
	schemaOpt := baseapp.SetSubstores(capKey, capKey2)
	app := baseapp.NewBaseApp(name, logger, db, schemaOpt)

	// set a value in the store on init chain
	key, value := []byte("hello"), []byte("goodbye")
	var initChainer sdk.InitChainer = func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		store := ctx.KVStore(capKey)
		store.Set(key, value)
		return abci.ResponseInitChain{}
	}

	query := abci.RequestQuery{
		Path: "/store/main/key",
		Data: key,
	}

	// initChainer is nil - nothing happens
	app.InitChain(abci.RequestInitChain{})
	res := app.Query(query)
	require.Nil(t, res.Value)

	// set initChainer and try again - should see the value
	app.SetInitChainer(initChainer)

	// stores are mounted and private members are set - sealing baseapp
	err := app.Init()
	require.Nil(t, err)
	require.Equal(t, int64(0), app.LastBlockHeight())

	initChainRes := app.InitChain(abci.RequestInitChain{AppStateBytes: []byte("{}"), ChainId: "test-chain-id"}) // must have valid JSON genesis file, even if empty

	// The AppHash returned by a new chain is the sha256 hash of "".
	// $ echo -n '' | sha256sum
	// e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
	require.Equal(
		t,
		[]byte{0xe3, 0xb0, 0xc4, 0x42, 0x98, 0xfc, 0x1c, 0x14, 0x9a, 0xfb, 0xf4, 0xc8, 0x99, 0x6f, 0xb9, 0x24, 0x27, 0xae, 0x41, 0xe4, 0x64, 0x9b, 0x93, 0x4c, 0xa4, 0x95, 0x99, 0x1b, 0x78, 0x52, 0xb8, 0x55},
		initChainRes.AppHash,
	)

	// assert that chainID is set correctly in InitChain
	chainID := app.DeliverState().Context().ChainID()
	require.Equal(t, "test-chain-id", chainID, "ChainID in deliverState not set correctly in InitChain")

	chainID = app.CheckState().Context().ChainID()
	require.Equal(t, "test-chain-id", chainID, "ChainID in checkState not set correctly in InitChain")

	app.Commit()
	res = app.Query(query)
	require.True(t, res.IsOK(), res.Log)
	require.Equal(t, int64(1), app.LastBlockHeight())
	require.Equal(t, value, res.Value)
	require.NoError(t, app.CloseStore())

	// reload app
	app = baseapp.NewBaseApp(name, logger, db, schemaOpt)
	app.SetInitChainer(initChainer)

	err = app.Init()
	require.Nil(t, err)
	require.Equal(t, int64(1), app.LastBlockHeight())

	// ensure we can still query after reloading
	res = app.Query(query)
	require.Equal(t, value, res.Value)

	// commit and ensure we can still query
	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	app.Commit()

	res = app.Query(query)
	require.Equal(t, value, res.Value)
}

func TestInitChain_WithInitialHeight(t *testing.T) {
	name := t.Name()
	db := memdb.NewDB()
	logger := defaultLogger()
	app := baseapp.NewBaseApp(name, logger, db)

	app.InitChain(
		abci.RequestInitChain{
			InitialHeight: 3,
		},
	)
	app.Commit()

	require.Equal(t, int64(3), app.LastBlockHeight())
}

func TestBeginBlock_WithInitialHeight(t *testing.T) {
	name := t.Name()
	db := memdb.NewDB()
	logger := defaultLogger()
	app := baseapp.NewBaseApp(name, logger, db)

	app.InitChain(
		abci.RequestInitChain{
			InitialHeight: 3,
		},
	)

	require.PanicsWithError(t, "invalid height: 4; expected: 3", func() {
		app.BeginBlock(abci.RequestBeginBlock{
			Header: tmproto.Header{
				Height: 4,
			},
		})
	})

	app.BeginBlock(abci.RequestBeginBlock{
		Header: tmproto.Header{
			Height: 3,
		},
	})
	app.Commit()

	require.Equal(t, int64(3), app.LastBlockHeight())
}

// Simple tx with a list of Msgs.
type txTest struct {
	Msgs       []sdk.Msg
	Counter    int64
	FailOnAnte bool
	GasLimit   uint64
}

func (tx *txTest) setFailOnAnte(fail bool) {
	tx.FailOnAnte = fail
}

func (tx *txTest) setFailOnHandler(fail bool) {
	for i, msg := range tx.Msgs {
		tx.Msgs[i] = &msgCounter{msg.(*msgCounter).Counter, fail}
	}
}

// Implements Tx
func (tx txTest) GetMsgs() []sdk.Msg   { return tx.Msgs }
func (tx txTest) ValidateBasic() error { return nil }

// Implements GasTx
func (tx txTest) GetGas() uint64 { return tx.GasLimit }

// Implements TxWithTimeoutHeight
func (tx txTest) GetTimeoutHeight() uint64 { return 0 }

const (
	routeMsgCounter  = "msgCounter"
	routeMsgCounter2 = "msgCounter2"
	routeMsgKeyValue = "msgKeyValue"
)

// ValidateBasic() fails on negative counters.
// Otherwise it's up to the handlers
type msgCounter struct {
	Counter       int64
	FailOnHandler bool
}

// dummy implementation of proto.Message
func (msg *msgCounter) Reset()         {}
func (msg *msgCounter) String() string { return "TODO" }
func (msg *msgCounter) ProtoMessage()  {}

// Implements Msg
func (msg *msgCounter) Route() string                { return routeMsgCounter }
func (msg *msgCounter) Type() string                 { return "counter1" }
func (msg *msgCounter) GetSignBytes() []byte         { return nil }
func (msg *msgCounter) GetSigners() []sdk.AccAddress { return nil }
func (msg *msgCounter) ValidateBasic() error {
	if msg.Counter >= 0 {
		return nil
	}
	return sdkerrors.Wrap(sdkerrors.ErrInvalidSequence, "counter should be a non-negative integer")
}

func newTxCounter(counter int64, msgCounters ...int64) txTest {
	msgs := make([]sdk.Msg, 0, len(msgCounters))
	for _, c := range msgCounters {
		msgs = append(msgs, &msgCounter{c, false})
	}

	return txTest{msgs, counter, false, math.MaxUint64}
}

// a msg we dont know how to route
type msgNoRoute struct {
	msgCounter
}

func (tx msgNoRoute) Route() string { return "noroute" }

// a msg we dont know how to decode
type msgNoDecode struct {
	msgCounter
}

func (tx msgNoDecode) Route() string { return routeMsgCounter }

// Another counter msg. Duplicate of msgCounter
type msgCounter2 struct {
	Counter int64
}

// dummy implementation of proto.Message
func (msg msgCounter2) Reset()         {}
func (msg msgCounter2) String() string { return "TODO" }
func (msg msgCounter2) ProtoMessage()  {}

// Implements Msg
func (msg msgCounter2) Route() string                { return routeMsgCounter2 }
func (msg msgCounter2) Type() string                 { return "counter2" }
func (msg msgCounter2) GetSignBytes() []byte         { return nil }
func (msg msgCounter2) GetSigners() []sdk.AccAddress { return nil }
func (msg msgCounter2) ValidateBasic() error {
	if msg.Counter >= 0 {
		return nil
	}
	return sdkerrors.Wrap(sdkerrors.ErrInvalidSequence, "counter should be a non-negative integer")
}

// A msg that sets a key/value pair.
type msgKeyValue struct {
	Key   []byte
	Value []byte
}

func (msg msgKeyValue) Reset()                       {}
func (msg msgKeyValue) String() string               { return "TODO" }
func (msg msgKeyValue) ProtoMessage()                {}
func (msg msgKeyValue) Route() string                { return routeMsgKeyValue }
func (msg msgKeyValue) Type() string                 { return "keyValue" }
func (msg msgKeyValue) GetSignBytes() []byte         { return nil }
func (msg msgKeyValue) GetSigners() []sdk.AccAddress { return nil }
func (msg msgKeyValue) ValidateBasic() error {
	if msg.Key == nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "key cannot be nil")
	}
	if msg.Value == nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "value cannot be nil")
	}
	return nil
}

// amino decode
func testTxDecoder(cdc *codec.LegacyAmino) sdk.TxDecoder {
	return func(txBytes []byte) (sdk.Tx, error) {
		var tx txTest
		if len(txBytes) == 0 {
			return nil, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "tx bytes are empty")
		}

		err := cdc.Unmarshal(txBytes, &tx)
		if err != nil {
			return nil, sdkerrors.ErrTxDecode
		}

		return tx, nil
	}
}

func customHandlerTxTest(t *testing.T, capKey stypes.StoreKey, storeKey []byte) handlerFun {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		store := ctx.KVStore(capKey)
		txTest := tx.(txTest)

		if txTest.FailOnAnte {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "ante handler failure")
		}

		_, err := incrementingCounter(t, store, storeKey, txTest.Counter)
		if err != nil {
			return ctx, err
		}

		ctx.EventManager().EmitEvents(
			counterEvent("post_handlers", txTest.Counter),
		)

		return ctx, nil
	}
}

func counterEvent(evType string, msgCount int64) sdk.Events {
	return sdk.Events{
		sdk.NewEvent(
			evType,
			sdk.NewAttribute("update_counter", fmt.Sprintf("%d", msgCount)),
		),
	}
}

func handlerMsgCounter(t *testing.T, capKey stypes.StoreKey, deliverKey []byte) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		store := ctx.KVStore(capKey)
		var msgCount int64

		switch m := msg.(type) {
		case *msgCounter:
			if m.FailOnHandler {
				return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "message handler failure")
			}

			msgCount = m.Counter
		case *msgCounter2:
			msgCount = m.Counter
		}

		ctx.EventManager().EmitEvents(
			counterEvent(sdk.EventTypeMessage, msgCount),
		)

		res, err := incrementingCounter(t, store, deliverKey, msgCount)
		if err != nil {
			return nil, err
		}

		res.Events = ctx.EventManager().Events().ToABCIEvents()

		any, err := codectypes.NewAnyWithValue(msg)
		if err != nil {
			return nil, err
		}

		res.MsgResponses = []*codectypes.Any{any}

		return res, nil
	}
}

func getIntFromStore(store sdk.KVStore, key []byte) int64 {
	bz := store.Get(key)
	if len(bz) == 0 {
		return 0
	}
	i, err := binary.ReadVarint(bytes.NewBuffer(bz))
	if err != nil {
		panic(err)
	}
	return i
}

func setIntOnStore(store sdk.KVStore, key []byte, i int64) {
	bz := make([]byte, 8)
	n := binary.PutVarint(bz, i)
	store.Set(key, bz[:n])
}

// check counter matches what's in store.
// increment and store
func incrementingCounter(t *testing.T, store sdk.KVStore, counterKey []byte, counter int64) (*sdk.Result, error) {
	storedCounter := getIntFromStore(store, counterKey)
	require.Equal(t, storedCounter, counter)
	setIntOnStore(store, counterKey, counter+1)
	return &sdk.Result{}, nil
}

// ---------------------------------------------------------------------
// Tx processing - CheckTx, DeliverTx, SimulateTx.
// These tests use the serialized tx as input, while most others will use the
// Check(), Deliver(), Simulate() methods directly.
// Ensure that Check/Deliver/Simulate work as expected with the store.

// Test that successive CheckTx can see each others' effects
// on the store within a block, and that the CheckTx state
// gets reset to the latest committed state during Commit
func TestCheckTx(t *testing.T) {
	// This ante handler reads the key and checks that the value matches the current counter.
	// This ensures changes to the kvstore persist across successive CheckTx.
	counterKey := []byte("counter-key")

	txHandlerOpt := baseapp.AppOptionFunc(func(bapp *baseapp.BaseApp) {
		legacyRouter := middleware.NewLegacyRouter()
		// TODO: can remove this once CheckTx doesnt process msgs.
		legacyRouter.AddRoute(sdk.NewRoute(routeMsgCounter, func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
			return &sdk.Result{}, nil
		}))
		txHandler := testTxHandler(
			middleware.TxHandlerOptions{
				LegacyRouter:     legacyRouter,
				MsgServiceRouter: middleware.NewMsgServiceRouter(encCfg.InterfaceRegistry),
				TxDecoder:        testTxDecoder(encCfg.Amino),
			},
			customHandlerTxTest(t, capKey1, counterKey),
		)
		bapp.SetTxHandler(txHandler)
	})

	app := setupBaseApp(t, txHandlerOpt)

	nTxs := int64(5)
	app.InitChain(abci.RequestInitChain{})

	for i := int64(0); i < nTxs; i++ {
		tx := newTxCounter(i, 0) // no messages
		txBytes, err := encCfg.Amino.Marshal(tx)
		require.NoError(t, err)
		r := app.CheckTx(abci.RequestCheckTx{Tx: txBytes})
		require.Empty(t, r.GetEvents())
		require.True(t, r.IsOK(), fmt.Sprintf("%+v", r))
	}

	checkStateStore := app.CheckState().Context().KVStore(capKey1)
	storedCounter := getIntFromStore(checkStateStore, counterKey)

	// Ensure storedCounter
	require.Equal(t, nTxs, storedCounter)

	// If a block is committed, CheckTx state should be reset.
	header := tmproto.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header, Hash: []byte("hash")})

	require.NotNil(t, app.CheckState().Context().BlockGasMeter(), "block gas meter should have been set to checkState")
	require.NotEmpty(t, app.CheckState().Context().HeaderHash())

	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	checkStateStore = app.CheckState().Context().KVStore(capKey1)
	storedBytes := checkStateStore.Get(counterKey)
	require.Nil(t, storedBytes)
}

// Test that successive DeliverTx can see each others' effects
// on the store, both within and across blocks.
func TestDeliverTx(t *testing.T) {
	// test increments in the post txHandler
	anteKey := []byte("ante-key")
	// test increments in the handler
	deliverKey := []byte("deliver-key")
	txHandlerOpt := baseapp.AppOptionFunc(func(bapp *baseapp.BaseApp) {
		legacyRouter := middleware.NewLegacyRouter()
		r := sdk.NewRoute(routeMsgCounter, handlerMsgCounter(t, capKey1, deliverKey))
		legacyRouter.AddRoute(r)
		txHandler := testTxHandler(
			middleware.TxHandlerOptions{
				LegacyRouter:     legacyRouter,
				MsgServiceRouter: middleware.NewMsgServiceRouter(encCfg.InterfaceRegistry),
				TxDecoder:        testTxDecoder(encCfg.Amino),
			},
			customHandlerTxTest(t, capKey1, anteKey),
		)
		bapp.SetTxHandler(txHandler)
	})
	app := setupBaseApp(t, txHandlerOpt)
	app.InitChain(abci.RequestInitChain{})

	nBlocks := 3
	txPerHeight := 5

	for blockN := 0; blockN < nBlocks; blockN++ {
		header := tmproto.Header{Height: int64(blockN) + 1}
		app.BeginBlock(abci.RequestBeginBlock{Header: header})

		for i := 0; i < txPerHeight; i++ {
			counter := int64(blockN*txPerHeight + i)
			tx := newTxCounter(counter, counter)

			txBytes, err := encCfg.Amino.Marshal(tx)
			require.NoError(t, err)

			res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
			require.True(t, res.IsOK(), fmt.Sprintf("%v", res))
			events := res.GetEvents()
			require.Len(t, events, 3, "should contain ante handler, message type and counter events respectively")
			require.Equal(t, sdk.MarkEventsToIndex(counterEvent("post_handlers", counter).ToABCIEvents(), map[string]struct{}{})[0], events[0], "ante handler event")
			require.Equal(t, sdk.MarkEventsToIndex(counterEvent(sdk.EventTypeMessage, counter).ToABCIEvents(), map[string]struct{}{})[0], events[2], "msg handler update counter event")
		}

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}
}

// Number of messages doesn't matter to CheckTx.
func TestMultiMsgCheckTx(t *testing.T) {
	// TODO: ensure we get the same results
	// with one message or many
}

// One call to DeliverTx should process all the messages, in order.
func TestMultiMsgDeliverTx(t *testing.T) {
	// increment the tx counter
	anteKey := []byte("ante-key")
	// increment the msg counter
	deliverKey := []byte("deliver-key")
	deliverKey2 := []byte("deliver-key2")

	txHandlerOpt := func(bapp *baseapp.BaseApp) {
		legacyRouter := middleware.NewLegacyRouter()
		r1 := sdk.NewRoute(routeMsgCounter, handlerMsgCounter(t, capKey1, deliverKey))
		r2 := sdk.NewRoute(routeMsgCounter2, handlerMsgCounter(t, capKey1, deliverKey2))
		legacyRouter.AddRoute(r1)
		legacyRouter.AddRoute(r2)
		txHandler := testTxHandler(
			middleware.TxHandlerOptions{
				LegacyRouter:     legacyRouter,
				MsgServiceRouter: middleware.NewMsgServiceRouter(encCfg.InterfaceRegistry),
				TxDecoder:        testTxDecoder(encCfg.Amino),
			},
			customHandlerTxTest(t, capKey1, anteKey),
		)
		bapp.SetTxHandler(txHandler)
	}
	app := setupBaseApp(t, baseapp.AppOptionFunc(txHandlerOpt))

	// run a multi-msg tx
	// with all msgs the same route

	header := tmproto.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	tx := newTxCounter(0, 0, 1, 2)
	txBytes, err := encCfg.Amino.Marshal(tx)
	require.NoError(t, err)
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.True(t, res.IsOK(), fmt.Sprintf("%v", res))

	store := app.DeliverState().Context().KVStore(capKey1)

	// tx counter only incremented once
	txCounter := getIntFromStore(store, anteKey)
	require.Equal(t, int64(1), txCounter)

	// msg counter incremented three times
	msgCounter := getIntFromStore(store, deliverKey)
	require.Equal(t, int64(3), msgCounter)

	// replace the second message with a msgCounter2

	tx = newTxCounter(1, 3)
	tx.Msgs = append(tx.Msgs, msgCounter2{0})
	tx.Msgs = append(tx.Msgs, msgCounter2{1})
	txBytes, err = encCfg.Amino.Marshal(tx)
	require.NoError(t, err)
	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.True(t, res.IsOK(), fmt.Sprintf("%v", res))

	store = app.DeliverState().Context().KVStore(capKey1)

	// tx counter only incremented once
	txCounter = getIntFromStore(store, anteKey)
	require.Equal(t, int64(2), txCounter)

	// original counter increments by one
	// new counter increments by two
	msgCounter = getIntFromStore(store, deliverKey)
	require.Equal(t, int64(4), msgCounter)
	msgCounter2 := getIntFromStore(store, deliverKey2)
	require.Equal(t, int64(2), msgCounter2)
}

// Interleave calls to Check and Deliver and ensure
// that there is no cross-talk. Check sees results of the previous Check calls
// and Deliver sees that of the previous Deliver calls, but they don't see eachother.
func TestConcurrentCheckDeliver(t *testing.T) {
	// TODO
}

// Simulate a transaction that uses gas to compute the gas.
// Simulate() and Query("/app/simulate", txBytes) should give
// the same results.
func TestSimulateTx(t *testing.T) {
	gasConsumed := uint64(5)

	txHandlerOpt := func(bapp *baseapp.BaseApp) {
		legacyRouter := middleware.NewLegacyRouter()
		r := sdk.NewRoute(routeMsgCounter, func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
			ctx.GasMeter().ConsumeGas(gasConsumed, "test")
			// Return dummy MsgResponse for msgCounter.
			any, err := codectypes.NewAnyWithValue(&testdata.Dog{})
			if err != nil {
				return nil, err
			}

			return &sdk.Result{
				MsgResponses: []*codectypes.Any{any},
			}, nil
		})
		legacyRouter.AddRoute(r)
		txHandler := testTxHandler(
			middleware.TxHandlerOptions{
				LegacyRouter:     legacyRouter,
				MsgServiceRouter: middleware.NewMsgServiceRouter(encCfg.InterfaceRegistry),
				TxDecoder:        testTxDecoder(encCfg.Amino),
			},
			func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) { return ctx, nil },
		)
		bapp.SetTxHandler(txHandler)
	}
	app := setupBaseApp(t, baseapp.AppOptionFunc(txHandlerOpt))

	app.InitChain(abci.RequestInitChain{})

	nBlocks := 3
	for blockN := 0; blockN < nBlocks; blockN++ {
		count := int64(blockN + 1)
		header := tmproto.Header{Height: count}
		app.BeginBlock(abci.RequestBeginBlock{Header: header})

		tx := newTxCounter(count, count)
		tx.GasLimit = gasConsumed
		txBytes, err := encCfg.Amino.Marshal(tx)
		require.Nil(t, err)

		// simulate a message, check gas reported
		gInfo, result, err := app.Simulate(txBytes)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, gasConsumed, gInfo.GasUsed)

		// simulate again, same result
		gInfo, result, err = app.Simulate(txBytes)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, gasConsumed, gInfo.GasUsed)

		// simulate by calling Query with encoded tx
		query := abci.RequestQuery{
			Path: "/app/simulate",
			Data: txBytes,
		}
		queryResult := app.Query(query)
		require.True(t, queryResult.IsOK(), queryResult.Log)

		var simRes sdk.SimulationResponse
		require.NoError(t, jsonpb.Unmarshal(strings.NewReader(string(queryResult.Value)), &simRes))

		require.Equal(t, gInfo, simRes.GasInfo)
		require.Equal(t, result.Log, simRes.Result.Log)
		require.Equal(t, result.Events, simRes.Result.Events)
		require.True(t, bytes.Equal(result.Data, simRes.Result.Data))

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}
}

func TestRunInvalidTransaction(t *testing.T) {
	txHandlerOpt := func(bapp *baseapp.BaseApp) {
		legacyRouter := middleware.NewLegacyRouter()
		r := sdk.NewRoute(routeMsgCounter, func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
			any, err := codectypes.NewAnyWithValue(msg)
			if err != nil {
				return nil, err
			}

			return &sdk.Result{
				MsgResponses: []*codectypes.Any{any},
			}, nil
		})
		legacyRouter.AddRoute(r)
		txHandler := testTxHandler(
			middleware.TxHandlerOptions{
				LegacyRouter:     legacyRouter,
				MsgServiceRouter: middleware.NewMsgServiceRouter(encCfg.InterfaceRegistry),
				TxDecoder:        testTxDecoder(encCfg.Amino),
			},
			func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
				return
			},
		)
		bapp.SetTxHandler(txHandler)
	}
	app := setupBaseApp(t, baseapp.AppOptionFunc(txHandlerOpt))

	header := tmproto.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	// transaction with no messages
	{
		emptyTx := &txTest{}
		_, result, err := app.SimDeliver(aminoTxEncoder(encCfg.Amino), emptyTx)
		require.Nil(t, result)

		space, code, _ := sdkerrors.ABCIInfo(err, false)
		require.EqualValues(t, sdkerrors.ErrInvalidRequest.Codespace(), space, err)
		require.EqualValues(t, sdkerrors.ErrInvalidRequest.ABCICode(), code, err)
	}

	// transaction where ValidateBasic fails
	{
		testCases := []struct {
			tx   txTest
			fail bool
		}{
			{newTxCounter(0, 0), false},
			{newTxCounter(-1, 0), false},
			{newTxCounter(100, 100), false},
			{newTxCounter(100, 5, 4, 3, 2, 1), false},

			{newTxCounter(0, -1), true},
			{newTxCounter(0, 1, -2), true},
			{newTxCounter(0, 1, 2, -10, 5), true},
		}

		for _, testCase := range testCases {
			tx := testCase.tx
			_, _, err := app.SimDeliver(aminoTxEncoder(encCfg.Amino), tx)

			if testCase.fail {
				require.Error(t, err)

				space, code, _ := sdkerrors.ABCIInfo(err, false)
				require.EqualValues(t, sdkerrors.ErrInvalidSequence.Codespace(), space, err)
				require.EqualValues(t, sdkerrors.ErrInvalidSequence.ABCICode(), code, err)
			} else {
				require.NoError(t, err)
			}
		}
	}

	// transaction with no known route
	{
		unknownRouteTx := txTest{[]sdk.Msg{&msgNoRoute{}}, 0, false, math.MaxUint64}
		_, result, err := app.SimDeliver(aminoTxEncoder(encCfg.Amino), unknownRouteTx)
		require.Error(t, err)
		require.Nil(t, result)

		space, code, _ := sdkerrors.ABCIInfo(err, false)
		require.EqualValues(t, sdkerrors.ErrUnknownRequest.Codespace(), space, err)
		require.EqualValues(t, sdkerrors.ErrUnknownRequest.ABCICode(), code, err)

		unknownRouteTx = txTest{[]sdk.Msg{&msgCounter{}, &msgNoRoute{}}, 0, false, math.MaxUint64}
		_, result, err = app.SimDeliver(aminoTxEncoder(encCfg.Amino), unknownRouteTx)
		require.Error(t, err)
		require.Nil(t, result)

		space, code, _ = sdkerrors.ABCIInfo(err, false)
		require.EqualValues(t, sdkerrors.ErrUnknownRequest.Codespace(), space, err)
		require.EqualValues(t, sdkerrors.ErrUnknownRequest.ABCICode(), code, err)
	}

	// Transaction with an unregistered message
	{
		tx := newTxCounter(0, 0)
		tx.Msgs = append(tx.Msgs, &msgNoDecode{})

		// new codec so we can encode the tx, but we shouldn't be able to decode,
		// because baseapp's codec is not aware of msgNoDecode.
		newCdc := codec.NewLegacyAmino()
		sdk.RegisterLegacyAminoCodec(newCdc) // register Tx, Msg
		registerTestCodec(newCdc)
		newCdc.RegisterConcrete(&msgNoDecode{}, "cosmos-sdk/baseapp/msgNoDecode", nil)

		txBytes, err := newCdc.Marshal(tx)
		require.NoError(t, err)

		res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
		require.EqualValues(t, sdkerrors.ErrTxDecode.ABCICode(), res.Code)
		require.EqualValues(t, sdkerrors.ErrTxDecode.Codespace(), res.Codespace)
	}
}

// Test that transactions exceeding gas limits fail
func TestTxGasLimits(t *testing.T) {
	gasGranted := uint64(10)

	ante := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		count := tx.(txTest).Counter
		ctx.GasMeter().ConsumeGas(uint64(count), "counter-ante")
		return ctx, nil
	}

	txHandlerOpt := func(bapp *baseapp.BaseApp) {
		legacyRouter := middleware.NewLegacyRouter()
		r := sdk.NewRoute(routeMsgCounter, func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
			count := msg.(*msgCounter).Counter
			ctx.GasMeter().ConsumeGas(uint64(count), "counter-handler")
			any, err := codectypes.NewAnyWithValue(msg)
			if err != nil {
				return nil, err
			}

			return &sdk.Result{
				MsgResponses: []*codectypes.Any{any},
			}, nil
		})
		legacyRouter.AddRoute(r)
		txHandler := testTxHandler(
			middleware.TxHandlerOptions{
				LegacyRouter:     legacyRouter,
				MsgServiceRouter: middleware.NewMsgServiceRouter(encCfg.InterfaceRegistry),
				TxDecoder:        testTxDecoder(encCfg.Amino),
			},
			ante,
		)

		bapp.SetTxHandler(txHandler)
	}
	app := setupBaseApp(t, baseapp.AppOptionFunc(txHandlerOpt))

	header := tmproto.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	testCases := []struct {
		tx      txTest
		gasUsed uint64
		fail    bool
	}{
		{newTxCounter(0, 0), 0, false},
		{newTxCounter(1, 1), 2, false},
		{newTxCounter(9, 1), 10, false},
		{newTxCounter(1, 9), 10, false},
		{newTxCounter(10, 0), 10, false},
		{newTxCounter(0, 10), 10, false},
		{newTxCounter(0, 8, 2), 10, false},
		{newTxCounter(0, 5, 1, 1, 1, 1, 1), 10, false},
		{newTxCounter(0, 5, 1, 1, 1, 1), 9, false},

		{newTxCounter(9, 2), 11, true},
		{newTxCounter(2, 9), 11, true},
		{newTxCounter(9, 1, 1), 11, true},
		{newTxCounter(1, 8, 1, 1), 11, true},
		{newTxCounter(11, 0), 11, true},
		{newTxCounter(0, 11), 11, true},
		{newTxCounter(0, 5, 11), 16, true},
	}

	for i, tc := range testCases {
		tx := tc.tx
		tx.GasLimit = gasGranted
		gInfo, result, err := app.SimDeliver(aminoTxEncoder(encCfg.Amino), tx)

		// check gas used and wanted
		require.Equal(t, tc.gasUsed, gInfo.GasUsed, fmt.Sprintf("tc #%d; gas: %v, result: %v, err: %s", i, gInfo, result, err))

		// check for out of gas
		if !tc.fail {
			require.NotNil(t, result, fmt.Sprintf("%d: %v, %v", i, tc, err))
		} else {
			require.Error(t, err)
			require.Nil(t, result)

			space, code, _ := sdkerrors.ABCIInfo(err, false)
			require.EqualValues(t, sdkerrors.ErrOutOfGas.Codespace(), space, err)
			require.EqualValues(t, sdkerrors.ErrOutOfGas.ABCICode(), code, err)
		}
	}
}

// Test that transactions exceeding gas limits fail
func TestMaxBlockGasLimits(t *testing.T) {
	gasGranted := uint64(10)
	ante := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		count := tx.(txTest).Counter
		ctx.GasMeter().ConsumeGas(uint64(count), "counter-ante")

		return ctx, nil
	}

	txHandlerOpt := func(bapp *baseapp.BaseApp) {
		legacyRouter := middleware.NewLegacyRouter()
		r := sdk.NewRoute(routeMsgCounter, func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
			count := msg.(*msgCounter).Counter
			ctx.GasMeter().ConsumeGas(uint64(count), "counter-handler")

			any, err := codectypes.NewAnyWithValue(msg)
			if err != nil {
				return nil, err
			}

			return &sdk.Result{
				MsgResponses: []*codectypes.Any{any},
			}, nil
		})
		legacyRouter.AddRoute(r)
		txHandler := testTxHandler(
			middleware.TxHandlerOptions{
				LegacyRouter:     legacyRouter,
				MsgServiceRouter: middleware.NewMsgServiceRouter(encCfg.InterfaceRegistry),
				TxDecoder:        testTxDecoder(encCfg.Amino),
			},
			ante,
		)
		bapp.SetTxHandler(txHandler)
	}
	app := setupBaseApp(t, baseapp.AppOptionFunc(txHandlerOpt))

	app.InitChain(abci.RequestInitChain{
		ConsensusParams: &tmproto.ConsensusParams{
			Block: &tmproto.BlockParams{
				MaxGas: 100,
			},
		},
	})

	testCases := []struct {
		tx                txTest
		numDelivers       int
		gasUsedPerDeliver uint64
		fail              bool
		failAfterDeliver  int
	}{
		{newTxCounter(0, 0), 0, 0, false, 0},
		{newTxCounter(9, 1), 2, 10, false, 0},
		{newTxCounter(10, 0), 3, 10, false, 0},
		{newTxCounter(10, 0), 10, 10, false, 0},
		{newTxCounter(2, 7), 11, 9, false, 0},
		{newTxCounter(10, 0), 10, 10, false, 0}, // hit the limit but pass

		{newTxCounter(10, 0), 11, 10, true, 10},
		{newTxCounter(10, 0), 15, 10, true, 10},
		{newTxCounter(9, 0), 12, 9, true, 11}, // fly past the limit
	}

	for i, tc := range testCases {
		tx := tc.tx
		tx.GasLimit = gasGranted

		// reset the block gas
		header := tmproto.Header{Height: app.LastBlockHeight() + 1}
		app.BeginBlock(abci.RequestBeginBlock{Header: header})

		// execute the transaction multiple times
		for j := 0; j < tc.numDelivers; j++ {
			_, result, err := app.SimDeliver(aminoTxEncoder(encCfg.Amino), tx)

			ctx := app.DeliverState().Context()

			// check for failed transactions
			if tc.fail && (j+1) > tc.failAfterDeliver {
				require.Error(t, err, fmt.Sprintf("tc #%d; result: %v, err: %s", i, result, err))
				require.Nil(t, result, fmt.Sprintf("tc #%d; result: %v, err: %s", i, result, err))

				space, code, _ := sdkerrors.ABCIInfo(err, false)
				require.EqualValues(t, sdkerrors.ErrOutOfGas.Codespace(), space, err)
				require.EqualValues(t, sdkerrors.ErrOutOfGas.ABCICode(), code, err)
				require.True(t, ctx.BlockGasMeter().IsOutOfGas())
			} else {
				// check gas used and wanted
				blockGasUsed := ctx.BlockGasMeter().GasConsumed()
				expBlockGasUsed := tc.gasUsedPerDeliver * uint64(j+1)
				require.Equal(
					t, expBlockGasUsed, blockGasUsed,
					fmt.Sprintf("%d,%d: %v, %v, %v, %v", i, j, tc, expBlockGasUsed, blockGasUsed, result),
				)

				require.NotNil(t, result, fmt.Sprintf("tc #%d; currDeliver: %d, result: %v, err: %s", i, j, result, err))
				require.False(t, ctx.BlockGasMeter().IsPastLimit())
			}
		}
	}
}

func TestBaseAppMiddleware(t *testing.T) {
	anteKey := []byte("ante-key")
	deliverKey := []byte("deliver-key")

	txHandlerOpt := func(bapp *baseapp.BaseApp) {
		legacyRouter := middleware.NewLegacyRouter()
		r := sdk.NewRoute(routeMsgCounter, handlerMsgCounter(t, capKey1, deliverKey))
		legacyRouter.AddRoute(r)
		txHandler := testTxHandler(
			middleware.TxHandlerOptions{
				LegacyRouter:     legacyRouter,
				MsgServiceRouter: middleware.NewMsgServiceRouter(encCfg.InterfaceRegistry),
				TxDecoder:        testTxDecoder(encCfg.Amino),
			},
			customHandlerTxTest(t, capKey1, anteKey),
		)
		bapp.SetTxHandler(txHandler)
	}
	app := setupBaseApp(t, baseapp.AppOptionFunc(txHandlerOpt))

	app.InitChain(abci.RequestInitChain{})

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	// execute a tx that will fail ante handler execution
	//
	// NOTE: State should not be mutated here. This will be implicitly checked by
	// the next txs ante handler execution (customHandlerTxTest).
	tx := newTxCounter(0, 0)
	tx.setFailOnAnte(true)
	txBytes, err := encCfg.Amino.Marshal(tx)
	require.NoError(t, err)
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.Empty(t, res.Events)
	require.False(t, res.IsOK(), fmt.Sprintf("%v", res))

	ctx := app.DeliverState().Context()
	store := ctx.KVStore(capKey1)
	require.Equal(t, int64(0), getIntFromStore(store, anteKey))

	// execute at tx that will pass the ante handler (the checkTx state should
	// mutate) but will fail the message handler
	tx = newTxCounter(0, 0)
	tx.setFailOnHandler(true)

	txBytes, err = encCfg.Amino.Marshal(tx)
	require.NoError(t, err)

	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.Empty(t, res.Events)
	require.False(t, res.IsOK(), fmt.Sprintf("%v", res))

	ctx = app.DeliverState().Context()
	store = ctx.KVStore(capKey1)
	require.Equal(t, int64(1), getIntFromStore(store, anteKey))
	require.Equal(t, int64(0), getIntFromStore(store, deliverKey))

	// execute a successful ante handler and message execution where state is
	// implicitly checked by previous tx executions
	tx = newTxCounter(1, 0)

	txBytes, err = encCfg.Amino.Marshal(tx)
	require.NoError(t, err)

	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.NotEmpty(t, res.Events)
	require.True(t, res.IsOK(), fmt.Sprintf("%v", res))

	ctx = app.DeliverState().Context()
	store = ctx.KVStore(capKey1)
	require.Equal(t, int64(2), getIntFromStore(store, anteKey))
	require.Equal(t, int64(1), getIntFromStore(store, deliverKey))

	// commit
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()
}

func TestGasConsumptionBadTx(t *testing.T) {
	gasWanted := uint64(5)
	ante := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		txTest := tx.(txTest)
		ctx.GasMeter().ConsumeGas(uint64(txTest.Counter), "counter-ante")
		if txTest.FailOnAnte {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "ante handler failure")
		}

		return ctx, nil
	}

	txHandlerOpt := func(bapp *baseapp.BaseApp) {
		legacyRouter := middleware.NewLegacyRouter()
		r := sdk.NewRoute(routeMsgCounter, func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
			count := msg.(*msgCounter).Counter
			ctx.GasMeter().ConsumeGas(uint64(count), "counter-handler")
			return &sdk.Result{}, nil
		})
		legacyRouter.AddRoute(r)
		txHandler := testTxHandler(
			middleware.TxHandlerOptions{
				LegacyRouter:     legacyRouter,
				MsgServiceRouter: middleware.NewMsgServiceRouter(encCfg.InterfaceRegistry),
				TxDecoder:        testTxDecoder(encCfg.Amino),
			},
			ante,
		)
		bapp.SetTxHandler(txHandler)
	}
	app := setupBaseApp(t, baseapp.AppOptionFunc(txHandlerOpt))

	app.InitChain(abci.RequestInitChain{
		ConsensusParams: &tmproto.ConsensusParams{
			Block: &tmproto.BlockParams{
				MaxGas: 9,
			},
		},
	})

	app.InitChain(abci.RequestInitChain{})

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	tx := newTxCounter(5, 0)
	tx.GasLimit = gasWanted
	tx.setFailOnAnte(true)
	txBytes, err := encCfg.Amino.Marshal(tx)
	require.NoError(t, err)

	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.False(t, res.IsOK(), fmt.Sprintf("%v", res))

	// require next tx to fail due to black gas limit
	tx = newTxCounter(5, 0)
	txBytes, err = encCfg.Amino.Marshal(tx)
	require.NoError(t, err)

	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.False(t, res.IsOK(), fmt.Sprintf("%v", res))
}

// Test that we can only query from the latest committed state.
func TestQuery(t *testing.T) {
	key, value := []byte("hello"), []byte("goodbye")

	txHandlerOpt := func(bapp *baseapp.BaseApp) {
		legacyRouter := middleware.NewLegacyRouter()
		r := sdk.NewRoute(routeMsgCounter, func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
			store := ctx.KVStore(capKey1)
			store.Set(key, value)

			any, err := codectypes.NewAnyWithValue(msg)
			if err != nil {
				return nil, err
			}
			return &sdk.Result{
				MsgResponses: []*codectypes.Any{any},
			}, nil
		})
		legacyRouter.AddRoute(r)
		txHandler := testTxHandler(
			middleware.TxHandlerOptions{
				LegacyRouter:     legacyRouter,
				MsgServiceRouter: middleware.NewMsgServiceRouter(encCfg.InterfaceRegistry),
				TxDecoder:        testTxDecoder(encCfg.Amino),
			},
			func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
				store := ctx.KVStore(capKey1)
				store.Set(key, value)
				return
			},
		)
		bapp.SetTxHandler(txHandler)
	}
	app := setupBaseApp(t, baseapp.AppOptionFunc(txHandlerOpt))

	app.InitChain(abci.RequestInitChain{})

	// NOTE: "/store/key1" tells us KVStore
	// and the final "/key" says to use the data as the
	// key in the given KVStore ...
	query := abci.RequestQuery{
		Path: "/store/key1/key",
		Data: key,
	}
	tx := newTxCounter(0, 0)

	// query is empty before we do anything
	res := app.Query(query)
	require.Equal(t, 0, len(res.Value))

	// query is still empty after a CheckTx
	_, resTx, err := app.SimCheck(aminoTxEncoder(encCfg.Amino), tx)
	require.NoError(t, err)
	require.NotNil(t, resTx)
	res = app.Query(query)
	require.Equal(t, 0, len(res.Value))

	// query is still empty after a DeliverTx before we commit
	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	_, resTx, err = app.SimDeliver(aminoTxEncoder(encCfg.Amino), tx)
	require.NoError(t, err)
	require.NotNil(t, resTx)
	res = app.Query(query)
	require.Equal(t, 0, len(res.Value))

	// query returns correct value after Commit
	app.Commit()
	res = app.Query(query)
	require.Equal(t, value, res.Value)
}

func TestGRPCQuery(t *testing.T) {
	grpcQueryOpt := func(bapp *baseapp.BaseApp) {
		testdata.RegisterQueryServer(
			bapp.GRPCQueryRouter(),
			testdata.QueryImpl{},
		)
	}

	app := setupBaseApp(t, baseapp.AppOptionFunc(grpcQueryOpt))
	app.GRPCQueryRouter().SetInterfaceRegistry(codectypes.NewInterfaceRegistry())

	app.InitChain(abci.RequestInitChain{})
	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	app.Commit()

	req := testdata.SayHelloRequest{Name: "foo"}
	reqBz, err := req.Marshal()
	require.NoError(t, err)

	reqQuery := abci.RequestQuery{
		Data: reqBz,
		Path: "/testdata.Query/SayHello",
	}

	resQuery := app.Query(reqQuery)

	require.Equal(t, abci.CodeTypeOK, resQuery.Code, resQuery)

	var res testdata.SayHelloResponse
	err = res.Unmarshal(resQuery.Value)
	require.NoError(t, err)
	require.Equal(t, "Hello foo!", res.Greeting)
}

func TestGRPCQueryPulsar(t *testing.T) {
	grpcQueryOpt := func(bapp *baseapp.BaseApp) {
		testdata_pulsar.RegisterQueryServer(
			bapp.GRPCQueryRouter(),
			testdata_pulsar.QueryImpl{},
		)
	}

	app := setupBaseApp(t, baseapp.AppOptionFunc(grpcQueryOpt))
	app.GRPCQueryRouter().SetInterfaceRegistry(codectypes.NewInterfaceRegistry())

	app.InitChain(abci.RequestInitChain{})
	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	app.Commit()

	req := &testdata_pulsar.SayHelloRequest{Name: "foo"}
	reqBz, err := proto.Marshal(req)
	require.NoError(t, err)

	reqQuery := abci.RequestQuery{
		Data: reqBz,
		Path: "/testdata.Query/SayHello",
	}

	resQuery := app.Query(reqQuery)

	require.Equal(t, abci.CodeTypeOK, resQuery.Code, resQuery)

	var res testdata_pulsar.SayHelloResponse
	err = proto.Unmarshal(resQuery.Value, &res)
	require.NoError(t, err)
	require.Equal(t, "Hello foo!", res.Greeting)
}

// Test p2p filter queries
func TestP2PQuery(t *testing.T) {
	addrPeerFilterOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAddrPeerFilter(func(addrport string) abci.ResponseQuery {
			require.Equal(t, "1.1.1.1:8000", addrport)
			return abci.ResponseQuery{Code: uint32(3)}
		})
	}

	idPeerFilterOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetIDPeerFilter(func(id string) abci.ResponseQuery {
			require.Equal(t, "testid", id)
			return abci.ResponseQuery{Code: uint32(4)}
		})
	}

	app := setupBaseApp(t, baseapp.AppOptionFunc(addrPeerFilterOpt), baseapp.AppOptionFunc(idPeerFilterOpt))

	addrQuery := abci.RequestQuery{
		Path: "/p2p/filter/addr/1.1.1.1:8000",
	}
	res := app.Query(addrQuery)
	require.Equal(t, uint32(3), res.Code)

	idQuery := abci.RequestQuery{
		Path: "/p2p/filter/id/testid",
	}
	res = app.Query(idQuery)
	require.Equal(t, uint32(4), res.Code)
}

func TestGetMaximumBlockGas(t *testing.T) {
	app := setupBaseApp(t)
	app.InitChain(abci.RequestInitChain{})
	ctx := app.NewContext(true, tmproto.Header{})

	app.StoreConsensusParams(ctx, &tmproto.ConsensusParams{Block: &tmproto.BlockParams{MaxGas: 0}})
	require.Equal(t, uint64(0), app.GetMaximumBlockGas(ctx))

	app.StoreConsensusParams(ctx, &tmproto.ConsensusParams{Block: &tmproto.BlockParams{MaxGas: -1}})
	require.Equal(t, uint64(0), app.GetMaximumBlockGas(ctx))

	app.StoreConsensusParams(ctx, &tmproto.ConsensusParams{Block: &tmproto.BlockParams{MaxGas: 5000000}})
	require.Equal(t, uint64(5000000), app.GetMaximumBlockGas(ctx))

	app.StoreConsensusParams(ctx, &tmproto.ConsensusParams{Block: &tmproto.BlockParams{MaxGas: -5000000}})
	require.Panics(t, func() { app.GetMaximumBlockGas(ctx) })
}

func TestListSnapshots(t *testing.T) {
	t.Skip("disabled pending MultiStore snapshot implementation")

	app, teardown := setupBaseAppWithSnapshots(t, 5, 4)
	defer teardown()

	resp := app.ListSnapshots(abci.RequestListSnapshots{})
	for _, s := range resp.Snapshots {
		assert.NotEmpty(t, s.Hash)
		assert.NotEmpty(t, s.Metadata)
		s.Hash = nil
		s.Metadata = nil
	}
	assert.Equal(t, abci.ResponseListSnapshots{Snapshots: []*abci.Snapshot{
		{Height: 4, Format: 1, Chunks: 2},
		{Height: 2, Format: 1, Chunks: 1},
	}}, resp)
}

func TestLoadSnapshotChunk(t *testing.T) {
	t.Skip("disabled pending MultiStore snapshot implementation")

	app, teardown := setupBaseAppWithSnapshots(t, 2, 5)
	defer teardown()

	testcases := map[string]struct {
		height      uint64
		format      uint32
		chunk       uint32
		expectEmpty bool
	}{
		"Existing snapshot": {2, 1, 1, false},
		"Missing height":    {100, 1, 1, true},
		"Missing format":    {2, 2, 1, true},
		"Missing chunk":     {2, 1, 9, true},
		"Zero height":       {0, 1, 1, true},
		"Zero format":       {2, 0, 1, true},
		"Zero chunk":        {2, 1, 0, false},
	}
	for name, tc := range testcases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			resp := app.LoadSnapshotChunk(abci.RequestLoadSnapshotChunk{
				Height: tc.height,
				Format: tc.format,
				Chunk:  tc.chunk,
			})
			if tc.expectEmpty {
				assert.Equal(t, abci.ResponseLoadSnapshotChunk{}, resp)
				return
			}
			assert.NotEmpty(t, resp.Chunk)
		})
	}
}

func TestOfferSnapshot_Errors(t *testing.T) {
	t.Skip("disabled pending MultiStore snapshot implementation")

	// Set up app before test cases, since it's fairly expensive.
	app, teardown := setupBaseAppWithSnapshots(t, 0, 0)
	defer teardown()

	m := snapshottypes.Metadata{ChunkHashes: [][]byte{{1}, {2}, {3}}}
	metadata, err := m.Marshal()
	require.NoError(t, err)
	hash := []byte{1, 2, 3}

	testcases := map[string]struct {
		snapshot *abci.Snapshot
		result   abci.ResponseOfferSnapshot_Result
	}{
		"nil snapshot": {nil, abci.ResponseOfferSnapshot_REJECT},
		"invalid format": {&abci.Snapshot{
			Height: 1, Format: 9, Chunks: 3, Hash: hash, Metadata: metadata,
		}, abci.ResponseOfferSnapshot_REJECT_FORMAT},
		"incorrect chunk count": {&abci.Snapshot{
			Height: 1, Format: 1, Chunks: 2, Hash: hash, Metadata: metadata,
		}, abci.ResponseOfferSnapshot_REJECT},
		"no chunks": {&abci.Snapshot{
			Height: 1, Format: 1, Chunks: 0, Hash: hash, Metadata: metadata,
		}, abci.ResponseOfferSnapshot_REJECT},
		"invalid metadata serialization": {&abci.Snapshot{
			Height: 1, Format: 1, Chunks: 0, Hash: hash, Metadata: []byte{3, 1, 4},
		}, abci.ResponseOfferSnapshot_REJECT},
	}
	for name, tc := range testcases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			resp := app.OfferSnapshot(abci.RequestOfferSnapshot{Snapshot: tc.snapshot})
			assert.Equal(t, tc.result, resp.Result)
		})
	}

	// Offering a snapshot after one has been accepted should error
	resp := app.OfferSnapshot(abci.RequestOfferSnapshot{Snapshot: &abci.Snapshot{
		Height:   1,
		Format:   snapshottypes.CurrentFormat,
		Chunks:   3,
		Hash:     []byte{1, 2, 3},
		Metadata: metadata,
	}})
	require.Equal(t, abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_ACCEPT}, resp)

	resp = app.OfferSnapshot(abci.RequestOfferSnapshot{Snapshot: &abci.Snapshot{
		Height:   2,
		Format:   snapshottypes.CurrentFormat,
		Chunks:   3,
		Hash:     []byte{1, 2, 3},
		Metadata: metadata,
	}})
	require.Equal(t, abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_ABORT}, resp)
}

func TestApplySnapshotChunk(t *testing.T) {
	t.Skip("disabled pending MultiStore snapshot implementation")

	source, teardown := setupBaseAppWithSnapshots(t, 4, 10)
	defer teardown()

	target, teardown := setupBaseAppWithSnapshots(t, 0, 0)
	defer teardown()

	// Fetch latest snapshot to restore
	respList := source.ListSnapshots(abci.RequestListSnapshots{})
	require.NotEmpty(t, respList.Snapshots)
	snapshot := respList.Snapshots[0]

	// Make sure the snapshot has at least 3 chunks
	require.GreaterOrEqual(t, snapshot.Chunks, uint32(3), "Not enough snapshot chunks")

	// Begin a snapshot restoration in the target
	respOffer := target.OfferSnapshot(abci.RequestOfferSnapshot{Snapshot: snapshot})
	require.Equal(t, abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_ACCEPT}, respOffer)

	// We should be able to pass an invalid chunk and get a verify failure, before reapplying it.
	respApply := target.ApplySnapshotChunk(abci.RequestApplySnapshotChunk{
		Index:  0,
		Chunk:  []byte{9},
		Sender: "sender",
	})
	require.Equal(t, abci.ResponseApplySnapshotChunk{
		Result:        abci.ResponseApplySnapshotChunk_RETRY,
		RefetchChunks: []uint32{0},
		RejectSenders: []string{"sender"},
	}, respApply)

	// Fetch each chunk from the source and apply it to the target
	for index := uint32(0); index < snapshot.Chunks; index++ {
		respChunk := source.LoadSnapshotChunk(abci.RequestLoadSnapshotChunk{
			Height: snapshot.Height,
			Format: snapshot.Format,
			Chunk:  index,
		})
		require.NotNil(t, respChunk.Chunk)
		respApply := target.ApplySnapshotChunk(abci.RequestApplySnapshotChunk{
			Index: index,
			Chunk: respChunk.Chunk,
		})
		require.Equal(t, abci.ResponseApplySnapshotChunk{
			Result: abci.ResponseApplySnapshotChunk_ACCEPT,
		}, respApply)
	}

	// The target should now have the same hash as the source
	assert.Equal(t, source.LastCommitID(), target.LastCommitID())
}

// NOTE: represents a new custom router for testing purposes of WithRouter()
type testCustomRouter struct {
	routes sync.Map
}

func (rtr *testCustomRouter) AddRoute(route sdk.Route) sdk.Router {
	rtr.routes.Store(route.Path(), route.Handler())
	return rtr
}

func (rtr *testCustomRouter) Route(ctx sdk.Context, path string) sdk.Handler {
	if v, ok := rtr.routes.Load(path); ok {
		if h, ok := v.(sdk.Handler); ok {
			return h
		}
	}
	return nil
}

func TestWithRouter(t *testing.T) {
	// test increments in the handler
	deliverKey := []byte("deliver-key")

	txHandlerOpt := func(bapp *baseapp.BaseApp) {
		customRouter := &testCustomRouter{routes: sync.Map{}}
		r := sdk.NewRoute(routeMsgCounter, handlerMsgCounter(t, capKey1, deliverKey))
		customRouter.AddRoute(r)
		txHandler := middleware.ComposeMiddlewares(
			middleware.NewRunMsgsTxHandler(middleware.NewMsgServiceRouter(encCfg.InterfaceRegistry), customRouter),
			middleware.NewTxDecoderMiddleware(testTxDecoder(encCfg.Amino)),
		)
		bapp.SetTxHandler(txHandler)
	}
	app := setupBaseApp(t, baseapp.AppOptionFunc(txHandlerOpt))
	app.InitChain(abci.RequestInitChain{})

	nBlocks := 3
	txPerHeight := 5

	for blockN := 0; blockN < nBlocks; blockN++ {
		header := tmproto.Header{Height: int64(blockN) + 1}
		app.BeginBlock(abci.RequestBeginBlock{Header: header})

		for i := 0; i < txPerHeight; i++ {
			counter := int64(blockN*txPerHeight + i)
			tx := newTxCounter(counter, counter)

			txBytes, err := encCfg.Amino.Marshal(tx)
			require.NoError(t, err)

			res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
			require.True(t, res.IsOK(), fmt.Sprintf("%v", res))
		}

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}
}

func TestBaseApp_EndBlock(t *testing.T) {
	db := memdb.NewDB()
	name := t.Name()
	logger := defaultLogger()

	cp := &tmproto.ConsensusParams{
		Block: &tmproto.BlockParams{
			MaxGas: 5000000,
		},
	}

	app := baseapp.NewBaseApp(name, logger, db)
	app.SetParamStore(newParamStore(memdb.NewDB()))
	app.InitChain(abci.RequestInitChain{
		ConsensusParams: cp,
	})

	app.SetEndBlocker(func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
		return abci.ResponseEndBlock{
			ValidatorUpdates: []abci.ValidatorUpdate{
				{Power: 100},
			},
		}
	})
	app.Seal()

	res := app.EndBlock(abci.RequestEndBlock{})
	require.Len(t, res.GetValidatorUpdates(), 1)
	require.Equal(t, int64(100), res.GetValidatorUpdates()[0].Power)
	require.Equal(t, cp.Block.MaxGas, res.ConsensusParamUpdates.Block.MaxGas)
}
