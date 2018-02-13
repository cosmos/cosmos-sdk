package baseapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// A mock transaction to update a validator's voting power.
type testUpdatePowerTx struct {
	Addr     []byte
	NewPower int64
}

const msgType = "testUpdatePowerTx"

func (tx testUpdatePowerTx) Type() string                            { return msgType }
func (tx testUpdatePowerTx) Get(key interface{}) (value interface{}) { return nil }
func (tx testUpdatePowerTx) GetMsg() sdk.Msg                         { return tx }
func (tx testUpdatePowerTx) GetSignBytes() []byte                    { return nil }
func (tx testUpdatePowerTx) ValidateBasic() sdk.Error                { return nil }
func (tx testUpdatePowerTx) GetSigners() []crypto.Address            { return nil }
func (tx testUpdatePowerTx) GetFeePayer() crypto.Address             { return nil }
func (tx testUpdatePowerTx) GetSignatures() []sdk.StdSignature       { return nil }

func TestBasic(t *testing.T) {

	// Create app.
	app := NewBaseApp(t.Name())
	storeKeys := createMounts(app.cms)
	app.SetTxDecoder(func(txBytes []byte) (sdk.Tx, sdk.Error) {
		var ttx testUpdatePowerTx
		fromJSON(txBytes, &ttx)
		return ttx, nil
	})

	app.SetDefaultAnteHandler(func(ctx sdk.Context, tx sdk.Tx) (newCtx sdk.Context, res sdk.Result, abort bool) { return })
	app.Router().AddRoute(msgType, func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		// TODO
		return sdk.Result{}
	})

	// Load latest state, which should be empty.
	err := app.LoadLatestVersion(storeKeys["main"])
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

// Mounts stores to CommitMultiStore and returns a map of keys.
func createMounts(ms sdk.CommitMultiStore) map[string]sdk.StoreKey {
	dbMain := dbm.NewMemDB()
	dbXtra := dbm.NewMemDB()
	keyMain := sdk.NewKVStoreKey("main")
	keyXtra := sdk.NewKVStoreKey("xtra")
	ms.MountStoreWithDB(keyMain, sdk.StoreTypeIAVL, dbMain)
	ms.MountStoreWithDB(keyXtra, sdk.StoreTypeIAVL, dbXtra)
	return map[string]sdk.StoreKey{
		"main": keyMain,
		"xtra": keyXtra,
	}
}
