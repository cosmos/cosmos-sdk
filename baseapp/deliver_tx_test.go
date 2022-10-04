package baseapp_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
	"unsafe"

	"cosmossdk.io/depinject"
	"github.com/cosmos/gogoproto/jsonpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	pruningtypes "github.com/cosmos/cosmos-sdk/pruning/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/snapshots"
	snapshottypes "github.com/cosmos/cosmos-sdk/snapshots/types"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
)

var (
	capKey1 = sdk.NewKVStoreKey("key1")
	capKey2 = sdk.NewKVStoreKey("key2")

	// testTxPriority is the CheckTx priority that we set in the test
	// antehandler.
	testTxPriority = int64(42)
)

type setupConfig struct {
	blocks             uint64
	blockTxs           int
	snapshotInterval   uint64
	snapshotKeepRecent uint32
	pruningOpts        pruningtypes.PruningOptions
}

func defaultLogger() log.Logger {
	return log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
}

// simple one store baseapp
func setupBaseApp(t *testing.T, options ...func(*baseapp.BaseApp)) *baseapp.BaseApp {
	cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
	baseapptestutil.RegisterInterfaces(cdc.InterfaceRegistry())
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)

	logger := defaultLogger()
	db := dbm.NewMemDB()
	app := baseapp.NewBaseApp(t.Name(), logger, db, txConfig.TxDecoder(), options...)
	require.Equal(t, t.Name(), app.Name())

	app.MountStores(capKey1, capKey2)
	app.SetParamStore(&paramStore{db: dbm.NewMemDB()})

	// stores are mounted
	err := app.LoadLatestVersion()
	require.Nil(t, err)
	return app
}

// simple one store baseapp with data and snapshots. Each tx is 1 MB in size (uncompressed).
func setupBaseAppWithSnapshots(t *testing.T, config *setupConfig) (*baseapp.BaseApp, error) {
	snapshotTimeout := 1 * time.Minute
	snapshotStore, err := snapshots.NewStore(dbm.NewMemDB(), testutil.GetTempDir(t))
	require.NoError(t, err)

	app := setupBaseApp(t, baseapp.SetSnapshot(snapshotStore, snapshottypes.NewSnapshotOptions(config.snapshotInterval, uint32(config.snapshotKeepRecent))), baseapp.SetPruning(config.pruningOpts))

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	baseapptestutil.RegisterInterfaces(cdc.InterfaceRegistry())
	app.SetMsgServiceRouter(baseapp.NewMsgServiceRouter())
	app.SetInterfaceRegistry(registry)

	baseapptestutil.RegisterKeyValueServer(app.MsgServiceRouter(), MsgKeyValueImpl{})

	// patch in TxConfig instead of using an output from x/auth/tx
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)

	// set the TxDecoder in the BaseApp for minimal tx simulations
	app.SetTxDecoder(txConfig.TxDecoder())

	app.InitChain(abci.RequestInitChain{
		ConsensusParams: &tmproto.ConsensusParams{},
	})

	r := rand.New(rand.NewSource(3920758213583))
	keyCounter := 0
	for height := int64(1); height <= int64(config.blocks); height++ {
		app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: height}})
		for txNum := 0; txNum < config.blockTxs; txNum++ {
			msgs := []sdk.Msg{}
			for msgNum := 0; msgNum < 100; msgNum++ {
				key := []byte(fmt.Sprintf("%v", keyCounter))
				value := make([]byte, 10000)
				_, err := r.Read(value)
				require.NoError(t, err)
				msgs = append(msgs, &baseapptestutil.MsgKeyValue{Key: key, Value: value})
				keyCounter++
			}

			builder := txConfig.NewTxBuilder()
			builder.SetMsgs(msgs...)

			txBytes, err := txConfig.TxEncoder()(builder.GetTx())
			require.NoError(t, err)
			resp := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
			require.True(t, resp.IsOK(), "%v", resp.String())
		}
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()

		// Wait for snapshot to be taken, since it happens asynchronously.
		if config.snapshotInterval > 0 && uint64(height)%config.snapshotInterval == 0 {
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

	return app, nil
}

// Test that we can make commits and then reload old versions.
// Test that LoadLatestVersion actually does.
func TestLoadVersion(t *testing.T) {
	logger := defaultLogger()
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
	header := tmproto.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	res := app.Commit()
	commitID1 := storetypes.CommitID{Version: 1, Hash: res.Data}

	// execute a block, collect commit ID
	header = tmproto.Header{Height: 2}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	res = app.Commit()
	commitID2 := storetypes.CommitID{Version: 2, Hash: res.Data}

	// reload with LoadLatestVersion
	app = baseapp.NewBaseApp(name, logger, db, nil, pruningOpt)
	app.MountStores()
	err = app.LoadLatestVersion()
	require.Nil(t, err)
	testLoadVersionHelper(t, app, int64(2), commitID2)

	// reload with LoadVersion, see if you can commit the same block and get
	// the same result
	app = baseapp.NewBaseApp(name, logger, db, nil, pruningOpt)
	err = app.LoadVersion(1)
	require.Nil(t, err)
	testLoadVersionHelper(t, app, int64(1), commitID1)
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	app.Commit()
	testLoadVersionHelper(t, app, int64(2), commitID2)
}

func useDefaultLoader(app *baseapp.BaseApp) {
	app.SetStoreLoader(baseapp.DefaultStoreLoader)
}

func initStore(t *testing.T, db dbm.DB, storeKey string, k, v []byte) {
	rs := rootmulti.NewStore(db, log.NewNopLogger())
	rs.SetPruning(pruningtypes.NewPruningOptions(pruningtypes.PruningNothing))
	key := sdk.NewKVStoreKey(storeKey)
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

func checkStore(t *testing.T, db dbm.DB, ver int64, storeKey string, k, v []byte) {
	rs := rootmulti.NewStore(db, log.NewNopLogger())
	rs.SetPruning(pruningtypes.NewPruningOptions(pruningtypes.PruningDefault))
	key := sdk.NewKVStoreKey(storeKey)
	rs.MountStoreWithDB(key, storetypes.StoreTypeIAVL, nil)
	err := rs.LoadLatestVersion()
	require.Nil(t, err)
	require.Equal(t, ver, rs.LastCommitID().Version)

	// query data in substore
	kv, _ := rs.GetStore(key).(storetypes.KVStore)
	require.NotNil(t, kv)
	require.Equal(t, v, kv.Get(k))
}

// Test that we can make commits and then reload old versions.
// Test that LoadLatestVersion actually does.
func TestSetLoader(t *testing.T) {
	cases := map[string]struct {
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

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			// prepare a db with some data
			db := dbm.NewMemDB()
			initStore(t, db, tc.origStoreKey, k, v)

			// load the app with the existing db
			opts := []func(*baseapp.BaseApp){baseapp.SetPruning(pruningtypes.NewPruningOptions(pruningtypes.PruningNothing))}
			if tc.setLoader != nil {
				opts = append(opts, tc.setLoader)
			}
			app := baseapp.NewBaseApp(t.Name(), defaultLogger(), db, nil, opts...)
			app.MountStores(sdk.NewKVStoreKey(tc.loadStoreKey))
			err := app.LoadLatestVersion()
			require.Nil(t, err)

			// "execute" one block
			app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: 2}})
			res := app.Commit()
			require.NotNil(t, res.Data)

			// check db is properly updated
			checkStore(t, db, 2, tc.loadStoreKey, k, v)
			checkStore(t, db, 2, tc.loadStoreKey, []byte("foo"), nil)
		})
	}
}

func TestVersionSetterGetter(t *testing.T) {
	logger := defaultLogger()
	pruningOpt := baseapp.SetPruning(pruningtypes.NewPruningOptions(pruningtypes.PruningDefault))
	db := dbm.NewMemDB()
	name := t.Name()
	app := baseapp.NewBaseApp(name, logger, db, nil, pruningOpt)

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

	header := tmproto.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	res := app.Commit()
	commitID1 := storetypes.CommitID{Version: 1, Hash: res.Data}

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

func testLoadVersionHelper(t *testing.T, app *baseapp.BaseApp, expectedHeight int64, expectedID storetypes.CommitID) {
	lastHeight := app.LastBlockHeight()
	lastID := app.LastCommitID()
	require.Equal(t, expectedHeight, lastHeight)
	require.Equal(t, expectedID, lastID)
}

func TestOptionFunction(t *testing.T) {
	logger := defaultLogger()
	db := dbm.NewMemDB()
	bap := baseapp.NewBaseApp("starting name", logger, db, nil, testChangeNameHelper("new name"))
	require.Equal(t, bap.Name(), "new name", "BaseApp should have had name changed via option function")
}

func testChangeNameHelper(name string) func(*baseapp.BaseApp) {
	return func(bap *baseapp.BaseApp) {
		bap.SetName(name)
	}
}

// Test that Info returns the latest committed state.
func TestInfo(t *testing.T) {
	app := setupBaseApp(t)

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
		app.SetDB(nil)
	})
	require.Panics(t, func() {
		app.SetCMS(nil)
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
		app.SetAnteHandler(nil)
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

func TestInitChainer(t *testing.T) {
	name := t.Name()
	// keep the db and logger ourselves so
	// we can reload the same  app later
	db := dbm.NewMemDB()
	logger := defaultLogger()
	app := baseapp.NewBaseApp(name, logger, db, nil)
	capKey := sdk.NewKVStoreKey("main")
	capKey2 := sdk.NewKVStoreKey("key2")
	app.MountStores(capKey, capKey2)

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
	require.Equal(t, 0, len(res.Value))

	// set initChainer and try again - should see the value
	app.SetInitChainer(initChainer)

	// stores are mounted and private members are set - sealing baseapp
	err := app.LoadLatestVersion() // needed to make stores non-nil
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

	chainID := getDeliverStateCtx(app).ChainID()
	require.Equal(t, "test-chain-id", chainID, "ChainID in deliverState not set correctly in InitChain")

	chainID = getCheckStateCtx(app).ChainID()
	require.Equal(t, "test-chain-id", chainID, "ChainID in checkState not set correctly in InitChain")

	app.Commit()
	res = app.Query(query)
	require.Equal(t, int64(1), app.LastBlockHeight())
	require.Equal(t, value, res.Value)

	// reload app
	app = baseapp.NewBaseApp(name, logger, db, nil)
	app.SetInitChainer(initChainer)
	app.MountStores(capKey, capKey2)
	err = app.LoadLatestVersion() // needed to make stores non-nil
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
	db := dbm.NewMemDB()
	logger := defaultLogger()
	app := baseapp.NewBaseApp(name, logger, db, nil)

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
	db := dbm.NewMemDB()
	logger := defaultLogger()
	app := baseapp.NewBaseApp(name, logger, db, nil)

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

// Number of messages doesn't matter to CheckTx.
func TestMultiMsgCheckTx(t *testing.T) {
	// TODO: ensure we get the same results
	// with one message or many
}

// Interleave calls to Check and Deliver and ensure
// that there is no cross-talk. Check sees results of the previous Check calls
// and Deliver sees that of the previous Deliver calls, but they don't see eachother.
func TestConcurrentCheckDeliver(t *testing.T) {
	// TODO
}

func TestGRPCQuery(t *testing.T) {
	grpcQueryOpt := func(bapp *baseapp.BaseApp) {
		testdata.RegisterQueryServer(
			bapp.GRPCQueryRouter(),
			testdata.QueryImpl{},
		)
	}

	app := setupBaseApp(t, grpcQueryOpt)
	app.GRPCQueryRouter().SetInterfaceRegistry(codectypes.NewInterfaceRegistry())

	app.InitChain(abci.RequestInitChain{
		ConsensusParams: &tmproto.ConsensusParams{},
	})

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

	app := setupBaseApp(t, addrPeerFilterOpt, idPeerFilterOpt)

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

func TestListSnapshots(t *testing.T) {
	setupConfig := &setupConfig{
		blocks:             5,
		blockTxs:           4,
		snapshotInterval:   2,
		snapshotKeepRecent: 2,
		pruningOpts:        pruningtypes.NewPruningOptions(pruningtypes.PruningNothing),
	}

	app, err := setupBaseAppWithSnapshots(t, setupConfig)
	require.NoError(t, err)

	resp := app.ListSnapshots(abci.RequestListSnapshots{})
	for _, s := range resp.Snapshots {
		assert.NotEmpty(t, s.Hash)
		assert.NotEmpty(t, s.Metadata)
		s.Hash = nil
		s.Metadata = nil
	}
	assert.Equal(t, abci.ResponseListSnapshots{Snapshots: []*abci.Snapshot{
		{Height: 4, Format: snapshottypes.CurrentFormat, Chunks: 2},
		{Height: 2, Format: snapshottypes.CurrentFormat, Chunks: 1},
	}}, resp)
}

func TestSnapshotWithPruning(t *testing.T) {
	testcases := map[string]struct {
		config            *setupConfig
		expectedSnapshots []*abci.Snapshot
		expectedErr       error
	}{
		"prune nothing with snapshot": {
			config: &setupConfig{
				blocks:             20,
				blockTxs:           2,
				snapshotInterval:   5,
				snapshotKeepRecent: 1,
				pruningOpts:        pruningtypes.NewPruningOptions(pruningtypes.PruningNothing),
			},
			expectedSnapshots: []*abci.Snapshot{
				{Height: 20, Format: snapshottypes.CurrentFormat, Chunks: 5},
			},
		},
		"prune everything with snapshot": {
			config: &setupConfig{
				blocks:             20,
				blockTxs:           2,
				snapshotInterval:   5,
				snapshotKeepRecent: 1,
				pruningOpts:        pruningtypes.NewPruningOptions(pruningtypes.PruningEverything),
			},
			expectedSnapshots: []*abci.Snapshot{
				{Height: 20, Format: snapshottypes.CurrentFormat, Chunks: 5},
			},
		},
		"default pruning with snapshot": {
			config: &setupConfig{
				blocks:             20,
				blockTxs:           2,
				snapshotInterval:   5,
				snapshotKeepRecent: 1,
				pruningOpts:        pruningtypes.NewPruningOptions(pruningtypes.PruningDefault),
			},
			expectedSnapshots: []*abci.Snapshot{
				{Height: 20, Format: snapshottypes.CurrentFormat, Chunks: 5},
			},
		},
		"custom": {
			config: &setupConfig{
				blocks:             25,
				blockTxs:           2,
				snapshotInterval:   5,
				snapshotKeepRecent: 2,
				pruningOpts:        pruningtypes.NewCustomPruningOptions(12, 12),
			},
			expectedSnapshots: []*abci.Snapshot{
				{Height: 25, Format: snapshottypes.CurrentFormat, Chunks: 6},
				{Height: 20, Format: snapshottypes.CurrentFormat, Chunks: 5},
			},
		},
		"no snapshots": {
			config: &setupConfig{
				blocks:           10,
				blockTxs:         2,
				snapshotInterval: 0, // 0 implies disable snapshots
				pruningOpts:      pruningtypes.NewPruningOptions(pruningtypes.PruningNothing),
			},
			expectedSnapshots: []*abci.Snapshot{},
		},
		"keep all snapshots": {
			config: &setupConfig{
				blocks:             10,
				blockTxs:           2,
				snapshotInterval:   3,
				snapshotKeepRecent: 0, // 0 implies keep all snapshots
				pruningOpts:        pruningtypes.NewPruningOptions(pruningtypes.PruningNothing),
			},
			expectedSnapshots: []*abci.Snapshot{
				{Height: 9, Format: snapshottypes.CurrentFormat, Chunks: 2},
				{Height: 6, Format: snapshottypes.CurrentFormat, Chunks: 2},
				{Height: 3, Format: snapshottypes.CurrentFormat, Chunks: 1},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			app, err := setupBaseAppWithSnapshots(t, tc.config)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				return
			}
			require.NoError(t, err)

			resp := app.ListSnapshots(abci.RequestListSnapshots{})
			for _, s := range resp.Snapshots {
				assert.NotEmpty(t, s.Hash)
				assert.NotEmpty(t, s.Metadata)
				s.Hash = nil
				s.Metadata = nil
			}
			fmt.Println(resp)
			assert.Equal(t, abci.ResponseListSnapshots{Snapshots: tc.expectedSnapshots}, resp)

			// Validate that heights were pruned correctly by querying the state at the last height that should be present relative to latest
			// and the first height that should be pruned.
			//
			// Exceptions:
			//   * Prune nothing: should be able to query all heights (we only test first and latest)
			//   * Prune default: should be able to query all heights (we only test first and latest)
			//      * The reason for default behaving this way is that we only commit 20 heights but default has 100_000 keep-recent
			var lastExistingHeight int64
			if tc.config.pruningOpts.GetPruningStrategy() == pruningtypes.PruningNothing || tc.config.pruningOpts.GetPruningStrategy() == pruningtypes.PruningDefault {
				lastExistingHeight = 1
			} else {
				// Integer division rounds down so by multiplying back we get the last height at which we pruned
				lastExistingHeight = int64((tc.config.blocks/tc.config.pruningOpts.Interval)*tc.config.pruningOpts.Interval - tc.config.pruningOpts.KeepRecent)
			}

			// Query 1
			res := app.Query(abci.RequestQuery{Path: fmt.Sprintf("/store/%s/key", capKey2.Name()), Data: []byte("0"), Height: lastExistingHeight})
			require.NotNil(t, res, "height: %d", lastExistingHeight)
			require.NotNil(t, res.Value, "height: %d", lastExistingHeight)

			// Query 2
			res = app.Query(abci.RequestQuery{Path: fmt.Sprintf("/store/%s/key", capKey2.Name()), Data: []byte("0"), Height: lastExistingHeight - 1})
			require.NotNil(t, res, "height: %d", lastExistingHeight-1)
			if tc.config.pruningOpts.GetPruningStrategy() == pruningtypes.PruningNothing || tc.config.pruningOpts.GetPruningStrategy() == pruningtypes.PruningDefault {
				// With prune nothing or default, we query height 0 which translates to the latest height.
				require.NotNil(t, res.Value, "height: %d", lastExistingHeight-1)
			}
		})
	}
}

func TestLoadSnapshotChunk(t *testing.T) {
	setupConfig := &setupConfig{
		blocks:             2,
		blockTxs:           5,
		snapshotInterval:   2,
		snapshotKeepRecent: snapshottypes.CurrentFormat,
		pruningOpts:        pruningtypes.NewPruningOptions(pruningtypes.PruningNothing),
	}
	app, err := setupBaseAppWithSnapshots(t, setupConfig)
	require.NoError(t, err)

	testcases := map[string]struct {
		height      uint64
		format      uint32
		chunk       uint32
		expectEmpty bool
	}{
		"Existing snapshot": {2, snapshottypes.CurrentFormat, 1, false},
		"Missing height":    {100, snapshottypes.CurrentFormat, 1, true},
		"Missing format":    {2, snapshottypes.CurrentFormat + 1, 1, true},
		"Missing chunk":     {2, snapshottypes.CurrentFormat, 9, true},
		"Zero height":       {0, snapshottypes.CurrentFormat, 1, true},
		"Zero format":       {2, 0, 1, true},
		"Zero chunk":        {2, snapshottypes.CurrentFormat, 0, false},
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
	// Set up app before test cases, since it's fairly expensive.
	setupConfig := &setupConfig{
		blocks:             0,
		blockTxs:           0,
		snapshotInterval:   2,
		snapshotKeepRecent: 2,
		pruningOpts:        pruningtypes.NewPruningOptions(pruningtypes.PruningNothing),
	}
	app, err := setupBaseAppWithSnapshots(t, setupConfig)
	require.NoError(t, err)

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
			Height: 1, Format: snapshottypes.CurrentFormat, Chunks: 2, Hash: hash, Metadata: metadata,
		}, abci.ResponseOfferSnapshot_REJECT},
		"no chunks": {&abci.Snapshot{
			Height: 1, Format: snapshottypes.CurrentFormat, Chunks: 0, Hash: hash, Metadata: metadata,
		}, abci.ResponseOfferSnapshot_REJECT},
		"invalid metadata serialization": {&abci.Snapshot{
			Height: 1, Format: snapshottypes.CurrentFormat, Chunks: 0, Hash: hash, Metadata: []byte{3, 1, 4},
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
	setupConfig1 := &setupConfig{
		blocks:             4,
		blockTxs:           10,
		snapshotInterval:   2,
		snapshotKeepRecent: 2,
		pruningOpts:        pruningtypes.NewPruningOptions(pruningtypes.PruningNothing),
	}
	source, err := setupBaseAppWithSnapshots(t, setupConfig1)
	require.NoError(t, err)

	setupConfig2 := &setupConfig{
		blocks:             0,
		blockTxs:           0,
		snapshotInterval:   2,
		snapshotKeepRecent: 2,
		pruningOpts:        pruningtypes.NewPruningOptions(pruningtypes.PruningNothing),
	}
	target, err := setupBaseAppWithSnapshots(t, setupConfig2)
	require.NoError(t, err)

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

func TestBaseApp_EndBlock(t *testing.T) {
	db := dbm.NewMemDB()
	name := t.Name()
	logger := defaultLogger()

	cp := &tmproto.ConsensusParams{
		Block: &tmproto.BlockParams{
			MaxGas: 5000000,
		},
	}

	app := baseapp.NewBaseApp(name, logger, db, nil)
	app.SetParamStore(&paramStore{db: dbm.NewMemDB()})
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

// Test that txs can be unmarshalled and read and that
// correct error codes are returned when not
func TestTxDecoder(t *testing.T) {
	cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
	baseapptestutil.RegisterInterfaces(cdc.InterfaceRegistry())

	// patch in TxConfig instead of using an output from x/auth/tx
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)

	tx := newTxCounter(txConfig, 1, 0)
	txBytes, err := txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	dTx, err := txConfig.TxDecoder()(txBytes)
	require.NoError(t, err)

	counter, _ := parseTxMemo(tx)
	dTxCounter, _ := parseTxMemo(dTx)
	require.Equal(t, counter, dTxCounter)
}

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

	anteOpt := func(bapp *baseapp.BaseApp) { bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, counterKey)) }

	// Setup baseapp.
	var (
		appBuilder *runtime.AppBuilder
		cdc        codec.ProtoCodecMarshaler
	)
	err := depinject.Inject(makeMinimalConfig(), &appBuilder, &cdc)
	require.NoError(t, err)

	testCtx := testutil.DefaultContextWithDB(t, capKey1, sdk.NewTransientStoreKey("transient_test"))

	app := appBuilder.Build(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), testCtx.DB, nil, anteOpt)
	app.SetCMS(testCtx.CMS)
	baseapptestutil.RegisterInterfaces(cdc.InterfaceRegistry())
	baseapptestutil.RegisterCounterServer(app.MsgServiceRouter(), CounterServerImpl{t, capKey1, counterKey})

	// patch in TxConfig instead of using an output from x/auth/tx
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)
	// set the TxDecoder in the BaseApp for minimal tx simulations
	app.SetTxDecoder(txConfig.TxDecoder())

	nTxs := int64(5)
	app.InitChain(abci.RequestInitChain{})

	for i := int64(0); i < nTxs; i++ {
		tx := newTxCounter(txConfig, i, 0) // no messages
		txBytes, err := txConfig.TxEncoder()(tx)
		require.NoError(t, err)

		require.NoError(t, err)
		r := app.CheckTx(abci.RequestCheckTx{Tx: txBytes})
		require.Equal(t, testTxPriority, r.Priority)
		require.Empty(t, r.GetEvents())
		require.True(t, r.IsOK(), fmt.Sprintf("%v", r))
	}

	checkStateStore := getCheckStateCtx(app.BaseApp).KVStore(capKey1)
	storedCounter := getIntFromStore(checkStateStore, counterKey)

	// Ensure AnteHandler ran
	require.Equal(t, nTxs, storedCounter)

	// If a block is committed, CheckTx state should be reset.
	header := tmproto.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header, Hash: []byte("hash")})

	require.NotNil(t, getCheckStateCtx(app.BaseApp).BlockGasMeter(), "block gas meter should have been set to checkState")
	require.NotEmpty(t, getCheckStateCtx(app.BaseApp).HeaderHash())

	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	checkStateStore = getCheckStateCtx(app.BaseApp).KVStore(capKey1)
	storedBytes := checkStateStore.Get(counterKey)
	require.Nil(t, storedBytes)
}

// Test that successive DeliverTx can see each others' effects
// on the store, both within and across blocks.
func TestDeliverTx(t *testing.T) {
	// test increments in the ante
	anteKey := []byte("ante-key")
	anteOpt := func(bapp *baseapp.BaseApp) { bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, anteKey)) }

	// Setup baseapp.
	var (
		appBuilder *runtime.AppBuilder
		cdc        codec.ProtoCodecMarshaler
	)
	err := depinject.Inject(makeMinimalConfig(), &appBuilder, &cdc)
	require.NoError(t, err)

	testCtx := testutil.DefaultContextWithDB(t, capKey1, sdk.NewTransientStoreKey("transient_test"))

	app := appBuilder.Build(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), testCtx.DB, nil, anteOpt)
	app.SetCMS(testCtx.CMS)
	baseapptestutil.RegisterInterfaces(cdc.InterfaceRegistry())

	// patch in TxConfig instead of using an output from x/auth/tx
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)
	// set the TxDecoder in the BaseApp for minimal tx simulations
	app.SetTxDecoder(txConfig.TxDecoder())

	app.InitChain(abci.RequestInitChain{})

	deliverKey := []byte("deliver-key")
	baseapptestutil.RegisterCounterServer(app.MsgServiceRouter(), CounterServerImpl{t, capKey1, deliverKey})

	nBlocks := 3
	txPerHeight := 5

	for blockN := 0; blockN < nBlocks; blockN++ {
		header := tmproto.Header{Height: int64(blockN) + 1}
		app.BeginBlock(abci.RequestBeginBlock{Header: header})

		for i := 0; i < txPerHeight; i++ {
			counter := int64(blockN*txPerHeight + i)
			tx := newTxCounter(txConfig, counter, counter)
			txBytes, err := txConfig.TxEncoder()(tx)
			require.NoError(t, err)

			res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
			require.True(t, res.IsOK(), fmt.Sprintf("%v", res))
			events := res.GetEvents()
			require.Len(t, events, 3, "should contain ante handler, message type and counter events respectively")
			require.Equal(t, sdk.MarkEventsToIndex(counterEvent("ante_handler", counter).ToABCIEvents(), map[string]struct{}{})[0], events[0], "ante handler event")
			require.Equal(t, sdk.MarkEventsToIndex(counterEvent(sdk.EventTypeMessage, counter).ToABCIEvents(), map[string]struct{}{})[0], events[2], "msg handler update counter event")
		}

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}
}

// One call to DeliverTx should process all the messages, in order.
func TestMultiMsgDeliverTx(t *testing.T) {
	// test increments in the ante
	anteKey := []byte("ante-key")
	anteOpt := func(bapp *baseapp.BaseApp) { bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, anteKey)) }

	// Setup baseapp.
	var (
		appBuilder *runtime.AppBuilder
		cdc        codec.ProtoCodecMarshaler
	)
	err := depinject.Inject(makeMinimalConfig(), &appBuilder, &cdc)
	require.NoError(t, err)

	testCtx := testutil.DefaultContextWithDB(t, capKey1, sdk.NewTransientStoreKey("transient_test"))

	app := appBuilder.Build(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), testCtx.DB, nil, anteOpt)
	app.SetCMS(testCtx.CMS)
	baseapptestutil.RegisterInterfaces(cdc.InterfaceRegistry())

	// patch in TxConfig instead of using an output from x/auth/tx
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)
	// set the TxDecoder in the BaseApp for minimal tx simulations
	app.SetTxDecoder(txConfig.TxDecoder())

	app.InitChain(abci.RequestInitChain{})

	deliverKey := []byte("deliver-key")
	baseapptestutil.RegisterCounterServer(app.MsgServiceRouter(), CounterServerImpl{t, capKey1, deliverKey})

	deliverKey2 := []byte("deliver-key2")
	baseapptestutil.RegisterCounter2Server(app.MsgServiceRouter(), Counter2ServerImpl{t, capKey1, deliverKey2})

	// run a multi-msg tx
	// with all msgs the same route

	header := tmproto.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	tx := newTxCounter(txConfig, 0, 0, 1, 2)
	txBytes, err := txConfig.TxEncoder()(tx)
	require.NoError(t, err)
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.True(t, res.IsOK(), fmt.Sprintf("%v", res))

	store := getDeliverStateCtx(app.BaseApp).KVStore(capKey1)

	// tx counter only incremented once
	txCounter := getIntFromStore(store, anteKey)
	require.Equal(t, int64(1), txCounter)

	// msg counter incremented three times
	msgCounter := getIntFromStore(store, deliverKey)
	require.Equal(t, int64(3), msgCounter)

	// replace the second message with a Counter2
	tx = newTxCounter(txConfig, 1, 3)

	builder := txConfig.NewTxBuilder()
	msgs := tx.GetMsgs()
	msgs = append(msgs, &baseapptestutil.MsgCounter2{Counter: 0})
	msgs = append(msgs, &baseapptestutil.MsgCounter2{Counter: 1})

	builder.SetMsgs(msgs...)
	builder.SetMemo(tx.GetMemo())

	txBytes, err = txConfig.TxEncoder()(builder.GetTx())
	require.NoError(t, err)
	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.True(t, res.IsOK(), fmt.Sprintf("%v", res))

	store = getDeliverStateCtx(app.BaseApp).KVStore(capKey1)

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

// Simulate a transaction that uses gas to compute the gas.
// Simulate() and Query("/app/simulate", txBytes) should give
// the same results.
func TestSimulateTx(t *testing.T) {
	gasConsumed := uint64(5)

	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			newCtx = ctx.WithGasMeter(sdk.NewGasMeter(gasConsumed))
			return
		})
	}

	// Setup baseapp.
	var (
		appBuilder *runtime.AppBuilder
		cdc        codec.ProtoCodecMarshaler
	)
	err := depinject.Inject(makeMinimalConfig(), &appBuilder, &cdc)
	require.NoError(t, err)

	testCtx := testutil.DefaultContextWithDB(t, capKey1, sdk.NewTransientStoreKey("transient_test"))

	app := appBuilder.Build(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), testCtx.DB, nil, anteOpt)
	app.SetCMS(testCtx.CMS)
	baseapptestutil.RegisterInterfaces(cdc.InterfaceRegistry())

	// patch in TxConfig instead of using an output from x/auth/tx
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)
	// set the TxDecoder in the BaseApp for minimal tx simulations
	app.SetTxDecoder(txConfig.TxDecoder())

	app.InitChain(abci.RequestInitChain{})

	baseapptestutil.RegisterCounterServer(app.MsgServiceRouter(), CounterServerImplGasMeterOnly{gasConsumed})

	nBlocks := 3
	for blockN := 0; blockN < nBlocks; blockN++ {
		count := int64(blockN + 1)
		header := tmproto.Header{Height: count}
		app.BeginBlock(abci.RequestBeginBlock{Header: header})

		tx := newTxCounter(txConfig, count, count)

		txBytes, err := txConfig.TxEncoder()(tx)
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
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			return
		})
	}

	// Setup baseapp.
	var (
		appBuilder *runtime.AppBuilder
		cdc        codec.ProtoCodecMarshaler
	)
	err := depinject.Inject(makeMinimalConfig(), &appBuilder, &cdc)
	require.NoError(t, err)

	testCtx := testutil.DefaultContextWithDB(t, capKey1, sdk.NewTransientStoreKey("transient_test"))

	app := appBuilder.Build(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), testCtx.DB, nil, anteOpt)
	app.SetCMS(testCtx.CMS)

	baseapptestutil.RegisterInterfaces(cdc.InterfaceRegistry())
	baseapptestutil.RegisterCounterServer(app.MsgServiceRouter(), CounterServerImplGasMeterOnly{})

	// patch in TxConfig instead of using an output from x/auth/tx
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)
	// set the TxDecoder in the BaseApp for minimal tx simulations
	app.SetTxDecoder(txConfig.TxDecoder())

	app.InitChain(abci.RequestInitChain{})

	header := tmproto.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	// transaction with no messages
	{
		emptyTx := txConfig.NewTxBuilder().GetTx()
		_, result, err := app.SimDeliver(txConfig.TxEncoder(), emptyTx)
		require.Error(t, err)
		require.Nil(t, result)

		space, code, _ := sdkerrors.ABCIInfo(err, false)
		require.EqualValues(t, sdkerrors.ErrInvalidRequest.Codespace(), space, err)
		require.EqualValues(t, sdkerrors.ErrInvalidRequest.ABCICode(), code, err)
	}

	// transaction where ValidateBasic fails
	{
		testCases := []struct {
			tx   signing.Tx
			fail bool
		}{
			{newTxCounter(txConfig, 0, 0), false},
			{newTxCounter(txConfig, -1, 0), false},
			{newTxCounter(txConfig, 100, 100), false},
			{newTxCounter(txConfig, 100, 5, 4, 3, 2, 1), false},

			{newTxCounter(txConfig, 0, -1), true},
			{newTxCounter(txConfig, 0, 1, -2), true},
			{newTxCounter(txConfig, 0, 1, 2, -10, 5), true},
		}

		for _, testCase := range testCases {
			tx := testCase.tx
			_, result, err := app.SimDeliver(txConfig.TxEncoder(), tx)

			if testCase.fail {
				require.Error(t, err)

				space, code, _ := sdkerrors.ABCIInfo(err, false)
				require.EqualValues(t, sdkerrors.ErrInvalidSequence.Codespace(), space, err)
				require.EqualValues(t, sdkerrors.ErrInvalidSequence.ABCICode(), code, err)
			} else {
				require.NotNil(t, result)
			}
		}
	}

	// transaction with no known route
	{
		txBuilder := txConfig.NewTxBuilder()
		txBuilder.SetMsgs(&baseapptestutil.MsgCounter2{})
		unknownRouteTx := txBuilder.GetTx()

		_, result, err := app.SimDeliver(txConfig.TxEncoder(), unknownRouteTx)
		require.Error(t, err)
		require.Nil(t, result)

		space, code, _ := sdkerrors.ABCIInfo(err, false)
		require.EqualValues(t, sdkerrors.ErrUnknownRequest.Codespace(), space, err)
		require.EqualValues(t, sdkerrors.ErrUnknownRequest.ABCICode(), code, err)

		txBuilder = txConfig.NewTxBuilder()
		txBuilder.SetMsgs(&baseapptestutil.MsgCounter{}, &baseapptestutil.MsgCounter2{})
		unknownRouteTx = txBuilder.GetTx()
		_, result, err = app.SimDeliver(txConfig.TxEncoder(), unknownRouteTx)
		require.Error(t, err)
		require.Nil(t, result)

		space, code, _ = sdkerrors.ABCIInfo(err, false)
		require.EqualValues(t, sdkerrors.ErrUnknownRequest.Codespace(), space, err)
		require.EqualValues(t, sdkerrors.ErrUnknownRequest.ABCICode(), code, err)
	}

	// Transaction with an unregistered message
	{
		txBuilder := txConfig.NewTxBuilder()
		txBuilder.SetMsgs(&testdata.MsgCreateDog{})
		tx := txBuilder.GetTx()

		txBytes, err := txConfig.TxEncoder()(tx)
		require.NoError(t, err)

		res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
		require.EqualValues(t, sdkerrors.ErrTxDecode.ABCICode(), res.Code)
		require.EqualValues(t, sdkerrors.ErrTxDecode.Codespace(), res.Codespace)
	}
}

// Test that transactions exceeding gas limits fail
func TestTxGasLimits(t *testing.T) {
	gasGranted := uint64(10)
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			newCtx = ctx.WithGasMeter(sdk.NewGasMeter(gasGranted))

			// AnteHandlers must have their own defer/recover in order for the BaseApp
			// to know how much gas was used! This is because the GasMeter is created in
			// the AnteHandler, but if it panics the context won't be set properly in
			// runTx's recover call.
			defer func() {
				if r := recover(); r != nil {
					switch rType := r.(type) {
					case sdk.ErrorOutOfGas:
						err = sdkerrors.Wrapf(sdkerrors.ErrOutOfGas, "out of gas in location: %v", rType.Descriptor)
					default:
						panic(r)
					}
				}
			}()

			count, _ := parseTxMemo(tx)
			newCtx.GasMeter().ConsumeGas(uint64(count), "counter-ante")

			return newCtx, nil
		})
	}

	// Setup baseapp.
	var (
		appBuilder *runtime.AppBuilder
		cdc        codec.ProtoCodecMarshaler
	)
	err := depinject.Inject(makeMinimalConfig(), &appBuilder, &cdc)
	require.NoError(t, err)

	testCtx := testutil.DefaultContextWithDB(t, capKey1, sdk.NewTransientStoreKey("transient_test"))

	app := appBuilder.Build(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), testCtx.DB, nil, anteOpt)
	app.SetCMS(testCtx.CMS)

	baseapptestutil.RegisterInterfaces(cdc.InterfaceRegistry())
	baseapptestutil.RegisterCounterServer(app.MsgServiceRouter(), CounterServerImplGasMeterOnly{})

	// patch in TxConfig instead of using an output from x/auth/tx
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)
	// set the TxDecoder in the BaseApp for minimal tx simulations
	app.SetTxDecoder(txConfig.TxDecoder())

	app.InitChain(abci.RequestInitChain{})

	header := tmproto.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	testCases := []struct {
		tx      signing.Tx
		gasUsed uint64
		fail    bool
	}{
		{newTxCounter(txConfig, 0, 0), 0, false},
		{newTxCounter(txConfig, 1, 1), 2, false},
		{newTxCounter(txConfig, 9, 1), 10, false},
		{newTxCounter(txConfig, 1, 9), 10, false},
		{newTxCounter(txConfig, 10, 0), 10, false},
		{newTxCounter(txConfig, 0, 10), 10, false},
		{newTxCounter(txConfig, 0, 8, 2), 10, false},
		{newTxCounter(txConfig, 0, 5, 1, 1, 1, 1, 1), 10, false},
		{newTxCounter(txConfig, 0, 5, 1, 1, 1, 1), 9, false},

		{newTxCounter(txConfig, 9, 2), 11, true},
		{newTxCounter(txConfig, 2, 9), 11, true},
		{newTxCounter(txConfig, 9, 1, 1), 11, true},
		{newTxCounter(txConfig, 1, 8, 1, 1), 11, true},
		{newTxCounter(txConfig, 11, 0), 11, true},
		{newTxCounter(txConfig, 0, 11), 11, true},
		{newTxCounter(txConfig, 0, 5, 11), 16, true},
	}

	for i, tc := range testCases {
		tx := tc.tx
		gInfo, result, err := app.SimDeliver(txConfig.TxEncoder(), tx)

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
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			newCtx = ctx.WithGasMeter(sdk.NewGasMeter(gasGranted))

			defer func() {
				if r := recover(); r != nil {
					switch rType := r.(type) {
					case sdk.ErrorOutOfGas:
						err = sdkerrors.Wrapf(sdkerrors.ErrOutOfGas, "out of gas in location: %v", rType.Descriptor)
					default:
						panic(r)
					}
				}
			}()

			count, _ := parseTxMemo(tx)
			newCtx.GasMeter().ConsumeGas(uint64(count), "counter-ante")

			return
		})
	}

	// Setup baseapp.
	var (
		appBuilder *runtime.AppBuilder
		cdc        codec.ProtoCodecMarshaler
	)
	err := depinject.Inject(makeMinimalConfig(), &appBuilder, &cdc)
	require.NoError(t, err)

	testCtx := testutil.DefaultContextWithDB(t, capKey1, sdk.NewTransientStoreKey("transient_test"))

	app := appBuilder.Build(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), testCtx.DB, nil, anteOpt)
	app.SetCMS(testCtx.CMS)

	baseapptestutil.RegisterInterfaces(cdc.InterfaceRegistry())
	baseapptestutil.RegisterCounterServer(app.MsgServiceRouter(), CounterServerImplGasMeterOnly{})

	// patch in TxConfig instead of using an output from x/auth/tx
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)
	// set the TxDecoder in the BaseApp for minimal tx simulations
	app.SetTxDecoder(txConfig.TxDecoder())
	app.SetParamStore(&paramStore{db: dbm.NewMemDB()})

	app.InitChain(abci.RequestInitChain{
		ConsensusParams: &tmproto.ConsensusParams{
			Block: &tmproto.BlockParams{
				MaxGas: 100,
			},
		},
	})

	header := tmproto.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	testCases := []struct {
		tx                signing.Tx
		numDelivers       int
		gasUsedPerDeliver uint64
		fail              bool
		failAfterDeliver  int
	}{
		{newTxCounter(txConfig, 0, 0), 0, 0, false, 0},
		{newTxCounter(txConfig, 9, 1), 2, 10, false, 0},
		{newTxCounter(txConfig, 10, 0), 3, 10, false, 0},
		{newTxCounter(txConfig, 10, 0), 10, 10, false, 0},
		{newTxCounter(txConfig, 2, 7), 11, 9, false, 0},
		{newTxCounter(txConfig, 10, 0), 10, 10, false, 0}, // hit the limit but pass

		{newTxCounter(txConfig, 10, 0), 11, 10, true, 10},
		{newTxCounter(txConfig, 10, 0), 15, 10, true, 10},
		{newTxCounter(txConfig, 9, 0), 12, 9, true, 11}, // fly past the limit
	}

	for i, tc := range testCases {
		tx := tc.tx

		// reset the block gas
		header := tmproto.Header{Height: app.LastBlockHeight() + 1}
		app.BeginBlock(abci.RequestBeginBlock{Header: header})

		// execute the transaction multiple times
		for j := 0; j < tc.numDelivers; j++ {
			_, result, err := app.SimDeliver(txConfig.TxEncoder(), tx)

			ctx := getDeliverStateCtx(app.BaseApp)

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

// Test custom panic handling within app.DeliverTx method
func TestCustomRunTxPanicHandler(t *testing.T) {
	const customPanicMsg = "test panic"
	anteErr := sdkerrors.Register("fakeModule", 100500, "fakeError")

	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			panic(sdkerrors.Wrap(anteErr, "anteHandler"))
		})
	}

	// Setup baseapp.
	var (
		appBuilder *runtime.AppBuilder
		cdc        codec.ProtoCodecMarshaler
	)
	err := depinject.Inject(makeMinimalConfig(), &appBuilder, &cdc)
	require.NoError(t, err)

	testCtx := testutil.DefaultContextWithDB(t, capKey1, sdk.NewTransientStoreKey("transient_test"))

	app := appBuilder.Build(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), testCtx.DB, nil, anteOpt)
	app.SetCMS(testCtx.CMS)
	baseapptestutil.RegisterInterfaces(cdc.InterfaceRegistry())

	// patch in TxConfig instead of using an output from x/auth/tx
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)
	// set the TxDecoder in the BaseApp for minimal tx simulations
	app.SetTxDecoder(txConfig.TxDecoder())

	app.InitChain(abci.RequestInitChain{})

	header := tmproto.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	app.AddRunTxRecoveryHandler(func(recoveryObj interface{}) error {
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

	// Transaction should panic with custom handler above
	{
		tx := newTxCounter(txConfig, 0, 0)

		require.PanicsWithValue(t, customPanicMsg, func() { app.SimDeliver(txConfig.TxEncoder(), tx) })
	}
}

func TestBaseAppAnteHandler(t *testing.T) {
	anteKey := []byte("ante-key")
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, anteKey))
	}

	// Setup baseapp.
	var (
		appBuilder *runtime.AppBuilder
		cdc        codec.ProtoCodecMarshaler
	)
	err := depinject.Inject(makeMinimalConfig(), &appBuilder, &cdc)
	require.NoError(t, err)

	testCtx := testutil.DefaultContextWithDB(t, capKey1, sdk.NewTransientStoreKey("transient_test"))

	app := appBuilder.Build(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), testCtx.DB, nil, anteOpt)
	app.SetCMS(testCtx.CMS)
	baseapptestutil.RegisterInterfaces(cdc.InterfaceRegistry())
	deliverKey := []byte("deliver-key")
	baseapptestutil.RegisterCounterServer(app.MsgServiceRouter(), CounterServerImpl{t, capKey1, deliverKey})

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	// patch in TxConfig instead of using an output from x/auth/tx
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)
	// set the TxDecoder in the BaseApp for minimal tx simulations
	app.SetTxDecoder(txConfig.TxDecoder())

	// execute a tx that will fail ante handler execution
	//
	// NOTE: State should not be mutated here. This will be implicitly checked by
	// the next txs ante handler execution (anteHandlerTxTest).
	tx := newTxCounter(txConfig, 0, 0)
	tx = setFailOnAnte(txConfig, tx, true)
	txBytes, err := txConfig.TxEncoder()(tx)
	require.NoError(t, err)
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.Empty(t, res.Events)
	require.False(t, res.IsOK(), fmt.Sprintf("%v", res))

	ctx := getDeliverStateCtx(app.BaseApp)
	store := ctx.KVStore(capKey1)
	require.Equal(t, int64(0), getIntFromStore(store, anteKey))

	// execute at tx that will pass the ante handler (the checkTx state should
	// mutate) but will fail the message handler
	tx = newTxCounter(txConfig, 0, 0)
	tx = setFailOnHandler(txConfig, tx, true)

	txBytes, err = txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	// should emit ante event
	require.NotEmpty(t, res.Events)
	require.False(t, res.IsOK(), fmt.Sprintf("%v", res))

	ctx = getDeliverStateCtx(app.BaseApp)
	store = ctx.KVStore(capKey1)
	require.Equal(t, int64(1), getIntFromStore(store, anteKey))
	require.Equal(t, int64(0), getIntFromStore(store, deliverKey))

	// execute a successful ante handler and message execution where state is
	// implicitly checked by previous tx executions
	tx = newTxCounter(txConfig, 1, 0)

	txBytes, err = txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.NotEmpty(t, res.Events)
	require.True(t, res.IsOK(), fmt.Sprintf("%v", res))

	ctx = getDeliverStateCtx(app.BaseApp)
	store = ctx.KVStore(capKey1)
	require.Equal(t, int64(2), getIntFromStore(store, anteKey))
	require.Equal(t, int64(1), getIntFromStore(store, deliverKey))

	// commit
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()
}

func TestGasConsumptionBadTx(t *testing.T) {
	gasWanted := uint64(5)
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			newCtx = ctx.WithGasMeter(sdk.NewGasMeter(gasWanted))

			defer func() {
				if r := recover(); r != nil {
					switch rType := r.(type) {
					case sdk.ErrorOutOfGas:
						log := fmt.Sprintf("out of gas in location: %v", rType.Descriptor)
						err = sdkerrors.Wrap(sdkerrors.ErrOutOfGas, log)
					default:
						panic(r)
					}
				}
			}()

			counter, failOnAnte := parseTxMemo(tx)
			newCtx.GasMeter().ConsumeGas(uint64(counter), "counter-ante")
			if failOnAnte {
				return newCtx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "ante handler failure")
			}

			return
		})
	}

	// Setup baseapp.
	var (
		appBuilder *runtime.AppBuilder
		cdc        codec.ProtoCodecMarshaler
	)
	err := depinject.Inject(makeMinimalConfig(), &appBuilder, &cdc)
	require.NoError(t, err)

	testCtx := testutil.DefaultContextWithDB(t, capKey1, sdk.NewTransientStoreKey("transient_test"))

	app := appBuilder.Build(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), testCtx.DB, nil, anteOpt)
	app.SetCMS(testCtx.CMS)

	baseapptestutil.RegisterInterfaces(cdc.InterfaceRegistry())
	baseapptestutil.RegisterCounterServer(app.MsgServiceRouter(), CounterServerImplGasMeterOnly{})

	// patch in TxConfig instead of using an output from x/auth/tx
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)
	// set the TxDecoder in the BaseApp for minimal tx simulations
	app.SetTxDecoder(txConfig.TxDecoder())
	app.SetParamStore(&paramStore{db: dbm.NewMemDB()})

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

	tx := newTxCounter(txConfig, 5, 0)
	tx = setFailOnAnte(txConfig, tx, true)
	txBytes, err := txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.False(t, res.IsOK(), fmt.Sprintf("%v", res))

	// require next tx to fail due to black gas limit
	tx = newTxCounter(txConfig, 5, 0)
	txBytes, err = txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.False(t, res.IsOK(), fmt.Sprintf("%v", res))
}

// Test that we can only query from the latest committed state.
func TestQuery(t *testing.T) {
	key, value := []byte("hello"), []byte("goodbye")
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			store := ctx.KVStore(capKey1)
			store.Set(key, value)
			return
		})
	}

	// Setup baseapp.
	var (
		appBuilder *runtime.AppBuilder
		cdc        codec.ProtoCodecMarshaler
	)
	err := depinject.Inject(makeMinimalConfig(), &appBuilder, &cdc)
	require.NoError(t, err)

	testCtx := testutil.DefaultContextWithDB(t, capKey1, sdk.NewTransientStoreKey("transient_test"))

	app := appBuilder.Build(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), testCtx.DB, nil, anteOpt)
	app.SetCMS(testCtx.CMS)

	baseapptestutil.RegisterInterfaces(cdc.InterfaceRegistry())
	baseapptestutil.RegisterCounterServer(app.MsgServiceRouter(), CounterServerImplGasMeterOnly{})

	// patch in TxConfig instead of using an output from x/auth/tx
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)
	// set the TxDecoder in the BaseApp for minimal tx simulations
	app.SetTxDecoder(txConfig.TxDecoder())
	app.SetParamStore(&paramStore{db: dbm.NewMemDB()})

	app.InitChain(abci.RequestInitChain{
		ConsensusParams: &tmproto.ConsensusParams{},
	})

	// NOTE: "/store/key1" tells us KVStore
	// and the final "/key" says to use the data as the
	// key in the given KVStore ...
	query := abci.RequestQuery{
		Path: "/store/key1/key",
		Data: key,
	}
	tx := newTxCounter(txConfig, 0, 0)

	// query is empty before we do anything
	res := app.Query(query)
	require.Equal(t, 0, len(res.Value))

	// query is still empty after a CheckTx
	_, resTx, err := app.SimCheck(txConfig.TxEncoder(), tx)
	require.NoError(t, err)
	require.NotNil(t, resTx)
	res = app.Query(query)
	require.Equal(t, 0, len(res.Value))

	// query is still empty after a DeliverTx before we commit
	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	_, resTx, err = app.SimDeliver(txConfig.TxEncoder(), tx)
	require.NoError(t, err)
	require.NotNil(t, resTx)
	res = app.Query(query)
	require.Equal(t, 0, len(res.Value))

	// query returns correct value after Commit
	app.Commit()
	res = app.Query(query)
	require.Equal(t, value, res.Value)
}

func getCheckStateCtx(app *baseapp.BaseApp) sdk.Context {
	v := reflect.ValueOf(app).Elem()
	f := v.FieldByName("checkState")
	rf := reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
	return rf.MethodByName("Context").Call(nil)[0].Interface().(sdk.Context)
}

func getDeliverStateCtx(app *baseapp.BaseApp) sdk.Context {
	v := reflect.ValueOf(app).Elem()
	f := v.FieldByName("deliverState")
	rf := reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
	return rf.MethodByName("Context").Call(nil)[0].Interface().(sdk.Context)
}

func parseTxMemo(tx sdk.Tx) (counter int64, failOnAnte bool) {
	txWithMemo, ok := tx.(sdk.TxWithMemo)
	if !ok {
		panic("not a sdk.TxWithMemo")
	}
	memo := txWithMemo.GetMemo()
	vals, err := url.ParseQuery(memo)
	if err != nil {
		panic("invalid memo")
	}

	counter, err = strconv.ParseInt(vals.Get("counter"), 10, 64)
	if err != nil {
		panic("invalid counter")
	}

	failOnAnte = vals.Get("failOnAnte") == "true"

	return counter, failOnAnte
}

func counterEvent(evType string, msgCount int64) sdk.Events {
	return sdk.Events{
		sdk.NewEvent(
			evType,
			sdk.NewAttribute("update_counter", fmt.Sprintf("%d", msgCount)),
		),
	}
}

func newTxCounter(cfg client.TxConfig, counter int64, msgCounters ...int64) signing.Tx {
	msgs := make([]sdk.Msg, 0, len(msgCounters))
	for _, c := range msgCounters {
		msg := &baseapptestutil.MsgCounter{Counter: c, FailOnHandler: false}
		msgs = append(msgs, msg)
	}

	builder := cfg.NewTxBuilder()
	builder.SetMsgs(msgs...)
	builder.SetMemo("counter=" + strconv.FormatInt(counter, 10) + "&failOnAnte=false")

	return builder.GetTx()
}

func setFailOnAnte(cfg client.TxConfig, tx signing.Tx, failOnAnte bool) signing.Tx {
	builder := cfg.NewTxBuilder()
	builder.SetMsgs(tx.GetMsgs()...)

	memo := tx.GetMemo()
	vals, err := url.ParseQuery(memo)
	if err != nil {
		panic("invalid memo")
	}

	vals.Set("failOnAnte", strconv.FormatBool(failOnAnte))
	memo = vals.Encode()
	builder.SetMemo(memo)

	return builder.GetTx()
}

func setFailOnHandler(cfg client.TxConfig, tx signing.Tx, fail bool) signing.Tx {
	builder := cfg.NewTxBuilder()
	builder.SetMemo(tx.GetMemo())

	msgs := tx.GetMsgs()
	for i, msg := range msgs {
		msgs[i] = &baseapptestutil.MsgCounter{
			Counter:       msg.(*baseapptestutil.MsgCounter).Counter,
			FailOnHandler: fail,
		}
	}

	builder.SetMsgs(msgs...)
	return builder.GetTx()
}

func anteHandlerTxTest(t *testing.T, capKey storetypes.StoreKey, storeKey []byte) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		store := ctx.KVStore(capKey)
		counter, failOnAnte := parseTxMemo(tx)

		if failOnAnte {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "ante handler failure")
		}

		_, err := incrementingCounter(t, store, storeKey, counter)
		if err != nil {
			return ctx, err
		}

		ctx.EventManager().EmitEvents(
			counterEvent("ante_handler", counter),
		)

		ctx = ctx.WithPriority(testTxPriority)

		return ctx, nil
	}
}

// check counter matches what's in store.
// increment and store
func incrementingCounter(t *testing.T, store sdk.KVStore, counterKey []byte, counter int64) (*sdk.Result, error) {
	storedCounter := getIntFromStore(store, counterKey)
	require.Equal(t, storedCounter, counter)
	setIntOnStore(store, counterKey, counter+1)
	return &sdk.Result{}, nil
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

type MsgKeyValueImpl struct{}

func (m MsgKeyValueImpl) Set(ctx context.Context, msg *baseapptestutil.MsgKeyValue) (*baseapptestutil.MsgCreateKeyValueResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.KVStore(capKey2).Set(msg.Key, msg.Value)
	return &baseapptestutil.MsgCreateKeyValueResponse{}, nil
}

type CounterServerImpl struct {
	t          *testing.T
	capKey     storetypes.StoreKey
	deliverKey []byte
}

func (m CounterServerImpl) IncrementCounter(ctx context.Context, msg *baseapptestutil.MsgCounter) (*baseapptestutil.MsgCreateCounterResponse, error) {
	return incrementCounter(ctx, m.t, m.capKey, m.deliverKey, msg)
}

type Counter2ServerImpl struct {
	t          *testing.T
	capKey     storetypes.StoreKey
	deliverKey []byte
}

func (m Counter2ServerImpl) IncrementCounter(ctx context.Context, msg *baseapptestutil.MsgCounter2) (*baseapptestutil.MsgCreateCounterResponse, error) {
	return incrementCounter(ctx, m.t, m.capKey, m.deliverKey, msg)
}

func incrementCounter(ctx context.Context,
	t *testing.T,
	capKey storetypes.StoreKey,
	deliverKey []byte,
	msg sdk.Msg,
) (*baseapptestutil.MsgCreateCounterResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(capKey)

	sdkCtx.GasMeter().ConsumeGas(5, "test")

	var msgCount int64

	switch m := msg.(type) {
	case *baseapptestutil.MsgCounter:
		if m.FailOnHandler {
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "message handler failure")
		}
		msgCount = m.Counter
	case *baseapptestutil.MsgCounter2:
		if m.FailOnHandler {
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "message handler failure")
		}
		msgCount = m.Counter
	}

	sdkCtx.EventManager().EmitEvents(
		counterEvent(sdk.EventTypeMessage, msgCount),
	)

	_, err := incrementingCounter(t, store, deliverKey, msgCount)
	if err != nil {
		return nil, err
	}

	return &baseapptestutil.MsgCreateCounterResponse{}, nil
}

type CounterServerImplGasMeterOnly struct {
	gas uint64
}

func (m CounterServerImplGasMeterOnly) IncrementCounter(ctx context.Context, msg *baseapptestutil.MsgCounter) (*baseapptestutil.MsgCreateCounterResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	gas := m.gas

	// if no gas is provided, use the counter as gas. This is useful for testing
	if gas == 0 {
		gas = uint64(msg.Counter)
	}
	sdkCtx.GasMeter().ConsumeGas(gas, "test")
	return &baseapptestutil.MsgCreateCounterResponse{}, nil
}

type paramStore struct {
	db *dbm.MemDB
}

var ParamstoreKey = []byte("paramstore")

func (ps *paramStore) Set(_ sdk.Context, value *tmproto.ConsensusParams) {
	bz, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}

	ps.db.Set(ParamstoreKey, bz)
}

func (ps *paramStore) Has(_ sdk.Context) bool {
	ok, err := ps.db.Has(ParamstoreKey)
	if err != nil {
		panic(err)
	}

	return ok
}

func (ps paramStore) Get(ctx sdk.Context) (*tmproto.ConsensusParams, error) {
	bz, err := ps.db.Get(ParamstoreKey)
	if err != nil {
		panic(err)
	}

	if len(bz) == 0 {
		return nil, errors.New("params not found")
	}

	var params tmproto.ConsensusParams

	if err := json.Unmarshal(bz, &params); err != nil {
		panic(err)
	}

	return &params, nil
}
