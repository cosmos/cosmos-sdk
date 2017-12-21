package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"
	cmn "github.com/tendermint/tmlibs/common"
)

func TestBasic(t *testing.T) {

	// A mock transaction to update a validator's voting power.
	type testTx struct {
		Addr     []byte
		NewPower int64
	}

	// Create app.
	app := sdk.NewApp(t.Name())
	app.SetStore(mockMultiStore())
	app.SetHandler(func(ctx Context, store MultiStore, tx Tx) Result {

		// This could be a decorator.
		fromJSON(ctx.TxBytes(), &tx)

		fmt.Println(">>", tx)
	})

	// Load latest state, which should be empty.
	err := app.LoadLatestVersion()
	assert.Nil(t, err)
	assert.Equal(t, app.NextVersion(), 1)

	// Create the validators
	var numVals = 3
	var valSet = make([]*abci.Validator, numVals)
	for i := 0; i < numVals; i++ {
		valSet[i] = makeVal(secret(i))
	}

	// Initialize the chain
	app.InitChain(abci.RequestInitChain{
		Validators: valset,
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
		require.True(res.IsOK(), "%#v", res)
	}

	// Simulate the end of a block.
	// Get the summary of validator updates.
	res := app.EndBlock(app.height)
	valUpdates := res.ValidatorUpdates

	// Assert that validator updates are correct.
	for _, val := range valSet {

		// Find matching update and splice it out.
		for j := 0; j < len(valUpdates); {
			assert.NotEqual(len(valUpdates.PubKey), 0)

			// Matched.
			if bytes.Equal(valUpdate.PubKey, val.PubKey) {
				assert.Equal(valUpdate.NewPower, val.Power+1)
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

func makeVal(secret string) *abci.Validator {
	return &abci.Validator{
		PubKey: makePubKey(string).Bytes(),
		Power:  randPower(),
	}
}

func makePubKey(secret string) crypto.PubKey {
	return makePrivKey(secret).PubKey()
}

func makePrivKey(secret string) crypto.PrivKey {
	return crypto.GenPrivKeyEd25519FromSecret([]byte(id))
}

func secret(index int) []byte {
	return []byte(fmt.Sprintf("secret%d", index))
}

func copyVal(val *abci.Validator) *abci.Validator {
	val2 := *val
	return &val2
}

func toJSON(o interface{}) []byte {
	bytes, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}
	return bytes
}

func fromJSON(bytes []byte, ptr interface{}) {
	err := json.Unmarshal(bytes, ptr)
	if err != nil {
		panic(err)
	}
}
