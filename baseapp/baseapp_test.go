package baseapp

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	tmtypes "github.com/tendermint/tendermint/types"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

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
	wire.RegisterCrypto(codec)
	return NewBaseApp(name, codec, logger, db)
}

func TestMountStores(t *testing.T) {
	name := t.Name()
	app := newBaseApp(name)
	assert.Equal(t, name, app.Name())

	// make some cap keys
	capKey1 := sdk.NewKVStoreKey("key1")
	capKey2 := sdk.NewKVStoreKey("key2")

	// no stores are mounted
	assert.Panics(t, func() { app.LoadLatestVersion(capKey1) })

	app.MountStoresIAVL(capKey1, capKey2)

	// stores are mounted
	err := app.LoadLatestVersion(capKey1)
	assert.Nil(t, err)

	// check both stores
	store1 := app.cms.GetCommitKVStore(capKey1)
	assert.NotNil(t, store1)
	store2 := app.cms.GetCommitKVStore(capKey2)
	assert.NotNil(t, store2)
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
	assert.Nil(t, err)

	emptyCommitID := sdk.CommitID{}

	lastHeight := app.LastBlockHeight()
	lastID := app.LastCommitID()
	assert.Equal(t, int64(0), lastHeight)
	assert.Equal(t, emptyCommitID, lastID)

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
	assert.Nil(t, err)
	testLoadVersionHelper(t, app, int64(2), commitID2)

	// reload with LoadVersion, see if you can commit the same block and get
	// the same result
	app = NewBaseApp(name, nil, logger, db)
	app.MountStoresIAVL(capKey)
	err = app.LoadVersion(1, capKey)
	assert.Nil(t, err)
	testLoadVersionHelper(t, app, int64(1), commitID1)
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	app.Commit()
	testLoadVersionHelper(t, app, int64(2), commitID2)
}

func testLoadVersionHelper(t *testing.T, app *BaseApp, expectedHeight int64, expectedID sdk.CommitID) {
	lastHeight := app.LastBlockHeight()
	lastID := app.LastCommitID()
	assert.Equal(t, expectedHeight, lastHeight)
	assert.Equal(t, expectedID, lastID)
}

// Test that the app hash is static
// TODO: https://github.com/cosmos/cosmos-sdk/issues/520
/*func TestStaticAppHash(t *testing.T) {
	app := newBaseApp(t.Name())

	// make a cap key and mount the store
	capKey := sdk.NewKVStoreKey("main")
	app.MountStoresIAVL(capKey)
	err := app.LoadLatestVersion(capKey) // needed to make stores non-nil
	assert.Nil(t, err)

	// execute some blocks
	header := abci.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	res := app.Commit()
	commitID1 := sdk.CommitID{1, res.Data}

	header = abci.Header{Height: 2}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	res = app.Commit()
	commitID2 := sdk.CommitID{2, res.Data}

	assert.Equal(t, commitID1.Hash, commitID2.Hash)
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
	assert.Equal(t, []uint8(nil), res.LastBlockAppHash)

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
	assert.Nil(t, err)

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
	assert.Equal(t, 0, len(res.Value))

	// set initChainer and try again - should see the value
	app.SetInitChainer(initChainer)
	app.InitChain(abci.RequestInitChain{AppStateBytes: []byte("{}")}) // must have valid JSON genesis file, even if empty
	app.Commit()
	res = app.Query(query)
	assert.Equal(t, value, res.Value)

	// reload app
	app = NewBaseApp(name, nil, logger, db)
	app.MountStoresIAVL(capKey, capKey2)
	err = app.LoadLatestVersion(capKey) // needed to make stores non-nil
	assert.Nil(t, err)
	app.SetInitChainer(initChainer)

	// ensure we can still query after reloading
	res = app.Query(query)
	assert.Equal(t, value, res.Value)

	// commit and ensure we can still query
	app.BeginBlock(abci.RequestBeginBlock{})
	app.Commit()
	res = app.Query(query)
	assert.Equal(t, value, res.Value)
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
			assert.Equal(t, counterBytes, prevBytes)
		}

		// set the current counter in the store
		counterBytes := []byte{byte(counter)}
		store.Set(counterBytes, counterBytes)

		// check that we can see the current header
		// wrapped in an if, so it can be reused between CheckTx and DeliverTx tests.
		if checkHeader {
			thisHeader := ctx.BlockHeader()
			height := int64((counter / txPerHeight) + 1)
			assert.Equal(t, height, thisHeader.Height)
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
func (tx testTx) GetMsg() sdk.Msg                    { return tx }
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
	assert.Nil(t, err)
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
		assert.Nil(t, storedValue)
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
	assert.Nil(t, err)

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
	assert.Nil(t, err)

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
		assert.Equal(t, height, thisHeader.Height)
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

	nBlocks := 3
	for blockN := 0; blockN < nBlocks; blockN++ {
		// block1
		header.Height = int64(blockN + 1)
		app.BeginBlock(abci.RequestBeginBlock{Header: header})
		result := app.Simulate(tx)
		require.Equal(t, result.Code, sdk.ABCICodeOK)
		require.Equal(t, int64(80), result.GasUsed)
		counter--
		encoded, err := json.Marshal(tx)
		require.Nil(t, err)
		query := abci.RequestQuery{
			Path: "/app/simulate",
			Data: encoded,
		}
		queryResult := app.Query(query)
		require.Equal(t, queryResult.Code, uint32(sdk.ABCICodeOK))
		var res sdk.Result
		app.cdc.MustUnmarshalBinary(queryResult.Value, &res)
		require.Equal(t, sdk.ABCICodeOK, res.Code)
		require.Equal(t, int64(160), res.GasUsed)
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
	assert.Nil(t, err)
	app.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx) (newCtx sdk.Context, res sdk.Result, abort bool) { return })
	app.Router().AddRoute(msgType2, func(ctx sdk.Context, msg sdk.Msg) (res sdk.Result) { return })
	app.BeginBlock(abci.RequestBeginBlock{})
	// Transaction where validate fails
	invalidTx := testTx{-1}
	err1 := app.Deliver(invalidTx)
	assert.Equal(t, sdk.ToABCICode(sdk.CodespaceRoot, sdk.CodeTxDecode), err1.Code)
	// Transaction with no known route
	unknownRouteTx := testUpdatePowerTx{}
	err2 := app.Deliver(unknownRouteTx)
	assert.Equal(t, sdk.ToABCICode(sdk.CodespaceRoot, sdk.CodeUnknownRequest), err2.Code)
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
	assert.Nil(t, err)

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
	assert.Equal(t, res.Code, sdk.ToABCICode(sdk.CodespaceRoot, sdk.CodeOutOfGas), "Expected transaction to run out of gas")
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
	assert.Nil(t, err)

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
	assert.Equal(t, 0, len(res.Value))

	tx := testUpdatePowerTx{} // doesn't matter

	// query is still empty after a CheckTx
	app.Check(tx)
	res = app.Query(query)
	assert.Equal(t, 0, len(res.Value))

	// query is still empty after a DeliverTx before we commit
	app.BeginBlock(abci.RequestBeginBlock{})
	app.Deliver(tx)
	res = app.Query(query)
	assert.Equal(t, 0, len(res.Value))

	// query returns correct value after Commit
	app.Commit()
	res = app.Query(query)
	assert.Equal(t, value, res.Value)
}

// Test p2p filter queries
func TestP2PQuery(t *testing.T) {
	app := newBaseApp(t.Name())

	// make a cap key and mount the store
	capKey := sdk.NewKVStoreKey("main")
	app.MountStoresIAVL(capKey)
	err := app.LoadLatestVersion(capKey) // needed to make stores non-nil
	assert.Nil(t, err)

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
func (tx testUpdatePowerTx) GetMsg() sdk.Msg                    { return tx }
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
	assert.Nil(t, err)
	assert.Equal(t, app.LastBlockHeight(), int64(0))

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
		assert.True(t, res.IsOK(), "%#v\nABCI log: %s", res, res.Log)
	}

	// Simulate the end of a block.
	// Get the summary of validator updates.
	res := app.EndBlock(abci.RequestEndBlock{})
	valUpdates := res.ValidatorUpdates

	// Assert that validator updates are correct.
	for _, val := range valSet {

		pubkey, err := tmtypes.PB2TM.PubKey(val.PubKey)
		// Sanity
		assert.Nil(t, err)

		// Find matching update and splice it out.
		for j := 0; j < len(valUpdates); j++ {
			valUpdate := valUpdates[j]

			updatePubkey, err := tmtypes.PB2TM.PubKey(valUpdate.PubKey)
			assert.Nil(t, err)

			// Matched.
			if updatePubkey.Equals(pubkey) {
				assert.Equal(t, valUpdate.Power, val.Power+1)
				if j < len(valUpdates)-1 {
					// Splice it out.
					valUpdates = append(valUpdates[:j], valUpdates[j+1:]...)
				}
				break
			}

			// Not matched.
		}
	}
	assert.Equal(t, len(valUpdates), 0, "Some validator updates were unexpected")
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
	bz, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}
	// fmt.Println(">> toJSON:", string(bz))
	return bz
}

func fromJSON(bz []byte, ptr interface{}) {
	// fmt.Println(">> fromJSON:", string(bz))
	err := json.Unmarshal(bz, ptr)
	if err != nil {
		panic(err)
	}
}
