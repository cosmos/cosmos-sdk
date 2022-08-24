package baseapp_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"cosmossdk.io/depinject"
	"github.com/cosmos/cosmos-sdk/baseapp"
	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

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

	checkStateStore := getCheckStateCtx(app).KVStore(capKey1)
	storedCounter := getIntFromStore(checkStateStore, counterKey)

	// Ensure AnteHandler ran
	require.Equal(t, nTxs, storedCounter)

	// If a block is committed, CheckTx state should be reset.
	header := tmproto.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header, Hash: []byte("hash")})

	require.NotNil(t, getCheckStateCtx(app).BlockGasMeter(), "block gas meter should have been set to checkState")
	require.NotEmpty(t, getCheckStateCtx(app).HeaderHash())

	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	checkStateStore = getCheckStateCtx(app).KVStore(capKey1)
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

	store := getDeliverStateCtx(app).KVStore(capKey1)

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

	store = getDeliverStateCtx(app).KVStore(capKey1)

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
		ConsensusParams: &abci.ConsensusParams{
			Block: &abci.BlockParams{
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

			ctx := getDeliverStateCtx(app)

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

	ctx := getDeliverStateCtx(app)
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

	ctx = getDeliverStateCtx(app)
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

	ctx = getDeliverStateCtx(app)
	store = ctx.KVStore(capKey1)
	require.Equal(t, int64(2), getIntFromStore(store, anteKey))
	require.Equal(t, int64(1), getIntFromStore(store, deliverKey))

	// commit
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()
}

func getCheckStateCtx(app *runtime.App) sdk.Context {
	v := reflect.ValueOf(app.BaseApp).Elem()
	f := v.FieldByName("checkState")
	rf := reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
	return rf.MethodByName("Context").Call(nil)[0].Interface().(sdk.Context)
}

func getDeliverStateCtx(app *runtime.App) sdk.Context {
	v := reflect.ValueOf(app.BaseApp).Elem()
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

var (
	capKey1 = sdk.NewKVStoreKey("key1")
	capKey2 = sdk.NewKVStoreKey("key2")

	// testTxPriority is the CheckTx priority that we set in the test
	// antehandler.
	testTxPriority = int64(42)
)

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

type CounterServerImpl struct {
	t          *testing.T
	capKey     storetypes.StoreKey
	deliverKey []byte
}

type Counter2ServerImpl struct {
	t          *testing.T
	capKey     storetypes.StoreKey
	deliverKey []byte
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

func (m CounterServerImpl) IncrementCounter(ctx context.Context, msg *baseapptestutil.MsgCounter) (*baseapptestutil.MsgCreateCounterResponse, error) {
	return incrementCounter(ctx, m.t, m.capKey, m.deliverKey, msg)
}

func (m Counter2ServerImpl) IncrementCounter(ctx context.Context, msg *baseapptestutil.MsgCounter2) (*baseapptestutil.MsgCreateCounterResponse, error) {
	return incrementCounter(ctx, m.t, m.capKey, m.deliverKey, msg)
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

func (ps *paramStore) Set(_ sdk.Context, key []byte, value interface{}) {
	bz, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}

	ps.db.Set(key, bz)
}

func (ps *paramStore) Has(_ sdk.Context, key []byte) bool {
	ok, err := ps.db.Has(key)
	if err != nil {
		panic(err)
	}

	return ok
}

func (ps *paramStore) Get(_ sdk.Context, key []byte, ptr interface{}) {
	bz, err := ps.db.Get(key)
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
