package baseapp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// make some cap keys
	capKey1 = sdk.NewKVStoreKey("key1")
	capKey2 = sdk.NewKVStoreKey("key2")
)

//------------------------------------------------------------------------------------------
// Helpers for setup. Most tests should be able to use setupBaseApp

func defaultLogger() log.Logger {
	return log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
}

func newBaseApp(name string, options ...func(*BaseApp)) *BaseApp {
	logger := defaultLogger()
	db := dbm.NewMemDB()
	codec := codec.New()
	registerTestCodec(codec)
	return NewBaseApp(name, logger, db, testTxDecoder(codec), options...)
}

func registerTestCodec(cdc *codec.Codec) {
	// register Tx, Msg
	sdk.RegisterCodec(cdc)

	// register test types
	cdc.RegisterConcrete(&txTest{}, "cosmos-sdk/baseapp/txTest", nil)
	cdc.RegisterConcrete(&msgCounter{}, "cosmos-sdk/baseapp/msgCounter", nil)
	cdc.RegisterConcrete(&msgCounter2{}, "cosmos-sdk/baseapp/msgCounter2", nil)
	cdc.RegisterConcrete(&msgNoRoute{}, "cosmos-sdk/baseapp/msgNoRoute", nil)
}

// simple one store baseapp
func setupBaseApp(t *testing.T, options ...func(*BaseApp)) *BaseApp {
	app := newBaseApp(t.Name(), options...)
	require.Equal(t, t.Name(), app.Name())

	// no stores are mounted
	require.Panics(t, func() { app.LoadLatestVersion(capKey1) })

	app.MountStoresIAVL(capKey1, capKey2)

	// stores are mounted
	err := app.LoadLatestVersion(capKey1)
	require.Nil(t, err)
	return app
}

//------------------------------------------------------------------------------------------
// test mounting and loading stores

func TestMountStores(t *testing.T) {
	app := setupBaseApp(t)

	// check both stores
	store1 := app.cms.GetCommitKVStore(capKey1)
	require.NotNil(t, store1)
	store2 := app.cms.GetCommitKVStore(capKey2)
	require.NotNil(t, store2)
}

// Test that we can make commits and then reload old versions.
// Test that LoadLatestVersion actually does.
func TestLoadVersion(t *testing.T) {
	logger := defaultLogger()
	db := dbm.NewMemDB()
	name := t.Name()
	app := NewBaseApp(name, logger, db, nil)

	// make a cap key and mount the store
	capKey := sdk.NewKVStoreKey("main")
	app.MountStoresIAVL(capKey)
	err := app.LoadLatestVersion(capKey) // needed to make stores non-nil
	require.Nil(t, err)

	emptyCommitID := sdk.CommitID{}

	// fresh store has zero/empty last commit
	lastHeight := app.LastBlockHeight()
	lastID := app.LastCommitID()
	require.Equal(t, int64(0), lastHeight)
	require.Equal(t, emptyCommitID, lastID)

	// execute a block, collect commit ID
	header := abci.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	res := app.Commit()
	commitID1 := sdk.CommitID{1, res.Data}

	// execute a block, collect commit ID
	header = abci.Header{Height: 2}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	res = app.Commit()
	commitID2 := sdk.CommitID{2, res.Data}

	// reload with LoadLatestVersion
	app = NewBaseApp(name, logger, db, nil)
	app.MountStoresIAVL(capKey)
	err = app.LoadLatestVersion(capKey)
	require.Nil(t, err)
	testLoadVersionHelper(t, app, int64(2), commitID2)

	// reload with LoadVersion, see if you can commit the same block and get
	// the same result
	app = NewBaseApp(name, logger, db, nil)
	app.MountStoresIAVL(capKey)
	err = app.LoadVersion(1, capKey)
	require.Nil(t, err)
	testLoadVersionHelper(t, app, int64(1), commitID1)
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	app.Commit()
	testLoadVersionHelper(t, app, int64(2), commitID2)
}

func testLoadVersionHelper(t *testing.T, app *BaseApp, expectedHeight int64, expectedID sdk.CommitID) {
	lastHeight := app.LastBlockHeight()
	lastID := app.LastCommitID()
	require.Equal(t, expectedHeight, lastHeight)
	require.Equal(t, expectedID, lastID)
}

func TestOptionFunction(t *testing.T) {
	logger := defaultLogger()
	db := dbm.NewMemDB()
	bap := NewBaseApp("starting name", logger, db, nil, testChangeNameHelper("new name"))
	require.Equal(t, bap.name, "new name", "BaseApp should have had name changed via option function")
}

func testChangeNameHelper(name string) func(*BaseApp) {
	return func(bap *BaseApp) {
		bap.name = name
	}
}

// Test that the app hash is static
// TODO: https://github.com/cosmos/cosmos-sdk/issues/520
/*func TestStaticAppHash(t *testing.T) {
	app := newBaseApp(t.Name())

	// make a cap key and mount the store
	capKey := sdk.NewKVStoreKey("main")
	app.MountStoresIAVL(capKey)
	err := app.LoadLatestVersion(capKey) // needed to make stores non-nil
	require.Nil(t, err)

	// execute some blocks
	header := abci.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	res := app.Commit()
	commitID1 := sdk.CommitID{1, res.Data}

	header = abci.Header{Height: 2}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	res = app.Commit()
	commitID2 := sdk.CommitID{2, res.Data}

	require.Equal(t, commitID1.Hash, commitID2.Hash)
}
*/

//------------------------------------------------------------------------------------------
// test some basic abci/baseapp functionality

// Test that txs can be unmarshalled and read and that
// correct error codes are returned when not
func TestTxDecoder(t *testing.T) {
	// TODO
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

	// ----- test a proper response -------
	// TODO
}

//------------------------------------------------------------------------------------------
// InitChain, BeginBlock, EndBlock

func TestInitChainer(t *testing.T) {
	name := t.Name()
	// keep the db and logger ourselves so
	// we can reload the same  app later
	db := dbm.NewMemDB()
	logger := defaultLogger()
	app := NewBaseApp(name, logger, db, nil)
	capKey := sdk.NewKVStoreKey("main")
	capKey2 := sdk.NewKVStoreKey("key2")
	app.MountStoresIAVL(capKey, capKey2)

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
	err := app.LoadLatestVersion(capKey) // needed to make stores non-nil
	require.Nil(t, err)

	app.InitChain(abci.RequestInitChain{AppStateBytes: []byte("{}"), ChainId: "test-chain-id"}) // must have valid JSON genesis file, even if empty

	// assert that chainID is set correctly in InitChain
	chainID := app.deliverState.ctx.ChainID()
	require.Equal(t, "test-chain-id", chainID, "ChainID in deliverState not set correctly in InitChain")

	chainID = app.checkState.ctx.ChainID()
	require.Equal(t, "test-chain-id", chainID, "ChainID in checkState not set correctly in InitChain")

	app.Commit()
	res = app.Query(query)
	require.Equal(t, value, res.Value)

	// reload app
	app = NewBaseApp(name, logger, db, nil)
	app.SetInitChainer(initChainer)
	app.MountStoresIAVL(capKey, capKey2)
	err = app.LoadLatestVersion(capKey) // needed to make stores non-nil
	require.Nil(t, err)

	// ensure we can still query after reloading
	res = app.Query(query)
	require.Equal(t, value, res.Value)

	// commit and ensure we can still query
	app.BeginBlock(abci.RequestBeginBlock{})
	app.Commit()
	res = app.Query(query)
	require.Equal(t, value, res.Value)
}

//------------------------------------------------------------------------------------------
// Mock tx, msgs, and mapper for the baseapp tests.
// Self-contained, just uses counters.
// We don't care about signatures, coins, accounts, etc. in the baseapp.

// Simple tx with a list of Msgs.
type txTest struct {
	Msgs    []sdk.Msg
	Counter int64
}

// Implements Tx
func (tx txTest) GetMsgs() []sdk.Msg { return tx.Msgs }

const (
	typeMsgCounter  = "msgCounter"
	typeMsgCounter2 = "msgCounter2"
)

// ValidateBasic() fails on negative counters.
// Otherwise it's up to the handlers
type msgCounter struct {
	Counter int64
}

// Implements Msg
func (msg msgCounter) Type() string                 { return typeMsgCounter }
func (msg msgCounter) Name() string                 { return "counter1" }
func (msg msgCounter) GetSignBytes() []byte         { return nil }
func (msg msgCounter) GetSigners() []sdk.AccAddress { return nil }
func (msg msgCounter) ValidateBasic() sdk.Error {
	if msg.Counter >= 0 {
		return nil
	}
	return sdk.ErrInvalidSequence("counter should be a non-negative integer.")
}

func newTxCounter(txInt int64, msgInts ...int64) *txTest {
	var msgs []sdk.Msg
	for _, msgInt := range msgInts {
		msgs = append(msgs, msgCounter{msgInt})
	}
	return &txTest{msgs, txInt}
}

// a msg we dont know how to route
type msgNoRoute struct {
	msgCounter
}

func (tx msgNoRoute) Type() string { return "noroute" }

// a msg we dont know how to decode
type msgNoDecode struct {
	msgCounter
}

func (tx msgNoDecode) Type() string { return typeMsgCounter }

// Another counter msg. Duplicate of msgCounter
type msgCounter2 struct {
	Counter int64
}

// Implements Msg
func (msg msgCounter2) Type() string                 { return typeMsgCounter2 }
func (msg msgCounter2) Name() string                 { return "counter2" }
func (msg msgCounter2) GetSignBytes() []byte         { return nil }
func (msg msgCounter2) GetSigners() []sdk.AccAddress { return nil }
func (msg msgCounter2) ValidateBasic() sdk.Error {
	if msg.Counter >= 0 {
		return nil
	}
	return sdk.ErrInvalidSequence("counter should be a non-negative integer.")
}

// amino decode
func testTxDecoder(cdc *codec.Codec) sdk.TxDecoder {
	return func(txBytes []byte) (sdk.Tx, sdk.Error) {
		var tx txTest
		if len(txBytes) == 0 {
			return nil, sdk.ErrTxDecode("txBytes are empty")
		}
		err := cdc.UnmarshalBinary(txBytes, &tx)
		if err != nil {
			return nil, sdk.ErrTxDecode("").TraceSDK(err.Error())
		}
		return tx, nil
	}
}

func anteHandlerTxTest(t *testing.T, capKey *sdk.KVStoreKey, storeKey []byte) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, res sdk.Result, abort bool) {
		store := ctx.KVStore(capKey)
		msgCounter := tx.(txTest).Counter
		res = incrementingCounter(t, store, storeKey, msgCounter)
		return
	}
}

func handlerMsgCounter(t *testing.T, capKey *sdk.KVStoreKey, deliverKey []byte) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		store := ctx.KVStore(capKey)
		var msgCount int64
		switch m := msg.(type) {
		case *msgCounter:
			msgCount = m.Counter
		case *msgCounter2:
			msgCount = m.Counter
		}
		return incrementingCounter(t, store, deliverKey, msgCount)
	}
}

//-----------------------------------------------------------------
// simple int mapper

func i2b(i int64) []byte {
	return []byte{byte(i)}
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
func incrementingCounter(t *testing.T, store sdk.KVStore, counterKey []byte, counter int64) (res sdk.Result) {
	storedCounter := getIntFromStore(store, counterKey)
	require.Equal(t, storedCounter, counter)
	setIntOnStore(store, counterKey, counter+1)
	return
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
		bapp.Router().AddRoute(typeMsgCounter, func(ctx sdk.Context, msg sdk.Msg) sdk.Result { return sdk.Result{} })
	}

	app := setupBaseApp(t, anteOpt, routerOpt)

	nTxs := int64(5)

	app.InitChain(abci.RequestInitChain{})

	// Create same codec used in txDecoder
	codec := codec.New()
	registerTestCodec(codec)

	for i := int64(0); i < nTxs; i++ {
		tx := newTxCounter(i, 0)
		txBytes, err := codec.MarshalBinary(tx)
		require.NoError(t, err)
		r := app.CheckTx(txBytes)
		assert.True(t, r.IsOK(), fmt.Sprintf("%v", r))
	}

	checkStateStore := app.checkState.ctx.KVStore(capKey1)
	storedCounter := getIntFromStore(checkStateStore, counterKey)

	// Ensure AnteHandler ran
	require.Equal(t, nTxs, storedCounter)

	// If a block is committed, CheckTx state should be reset.
	app.BeginBlock(abci.RequestBeginBlock{})
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
	routerOpt := func(bapp *BaseApp) { bapp.Router().AddRoute(typeMsgCounter, handlerMsgCounter(t, capKey1, deliverKey)) }

	app := setupBaseApp(t, anteOpt, routerOpt)

	// Create same codec used in txDecoder
	codec := codec.New()
	registerTestCodec(codec)

	nBlocks := 3
	txPerHeight := 5
	for blockN := 0; blockN < nBlocks; blockN++ {
		app.BeginBlock(abci.RequestBeginBlock{})
		for i := 0; i < txPerHeight; i++ {
			counter := int64(blockN*txPerHeight + i)
			tx := newTxCounter(counter, counter)
			txBytes, err := codec.MarshalBinary(tx)
			require.NoError(t, err)
			res := app.DeliverTx(txBytes)
			require.True(t, res.IsOK(), fmt.Sprintf("%v", res))
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
		bapp.Router().AddRoute(typeMsgCounter, handlerMsgCounter(t, capKey1, deliverKey))
		bapp.Router().AddRoute(typeMsgCounter2, handlerMsgCounter(t, capKey1, deliverKey2))
	}

	app := setupBaseApp(t, anteOpt, routerOpt)

	// Create same codec used in txDecoder
	codec := codec.New()
	registerTestCodec(codec)

	// run a multi-msg tx
	// with all msgs the same type
	{
		app.BeginBlock(abci.RequestBeginBlock{})
		tx := newTxCounter(0, 0, 1, 2)
		txBytes, err := codec.MarshalBinary(tx)
		require.NoError(t, err)
		res := app.DeliverTx(txBytes)
		require.True(t, res.IsOK(), fmt.Sprintf("%v", res))

		store := app.deliverState.ctx.KVStore(capKey1)

		// tx counter only incremented once
		txCounter := getIntFromStore(store, anteKey)
		require.Equal(t, int64(1), txCounter)

		// msg counter incremented three times
		msgCounter := getIntFromStore(store, deliverKey)
		require.Equal(t, int64(3), msgCounter)
	}

	// replace the second message with a msgCounter2
	{
		tx := newTxCounter(1, 3)
		tx.Msgs = append(tx.Msgs, msgCounter2{0})
		tx.Msgs = append(tx.Msgs, msgCounter2{1})
		txBytes, err := codec.MarshalBinary(tx)
		require.NoError(t, err)
		res := app.DeliverTx(txBytes)
		require.True(t, res.IsOK(), fmt.Sprintf("%v", res))

		store := app.deliverState.ctx.KVStore(capKey1)

		// tx counter only incremented once
		txCounter := getIntFromStore(store, anteKey)
		require.Equal(t, int64(2), txCounter)

		// original counter increments by one
		// new counter increments by two
		msgCounter := getIntFromStore(store, deliverKey)
		require.Equal(t, int64(4), msgCounter)
		msgCounter2 := getIntFromStore(store, deliverKey2)
		require.Equal(t, int64(2), msgCounter2)
	}
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
	gasConsumed := int64(5)

	anteOpt := func(bapp *BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, res sdk.Result, abort bool) {
			newCtx = ctx.WithGasMeter(sdk.NewGasMeter(gasConsumed))
			return
		})
	}

	routerOpt := func(bapp *BaseApp) {
		bapp.Router().AddRoute(typeMsgCounter, func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
			ctx.GasMeter().ConsumeGas(gasConsumed, "test")
			return sdk.Result{GasUsed: ctx.GasMeter().GasConsumed()}
		})
	}

	app := setupBaseApp(t, anteOpt, routerOpt)

	app.InitChain(abci.RequestInitChain{})

	// Create same codec used in txDecoder
	cdc := codec.New()
	registerTestCodec(cdc)

	nBlocks := 3
	for blockN := 0; blockN < nBlocks; blockN++ {
		count := int64(blockN + 1)
		app.BeginBlock(abci.RequestBeginBlock{})

		tx := newTxCounter(count, count)

		// simulate a message, check gas reported
		result := app.Simulate(tx)
		require.True(t, result.IsOK(), result.Log)
		require.Equal(t, gasConsumed, result.GasUsed)

		// simulate again, same result
		result = app.Simulate(tx)
		require.True(t, result.IsOK(), result.Log)
		require.Equal(t, gasConsumed, result.GasUsed)

		// simulate by calling Query with encoded tx
		txBytes, err := cdc.MarshalBinary(tx)
		require.Nil(t, err)
		query := abci.RequestQuery{
			Path: "/app/simulate",
			Data: txBytes,
		}
		queryResult := app.Query(query)
		require.True(t, queryResult.IsOK(), queryResult.Log)

		var res sdk.Result
		codec.Cdc.MustUnmarshalBinary(queryResult.Value, &res)
		require.Nil(t, err, "Result unmarshalling failed")
		require.True(t, res.IsOK(), res.Log)
		require.Equal(t, gasConsumed, res.GasUsed, res.Log)
		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}
}

//-------------------------------------------------------------------------------------------
// Tx failure cases
// TODO: add more

func TestRunInvalidTransaction(t *testing.T) {
	anteOpt := func(bapp *BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, res sdk.Result, abort bool) {
			return
		})
	}
	routerOpt := func(bapp *BaseApp) {
		bapp.Router().AddRoute(typeMsgCounter, func(ctx sdk.Context, msg sdk.Msg) (res sdk.Result) { return })
	}

	app := setupBaseApp(t, anteOpt, routerOpt)
	app.BeginBlock(abci.RequestBeginBlock{})

	// Transaction with no messages
	{
		emptyTx := &txTest{}
		err := app.Deliver(emptyTx)
		require.Equal(t, sdk.ToABCICode(sdk.CodespaceRoot, sdk.CodeInternal), err.Code)
	}

	// Transaction where ValidateBasic fails
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
			res := app.Deliver(tx)
			if testCase.fail {
				require.Equal(t, sdk.ToABCICode(sdk.CodespaceRoot, sdk.CodeInvalidSequence), res.Code)
			} else {
				require.True(t, res.IsOK(), fmt.Sprintf("%v", res))
			}
		}
	}

	// Transaction with no known route
	{
		unknownRouteTx := txTest{[]sdk.Msg{msgNoRoute{}}, 0}
		err := app.Deliver(unknownRouteTx)
		require.Equal(t, sdk.ToABCICode(sdk.CodespaceRoot, sdk.CodeUnknownRequest), err.Code)

		unknownRouteTx = txTest{[]sdk.Msg{msgCounter{}, msgNoRoute{}}, 0}
		err = app.Deliver(unknownRouteTx)
		require.Equal(t, sdk.ToABCICode(sdk.CodespaceRoot, sdk.CodeUnknownRequest), err.Code)
	}

	// Transaction with an unregistered message
	{
		tx := newTxCounter(0, 0)
		tx.Msgs = append(tx.Msgs, msgNoDecode{})

		// new codec so we can encode the tx, but we shouldn't be able to decode
		newCdc := codec.New()
		registerTestCodec(newCdc)
		newCdc.RegisterConcrete(&msgNoDecode{}, "cosmos-sdk/baseapp/msgNoDecode", nil)

		txBytes, err := newCdc.MarshalBinary(tx)
		require.NoError(t, err)
		res := app.DeliverTx(txBytes)
		require.EqualValues(t, sdk.ToABCICode(sdk.CodespaceRoot, sdk.CodeTxDecode), res.Code)
	}
}

// Test that transactions exceeding gas limits fail
func TestTxGasLimits(t *testing.T) {
	gasGranted := int64(10)
	anteOpt := func(bapp *BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, res sdk.Result, abort bool) {
			newCtx = ctx.WithGasMeter(sdk.NewGasMeter(gasGranted))

			// NOTE/TODO/XXX:
			// AnteHandlers must have their own defer/recover in order
			// for the BaseApp to know how much gas was used used!
			// This is because the GasMeter is created in the AnteHandler,
			// but if it panics the context won't be set properly in runTx's recover ...
			defer func() {
				if r := recover(); r != nil {
					switch rType := r.(type) {
					case sdk.ErrorOutOfGas:
						log := fmt.Sprintf("out of gas in location: %v", rType.Descriptor)
						res = sdk.ErrOutOfGas(log).Result()
						res.GasWanted = gasGranted
						res.GasUsed = newCtx.GasMeter().GasConsumed()
					default:
						panic(r)
					}
				}
			}()

			count := tx.(*txTest).Counter
			newCtx.GasMeter().ConsumeGas(count, "counter-ante")
			res = sdk.Result{
				GasWanted: gasGranted,
			}
			return
		})

	}

	routerOpt := func(bapp *BaseApp) {
		bapp.Router().AddRoute(typeMsgCounter, func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
			count := msg.(msgCounter).Counter
			ctx.GasMeter().ConsumeGas(count, "counter-handler")
			return sdk.Result{}
		})
	}

	app := setupBaseApp(t, anteOpt, routerOpt)

	app.BeginBlock(abci.RequestBeginBlock{})

	testCases := []struct {
		tx      *txTest
		gasUsed int64
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
		res := app.Deliver(tx)

		// check gas used and wanted
		require.Equal(t, tc.gasUsed, res.GasUsed, fmt.Sprintf("%d: %v, %v", i, tc, res))

		// check for out of gas
		if !tc.fail {
			require.True(t, res.IsOK(), fmt.Sprintf("%d: %v, %v", i, tc, res))
		} else {
			require.Equal(t, res.Code, sdk.ToABCICode(sdk.CodespaceRoot, sdk.CodeOutOfGas), fmt.Sprintf("%d: %v, %v", i, tc, res))
		}
	}
}

//-------------------------------------------------------------------------------------------
// Queries

// Test that we can only query from the latest committed state.
func TestQuery(t *testing.T) {
	key, value := []byte("hello"), []byte("goodbye")
	anteOpt := func(bapp *BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, res sdk.Result, abort bool) {
			store := ctx.KVStore(capKey1)
			store.Set(key, value)
			return
		})
	}

	routerOpt := func(bapp *BaseApp) {
		bapp.Router().AddRoute(typeMsgCounter, func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
			store := ctx.KVStore(capKey1)
			store.Set(key, value)
			return sdk.Result{}
		})
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
	resTx := app.Check(tx)
	require.True(t, resTx.IsOK(), fmt.Sprintf("%v", resTx))
	res = app.Query(query)
	require.Equal(t, 0, len(res.Value))

	// query is still empty after a DeliverTx before we commit
	app.BeginBlock(abci.RequestBeginBlock{})
	resTx = app.Deliver(tx)
	require.True(t, resTx.IsOK(), fmt.Sprintf("%v", resTx))
	res = app.Query(query)
	require.Equal(t, 0, len(res.Value))

	// query returns correct value after Commit
	app.Commit()
	res = app.Query(query)
	require.Equal(t, value, res.Value)
}

// Test p2p filter queries
func TestP2PQuery(t *testing.T) {
	addrPeerFilterOpt := func(bapp *BaseApp) {
		bapp.SetAddrPeerFilter(func(addrport string) abci.ResponseQuery {
			require.Equal(t, "1.1.1.1:8000", addrport)
			return abci.ResponseQuery{Code: uint32(3)}
		})
	}

	pubkeyPeerFilterOpt := func(bapp *BaseApp) {
		bapp.SetPubKeyPeerFilter(func(pubkey string) abci.ResponseQuery {
			require.Equal(t, "testpubkey", pubkey)
			return abci.ResponseQuery{Code: uint32(4)}
		})
	}

	app := setupBaseApp(t, addrPeerFilterOpt, pubkeyPeerFilterOpt)

	addrQuery := abci.RequestQuery{
		Path: "/p2p/filter/addr/1.1.1.1:8000",
	}
	res := app.Query(addrQuery)
	require.Equal(t, uint32(3), res.Code)

	pubkeyQuery := abci.RequestQuery{
		Path: "/p2p/filter/pubkey/testpubkey",
	}
	res = app.Query(pubkeyQuery)
	require.Equal(t, uint32(4), res.Code)
}
