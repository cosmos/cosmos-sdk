package baseapp_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/gogoproto/jsonpb"
	"github.com/stretchr/testify/require"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	pruningtypes "cosmossdk.io/store/pruning/types"
	"cosmossdk.io/store/snapshots"
	snapshottypes "cosmossdk.io/store/snapshots/types"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

func TestABCI_Info(t *testing.T) {
	suite := NewBaseAppSuite(t)

	reqInfo := abci.RequestInfo{}
	res, err := suite.baseApp.Info(&reqInfo)
	require.NoError(t, err)

	require.Equal(t, "", res.Version)
	require.Equal(t, t.Name(), res.GetData())
	require.Equal(t, int64(0), res.LastBlockHeight)
	require.Equal(t, []uint8(nil), res.LastBlockAppHash)
	require.Equal(t, suite.baseApp.AppVersion(), res.AppVersion)
}

func TestABCI_First_block_Height(t *testing.T) {
	suite := NewBaseAppSuite(t, baseapp.SetChainID("test-chain-id"))
	app := suite.baseApp

	app.InitChain(&abci.RequestInitChain{
		ChainId:         "test-chain-id",
		ConsensusParams: &cmtproto.ConsensusParams{Block: &cmtproto.BlockParams{MaxGas: 5000000}},
		InitialHeight:   1,
	})
	app.Commit()

	ctx := app.GetContextForCheckTx(nil)
	require.Equal(t, int64(1), ctx.BlockHeight())
}

func TestABCI_InitChain(t *testing.T) {
	name := t.Name()
	db := dbm.NewMemDB()
	logger := log.NewTestLogger(t)
	app := baseapp.NewBaseApp(name, logger, db, nil, baseapp.SetChainID("test-chain-id"))

	capKey := storetypes.NewKVStoreKey("main")
	capKey2 := storetypes.NewKVStoreKey("key2")
	app.MountStores(capKey, capKey2)

	// set a value in the store on init chain
	key, value := []byte("hello"), []byte("goodbye")
	var initChainer sdk.InitChainer = func(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
		store := ctx.KVStore(capKey)
		store.Set(key, value)
		return &abci.ResponseInitChain{}, nil
	}

	query := abci.RequestQuery{
		Path: "/store/main/key",
		Data: key,
	}

	_, err := app.InitChain(&abci.RequestInitChain{ChainId: "wrong-chain-id"})
	// initChain is nil and chain ID is wrong - panics
	require.Error(t, err)

	// initChain is nil - nothing happens
	_, err = app.InitChain(&abci.RequestInitChain{ChainId: "test-chain-id"})
	require.NoError(t, err)
	resQ, err := app.Query(context.TODO(), &query)
	require.NoError(t, err)
	require.Equal(t, 0, len(resQ.Value))

	// set initChainer and try again - should see the value
	app.SetInitChainer(initChainer)

	// stores are mounted and private members are set - sealing baseapp
	err = app.LoadLatestVersion() // needed to make stores non-nil
	require.Nil(t, err)
	require.Equal(t, int64(0), app.LastBlockHeight())

	initChainRes, err := app.InitChain(&abci.RequestInitChain{AppStateBytes: []byte("{}"), ChainId: "test-chain-id"}) // must have valid JSON genesis file, even if empty
	require.NoError(t, err)

	// The AppHash returned by a new chain is the sha256 hash of "".
	// $ echo -n '' | sha256sum
	// e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
	require.Equal(
		t,
		[]byte{0xe3, 0xb0, 0xc4, 0x42, 0x98, 0xfc, 0x1c, 0x14, 0x9a, 0xfb, 0xf4, 0xc8, 0x99, 0x6f, 0xb9, 0x24, 0x27, 0xae, 0x41, 0xe4, 0x64, 0x9b, 0x93, 0x4c, 0xa4, 0x95, 0x99, 0x1b, 0x78, 0x52, 0xb8, 0x55},
		initChainRes.AppHash,
	)

	// assert that chainID is set correctly in InitChain
	chainID := getFinalizeBlockStateCtx(app).ChainID()
	require.Equal(t, "test-chain-id", chainID, "ChainID in deliverState not set correctly in InitChain")

	chainID = getCheckStateCtx(app).ChainID()
	require.Equal(t, "test-chain-id", chainID, "ChainID in checkState not set correctly in InitChain")

	app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Hash:   initChainRes.AppHash,
		Height: 1,
	})

	app.Commit()
	resQ, err = app.Query(context.TODO(), &query)
	require.NoError(t, err)
	require.Equal(t, int64(1), app.LastBlockHeight())
	require.Equal(t, value, resQ.Value)

	// reload app
	app = baseapp.NewBaseApp(name, logger, db, nil)
	app.SetInitChainer(initChainer)
	app.MountStores(capKey, capKey2)
	err = app.LoadLatestVersion() // needed to make stores non-nil
	require.Nil(t, err)
	require.Equal(t, int64(1), app.LastBlockHeight())

	// ensure we can still query after reloading
	resQ, err = app.Query(context.TODO(), &query)
	require.NoError(t, err)
	require.Equal(t, value, resQ.Value)

	// commit and ensure we can still query
	app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: app.LastBlockHeight() + 1})
	app.Commit()

	resQ, err = app.Query(context.TODO(), &query)
	require.NoError(t, err)
	require.Equal(t, value, resQ.Value)
}

func TestABCI_InitChain_WithInitialHeight(t *testing.T) {
	name := t.Name()
	db := dbm.NewMemDB()
	app := baseapp.NewBaseApp(name, log.NewTestLogger(t), db, nil)

	app.InitChain(
		&abci.RequestInitChain{
			InitialHeight: 3,
		},
	)
	app.Commit()

	require.Equal(t, int64(3), app.LastBlockHeight())
}

func TestABCI_FinalizeBlock_WithInitialHeight(t *testing.T) {
	name := t.Name()
	db := dbm.NewMemDB()
	app := baseapp.NewBaseApp(name, log.NewTestLogger(t), db, nil)

	app.InitChain(
		&abci.RequestInitChain{
			InitialHeight: 3,
		},
	)

	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 4})
	require.Error(t, err, "invalid height: 4; expected: 3")

	app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 3})
	app.Commit()

	require.Equal(t, int64(3), app.LastBlockHeight())
}

func TestABCI_GRPCQuery(t *testing.T) {
	grpcQueryOpt := func(bapp *baseapp.BaseApp) {
		testdata.RegisterQueryServer(
			bapp.GRPCQueryRouter(),
			testdata.QueryImpl{},
		)
	}

	suite := NewBaseAppSuite(t, grpcQueryOpt)

	suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})

	req := testdata.SayHelloRequest{Name: "foo"}
	reqBz, err := req.Marshal()
	require.NoError(t, err)

	resQuery, err := suite.baseApp.Query(context.TODO(), &abci.RequestQuery{
		Data: reqBz,
		Path: "/testpb.Query/SayHello",
	})
	require.NoError(t, err)
	require.Equal(t, sdkerrors.ErrInvalidHeight.ABCICode(), resQuery.Code, resQuery)
	require.Contains(t, resQuery.Log, "TestABCI_GRPCQuery is not ready; please wait for first block")

	suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: suite.baseApp.LastBlockHeight() + 1})
	suite.baseApp.Commit()

	reqQuery := abci.RequestQuery{
		Data: reqBz,
		Path: "/testpb.Query/SayHello",
	}

	resQuery, err = suite.baseApp.Query(context.TODO(), &reqQuery)
	require.NoError(t, err)
	require.Equal(t, abci.CodeTypeOK, resQuery.Code, resQuery)

	var res testdata.SayHelloResponse
	require.NoError(t, res.Unmarshal(resQuery.Value))
	require.Equal(t, "Hello foo!", res.Greeting)
}

func TestABCI_P2PQuery(t *testing.T) {
	addrPeerFilterOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAddrPeerFilter(func(addrport string) *abci.ResponseQuery {
			require.Equal(t, "1.1.1.1:8000", addrport)
			return &abci.ResponseQuery{Code: uint32(3)}
		})
	}

	idPeerFilterOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetIDPeerFilter(func(id string) *abci.ResponseQuery {
			require.Equal(t, "testid", id)
			return &abci.ResponseQuery{Code: uint32(4)}
		})
	}

	suite := NewBaseAppSuite(t, addrPeerFilterOpt, idPeerFilterOpt)

	addrQuery := abci.RequestQuery{
		Path: "/p2p/filter/addr/1.1.1.1:8000",
	}
	res, err := suite.baseApp.Query(context.TODO(), &addrQuery)
	require.NoError(t, err)
	require.Equal(t, uint32(3), res.Code)

	idQuery := abci.RequestQuery{
		Path: "/p2p/filter/id/testid",
	}
	res, err = suite.baseApp.Query(context.TODO(), &idQuery)
	require.NoError(t, err)
	require.Equal(t, uint32(4), res.Code)
}

func TestBaseApp_PrepareCheckState(t *testing.T) {
	db := dbm.NewMemDB()
	name := t.Name()
	logger := log.NewTestLogger(t)

	cp := &cmtproto.ConsensusParams{
		Block: &cmtproto.BlockParams{
			MaxGas: 5000000,
		},
	}

	app := baseapp.NewBaseApp(name, logger, db, nil)
	app.SetParamStore(&paramStore{db: dbm.NewMemDB()})
	app.InitChain(&abci.RequestInitChain{
		ConsensusParams: cp,
	})

	wasPrepareCheckStateCalled := false
	app.SetPrepareCheckStater(func(ctx sdk.Context) {
		wasPrepareCheckStateCalled = true
	})
	app.Seal()

	app.Commit()
	require.Equal(t, true, wasPrepareCheckStateCalled)
}

func TestBaseApp_Precommit(t *testing.T) {
	db := dbm.NewMemDB()
	name := t.Name()
	logger := log.NewTestLogger(t)

	cp := &cmtproto.ConsensusParams{
		Block: &cmtproto.BlockParams{
			MaxGas: 5000000,
		},
	}

	app := baseapp.NewBaseApp(name, logger, db, nil)
	app.SetParamStore(&paramStore{db: dbm.NewMemDB()})
	app.InitChain(&abci.RequestInitChain{
		ConsensusParams: cp,
	})

	wasPrecommiterCalled := false
	app.SetPrecommiter(func(ctx sdk.Context) {
		wasPrecommiterCalled = true
	})
	app.Seal()

	app.Commit()
	require.Equal(t, true, wasPrecommiterCalled)
}

func TestABCI_CheckTx(t *testing.T) {
	// This ante handler reads the key and checks that the value matches the
	// current counter. This ensures changes to the KVStore persist across
	// successive CheckTx runs.
	counterKey := []byte("counter-key")
	anteOpt := func(bapp *baseapp.BaseApp) { bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, counterKey)) }
	suite := NewBaseAppSuite(t, anteOpt)

	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), CounterServerImpl{t, capKey1, counterKey})

	nTxs := int64(5)
	suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})

	for i := int64(0); i < nTxs; i++ {
		tx := newTxCounter(t, suite.txConfig, i, 0) // no messages
		txBytes, err := suite.txConfig.TxEncoder()(tx)
		require.NoError(t, err)

		r, err := suite.baseApp.CheckTx(&abci.RequestCheckTx{Tx: txBytes})
		require.NoError(t, err)
		require.True(t, r.IsOK(), fmt.Sprintf("%v", r))
		require.Empty(t, r.GetEvents())
	}

	checkStateStore := getCheckStateCtx(suite.baseApp).KVStore(capKey1)
	storedCounter := getIntFromStore(t, checkStateStore, counterKey)

	// ensure AnteHandler ran
	require.Equal(t, nTxs, storedCounter)

	// if a block is committed, CheckTx state should be reset
	_, err := suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Hash:   []byte("hash"),
	})
	require.NoError(t, err)

	require.NotNil(t, getCheckStateCtx(suite.baseApp).BlockGasMeter(), "block gas meter should have been set to checkState")
	require.NotEmpty(t, getCheckStateCtx(suite.baseApp).HeaderHash())

	suite.baseApp.Commit()

	checkStateStore = getCheckStateCtx(suite.baseApp).KVStore(capKey1)
	storedBytes := checkStateStore.Get(counterKey)
	require.Nil(t, storedBytes)
}

func TestABCI_FinalizeBlock_DeliverTx(t *testing.T) {
	anteKey := []byte("ante-key")
	anteOpt := func(bapp *baseapp.BaseApp) { bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, anteKey)) }
	suite := NewBaseAppSuite(t, anteOpt)

	suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})

	deliverKey := []byte("deliver-key")
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), CounterServerImpl{t, capKey1, deliverKey})

	nBlocks := 3
	txPerHeight := 5

	for blockN := 0; blockN < nBlocks; blockN++ {

		txs := [][]byte{}
		for i := 0; i < txPerHeight; i++ {
			counter := int64(blockN*txPerHeight + i)
			tx := newTxCounter(t, suite.txConfig, counter, counter)

			txBytes, err := suite.txConfig.TxEncoder()(tx)
			require.NoError(t, err)

			txs = append(txs, txBytes)
		}

		res, err := suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: int64(blockN) + 1,
			Txs:    txs,
		})
		require.NoError(t, err)

		for i := 0; i < txPerHeight; i++ {
			counter := int64(blockN*txPerHeight + i)
			require.True(t, res.TxResults[i].IsOK(), fmt.Sprintf("%v", res))

			events := res.TxResults[i].GetEvents()
			require.Len(t, events, 3, "should contain ante handler, message type and counter events respectively")
			require.Equal(t, sdk.MarkEventsToIndex(counterEvent("ante_handler", counter).ToABCIEvents(), map[string]struct{}{})[0], events[0], "ante handler event")
			require.Equal(t, sdk.MarkEventsToIndex(counterEvent(sdk.EventTypeMessage, counter).ToABCIEvents(), map[string]struct{}{})[0].Attributes[0], events[2].Attributes[0], "msg handler update counter event")
		}

		suite.baseApp.Commit()
	}
}

func TestABCI_FinalizeBlock_MultiMsg(t *testing.T) {
	anteKey := []byte("ante-key")
	anteOpt := func(bapp *baseapp.BaseApp) { bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, anteKey)) }
	suite := NewBaseAppSuite(t, anteOpt)

	suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})

	deliverKey := []byte("deliver-key")
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), CounterServerImpl{t, capKey1, deliverKey})

	deliverKey2 := []byte("deliver-key2")
	baseapptestutil.RegisterCounter2Server(suite.baseApp.MsgServiceRouter(), Counter2ServerImpl{t, capKey1, deliverKey2})

	// run a multi-msg tx
	// with all msgs the same route
	tx := newTxCounter(t, suite.txConfig, 0, 0, 1, 2)
	txBytes, err := suite.txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Txs:    [][]byte{txBytes},
	})

	store := getFinalizeBlockStateCtx(suite.baseApp).KVStore(capKey1)

	// tx counter only incremented once
	txCounter := getIntFromStore(t, store, anteKey)
	require.Equal(t, int64(1), txCounter)

	// msg counter incremented three times
	msgCounter := getIntFromStore(t, store, deliverKey)
	require.Equal(t, int64(3), msgCounter)

	// replace the second message with a Counter2
	tx = newTxCounter(t, suite.txConfig, 1, 3)

	builder := suite.txConfig.NewTxBuilder()
	msgs := tx.GetMsgs()
	_, _, addr := testdata.KeyTestPubAddr()
	msgs = append(msgs, &baseapptestutil.MsgCounter2{Counter: 0, Signer: addr.String()})
	msgs = append(msgs, &baseapptestutil.MsgCounter2{Counter: 1, Signer: addr.String()})

	builder.SetMsgs(msgs...)
	builder.SetMemo(tx.GetMemo())
	setTxSignature(t, builder, 0)

	txBytes, err = suite.txConfig.TxEncoder()(builder.GetTx())
	require.NoError(t, err)

	_, err = suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Txs:    [][]byte{txBytes},
	})
	require.NoError(t, err)

	store = getFinalizeBlockStateCtx(suite.baseApp).KVStore(capKey1)

	// tx counter only incremented once
	txCounter = getIntFromStore(t, store, anteKey)
	require.Equal(t, int64(2), txCounter)

	// original counter increments by one
	// new counter increments by two
	msgCounter = getIntFromStore(t, store, deliverKey)
	require.Equal(t, int64(4), msgCounter)

	msgCounter2 := getIntFromStore(t, store, deliverKey2)
	require.Equal(t, int64(2), msgCounter2)
}

func TestABCI_Query_SimulateTx(t *testing.T) {
	gasConsumed := uint64(5)
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			newCtx = ctx.WithGasMeter(storetypes.NewGasMeter(gasConsumed))
			return
		})
	}
	suite := NewBaseAppSuite(t, anteOpt)

	suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})

	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), CounterServerImplGasMeterOnly{gasConsumed})

	nBlocks := 3
	for blockN := 0; blockN < nBlocks; blockN++ {
		count := int64(blockN + 1)

		tx := newTxCounter(t, suite.txConfig, count, count)

		txBytes, err := suite.txConfig.TxEncoder()(tx)
		require.Nil(t, err)

		// simulate a message, check gas reported
		gInfo, result, err := suite.baseApp.Simulate(txBytes)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, gasConsumed, gInfo.GasUsed)

		// simulate again, same result
		gInfo, result, err = suite.baseApp.Simulate(txBytes)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, gasConsumed, gInfo.GasUsed)

		// simulate by calling Query with encoded tx
		query := abci.RequestQuery{
			Path: "/app/simulate",
			Data: txBytes,
		}
		queryResult, err := suite.baseApp.Query(context.TODO(), &query)
		require.NoError(t, err)
		require.True(t, queryResult.IsOK(), queryResult.Log)

		var simRes sdk.SimulationResponse
		require.NoError(t, jsonpb.Unmarshal(strings.NewReader(string(queryResult.Value)), &simRes))

		require.Equal(t, gInfo, simRes.GasInfo)
		require.Equal(t, result.Log, simRes.Result.Log)
		require.Equal(t, result.Events, simRes.Result.Events)
		require.True(t, bytes.Equal(result.Data, simRes.Result.Data))

		suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: count})
		suite.baseApp.Commit()
	}
}

func TestABCI_InvalidTransaction(t *testing.T) {
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			return
		})
	}

	suite := NewBaseAppSuite(t, anteOpt)
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), CounterServerImplGasMeterOnly{})

	suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})

	suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
	})

	// transaction with no messages
	{
		emptyTx := suite.txConfig.NewTxBuilder().GetTx()
		bz, err := suite.txConfig.TxEncoder()(emptyTx)
		require.NoError(t, err)
		result, err := suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: 1,
			Txs:    [][]byte{bz},
		})
		require.EqualValues(t, sdkerrors.ErrInvalidRequest.Codespace(), result.TxResults[0].Codespace, err)
		require.EqualValues(t, sdkerrors.ErrInvalidRequest.ABCICode(), result.TxResults[0].Code, err)
	}

	// transaction where ValidateBasic fails
	{
		testCases := []struct {
			tx   signing.Tx
			fail bool
		}{
			{newTxCounter(t, suite.txConfig, 0, 0), false},
			{newTxCounter(t, suite.txConfig, -1, 0), false},
			{newTxCounter(t, suite.txConfig, 100, 100), false},
			{newTxCounter(t, suite.txConfig, 100, 5, 4, 3, 2, 1), false},

			{newTxCounter(t, suite.txConfig, 0, -1), true},
			{newTxCounter(t, suite.txConfig, 0, 1, -2), true},
			{newTxCounter(t, suite.txConfig, 0, 1, 2, -10, 5), true},
		}

		for _, testCase := range testCases {
			tx := testCase.tx
			_, result, err := suite.baseApp.SimDeliver(suite.txConfig.TxEncoder(), tx)

			if testCase.fail {
				require.Error(t, err)

				space, code, _ := errorsmod.ABCIInfo(err, false)
				require.EqualValues(t, sdkerrors.ErrInvalidSequence.Codespace(), space, err)
				require.EqualValues(t, sdkerrors.ErrInvalidSequence.ABCICode(), code, err)
			} else {
				require.NotNil(t, result)
			}
		}
	}

	// transaction with no known route
	{
		txBuilder := suite.txConfig.NewTxBuilder()
		_, _, addr := testdata.KeyTestPubAddr()
		txBuilder.SetMsgs(&baseapptestutil.MsgCounter2{Signer: addr.String()})
		setTxSignature(t, txBuilder, 0)
		unknownRouteTx := txBuilder.GetTx()

		_, result, err := suite.baseApp.SimDeliver(suite.txConfig.TxEncoder(), unknownRouteTx)
		require.Error(t, err)
		require.Nil(t, result)

		space, code, _ := errorsmod.ABCIInfo(err, false)
		require.EqualValues(t, sdkerrors.ErrUnknownRequest.Codespace(), space, err)
		require.EqualValues(t, sdkerrors.ErrUnknownRequest.ABCICode(), code, err)

		txBuilder = suite.txConfig.NewTxBuilder()
		txBuilder.SetMsgs(
			&baseapptestutil.MsgCounter{Signer: addr.String()},
			&baseapptestutil.MsgCounter2{Signer: addr.String()},
		)
		setTxSignature(t, txBuilder, 0)
		unknownRouteTx = txBuilder.GetTx()

		_, result, err = suite.baseApp.SimDeliver(suite.txConfig.TxEncoder(), unknownRouteTx)
		require.Error(t, err)
		require.Nil(t, result)

		space, code, _ = errorsmod.ABCIInfo(err, false)
		require.EqualValues(t, sdkerrors.ErrUnknownRequest.Codespace(), space, err)
		require.EqualValues(t, sdkerrors.ErrUnknownRequest.ABCICode(), code, err)
	}

	// Transaction with an unregistered message
	{
		txBuilder := suite.txConfig.NewTxBuilder()
		txBuilder.SetMsgs(&testdata.MsgCreateDog{})
		tx := txBuilder.GetTx()

		_, _, err := suite.baseApp.SimDeliver(suite.txConfig.TxEncoder(), tx)
		require.Error(t, err)
		space, code, _ := errorsmod.ABCIInfo(err, false)
		require.EqualValues(t, sdkerrors.ErrTxDecode.ABCICode(), code)
		require.EqualValues(t, sdkerrors.ErrTxDecode.Codespace(), space)
	}
}

func TestABCI_TxGasLimits(t *testing.T) {
	gasGranted := uint64(10)
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			newCtx = ctx.WithGasMeter(storetypes.NewGasMeter(gasGranted))

			// AnteHandlers must have their own defer/recover in order for the BaseApp
			// to know how much gas was used! This is because the GasMeter is created in
			// the AnteHandler, but if it panics the context won't be set properly in
			// runTx's recover call.
			defer func() {
				if r := recover(); r != nil {
					switch rType := r.(type) {
					case storetypes.ErrorOutOfGas:
						err = errorsmod.Wrapf(sdkerrors.ErrOutOfGas, "out of gas in location: %v", rType.Descriptor)
					default:
						panic(r)
					}
				}
			}()

			count, _ := parseTxMemo(t, tx)
			newCtx.GasMeter().ConsumeGas(uint64(count), "counter-ante")

			return newCtx, nil
		})
	}

	suite := NewBaseAppSuite(t, anteOpt)
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), CounterServerImplGasMeterOnly{})

	suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})

	suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
	})

	suite.baseApp.Commit()

	testCases := []struct {
		tx      signing.Tx
		gasUsed int64
		fail    bool
	}{
		{newTxCounter(t, suite.txConfig, 0, 0), 0, false},
		{newTxCounter(t, suite.txConfig, 1, 1), 2, false},
		{newTxCounter(t, suite.txConfig, 9, 1), 10, false},
		{newTxCounter(t, suite.txConfig, 1, 9), 10, false},
		{newTxCounter(t, suite.txConfig, 10, 0), 10, false},

		{newTxCounter(t, suite.txConfig, 9, 2), 11, true},
		{newTxCounter(t, suite.txConfig, 2, 9), 11, true},
		// {newTxCounter(t, suite.txConfig, 9, 1, 1), 11, true},
		// {newTxCounter(t, suite.txConfig, 1, 8, 1, 1), 11, true},
		//  {newTxCounter(t, suite.txConfig, 11, 0), 11, true},
		//  {newTxCounter(t, suite.txConfig, 0, 11), 11, true},
		//  {newTxCounter(t, suite.txConfig, 0, 5, 11), 16, true},
	}

	txs := [][]byte{}
	for _, tc := range testCases {
		tx := tc.tx
		bz, err := suite.txConfig.TxEncoder()(tx)
		require.NoError(t, err)
		txs = append(txs, bz)
	}

	// Deliver the txs
	res, err := suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 2,
		Txs:    txs,
	})

	require.NoError(t, err)

	for i, tc := range testCases {

		result := res.TxResults[i]

		require.Equal(t, tc.gasUsed, result.GasUsed, fmt.Sprintf("tc #%d; gas: %v, result: %v, err: %s", i, result.GasUsed, result, err))

		// check for out of gas
		if !tc.fail {
			require.NotNil(t, result, fmt.Sprintf("%d: %v, %v", i, tc, err))
		} else {
			require.EqualValues(t, sdkerrors.ErrOutOfGas.Codespace(), result.Codespace, err)
			require.EqualValues(t, sdkerrors.ErrOutOfGas.ABCICode(), result.Code, err)
		}
	}
}

func TestABCI_MaxBlockGasLimits(t *testing.T) {
	gasGranted := uint64(10)
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			newCtx = ctx.WithGasMeter(storetypes.NewGasMeter(gasGranted))

			defer func() {
				if r := recover(); r != nil {
					switch rType := r.(type) {
					case storetypes.ErrorOutOfGas:
						err = errorsmod.Wrapf(sdkerrors.ErrOutOfGas, "out of gas in location: %v", rType.Descriptor)
					default:
						panic(r)
					}
				}
			}()

			count, _ := parseTxMemo(t, tx)
			newCtx.GasMeter().ConsumeGas(uint64(count), "counter-ante")

			return
		})
	}

	suite := NewBaseAppSuite(t, anteOpt)
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), CounterServerImplGasMeterOnly{})

	suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{
			Block: &cmtproto.BlockParams{
				MaxGas: 100,
			},
		},
	})

	suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})

	testCases := []struct {
		tx                signing.Tx
		numDelivers       int
		gasUsedPerDeliver uint64
		fail              bool
		failAfterDeliver  int
	}{
		{newTxCounter(t, suite.txConfig, 0, 0), 0, 0, false, 0},
		{newTxCounter(t, suite.txConfig, 9, 1), 2, 10, false, 0},
		{newTxCounter(t, suite.txConfig, 10, 0), 3, 10, false, 0},
		{newTxCounter(t, suite.txConfig, 10, 0), 10, 10, false, 0},
		{newTxCounter(t, suite.txConfig, 2, 7), 11, 9, false, 0},
		// {newTxCounter(t, suite.txConfig, 10, 0), 10, 10, false, 0}, // hit the limit but pass

		// {newTxCounter(t, suite.txConfig, 10, 0), 11, 10, true, 10},
		// {newTxCounter(t, suite.txConfig, 10, 0), 15, 10, true, 10},
		// {newTxCounter(t, suite.txConfig, 9, 0), 12, 9, true, 11}, // fly past the limit
	}

	for i, tc := range testCases {
		tx := tc.tx

		// reset block gas
		suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: suite.baseApp.LastBlockHeight() + 1})

		// execute the transaction multiple times
		for j := 0; j < tc.numDelivers; j++ {

			_, result, err := suite.baseApp.SimDeliver(suite.txConfig.TxEncoder(), tx)

			ctx := getFinalizeBlockStateCtx(suite.baseApp)

			// check for failed transactions
			if tc.fail && (j+1) > tc.failAfterDeliver {
				require.Error(t, err, fmt.Sprintf("tc #%d; result: %v, err: %s", i, result, err))
				require.Nil(t, tx, fmt.Sprintf("tc #%d; result: %v, err: %s", i, result, err))

				space, code, _ := errorsmod.ABCIInfo(err, false)
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

				require.NotNil(t, tx, fmt.Sprintf("tc #%d; currDeliver: %d, result: %v, err: %s", i, j, result, err))
				require.False(t, ctx.BlockGasMeter().IsPastLimit())
			}
		}
	}
}

func TestABCI_GasConsumptionBadTx(t *testing.T) {
	gasWanted := uint64(5)
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			newCtx = ctx.WithGasMeter(storetypes.NewGasMeter(gasWanted))

			defer func() {
				if r := recover(); r != nil {
					switch rType := r.(type) {
					case storetypes.ErrorOutOfGas:
						log := fmt.Sprintf("out of gas in location: %v", rType.Descriptor)
						err = errorsmod.Wrap(sdkerrors.ErrOutOfGas, log)
					default:
						panic(r)
					}
				}
			}()

			counter, failOnAnte := parseTxMemo(t, tx)
			newCtx.GasMeter().ConsumeGas(uint64(counter), "counter-ante")
			if failOnAnte {
				return newCtx, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "ante handler failure")
			}

			return
		})
	}

	suite := NewBaseAppSuite(t, anteOpt)
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), CounterServerImplGasMeterOnly{})

	suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{
			Block: &cmtproto.BlockParams{
				MaxGas: 9,
			},
		},
	})

	tx := newTxCounter(t, suite.txConfig, 5, 0)
	tx = setFailOnAnte(t, suite.txConfig, tx, true)
	txBytes, err := suite.txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	// require next tx to fail due to black gas limit
	tx = newTxCounter(t, suite.txConfig, 5, 0)
	txBytes2, err := suite.txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	_, err = suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: suite.baseApp.LastBlockHeight() + 1,
		Txs:    [][]byte{txBytes, txBytes2},
	})
	require.NoError(t, err)
}

func TestABCI_Query(t *testing.T) {
	key, value := []byte("hello"), []byte("goodbye")
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			store := ctx.KVStore(capKey1)
			store.Set(key, value)
			return
		})
	}

	suite := NewBaseAppSuite(t, anteOpt)
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), CounterServerImplGasMeterOnly{})

	suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})

	// NOTE: "/store/key1" tells us KVStore
	// and the final "/key" says to use the data as the
	// key in the given KVStore ...
	query := abci.RequestQuery{
		Path: "/store/key1/key",
		Data: key,
	}
	tx := newTxCounter(t, suite.txConfig, 0, 0)

	// query is empty before we do anything
	res, err := suite.baseApp.Query(context.TODO(), &query)
	require.NoError(t, err)
	require.Equal(t, 0, len(res.Value))

	// query is still empty after a CheckTx
	_, resTx, err := suite.baseApp.SimCheck(suite.txConfig.TxEncoder(), tx)
	require.NoError(t, err)
	require.NotNil(t, resTx)

	res, err = suite.baseApp.Query(context.TODO(), &query)
	require.NoError(t, err)
	require.Equal(t, 0, len(res.Value))

	bz, err := suite.txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Txs:    [][]byte{bz},
	})
	require.NoError(t, err)

	res, err = suite.baseApp.Query(context.TODO(), &query)
	require.NoError(t, err)
	require.Equal(t, 0, len(res.Value))

	// query returns correct value after Commit
	suite.baseApp.Commit()

	res, err = suite.baseApp.Query(context.TODO(), &query)
	require.NoError(t, err)
	require.Equal(t, value, res.Value)
}

func TestABCI_GetBlockRetentionHeight(t *testing.T) {
	logger := log.NewTestLogger(t)
	db := dbm.NewMemDB()
	name := t.Name()

	snapshotStore, err := snapshots.NewStore(dbm.NewMemDB(), testutil.GetTempDir(t))
	require.NoError(t, err)

	testCases := map[string]struct {
		bapp         *baseapp.BaseApp
		maxAgeBlocks int64
		commitHeight int64
		expected     int64
	}{
		"defaults": {
			bapp:         baseapp.NewBaseApp(name, logger, db, nil),
			maxAgeBlocks: 0,
			commitHeight: 499000,
			expected:     0,
		},
		"pruning unbonding time only": {
			bapp:         baseapp.NewBaseApp(name, logger, db, nil, baseapp.SetMinRetainBlocks(1)),
			maxAgeBlocks: 362880,
			commitHeight: 499000,
			expected:     136120,
		},
		"pruning iavl snapshot only": {
			bapp: baseapp.NewBaseApp(
				name, logger, db, nil,
				baseapp.SetPruning(pruningtypes.NewPruningOptions(pruningtypes.PruningNothing)),
				baseapp.SetMinRetainBlocks(1),
				baseapp.SetSnapshot(snapshotStore, snapshottypes.NewSnapshotOptions(10000, 1)),
			),
			maxAgeBlocks: 0,
			commitHeight: 499000,
			expected:     489000,
		},
		"pruning state sync snapshot only": {
			bapp: baseapp.NewBaseApp(
				name, logger, db, nil,
				baseapp.SetSnapshot(snapshotStore, snapshottypes.NewSnapshotOptions(50000, 3)),
				baseapp.SetMinRetainBlocks(1),
			),
			maxAgeBlocks: 0,
			commitHeight: 499000,
			expected:     349000,
		},
		"pruning min retention only": {
			bapp: baseapp.NewBaseApp(
				name, logger, db, nil,
				baseapp.SetMinRetainBlocks(400000),
			),
			maxAgeBlocks: 0,
			commitHeight: 499000,
			expected:     99000,
		},
		"pruning all conditions": {
			bapp: baseapp.NewBaseApp(
				name, logger, db, nil,
				baseapp.SetPruning(pruningtypes.NewCustomPruningOptions(0, 0)),
				baseapp.SetMinRetainBlocks(400000),
				baseapp.SetSnapshot(snapshotStore, snapshottypes.NewSnapshotOptions(50000, 3)),
			),
			maxAgeBlocks: 362880,
			commitHeight: 499000,
			expected:     99000,
		},
		"no pruning due to no persisted state": {
			bapp: baseapp.NewBaseApp(
				name, logger, db, nil,
				baseapp.SetPruning(pruningtypes.NewCustomPruningOptions(0, 0)),
				baseapp.SetMinRetainBlocks(400000),
				baseapp.SetSnapshot(snapshotStore, snapshottypes.NewSnapshotOptions(50000, 3)),
			),
			maxAgeBlocks: 362880,
			commitHeight: 10000,
			expected:     0,
		},
		"disable pruning": {
			bapp: baseapp.NewBaseApp(
				name, logger, db, nil,
				baseapp.SetPruning(pruningtypes.NewCustomPruningOptions(0, 0)),
				baseapp.SetMinRetainBlocks(0),
				baseapp.SetSnapshot(snapshotStore, snapshottypes.NewSnapshotOptions(50000, 3)),
			),
			maxAgeBlocks: 362880,
			commitHeight: 499000,
			expected:     0,
		},
	}

	for name, tc := range testCases {
		tc := tc

		tc.bapp.SetParamStore(&paramStore{db: dbm.NewMemDB()})
		tc.bapp.InitChain(&abci.RequestInitChain{
			ConsensusParams: &cmtproto.ConsensusParams{
				Evidence: &cmtproto.EvidenceParams{
					MaxAgeNumBlocks: tc.maxAgeBlocks,
				},
			},
		})

		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.bapp.GetBlockRetentionHeight(tc.commitHeight))
		})
	}
}

// Verifies that PrepareCheckState is called with the checkState.
func TestPrepareCheckStateCalledWithCheckState(t *testing.T) {
	t.Parallel()

	logger := log.NewTestLogger(t)
	db := dbm.NewMemDB()
	name := t.Name()
	app := baseapp.NewBaseApp(name, logger, db, nil)

	wasPrepareCheckStateCalled := false
	app.SetPrepareCheckStater(func(ctx sdk.Context) {
		require.Equal(t, true, !ctx.InConsensus())
		wasPrepareCheckStateCalled = true
	})

	app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	app.Commit()

	require.Equal(t, true, wasPrepareCheckStateCalled)
}

// Verifies that the Precommiter is called with the deliverState.
func TestPrecommiterCalledWithDeliverState(t *testing.T) {
	t.Parallel()

	logger := log.NewTestLogger(t)
	db := dbm.NewMemDB()
	name := t.Name()
	app := baseapp.NewBaseApp(name, logger, db, nil)

	wasPrecommiterCalled := false
	app.SetPrecommiter(func(ctx sdk.Context) {
		require.Equal(t, false, !ctx.InConsensus())
		require.Equal(t, false, ctx.IsReCheckTx())
		wasPrecommiterCalled = true
	})

	app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	app.Commit()

	require.Equal(t, true, wasPrecommiterCalled)
}

func TestABCI_Proposal_HappyPath(t *testing.T) {
	anteKey := []byte("ante-key")
	pool := mempool.NewSenderNonceMempool()
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, anteKey))
	}

	suite := NewBaseAppSuite(t, anteOpt, baseapp.SetMempool(pool))
	baseapptestutil.RegisterKeyValueServer(suite.baseApp.MsgServiceRouter(), MsgKeyValueImpl{})
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), NoopCounterServerImpl{})

	suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})

	tx := newTxCounter(t, suite.txConfig, 0, 1)
	txBytes, err := suite.txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	reqCheckTx := abci.RequestCheckTx{
		Tx:   txBytes,
		Type: abci.CheckTxType_New,
	}
	_, err = suite.baseApp.CheckTx(&reqCheckTx)
	require.NoError(t, err)

	tx2 := newTxCounter(t, suite.txConfig, 1, 1)

	tx2Bytes, err := suite.txConfig.TxEncoder()(tx2)
	require.NoError(t, err)

	err = pool.Insert(sdk.Context{}, tx2)
	require.NoError(t, err)

	reqPrepareProposal := abci.RequestPrepareProposal{
		MaxTxBytes: 1000,
		Height:     1,
	}
	resPrepareProposal, err := suite.baseApp.PrepareProposal(&reqPrepareProposal)
	require.NoError(t, err)
	require.Equal(t, 2, len(resPrepareProposal.Txs))

	reqProposalTxBytes := [2][]byte{
		txBytes,
		tx2Bytes,
	}
	reqProcessProposal := abci.RequestProcessProposal{
		Txs:    reqProposalTxBytes[:],
		Height: reqPrepareProposal.Height,
	}

	resProcessProposal, err := suite.baseApp.ProcessProposal(&reqProcessProposal)
	require.NoError(t, err)
	require.Equal(t, abci.ResponseProcessProposal_ACCEPT, resProcessProposal.Status)

	res, err := suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: suite.baseApp.LastBlockHeight() + 1,
		Txs:    [][]byte{txBytes},
	})
	require.NoError(t, err)

	require.Equal(t, 1, pool.CountTx())

	require.NotEmpty(t, res.TxResults[0].Events)
	require.True(t, res.TxResults[0].IsOK(), fmt.Sprintf("%v", res))
}

func TestABCI_Proposal_Read_State_PrepareProposal(t *testing.T) {
	someKey := []byte("some-key")

	setInitChainerOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetInitChainer(func(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
			ctx.KVStore(capKey1).Set(someKey, []byte("foo"))
			return &abci.ResponseInitChain{}, nil
		})
	}

	prepareOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetPrepareProposal(func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
			value := ctx.KVStore(capKey1).Get(someKey)
			// We should be able to access any state written in InitChain
			require.Equal(t, "foo", string(value))
			return &abci.ResponsePrepareProposal{Txs: req.Txs}, nil
		})
	}

	suite := NewBaseAppSuite(t, setInitChainerOpt, prepareOpt)

	suite.baseApp.InitChain(&abci.RequestInitChain{
		InitialHeight:   1,
		ConsensusParams: &cmtproto.ConsensusParams{},
	})

	reqPrepareProposal := abci.RequestPrepareProposal{
		MaxTxBytes: 1000,
		Height:     1, // this value can't be 0
	}
	resPrepareProposal, err := suite.baseApp.PrepareProposal(&reqPrepareProposal)
	require.NoError(t, err)
	require.Equal(t, 0, len(resPrepareProposal.Txs))

	reqProposalTxBytes := [][]byte{}
	reqProcessProposal := abci.RequestProcessProposal{
		Txs:    reqProposalTxBytes,
		Height: reqPrepareProposal.Height,
	}

	resProcessProposal, err := suite.baseApp.ProcessProposal(&reqProcessProposal)
	require.NoError(t, err)
	require.Equal(t, abci.ResponseProcessProposal_ACCEPT, resProcessProposal.Status)

	// suite.baseApp.BeginBlock(abci.RequestBeginBlock{
	// 	Header: cmtproto.Header{Height: suite.baseApp.LastBlockHeight() + 1},
	// })
}

func TestABCI_PrepareProposal_ReachedMaxBytes(t *testing.T) {
	anteKey := []byte("ante-key")
	pool := mempool.NewSenderNonceMempool()
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, anteKey))
	}

	suite := NewBaseAppSuite(t, anteOpt, baseapp.SetMempool(pool))
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), NoopCounterServerImpl{})

	suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})

	for i := 0; i < 100; i++ {
		tx2 := newTxCounter(t, suite.txConfig, int64(i), int64(i))
		err := pool.Insert(sdk.Context{}, tx2)
		require.NoError(t, err)
	}

	reqPrepareProposal := abci.RequestPrepareProposal{
		MaxTxBytes: 1500,
		Height:     1,
	}
	resPrepareProposal, err := suite.baseApp.PrepareProposal(&reqPrepareProposal)
	require.NoError(t, err)
	require.Equal(t, 8, len(resPrepareProposal.Txs))
}

func TestABCI_PrepareProposal_BadEncoding(t *testing.T) {
	anteKey := []byte("ante-key")
	pool := mempool.NewSenderNonceMempool()
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, anteKey))
	}

	suite := NewBaseAppSuite(t, anteOpt, baseapp.SetMempool(pool))
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), NoopCounterServerImpl{})

	suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})

	tx := newTxCounter(t, suite.txConfig, 0, 0)
	err := pool.Insert(sdk.Context{}, tx)
	require.NoError(t, err)

	reqPrepareProposal := abci.RequestPrepareProposal{
		MaxTxBytes: 1000,
		Height:     1,
	}
	resPrepareProposal, err := suite.baseApp.PrepareProposal(&reqPrepareProposal)
	require.NoError(t, err)
	require.Equal(t, 1, len(resPrepareProposal.Txs))
}

func TestABCI_PrepareProposal_MaxGas(t *testing.T) {
	pool := mempool.NewSenderNonceMempool()
	suite := NewBaseAppSuite(t, baseapp.SetMempool(pool))
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), NoopCounterServerImpl{})

	// set max block gas limit to 100
	suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{
			Block: &cmtproto.BlockParams{MaxGas: 100},
		},
	})

	// insert 100 txs, each with a gas limit of 10
	_, _, addr := testdata.KeyTestPubAddr()
	for i := int64(0); i < 100; i++ {
		msg := &baseapptestutil.MsgCounter{Counter: i, FailOnHandler: false, Signer: addr.String()}
		msgs := []sdk.Msg{msg}

		builder := suite.txConfig.NewTxBuilder()
		builder.SetMsgs(msgs...)
		builder.SetMemo("counter=" + strconv.FormatInt(i, 10) + "&failOnAnte=false")
		builder.SetGasLimit(10)
		setTxSignature(t, builder, uint64(i))

		err := pool.Insert(sdk.Context{}, builder.GetTx())
		require.NoError(t, err)
	}

	// ensure we only select transactions that fit within the block gas limit
	res, err := suite.baseApp.PrepareProposal(&abci.RequestPrepareProposal{
		MaxTxBytes: 1_000_000, // large enough to ignore restriction
		Height:     1,
	})
	require.NoError(t, err)
	require.Len(t, res.Txs, 10, "invalid number of transactions returned")
}

func TestABCI_PrepareProposal_Failures(t *testing.T) {
	anteKey := []byte("ante-key")
	pool := mempool.NewSenderNonceMempool()
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(anteHandlerTxTest(t, capKey1, anteKey))
	}

	suite := NewBaseAppSuite(t, anteOpt, baseapp.SetMempool(pool))
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), NoopCounterServerImpl{})

	suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})

	tx := newTxCounter(t, suite.txConfig, 0, 0)
	txBytes, err := suite.txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	reqCheckTx := abci.RequestCheckTx{
		Tx:   txBytes,
		Type: abci.CheckTxType_New,
	}
	checkTxRes, err := suite.baseApp.CheckTx(&reqCheckTx)
	require.NoError(t, err)
	require.True(t, checkTxRes.IsOK())

	failTx := newTxCounter(t, suite.txConfig, 1, 1)
	failTx = setFailOnAnte(t, suite.txConfig, failTx, true)

	err = pool.Insert(sdk.Context{}, failTx)
	require.NoError(t, err)
	require.Equal(t, 2, pool.CountTx())

	req := abci.RequestPrepareProposal{
		MaxTxBytes: 1000,
		Height:     1,
	}
	res, err := suite.baseApp.PrepareProposal(&req)
	require.NoError(t, err)
	require.Equal(t, 1, len(res.Txs))
}

func TestABCI_PrepareProposal_PanicRecovery(t *testing.T) {
	prepareOpt := func(app *baseapp.BaseApp) {
		app.SetPrepareProposal(func(ctx sdk.Context, rpp *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
			panic(errors.New("test"))
		})
	}
	suite := NewBaseAppSuite(t, prepareOpt)

	suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})

	req := abci.RequestPrepareProposal{
		MaxTxBytes: 1000,
		Height:     1,
	}

	require.NotPanics(t, func() {
		res, err := suite.baseApp.PrepareProposal(&req)
		require.NoError(t, err)
		require.Equal(t, req.Txs, res.Txs)
	})
}

func TestABCI_ProcessProposal_PanicRecovery(t *testing.T) {
	processOpt := func(app *baseapp.BaseApp) {
		app.SetProcessProposal(func(ctx sdk.Context, rpp *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
			panic(errors.New("test"))
		})
	}
	suite := NewBaseAppSuite(t, processOpt)

	suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})

	require.NotPanics(t, func() {
		res, err := suite.baseApp.ProcessProposal(&abci.RequestProcessProposal{Height: 1})
		require.NoError(t, err)
		require.Equal(t, res.Status, abci.ResponseProcessProposal_REJECT)
	})
}

// TestABCI_Proposal_Reset_State ensures that state is reset between runs of
// PrepareProposal and ProcessProposal in case they are called multiple times.
// This is only valid for heights > 1, given that on height 1 we always set the
// state to be deliverState.
func TestABCI_Proposal_Reset_State_Between_Calls(t *testing.T) {
	someKey := []byte("some-key")

	prepareOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetPrepareProposal(func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
			// This key should not exist given that we reset the state on every call.
			require.False(t, ctx.KVStore(capKey1).Has(someKey))
			ctx.KVStore(capKey1).Set(someKey, someKey)
			return &abci.ResponsePrepareProposal{Txs: req.Txs}, nil
		})
	}

	processOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetProcessProposal(func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
			// This key should not exist given that we reset the state on every call.
			require.False(t, ctx.KVStore(capKey1).Has(someKey))
			ctx.KVStore(capKey1).Set(someKey, someKey)
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
		})
	}

	suite := NewBaseAppSuite(t, prepareOpt, processOpt)

	suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})

	reqPrepareProposal := abci.RequestPrepareProposal{
		MaxTxBytes: 1000,
		Height:     2, // this value can't be 0
	}

	// Let's pretend something happened and PrepareProposal gets called many
	// times, this must be safe to do.
	for i := 0; i < 5; i++ {
		resPrepareProposal, err := suite.baseApp.PrepareProposal(&reqPrepareProposal)
		require.NoError(t, err)
		require.Equal(t, 0, len(resPrepareProposal.Txs))
	}

	reqProposalTxBytes := [][]byte{}
	reqProcessProposal := abci.RequestProcessProposal{
		Txs:    reqProposalTxBytes,
		Height: 2,
	}

	// Let's pretend something happened and ProcessProposal gets called many
	// times, this must be safe to do.
	for i := 0; i < 5; i++ {
		resProcessProposal, err := suite.baseApp.ProcessProposal(&reqProcessProposal)
		require.NoError(t, err)
		require.Equal(t, abci.ResponseProcessProposal_ACCEPT, resProcessProposal.Status)
	}
}
