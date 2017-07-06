package stack

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/txs"
	"github.com/tendermint/basecoin/types"
)

func TestSignatureChecks(t *testing.T) {
	assert := assert.New(t)

	// generic args
	ctx := NewContext("test-chain", log.NewNopLogger())
	store := types.NewMemKVStore()
	raw := txs.NewRaw([]byte{1, 2, 3, 4})

	// let's make some keys....
	priv1 := crypto.GenPrivKeyEd25519().Wrap()
	actor1 := SigPerm(priv1.PubKey().Address())
	priv2 := crypto.GenPrivKeySecp256k1().Wrap()
	actor2 := SigPerm(priv2.PubKey().Address())

	// test cases to make sure signature checks are solid
	cases := []struct {
		useMultiSig bool
		keys        []crypto.PrivKey
		check       basecoin.Actor
		valid       bool
	}{
		// test with single sigs
		{false, []crypto.PrivKey{priv1}, actor1, true},
		{false, []crypto.PrivKey{priv1}, actor2, false},
		{false, []crypto.PrivKey{priv2}, actor2, true},
		{false, []crypto.PrivKey{}, actor2, false},

		// same with multi sigs
		{true, []crypto.PrivKey{priv1}, actor1, true},
		{true, []crypto.PrivKey{priv1}, actor2, false},
		{true, []crypto.PrivKey{priv2}, actor2, true},
		{true, []crypto.PrivKey{}, actor2, false},

		// make sure both match on a multisig
		{true, []crypto.PrivKey{priv1, priv2}, actor1, true},
		{true, []crypto.PrivKey{priv1, priv2}, actor2, true},
	}

	for i, tc := range cases {
		idx := strconv.Itoa(i)

		// make the stack check for the given permission
		app := New(
			Recovery{}, // we need this so panics turn to errors
			Signatures{},
			CheckMiddleware{Required: tc.check},
		).Use(OKHandler{})

		var tx basecoin.Tx
		// this does the signing as needed
		if tc.useMultiSig {
			mtx := txs.NewMulti(raw)
			for _, k := range tc.keys {
				err := txs.Sign(mtx, k)
				assert.Nil(err, "%d: %+v", i, err)
			}
			tx = mtx.Wrap()
		} else {
			otx := txs.NewSig(raw)
			for _, k := range tc.keys {
				err := txs.Sign(otx, k)
				assert.Nil(err, "%d: %+v", i, err)
			}
			tx = otx.Wrap()
		}

		// this will trivial expose the printing error...
		// _, err := app.CheckTx(ctx, store, raw)
		_, err := app.CheckTx(ctx, store, tx)
		if tc.valid {
			// TODO: why doen't tmerror print properly???
			assert.Nil(err, "%d: %+v", i, err)
		} else {
			assert.NotNil(err, idx)
		}

		_, err = app.DeliverTx(ctx, store, tx)
		if tc.valid {
			assert.Nil(err, "%d: %+v", i, err)
		} else {
			assert.NotNil(err, idx)
		}
	}
}
