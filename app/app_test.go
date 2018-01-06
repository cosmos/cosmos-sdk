package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
)

// A mock transaction to update a validator's voting power.
type testTx struct {
	Addr     []byte
	NewPower int64
}

func (tx testTx) Get(key interface{}) (value interface{}) { return nil }
func (tx testTx) SignBytes() []byte                       { return nil }
func (tx testTx) ValidateBasic() error                    { return nil }
func (tx testTx) Signers() []crypto.Address               { return nil }
func (tx testTx) TxBytes() []byte                         { return nil }
func (tx testTx) Signatures() []types.StdSignature        { return nil }

func TestBasic(t *testing.T) {

	// Create app.
	app := NewApp(t.Name())
	app.SetCommitMultiStore(newCommitMultiStore())
	app.SetTxParser(func(txBytes []byte) (types.Tx, error) {
		var ttx testTx
		fromJSON(txBytes, &ttx)
		return ttx, nil
	})
	app.SetHandler(func(ctx types.Context, store types.MultiStore, tx types.Tx) types.Result {
		// TODO
		return types.Result{}
	})

	// Load latest state, which should be empty.
	err := app.LoadLatestVersion()
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
		tx := testTx{
			Addr:     makePubKey(secret(i)).Address(),
			NewPower: val.Power + 1,
		}
		txBytes := toJSON(tx)
		res := app.DeliverTx(txBytes)
		assert.True(t, res.IsOK(), "%#v", res)
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
			j += 1
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
	return privKey.Wrap()
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

// Creates a sample CommitMultiStore
func newCommitMultiStore() types.CommitMultiStore {
	dbMain := dbm.NewMemDB()
	dbXtra := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(dbMain) // Also store rootMultiStore metadata here (it shouldn't clash)
	ms.SetSubstoreLoader("main", store.NewIAVLStoreLoader(dbMain, 0, 0))
	ms.SetSubstoreLoader("xtra", store.NewIAVLStoreLoader(dbXtra, 0, 0))
	return ms
}
