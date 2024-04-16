//go:build e2e
// +build e2e

package baseapp_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	pruningtypes "cosmossdk.io/store/pruning/types"
	"cosmossdk.io/store/snapshots"
	snapshottypes "cosmossdk.io/store/snapshots/types"
	storetypes "cosmossdk.io/store/types"
	authtx "cosmossdk.io/x/auth/tx"

	"github.com/cosmos/cosmos-sdk/baseapp"
	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errorsmod "cosmossdk.io/errors"
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

func NewBaseAppSuite(t *testing.T, opts ...func(*baseapp.BaseApp)) *BaseAppSuite {
	t.Helper()
	cdc := codectestutil.CodecOptions{}.NewCodec()
	baseapptestutil.RegisterInterfaces(cdc.InterfaceRegistry())
	signingCtx := cdc.InterfaceRegistry().SigningContext()

	txConfig := authtx.NewTxConfig(cdc, signingCtx.AddressCodec(), signingCtx.ValidatorAddressCodec(), authtx.DefaultSignModes)
	db := dbm.NewMemDB()
	logBuffer := new(bytes.Buffer)
	logger := log.NewLogger(logBuffer, log.ColorOption(false))

	app := baseapp.NewBaseApp(t.Name(), logger, db, txConfig.TxDecoder(), opts...)
	require.Equal(t, t.Name(), app.Name())

	app.SetInterfaceRegistry(cdc.InterfaceRegistry())
	app.MsgServiceRouter().SetInterfaceRegistry(cdc.InterfaceRegistry())
	app.MountStores(capKey1, capKey2)
	app.SetParamStore(ParamStore{Db: dbm.NewMemDB()})
	app.SetTxDecoder(txConfig.TxDecoder())
	app.SetTxEncoder(txConfig.TxEncoder())

	// mount stores and seal
	require.Nil(t, app.LoadLatestVersion())

	return &BaseAppSuite{
		baseApp:   app,
		cdc:       cdc,
		txConfig:  txConfig,
		logBuffer: logBuffer,
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

func TestAnteHandlerGasMeter(t *testing.T) {
	// run BeginBlock and assert that the gas meter passed into the first Txn is zeroed out
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			gasMeter := ctx.BlockGasMeter()
			require.NotNil(t, gasMeter)
			require.Equal(t, storetypes.Gas(0), gasMeter.GasConsumed())
			return ctx, nil
		})
	}
	// set the beginBlocker to use some gas
	beginBlockerOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetBeginBlocker(func(ctx sdk.Context) (sdk.BeginBlock, error) {
			ctx.BlockGasMeter().ConsumeGas(1, "beginBlocker gas consumption")
			return sdk.BeginBlock{}, nil
		})
	}

	suite := NewBaseAppSuite(t, anteOpt, beginBlockerOpt)
	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})
	require.NoError(t, err)

	deliverKey := []byte("deliver-key")
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), CounterServerImpl{t, capKey1, deliverKey})

	tx := newTxCounter(t, suite.txConfig, 0, 0)
	txBytes, err := suite.txConfig.TxEncoder()(tx)
	require.NoError(t, err)
	_, err = suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1, Txs: [][]byte{txBytes}})
	require.NoError(t, err)
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
	cdc := codectestutil.CodecOptions{}.NewCodec()
	baseapptestutil.RegisterInterfaces(cdc.InterfaceRegistry())
	signingCtx := cdc.InterfaceRegistry().SigningContext()

	// patch in TxConfig instead of using an output from x/auth/tx
	txConfig := authtx.NewTxConfig(cdc, signingCtx.AddressCodec(), signingCtx.ValidatorAddressCodec(), authtx.DefaultSignModes)

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

func TestBaseAppPostHandler(t *testing.T) {
	postHandlerRun := false
	anteOpt := func(bapp *baseapp.BaseApp) {
		bapp.SetPostHandler(func(ctx sdk.Context, tx sdk.Tx, simulate, success bool) (newCtx sdk.Context, err error) {
			postHandlerRun = true
			return ctx, nil
		})
	}

	suite := NewBaseAppSuite(t, anteOpt)
	baseapptestutil.RegisterCounterServer(suite.baseApp.MsgServiceRouter(), CounterServerImpl{t, capKey1, []byte("foo")})

	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: &cmtproto.ConsensusParams{},
	})
	require.NoError(t, err)

	// execute a tx that will fail ante handler execution
	//
	// NOTE: State should not be mutated here. This will be implicitly checked by
	// the next txs ante handler execution (anteHandlerTxTest).
	tx := newTxCounter(t, suite.txConfig, 0, 0)
	txBytes, err := suite.txConfig.TxEncoder()(tx)
	require.NoError(t, err)

	res, err := suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1, Txs: [][]byte{txBytes}})
	require.NoError(t, err)
	require.Empty(t, res.Events)
	require.True(t, res.TxResults[0].IsOK(), fmt.Sprintf("%v", res))

	// PostHandler runs on successful message execution
	require.True(t, postHandlerRun)

	// It should also run on failed message execution
	postHandlerRun = false
	tx = setFailOnHandler(t, suite.txConfig, tx, true)
	txBytes, err = suite.txConfig.TxEncoder()(tx)
	require.NoError(t, err)
	res, err = suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1, Txs: [][]byte{txBytes}})
	require.NoError(t, err)
	require.Empty(t, res.Events)
	require.False(t, res.TxResults[0].IsOK(), fmt.Sprintf("%v", res))

	require.True(t, postHandlerRun)

	// regression test, should not panic when runMsgs fails
	tx = wonkyMsg(t, suite.txConfig, tx)
	txBytes, err = suite.txConfig.TxEncoder()(tx)
	require.NoError(t, err)
	_, err = suite.baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1, Txs: [][]byte{txBytes}})
	require.NoError(t, err)
	require.NotContains(t, suite.logBuffer.String(), "panic recovered in runTx")
}

func TestSetMinGasPrices(t *testing.T) {
	minGasPrices := sdk.DecCoins{sdk.NewInt64DecCoin("stake", 5000)}
	suite := NewBaseAppSuite(t, baseapp.SetMinGasPrices(minGasPrices.String()))

	ctx := getCheckStateCtx(suite.baseApp)
	require.Equal(t, minGasPrices, ctx.MinGasPrices())
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

func TestGetEmptyConsensusParams(t *testing.T) {
	suite := NewBaseAppSuite(t)
	_, err := suite.baseApp.InitChain(&abci.RequestInitChain{})
	require.NoError(t, err)
	ctx := suite.baseApp.NewContext(true)

	cp := suite.baseApp.GetConsensusParams(ctx)
	require.Equal(t, cmtproto.ConsensusParams{}, cp)
	require.Equal(t, uint64(0), suite.baseApp.GetMaximumBlockGas(ctx))
}