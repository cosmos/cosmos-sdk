package baseapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func defaultLogger() log.Logger {
	return log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
}

func newBaseApp(name string) *BaseApp {
	logger := defaultLogger()
	db := dbm.NewMemDB()
	return NewBaseApp(name, nil, logger, db, 10000)
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
	app := NewBaseApp(name, nil, logger, db, 10000)

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
	commitID := sdk.CommitID{1, res.Data}

	// reload
	app = NewBaseApp(name, nil, logger, db, 10000)
	app.MountStoresIAVL(capKey)
	err = app.LoadLatestVersion(capKey) // needed to make stores non-nil
	assert.Nil(t, err)

	lastHeight = app.LastBlockHeight()
	lastID = app.LastCommitID()
	assert.Equal(t, int64(1), lastHeight)
	assert.Equal(t, commitID, lastID)
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
	app := NewBaseApp(name, nil, logger, db, 10000)
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
		Path: "/main/key",
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
	app = NewBaseApp(name, nil, logger, db, 10000)
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

// Test that successive CheckTx can see each others' effects
// on the store within a block, and that the CheckTx state
// gets reset to the latest Committed state during Commit
func TestCheckTx(t *testing.T) {
	// TODO
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

	counter := 0
	txPerHeight := 2
	app.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx) (newCtx sdk.Context, res sdk.Result, abort bool) { return })
	app.Router().AddRoute(msgType, func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		store := ctx.KVStore(capKey)
		if counter > 0 {
			// check previous value in store
			counterBytes := []byte{byte(counter - 1)}
			prevBytes := store.Get(counterBytes)
			assert.Equal(t, prevBytes, counterBytes)
		}

		// set the current counter in the store
		counterBytes := []byte{byte(counter)}
		store.Set(counterBytes, counterBytes)

		// check we can see the current header
		thisHeader := ctx.BlockHeader()
		height := int64((counter / txPerHeight) + 1)
		assert.Equal(t, height, thisHeader.Height)

		counter++
		return sdk.Result{}
	})

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
		Path: "/main/key",
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

//----------------------
// TODO: clean this up

// A mock transaction to update a validator's voting power.
type testUpdatePowerTx struct {
	Addr     []byte
	NewPower int64
}

const msgType = "testUpdatePowerTx"

func (tx testUpdatePowerTx) Type() string                      { return msgType }
func (tx testUpdatePowerTx) GetMsg() sdk.Msg                   { return tx }
func (tx testUpdatePowerTx) GetSignBytes() []byte              { return nil }
func (tx testUpdatePowerTx) ValidateBasic() sdk.Error          { return nil }
func (tx testUpdatePowerTx) GetSigners() []sdk.Address         { return nil }
func (tx testUpdatePowerTx) GetSignatures() []sdk.StdSignature { return nil }

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
		// Sanity
		assert.NotEqual(t, len(val.PubKey), 0)

		// Find matching update and splice it out.
		for j := 0; j < len(valUpdates); {
			valUpdate := valUpdates[j]

			// Matched.
			if bytes.Equal(valUpdate.PubKey, val.PubKey) {
				assert.Equal(t, valUpdate.Power, val.Power+1)
				if j < len(valUpdates)-1 {
					// Splice it out.
					valUpdates = append(valUpdates[:j], valUpdates[j+1:]...)
				}
				break
			}

			// Not matched.
			j++
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
		PubKey: makePubKey(secret).Bytes(),
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
