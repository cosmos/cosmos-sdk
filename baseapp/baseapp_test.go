package baseapp_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	coretesting "cosmossdk.io/core/testing"
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
	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
)

var (
	capKey1 = sdk.NewKVStoreKey("key1")
	capKey2 = sdk.NewKVStoreKey("key2")
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

func newBaseApp(name string, options ...func(*BaseApp)) *BaseApp {
	logger := defaultLogger()
	db := dbm.NewMemDB()
	codec := codec.NewLegacyAmino()
	registerTestCodec(codec)
	return NewBaseApp(name, logger, db, testTxDecoder(codec), options...)
}

func registerTestCodec(cdc *codec.LegacyAmino) {
	// register Tx, Msg
	sdk.RegisterLegacyAminoCodec(cdc)

	// register test types
	cdc.RegisterConcrete(&txTest{}, "cosmos-sdk/baseapp/txTest", nil)
	cdc.RegisterConcrete(&msgCounter{}, "cosmos-sdk/baseapp/msgCounter", nil)
	cdc.RegisterConcrete(&msgCounter2{}, "cosmos-sdk/baseapp/msgCounter2", nil)
	cdc.RegisterConcrete(&msgKeyValue{}, "cosmos-sdk/baseapp/msgKeyValue", nil)
	cdc.RegisterConcrete(&msgNoRoute{}, "cosmos-sdk/baseapp/msgNoRoute", nil)
}

// aminoTxEncoder creates a amino TxEncoder for testing purposes.
func aminoTxEncoder() sdk.TxEncoder {
	cdc := codec.NewLegacyAmino()
	registerTestCodec(cdc)

	return legacytx.StdTxConfig{Cdc: cdc}.TxEncoder()
}

	app.SetInterfaceRegistry(cdc.InterfaceRegistry())
	app.MsgServiceRouter().SetInterfaceRegistry(cdc.InterfaceRegistry())
	app.MountStores(capKey1, capKey2)
	app.SetParamStore(paramStore{db: coretesting.NewMemDB()})
	app.SetTxDecoder(txConfig.TxDecoder())
	app.SetTxEncoder(txConfig.TxEncoder())
	app.SetVersionModifier(newMockedVersionModifier(0))

	// mount stores and seal
	require.Nil(t, app.LoadLatestVersion())

	return &BaseAppSuite{
		baseApp:   app,
		cdc:       cdc,
		txConfig:  txConfig,
		logBuffer: logBuffer,
		ac:        cdc.InterfaceRegistry().SigningContext().AddressCodec(),
	}
}

func getQueryBaseapp(t *testing.T) *baseapp.BaseApp {
	t.Helper()

	db := coretesting.NewMemDB()
	name := t.Name()
	app := baseapp.NewBaseApp(name, log.NewTestLogger(t), db, nil)

	_, err := app.FinalizeBlock(&abci.FinalizeBlockRequest{Height: 1})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)

	_, err = app.FinalizeBlock(&abci.FinalizeBlockRequest{Height: 2})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)

	return app
}

func TestLoadVersionPruning(t *testing.T) {
	logger := log.NewNopLogger()
	pruningOptions := store.PruningOptions{
		KeepRecent: 2,
		KeepEvery:  3,
		Interval:   1,
	}
	pruningOpt := SetPruning(pruningOptions)
	db := dbm.NewMemDB()
	name := t.Name()
	app := NewBaseApp(name, logger, db, nil, pruningOpt)

	// make a cap key and mount the store
	capKey := sdk.NewKVStoreKey("key1")
	app.MountStores(capKey)

	err := app.LoadLatestVersion() // needed to make stores non-nil
	require.Nil(t, err)

	emptyCommitID := sdk.CommitID{}

	// fresh store has zero/empty last commit
	lastHeight := app.LastBlockHeight()
	lastID := app.LastCommitID()
	require.Equal(t, int64(0), lastHeight)
	require.Equal(t, emptyCommitID, lastID)

	var lastCommitID sdk.CommitID

	// Commit seven blocks, of which 7 (latest) is kept in addition to 6, 5
	// (keep recent) and 3 (keep every).
	for i := int64(1); i <= 7; i++ {
		app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: i}})
		res := app.Commit()
		lastCommitID = sdk.CommitID{Version: i, Hash: res.Data}
	}

	for _, v := range []int64{1, 2, 4} {
		_, err = app.cms.CacheMultiStoreWithVersion(v)
		require.Error(t, err)
	}

	for _, v := range []int64{3, 5, 6, 7} {
		_, err = app.cms.CacheMultiStoreWithVersion(v)
		require.NoError(t, err)
	}

	// reload with LoadLatestVersion, check it loads last version
	app = NewBaseApp(name, logger, db, nil, pruningOpt)
	app.MountStores(capKey)

	err = app.LoadLatestVersion()
	require.Nil(t, err)
	testLoadVersionHelper(t, app, int64(7), lastCommitID)
}

func testLoadVersionHelper(t *testing.T, app *BaseApp, expectedHeight int64, expectedID sdk.CommitID) {
	lastHeight := app.LastBlockHeight()
	lastID := app.LastCommitID()
	require.Equal(t, expectedHeight, lastHeight)
	require.Equal(t, expectedID, lastID)
}

func TestSetMinGasPrices(t *testing.T) {
	minGasPrices := sdk.DecCoins{sdk.NewInt64DecCoin("stake", 5000)}
	app := newBaseApp(t.Name(), SetMinGasPrices(minGasPrices.String()))
	require.Equal(t, minGasPrices, app.minGasPrices)
}

func TestGetMaximumBlockGas(t *testing.T) {
	app := setupBaseApp(t)
	app.InitChain(abci.RequestInitChain{})
	ctx := app.NewContext(true, tmproto.Header{})

	app.StoreConsensusParams(ctx, &abci.ConsensusParams{Block: &abci.BlockParams{MaxGas: 0}})
	require.Equal(t, uint64(0), app.getMaximumBlockGas(ctx))

	app.StoreConsensusParams(ctx, &abci.ConsensusParams{Block: &abci.BlockParams{MaxGas: -1}})
	require.Equal(t, uint64(0), app.getMaximumBlockGas(ctx))

	app.StoreConsensusParams(ctx, &abci.ConsensusParams{Block: &abci.BlockParams{MaxGas: 5000000}})
	require.Equal(t, uint64(5000000), app.getMaximumBlockGas(ctx))

	app.StoreConsensusParams(ctx, &abci.ConsensusParams{Block: &abci.BlockParams{MaxGas: -5000000}})
	require.Panics(t, func() { app.getMaximumBlockGas(ctx) })
}

func TestListSnapshots(t *testing.T) {
	type setupConfig struct {
		blocks            uint64
		blockTxs          int
		snapshotInterval  uint64
		snapshotKeepEvery uint32
	}

	app, _ := setupBaseAppWithSnapshots(t, 2, 5)

	expected := abci.ResponseListSnapshots{Snapshots: []*abci.Snapshot{
		{Height: 2, Format: 2, Chunks: 2},
	}}

	resp := app.ListSnapshots(abci.RequestListSnapshots{})
	queryResponse := app.Query(abci.RequestQuery{
		Path: "/app/snapshots",
	})

	queryListSnapshotsResp := abci.ResponseListSnapshots{}
	err := json.Unmarshal(queryResponse.Value, &queryListSnapshotsResp)
	require.NoError(t, err)

	for i, s := range resp.Snapshots {
		querySnapshot := queryListSnapshotsResp.Snapshots[i]
		// we check that the query snapshot and function snapshot are equal
		// Then we check that the hash and metadata are not empty. We atm
		// do not have a good way to generate the expected value for these.
		assert.Equal(t, *s, *querySnapshot)
		assert.NotEmpty(t, s.Hash)
		assert.NotEmpty(t, s.Metadata)
		// Set hash and metadata to nil, so we can check the other snapshot
		// fields against expected
		s.Hash = nil
		s.Metadata = nil
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

//---------------------------------------------------------------------
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

	anteOpt := func(bapp *BaseApp) { bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, counterKey)) }
	routerOpt := func(bapp *BaseApp) {
		// TODO: can remove this once CheckTx doesnt process msgs.
		bapp.Router().AddRoute(sdk.NewRoute(routeMsgCounter, func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
			return &sdk.Result{}, nil
		}))
	}

	app := setupBaseApp(t, anteOpt, routerOpt)

	nTxs := int64(5)
	app.InitChain(abci.RequestInitChain{})

	// Create same codec used in txDecoder
	codec := codec.NewLegacyAmino()
	registerTestCodec(codec)

	for i := int64(0); i < nTxs; i++ {
		tx := newTxCounter(i, 0) // no messages
		txBytes, err := codec.Marshal(tx)
		require.NoError(t, err)
		r := app.CheckTx(abci.RequestCheckTx{Tx: txBytes})
		require.Equal(t, testTxPriority, r.Priority)
		require.Empty(t, r.GetEvents())
		require.True(t, r.IsOK(), fmt.Sprintf("%v", r))
	}

	checkStateStore := app.checkState.ctx.KVStore(capKey1)
	storedCounter := getIntFromStore(checkStateStore, counterKey)

	// Ensure AnteHandler ran
	require.Equal(t, nTxs, storedCounter)

	// If a block is committed, CheckTx state should be reset.
	header := tmproto.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header, Hash: []byte("hash")})

	require.NotNil(t, app.checkState.ctx.BlockGasMeter(), "block gas meter should have been set to checkState")
	require.NotEmpty(t, app.checkState.ctx.HeaderHash())

	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	checkStateStore = app.checkState.ctx.KVStore(capKey1)
	storedBytes := checkStateStore.Get(counterKey)
	require.Nil(t, storedBytes)
}

// Test that successive DeliverTx can see each others' effects
// on the store, both within and across blocks.
func TestDeliverTx(t *testing.T) {
	// test increments in the ante
	anteKey := []byte("ante-key")
	anteOpt := func(bapp *BaseApp) { bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, anteKey)) }

	// test increments in the handler
	deliverKey := []byte("deliver-key")
	routerOpt := func(bapp *BaseApp) {
		r := sdk.NewRoute(routeMsgCounter, handlerMsgCounter(t, capKey1, deliverKey))
		bapp.Router().AddRoute(r)
	}

	app := setupBaseApp(t, anteOpt, routerOpt)
	app.InitChain(abci.RequestInitChain{})

	// Create same codec used in txDecoder
	codec := codec.NewLegacyAmino()
	registerTestCodec(codec)

	nBlocks := 3
	txPerHeight := 5

	for blockN := 0; blockN < nBlocks; blockN++ {
		header := tmproto.Header{Height: int64(blockN) + 1}
		app.BeginBlock(abci.RequestBeginBlock{Header: header})

		for i := 0; i < txPerHeight; i++ {
			counter := int64(blockN*txPerHeight + i)
			tx := newTxCounter(counter, counter)

			txBytes, err := codec.Marshal(tx)
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

// Number of messages doesn't matter to CheckTx.
func TestMultiMsgCheckTx(t *testing.T) {
	// TODO: ensure we get the same results
	// with one message or many
}

// One call to DeliverTx should process all the messages, in order.
func TestMultiMsgDeliverTx(t *testing.T) {
	// increment the tx counter
	anteKey := []byte("ante-key")
	anteOpt := func(bapp *BaseApp) { bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, anteKey)) }

	// increment the msg counter
	deliverKey := []byte("deliver-key")
	deliverKey2 := []byte("deliver-key2")
	routerOpt := func(bapp *BaseApp) {
		r1 := sdk.NewRoute(routeMsgCounter, handlerMsgCounter(t, capKey1, deliverKey))
		r2 := sdk.NewRoute(routeMsgCounter2, handlerMsgCounter(t, capKey1, deliverKey2))
		bapp.Router().AddRoute(r1)
		bapp.Router().AddRoute(r2)
	}

	app := setupBaseApp(t, anteOpt, routerOpt)

	// Create same codec used in txDecoder
	codec := codec.NewLegacyAmino()
	registerTestCodec(codec)

	// run a multi-msg tx
	// with all msgs the same route

	header := tmproto.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	tx := newTxCounter(0, 0, 1, 2)
	txBytes, err := codec.Marshal(tx)
	require.NoError(t, err)
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.True(t, res.IsOK(), fmt.Sprintf("%v", res))

	store := app.deliverState.ctx.KVStore(capKey1)

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
	txBytes, err = codec.Marshal(tx)
	require.NoError(t, err)
	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.True(t, res.IsOK(), fmt.Sprintf("%v", res))

	store = app.deliverState.ctx.KVStore(capKey1)

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

	anteOpt := func(bapp *BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			newCtx = ctx.WithGasMeter(sdk.NewGasMeter(gasConsumed))
			return
		})
	}

	routerOpt := func(bapp *BaseApp) {
		r := sdk.NewRoute(routeMsgCounter, func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
			ctx.GasMeter().ConsumeGas(gasConsumed, "test")
			return &sdk.Result{}, nil
		})
		bapp.Router().AddRoute(r)
	}

	app := setupBaseApp(t, anteOpt, routerOpt)

	app.InitChain(abci.RequestInitChain{})

	// Create same codec used in txDecoder
	cdc := codec.NewLegacyAmino()
	registerTestCodec(cdc)

	nBlocks := 3
	for blockN := 0; blockN < nBlocks; blockN++ {
		count := int64(blockN + 1)
		header := tmproto.Header{Height: count}
		app.BeginBlock(abci.RequestBeginBlock{Header: header})

		tx := newTxCounter(count, count)
		txBytes, err := cdc.Marshal(tx)
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

func TestEventHistory(t *testing.T) {
	// generates events in the ante
	anteKey := []byte("ante-key")
	anteOpt := func(bapp *BaseApp) { bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, anteKey)) }

	deliverKey := []byte("deliver-key")
	routerOpt := func(bapp *BaseApp) {
		r := sdk.NewRoute(routeMsgCounter, handlerMsgCounter(t, capKey1, deliverKey))
		bapp.Router().AddRoute(r)
	}

	// expand setupBaseApp() here so we can override the EndBlocker before app is sealed
	app := newBaseApp(t.Name(), anteOpt, routerOpt)
	require.Equal(t, t.Name(), app.Name())

	app.MountStores(capKey1, capKey2)
	app.SetParamStore(&paramStore{db: dbm.NewMemDB()})

	var history []abci.Event
	app.SetEndBlocker(func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
		history = ctx.EventManager().GetABCIEventHistory()
		return abci.ResponseEndBlock{}
	})

	// stores are mounted (seals the app)
	err := app.LoadLatestVersion()
	require.Nil(t, err)

	app.InitChain(abci.RequestInitChain{})

	// Create same codec used in txDecoder
	cdc := codec.NewLegacyAmino()
	registerTestCodec(cdc)

	header := tmproto.Header{Height: int64(1)}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	require.Zero(t, len(app.deliverState.eventHistory))

	counter := int64(0)
	tx := newTxCounter(counter, counter)
	txBytes, err := cdc.Marshal(tx)
	require.NoError(t, err)

	// simulation returns events, but none in deliverState history
	_, simResult, err := app.Simulate(txBytes)
	require.NoError(t, err)
	require.NotNil(t, simResult)
	require.Zero(t, len(app.deliverState.eventHistory))
	expectedEvents := simResult.Events
	require.NotZero(t, len(expectedEvents))

	// delivery results should be reflected in deliverState history
	deliverResult := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.True(t, deliverResult.IsOK(), fmt.Sprintf("%v", deliverResult))
	require.NotZero(t, len(app.deliverState.eventHistory))
	// simResult events are not indexed, but should otherwise be the same as deliverResult events
	for i, _ := range simResult.Events {
		simResult.Events[i].Attributes[0].Index = true
	}
	require.Equal(t, expectedEvents, deliverResult.Events)
	require.Equal(t, expectedEvents, app.deliverState.eventHistory)

	// the deliverState event history should be passed to the end blocker in the context, then cleared
	app.EndBlock(abci.RequestEndBlock{})
	require.Zero(t, len(app.deliverState.eventHistory))
	require.Equal(t, expectedEvents, history)
}

func TestRunInvalidTransaction(t *testing.T) {
	anteOpt := func(bapp *BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			return
		})
	}
	routerOpt := func(bapp *BaseApp) {
		r := sdk.NewRoute(routeMsgCounter, func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
			return &sdk.Result{}, nil
		})
		bapp.Router().AddRoute(r)
	}

	app := setupBaseApp(t, anteOpt, routerOpt)

	header := tmproto.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	// transaction with no messages
	{
		emptyTx := &txTest{}
		_, result, err := app.SimDeliver(aminoTxEncoder(), emptyTx)
		require.Error(t, err)
		require.Nil(t, result)

		space, code, _ := sdkerrors.ABCIInfo(err, false)
		require.EqualValues(t, sdkerrors.ErrInvalidRequest.Codespace(), space, err)
		require.EqualValues(t, sdkerrors.ErrInvalidRequest.ABCICode(), code, err)
	}

	// transaction where ValidateBasic fails
	{
		testCases := []struct {
			tx   *txTest
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
			_, result, err := app.SimDeliver(aminoTxEncoder(), tx)

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
		unknownRouteTx := txTest{[]sdk.Msg{msgNoRoute{}}, 0, false}
		_, result, err := app.SimDeliver(aminoTxEncoder(), unknownRouteTx)
		require.Error(t, err)
		require.Nil(t, result)

		space, code, _ := sdkerrors.ABCIInfo(err, false)
		require.EqualValues(t, sdkerrors.ErrUnknownRequest.Codespace(), space, err)
		require.EqualValues(t, sdkerrors.ErrUnknownRequest.ABCICode(), code, err)

		unknownRouteTx = txTest{[]sdk.Msg{msgCounter{}, msgNoRoute{}}, 0, false}
		_, result, err = app.SimDeliver(aminoTxEncoder(), unknownRouteTx)
		require.Error(t, err)
		require.Nil(t, result)

		space, code, _ = sdkerrors.ABCIInfo(err, false)
		require.EqualValues(t, sdkerrors.ErrUnknownRequest.Codespace(), space, err)
		require.EqualValues(t, sdkerrors.ErrUnknownRequest.ABCICode(), code, err)
	}

	// Transaction with an unregistered message
	{
		tx := newTxCounter(0, 0)
		tx.Msgs = append(tx.Msgs, msgNoDecode{})

		// new codec so we can encode the tx, but we shouldn't be able to decode
		newCdc := codec.NewLegacyAmino()
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
	anteOpt := func(bapp *BaseApp) {
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

			count := tx.(txTest).Counter
			newCtx.GasMeter().ConsumeGas(uint64(count), "counter-ante")

			return newCtx, nil
		})
	}

	routerOpt := func(bapp *BaseApp) {
		r := sdk.NewRoute(routeMsgCounter, func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
			count := msg.(*msgCounter).Counter
			ctx.GasMeter().ConsumeGas(uint64(count), "counter-handler")
			return &sdk.Result{}, nil
		})
		bapp.Router().AddRoute(r)
	}

	app := setupBaseApp(t, anteOpt, routerOpt)

	header := tmproto.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	testCases := []struct {
		tx      *txTest
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
		gInfo, result, err := app.SimDeliver(aminoTxEncoder(), tx)

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
	anteOpt := func(bapp *BaseApp) {
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

			count := tx.(txTest).Counter
			newCtx.GasMeter().ConsumeGas(uint64(count), "counter-ante")

			return
		})
	}

	routerOpt := func(bapp *BaseApp) {
		r := sdk.NewRoute(routeMsgCounter, func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
			count := msg.(*msgCounter).Counter
			ctx.GasMeter().ConsumeGas(uint64(count), "counter-handler")
			return &sdk.Result{}, nil
		})
		bapp.Router().AddRoute(r)
	}

	app := setupBaseApp(t, anteOpt, routerOpt)
	app.InitChain(abci.RequestInitChain{
		ConsensusParams: &abci.ConsensusParams{
			Block: &abci.BlockParams{
				MaxGas: 100,
			},
		},
	})

	testCases := []struct {
		tx                *txTest
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

		// reset the block gas
		header := tmproto.Header{Height: app.LastBlockHeight() + 1}
		app.BeginBlock(abci.RequestBeginBlock{Header: header})

		// execute the transaction multiple times
		for j := 0; j < tc.numDelivers; j++ {
			_, result, err := app.SimDeliver(aminoTxEncoder(), tx)

			ctx := app.getState(runTxModeDeliver).ctx

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

	anteOpt := func(bapp *BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			panic(sdkerrors.Wrap(anteErr, "anteHandler"))
		})
	}
	routerOpt := func(bapp *BaseApp) {
		r := sdk.NewRoute(routeMsgCounter, func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
			return &sdk.Result{}, nil
		})
		bapp.Router().AddRoute(r)
	}

	app := setupBaseApp(t, anteOpt, routerOpt)

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
		tx := newTxCounter(0, 0)

		require.PanicsWithValue(t, customPanicMsg, func() { app.SimDeliver(aminoTxEncoder(), tx) })
	}
}

func TestBaseAppAnteHandler(t *testing.T) {
	anteKey := []byte("ante-key")
	anteOpt := func(bapp *BaseApp) {
		bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, anteKey))
	}

	deliverKey := []byte("deliver-key")
	routerOpt := func(bapp *BaseApp) {
		r := sdk.NewRoute(routeMsgCounter, handlerMsgCounter(t, capKey1, deliverKey))
		bapp.Router().AddRoute(r)
	}

	cdc := codec.NewLegacyAmino()
	app := setupBaseApp(t, anteOpt, routerOpt)

	app.InitChain(abci.RequestInitChain{})
	registerTestCodec(cdc)

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	// execute a tx that will fail ante handler execution
	//
	// NOTE: State should not be mutated here. This will be implicitly checked by
	// the next txs ante handler execution (anteHandlerTxTest).
	tx := newTxCounter(0, 0)
	tx.setFailOnAnte(true)
	txBytes, err := cdc.Marshal(tx)
	require.NoError(t, err)
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.Empty(t, res.Events)
	require.False(t, res.IsOK(), fmt.Sprintf("%v", res))

	ctx := app.getState(runTxModeDeliver).ctx
	store := ctx.KVStore(capKey1)
	require.Equal(t, int64(0), getIntFromStore(store, anteKey))

	// execute at tx that will pass the ante handler (the checkTx state should
	// mutate) but will fail the message handler
	tx = newTxCounter(0, 0)
	tx.setFailOnHandler(true)

	txBytes, err = cdc.Marshal(tx)
	require.NoError(t, err)

	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	// should emit ante event
	require.NotEmpty(t, res.Events)
	require.False(t, res.IsOK(), fmt.Sprintf("%v", res))

	ctx = app.getState(runTxModeDeliver).ctx
	store = ctx.KVStore(capKey1)
	require.Equal(t, int64(1), getIntFromStore(store, anteKey))
	require.Equal(t, int64(0), getIntFromStore(store, deliverKey))

	// execute a successful ante handler and message execution where state is
	// implicitly checked by previous tx executions
	tx = newTxCounter(1, 0)

	txBytes, err = cdc.Marshal(tx)
	require.NoError(t, err)

	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.NotEmpty(t, res.Events)
	require.True(t, res.IsOK(), fmt.Sprintf("%v", res))

	ctx = app.getState(runTxModeDeliver).ctx
	store = ctx.KVStore(capKey1)
	require.Equal(t, int64(2), getIntFromStore(store, anteKey))
	require.Equal(t, int64(1), getIntFromStore(store, deliverKey))

	// commit
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()
}

func TestGasConsumptionBadTx(t *testing.T) {
	gasWanted := uint64(5)
	anteOpt := func(bapp *BaseApp) {
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

			txTest := tx.(txTest)
			newCtx.GasMeter().ConsumeGas(uint64(txTest.Counter), "counter-ante")
			if txTest.FailOnAnte {
				return newCtx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "ante handler failure")
			}

			return
		})
	}

	routerOpt := func(bapp *BaseApp) {
		r := sdk.NewRoute(routeMsgCounter, func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
			count := msg.(*msgCounter).Counter
			ctx.GasMeter().ConsumeGas(uint64(count), "counter-handler")
			return &sdk.Result{}, nil
		})
		bapp.Router().AddRoute(r)
	}

	cdc := codec.NewLegacyAmino()
	registerTestCodec(cdc)

	app := setupBaseApp(t, anteOpt, routerOpt)
	app.InitChain(abci.RequestInitChain{
		ConsensusParams: &abci.ConsensusParams{
			Block: &abci.BlockParams{
				MaxGas: 9,
			},
		},
	})

	app.InitChain(abci.RequestInitChain{})

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	tx := newTxCounter(5, 0)
	tx.setFailOnAnte(true)
	txBytes, err := cdc.Marshal(tx)
	require.NoError(t, err)

	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.False(t, res.IsOK(), fmt.Sprintf("%v", res))

	// require next tx to fail due to black gas limit
	tx = newTxCounter(5, 0)
	txBytes, err = cdc.Marshal(tx)
	require.NoError(t, err)

	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.False(t, res.IsOK(), fmt.Sprintf("%v", res))
}

// Test that we can only query from the latest committed state.
func TestQuery(t *testing.T) {
	key, value := []byte("hello"), []byte("goodbye")
	anteOpt := func(bapp *BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			store := ctx.KVStore(capKey1)
			store.Set(key, value)
			return
		})
	}

	routerOpt := func(bapp *BaseApp) {
		r := sdk.NewRoute(routeMsgCounter, func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
			store := ctx.KVStore(capKey1)
			store.Set(key, value)
			return &sdk.Result{}, nil
		})
		bapp.Router().AddRoute(r)
	}

	app := setupBaseApp(t, anteOpt, routerOpt)

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
	_, resTx, err := app.SimCheck(aminoTxEncoder(), tx)
	require.NoError(t, err)
	require.NotNil(t, resTx)
	res = app.Query(query)
	require.Equal(t, 0, len(res.Value))

	// query is still empty after a DeliverTx before we commit
	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	_, resTx, err = app.SimDeliver(aminoTxEncoder(), tx)
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
	grpcQueryOpt := func(bapp *BaseApp) {
		testdata.RegisterQueryServer(
			bapp.GRPCQueryRouter(),
			testdata.QueryImpl{},
		)
	}

	app := setupBaseApp(t, grpcQueryOpt)
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

// Test p2p filter queries
func TestP2PQuery(t *testing.T) {
	addrPeerFilterOpt := func(bapp *BaseApp) {
		bapp.SetAddrPeerFilter(func(addrport string) abci.ResponseQuery {
			require.Equal(t, "1.1.1.1:8000", addrport)
			return abci.ResponseQuery{Code: uint32(3)}
		})
	}

	idPeerFilterOpt := func(bapp *BaseApp) {
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

func TestGetMaximumBlockGas(t *testing.T) {
	app := setupBaseApp(t)
	app.InitChain(abci.RequestInitChain{})
	ctx := app.NewContext(true, tmproto.Header{})

	app.StoreConsensusParams(ctx, &abci.ConsensusParams{Block: &abci.BlockParams{MaxGas: 0}})
	require.Equal(t, uint64(0), app.getMaximumBlockGas(ctx))

	app.StoreConsensusParams(ctx, &abci.ConsensusParams{Block: &abci.BlockParams{MaxGas: -1}})
	require.Equal(t, uint64(0), app.getMaximumBlockGas(ctx))

	app.StoreConsensusParams(ctx, &abci.ConsensusParams{Block: &abci.BlockParams{MaxGas: 5000000}})
	require.Equal(t, uint64(5000000), app.getMaximumBlockGas(ctx))

	app.StoreConsensusParams(ctx, &abci.ConsensusParams{Block: &abci.BlockParams{MaxGas: -5000000}})
	require.Panics(t, func() { app.getMaximumBlockGas(ctx) })
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
		snapshotKeepRecent: 2,
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
	// test increments in the ante
	anteKey := []byte("ante-key")
	anteOpt := func(bapp *BaseApp) { bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, anteKey)) }

	// test increments in the handler
	deliverKey := []byte("deliver-key")
	routerOpt := func(bapp *BaseApp) {
		bapp.SetRouter(&testCustomRouter{routes: sync.Map{}})
		r := sdk.NewRoute(routeMsgCounter, handlerMsgCounter(t, capKey1, deliverKey))
		bapp.Router().AddRoute(r)
	}

	app := setupBaseApp(t, anteOpt, routerOpt)
	app.InitChain(abci.RequestInitChain{})

	// Create same codec used in txDecoder
	codec := codec.NewLegacyAmino()
	registerTestCodec(codec)

	nBlocks := 3
	txPerHeight := 5

	for blockN := 0; blockN < nBlocks; blockN++ {
		header := tmproto.Header{Height: int64(blockN) + 1}
		app.BeginBlock(abci.RequestBeginBlock{Header: header})

		for i := 0; i < txPerHeight; i++ {
			counter := int64(blockN*txPerHeight + i)
			tx := newTxCounter(counter, counter)

			txBytes, err := codec.Marshal(tx)
			require.NoError(t, err)

			res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
			require.True(t, res.IsOK(), fmt.Sprintf("%v", res))
		}

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}
}

func TestBaseApp_EndBlock(t *testing.T) {
	db := dbm.NewMemDB()
	name := t.Name()
	logger := defaultLogger()

	cp := &abci.ConsensusParams{
		Block: &abci.BlockParams{
			MaxGas: 5000000,
		},
	}

	app := NewBaseApp(name, logger, db, nil)
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
