package baseapp

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

func defaultLogger() log.Logger {
	return log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
}

func newBaseApp(name string) *BaseApp {
	logger := defaultLogger()
	db := dbm.NewMemDB()
	codec := wire.NewCodec()
	auth.RegisterBaseAccount(codec)
	return NewBaseApp(name, codec, logger, db)
}

func TestMountStores(t *testing.T) {
	name := t.Name()
	app := newBaseApp(name)
	require.Equal(t, name, app.Name())

	// make some cap keys
	capKey1 := sdk.NewKVStoreKey("key1")
	capKey2 := sdk.NewKVStoreKey("key2")

	// no stores are mounted
	require.Panics(t, func() { app.LoadLatestVersion(capKey1) })

	app.MountStoresIAVL(capKey1, capKey2)

	// stores are mounted
	err := app.LoadLatestVersion(capKey1)
	require.Nil(t, err)

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
	app := NewBaseApp(name, nil, logger, db)

	// make a cap key and mount the store
	capKey := sdk.NewKVStoreKey("main")
	app.MountStoresIAVL(capKey)
	err := app.LoadLatestVersion(capKey) // needed to make stores non-nil
	require.Nil(t, err)

	emptyCommitID := sdk.CommitID{}

	lastHeight := app.LastBlockHeight()
	lastID := app.LastCommitID()
	require.Equal(t, int64(0), lastHeight)
	require.Equal(t, emptyCommitID, lastID)

	// execute some blocks
	header := abci.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	res := app.Commit()
	commitID1 := sdk.CommitID{1, res.Data}
	header = abci.Header{Height: 2}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	res = app.Commit()
	commitID2 := sdk.CommitID{2, res.Data}

	// reload with LoadLatestVersion
	app = NewBaseApp(name, nil, logger, db)
	app.MountStoresIAVL(capKey)
	err = app.LoadLatestVersion(capKey)
	require.Nil(t, err)
	testLoadVersionHelper(t, app, int64(2), commitID2)

	// reload with LoadVersion, see if you can commit the same block and get
	// the same result
	app = NewBaseApp(name, nil, logger, db)
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

func TestInitChainer(t *testing.T) {
	name := t.Name()
	db := dbm.NewMemDB()
	logger := defaultLogger()
	app := NewBaseApp(name, nil, logger, db)
	// make cap keys and mount the stores
	// NOTE/TODO: mounting multiple stores is broken
	// see https://github.com/cosmos/cosmos-sdk/issues/532
	capKey := sdk.NewKVStoreKey("main")
	capKey2 := sdk.NewKVStoreKey("key2")
	app.MountStoresIAVL(capKey, capKey2)
	err := app.LoadLatestVersion(capKey) // needed to make stores non-nil
	require.Nil(t, err)

	key, value := []byte("hello"), []byte("goodbye")

	// initChainer sets a value in the store
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
	app = NewBaseApp(name, nil, logger, db)
	app.MountStoresIAVL(capKey, capKey2)
	err = app.LoadLatestVersion(capKey) // needed to make stores non-nil
	require.Nil(t, err)
	app.SetInitChainer(initChainer)

	// ensure we can still query after reloading
	res = app.Query(query)
	require.Equal(t, value, res.Value)

	// commit and ensure we can still query
	app.BeginBlock(abci.RequestBeginBlock{})
	app.Commit()
	res = app.Query(query)
	require.Equal(t, value, res.Value)
}

func getStateCheckingHandler(t *testing.T, capKey *sdk.KVStoreKey, txPerHeight int, checkHeader bool) func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
	counter := 0
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		store := ctx.KVStore(capKey)
		// Checking state gets updated between checkTx's / DeliverTx's
		// on the store within a block.
		if counter > 0 {
			// check previous value in store
			counterBytes := []byte{byte(counter - 1)}
			prevBytes := store.Get(counterBytes)
			require.Equal(t, counterBytes, prevBytes)
		}

		// set the current counter in the store
		counterBytes := []byte{byte(counter)}
		store.Set(counterBytes, counterBytes)

		// check that we can see the current header
		// wrapped in an if, so it can be reused between CheckTx and DeliverTx tests.
		if checkHeader {
			thisHeader := ctx.BlockHeader()
			height := int64((counter / txPerHeight) + 1)
			require.Equal(t, height, thisHeader.Height)
		}

		counter++
		return sdk.Result{}
	}
}

// A mock transaction that has a validation which can fail.
type testTx struct {
	positiveNum int64
}

const msgType2 = "testTx"

func (tx testTx) Type() string                       { return msgType2 }
func (tx testTx) GetMemo() string                    { return "" }
func (tx testTx) GetMsgs() []sdk.Msg                 { return []sdk.Msg{tx} }
func (tx testTx) GetSignBytes() []byte               { return nil }
func (tx testTx) GetSigners() []sdk.Address          { return nil }
func (tx testTx) GetSignatures() []auth.StdSignature { return nil }
func (tx testTx) ValidateBasic() sdk.Error {
	if tx.positiveNum >= 0 {
		return nil
	}
	return sdk.ErrTxDecode("positiveNum should be a non-negative integer.")
}

// Test that successive CheckTx can see each others' effects
// on the store within a block, and that the CheckTx state
// gets reset to the latest Committed state during Commit
func TestCheckTx(t *testing.T) {
	// Initialize an app for testing
	app := newBaseApp(t.Name())
	// make a cap key and mount the store
	capKey := sdk.NewKVStoreKey("main")
	app.MountStoresIAVL(capKey)
	err := app.LoadLatestVersion(capKey) // needed to make stores non-nil
	require.Nil(t, err)
	app.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx) (newCtx sdk.Context, res sdk.Result, abort bool) { return })

	txPerHeight := 3
	app.Router().AddRoute(msgType, getStateCheckingHandler(t, capKey, txPerHeight, false)).
		AddRoute(msgType2, func(ctx sdk.Context, msg sdk.Msg) (res sdk.Result) { return })
	tx := testUpdatePowerTx{} // doesn't matter
	for i := 0; i < txPerHeight; i++ {
		app.Check(tx)
	}
	// If it gets to this point, then successive CheckTx's can see the effects of
	// other CheckTx's on the block. The following checks that if another block
	// is committed, the CheckTx State will reset.
	app.BeginBlock(abci.RequestBeginBlock{})
	tx2 := testTx{}
	for i := 0; i < txPerHeight; i++ {
		app.Deliver(tx2)
	}
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	checkStateStore := app.checkState.ctx.KVStore(capKey)
	for i := 0; i < txPerHeight; i++ {
		storedValue := checkStateStore.Get([]byte{byte(i)})
		require.Nil(t, storedValue)
	}
}

// Test that successive DeliverTx can see each others' effects
// on the store, both within and across blocks.
func TestDeliverTx(t *testing.T) {
	app := newBaseApp(t.Name())

	// make a cap key and mount the store
	capKey := sdk.NewKVStoreKey("main")
	app.MountStoresIAVL(capKey)
	err := app.LoadLatestVersion(capKey) // needed to make stores non-nil
	require.Nil(t, err)

	txPerHeight := 2
	app.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx) (newCtx sdk.Context, res sdk.Result, abort bool) { return })
	app.Router().AddRoute(msgType, getStateCheckingHandler(t, capKey, txPerHeight, true))

	tx := testUpdatePowerTx{} // doesn't matter
	header := abci.Header{AppHash: []byte("apphash")}

	nBlocks := 3
	for blockN := 0; blockN < nBlocks; blockN++ {
		// block1
		header.Height = int64(blockN + 1)
		app.BeginBlock(abci.RequestBeginBlock{Header: header})
		for i := 0; i < txPerHeight; i++ {
			app.Deliver(tx)
		}
		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}
}

func TestSimulateTx(t *testing.T) {
	app := newBaseApp(t.Name())

	// make a cap key and mount the store
	capKey := sdk.NewKVStoreKey("main")
	app.MountStoresIAVL(capKey)
	err := app.LoadLatestVersion(capKey) // needed to make stores non-nil
	require.Nil(t, err)

	counter := 0
	app.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx) (newCtx sdk.Context, res sdk.Result, abort bool) { return })
	app.Router().AddRoute(msgType, func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx.GasMeter().ConsumeGas(10, "test")
		store := ctx.KVStore(capKey)
		// ensure store is never written
		require.Nil(t, store.Get([]byte("key")))
		store.Set([]byte("key"), []byte("value"))
		// check we can see the current header
		thisHeader := ctx.BlockHeader()
		height := int64(counter)
		require.Equal(t, height, thisHeader.Height)
		counter++
		return sdk.Result{}
	})

	tx := testUpdatePowerTx{} // doesn't matter
	header := abci.Header{AppHash: []byte("apphash")}

	app.SetTxDecoder(func(txBytes []byte) (sdk.Tx, sdk.Error) {
		var ttx testUpdatePowerTx
		fromJSON(txBytes, &ttx)
		return ttx, nil
	})

	app.InitChain(abci.RequestInitChain{})

	nBlocks := 3
	for blockN := 0; blockN < nBlocks; blockN++ {
		// block1
		header.Height = int64(blockN + 1)
		app.BeginBlock(abci.RequestBeginBlock{Header: header})
		result := app.Simulate(tx)
		require.Equal(t, result.Code, sdk.ABCICodeOK, result.Log)
		require.Equal(t, int64(80), result.GasUsed)
		counter--
		encoded, err := app.cdc.MarshalJSON(tx)
		require.Nil(t, err)
		query := abci.RequestQuery{
			Path: "/app/simulate",
			Data: encoded,
		}
		queryResult := app.Query(query)
		require.Equal(t, queryResult.Code, uint32(sdk.ABCICodeOK))
		var res sdk.Result
		app.cdc.MustUnmarshalBinary(queryResult.Value, &res)
		require.Equal(t, sdk.ABCICodeOK, res.Code, res.Log)
		require.Equal(t, int64(160), res.GasUsed, res.Log)
		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}
}

func TestRunInvalidTransaction(t *testing.T) {
	// Initialize an app for testing
	app := newBaseApp(t.Name())
	// make a cap key and mount the store
	capKey := sdk.NewKVStoreKey("main")
	app.MountStoresIAVL(capKey)
	err := app.LoadLatestVersion(capKey) // needed to make stores non-nil
	require.Nil(t, err)
	app.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx) (newCtx sdk.Context, res sdk.Result, abort bool) { return })
	app.Router().AddRoute(msgType2, func(ctx sdk.Context, msg sdk.Msg) (res sdk.Result) { return })
	app.BeginBlock(abci.RequestBeginBlock{})
	// Transaction where validate fails
	invalidTx := testTx{-1}
	err1 := app.Deliver(invalidTx)
	require.Equal(t, sdk.ToABCICode(sdk.CodespaceRoot, sdk.CodeTxDecode), err1.Code)
	// Transaction with no known route
	unknownRouteTx := testUpdatePowerTx{}
	err2 := app.Deliver(unknownRouteTx)
	require.Equal(t, sdk.ToABCICode(sdk.CodespaceRoot, sdk.CodeUnknownRequest), err2.Code)
}

// Test that transactions exceeding gas limits fail
func TestTxGasLimits(t *testing.T) {
	logger := defaultLogger()
	db := dbm.NewMemDB()
	app := NewBaseApp(t.Name(), nil, logger, db)

	// make a cap key and mount the store
	capKey := sdk.NewKVStoreKey("main")
	app.MountStoresIAVL(capKey)
	err := app.LoadLatestVersion(capKey) // needed to make stores non-nil
	require.Nil(t, err)

	app.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx) (newCtx sdk.Context, res sdk.Result, abort bool) {
		newCtx = ctx.WithGasMeter(sdk.NewGasMeter(0))
		return
	})
	app.Router().AddRoute(msgType, func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx.GasMeter().ConsumeGas(10, "counter")
		return sdk.Result{}
	})

	tx := testUpdatePowerTx{} // doesn't matter
	header := abci.Header{AppHash: []byte("apphash")}

	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	res := app.Deliver(tx)
	require.Equal(t, res.Code, sdk.ToABCICode(sdk.CodespaceRoot, sdk.CodeOutOfGas), "Expected transaction to run out of gas")
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()
}

// Test that we can only query from the latest committed state.
func TestQuery(t *testing.T) {
	app := newBaseApp(t.Name())

	// make a cap key and mount the store
	capKey := sdk.NewKVStoreKey("main")
	app.MountStoresIAVL(capKey)
	err := app.LoadLatestVersion(capKey) // needed to make stores non-nil
	require.Nil(t, err)

	key, value := []byte("hello"), []byte("goodbye")

	app.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx) (newCtx sdk.Context, res sdk.Result, abort bool) { return })
	app.Router().AddRoute(msgType, func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		store := ctx.KVStore(capKey)
		store.Set(key, value)
		return sdk.Result{}
	})

	query := abci.RequestQuery{
		Path: "/store/main/key",
		Data: key,
	}

	// query is empty before we do anything
	res := app.Query(query)
	require.Equal(t, 0, len(res.Value))

	tx := testUpdatePowerTx{} // doesn't matter

	// query is still empty after a CheckTx
	app.Check(tx)
	res = app.Query(query)
	require.Equal(t, 0, len(res.Value))

	// query is still empty after a DeliverTx before we commit
	app.BeginBlock(abci.RequestBeginBlock{})
	app.Deliver(tx)
	res = app.Query(query)
	require.Equal(t, 0, len(res.Value))

	// query returns correct value after Commit
	app.Commit()
	res = app.Query(query)
	require.Equal(t, value, res.Value)
}

// Test p2p filter queries
func TestP2PQuery(t *testing.T) {
	app := newBaseApp(t.Name())

	// make a cap key and mount the store
	capKey := sdk.NewKVStoreKey("main")
	app.MountStoresIAVL(capKey)
	err := app.LoadLatestVersion(capKey) // needed to make stores non-nil
	require.Nil(t, err)

	app.SetAddrPeerFilter(func(addrport string) abci.ResponseQuery {
		require.Equal(t, "1.1.1.1:8000", addrport)
		return abci.ResponseQuery{Code: uint32(3)}
	})

	app.SetPubKeyPeerFilter(func(pubkey string) abci.ResponseQuery {
		require.Equal(t, "testpubkey", pubkey)
		return abci.ResponseQuery{Code: uint32(4)}
	})

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

//----------------------
// TODO: clean this up

// A mock transaction to update a validator's voting power.
type testUpdatePowerTx struct {
	Addr     []byte
	NewPower int64
}

const msgType = "testUpdatePowerTx"

func (tx testUpdatePowerTx) Type() string                       { return msgType }
func (tx testUpdatePowerTx) GetMemo() string                    { return "" }
func (tx testUpdatePowerTx) GetMsgs() []sdk.Msg                 { return []sdk.Msg{tx} }
func (tx testUpdatePowerTx) GetSignBytes() []byte               { return nil }
func (tx testUpdatePowerTx) ValidateBasic() sdk.Error           { return nil }
func (tx testUpdatePowerTx) GetSigners() []sdk.Address          { return nil }
func (tx testUpdatePowerTx) GetSignatures() []auth.StdSignature { return nil }

func TestValidatorChange(t *testing.T) {

	// Create app.
	app := newBaseApp(t.Name())
	capKey := sdk.NewKVStoreKey("key")
	app.MountStoresIAVL(capKey)
	app.SetTxDecoder(func(txBytes []byte) (sdk.Tx, sdk.Error) {
		var ttx testUpdatePowerTx
		fromJSON(txBytes, &ttx)
		return ttx, nil
	})

	app.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx) (newCtx sdk.Context, res sdk.Result, abort bool) { return })
	app.Router().AddRoute(msgType, func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		// TODO
		return sdk.Result{}
	})

	// Load latest state, which should be empty.
	err := app.LoadLatestVersion(capKey)
	require.Nil(t, err)
	require.Equal(t, app.LastBlockHeight(), int64(0))

	// Create the validators
	var numVals = 3
	var valSet = make([]abci.Validator, numVals)
	for i := 0; i < numVals; i++ {
		valSet[i] = makeVal(secret(i))
	}

	// Initialize the chain
	app.InitChain(abci.RequestInitChain{
		Validators: valSet,
	})

	// Simulate the start of a block.
	app.BeginBlock(abci.RequestBeginBlock{})

	// Add 1 to each validator's voting power.
	for i, val := range valSet {
		tx := testUpdatePowerTx{
			Addr:     makePubKey(secret(i)).Address(),
			NewPower: val.Power + 1,
		}
		txBytes := toJSON(tx)
		res := app.DeliverTx(txBytes)
		require.True(t, res.IsOK(), "%#v\nABCI log: %s", res, res.Log)
	}

	// Simulate the end of a block.
	// Get the summary of validator updates.
	res := app.EndBlock(abci.RequestEndBlock{})
	valUpdates := res.ValidatorUpdates

	// Assert that validator updates are correct.
	for _, val := range valSet {

		pubkey, err := tmtypes.PB2TM.PubKey(val.PubKey)
		// Sanity
		require.Nil(t, err)

		// Find matching update and splice it out.
		for j := 0; j < len(valUpdates); j++ {
			valUpdate := valUpdates[j]

			updatePubkey, err := tmtypes.PB2TM.PubKey(valUpdate.PubKey)
			require.Nil(t, err)

			// Matched.
			if updatePubkey.Equals(pubkey) {
				require.Equal(t, valUpdate.Power, val.Power+1)
				if j < len(valUpdates)-1 {
					// Splice it out.
					valUpdates = append(valUpdates[:j], valUpdates[j+1:]...)
				}
				break
			}

			// Not matched.
		}
	}
	require.Equal(t, len(valUpdates), 0, "Some validator updates were unexpected")
}

//----------------------------------------

func randPower() int64 {
	return cmn.RandInt64()
}

func makeVal(secret string) abci.Validator {
	return abci.Validator{
		PubKey: tmtypes.TM2PB.PubKey(makePubKey(secret)),
		Power:  randPower(),
	}
}

func makePubKey(secret string) crypto.PubKey {
	return makePrivKey(secret).PubKey()
}

func makePrivKey(secret string) crypto.PrivKey {
	privKey := crypto.GenPrivKeyEd25519FromSecret([]byte(secret))
	return privKey
}

func secret(index int) string {
	return fmt.Sprintf("secret%d", index)
}

func copyVal(val abci.Validator) abci.Validator {
	// val2 := *val
	// return &val2
	return val
}

func toJSON(o interface{}) []byte {
	bz, err := wire.Cdc.MarshalJSON(o)
	if err != nil {
		panic(err)
	}
	return bz
}

func fromJSON(bz []byte, ptr interface{}) {
	err := wire.Cdc.UnmarshalJSON(bz, ptr)
	if err != nil {
		panic(err)
	}
}
