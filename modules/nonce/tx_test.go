package nonce

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
)

func TestNonce(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// generic args here...
	chainID := "my-chain"
	height := uint64(100)
	// rigel: use MockContext, so we can add permissions
	ctx := stack.MockContext(chainID, height)
	store := state.NewMemKVStore()

	// rigel: you can leave chainID blank for the actors, note the comment on Actor struct
	act1 := basecoin.Actor{ChainID: chainID, App: "fooz", Address: []byte{1, 2, 3, 4}}
	act2 := basecoin.Actor{ChainID: chainID, App: "fooz", Address: []byte{1, 1, 1, 1}}
	act3 := basecoin.Actor{ChainID: chainID, App: "fooz", Address: []byte{3, 3, 3, 3}}

	// let's construct some tests to make the table a bit less verbose
	set0 := []basecoin.Actor{}
	set1 := []basecoin.Actor{act1}
	set2 := []basecoin.Actor{act2}
	set12 := []basecoin.Actor{act1, act2}
	set21 := []basecoin.Actor{act2, act1}
	set123 := []basecoin.Actor{act1, act2, act3}
	set321 := []basecoin.Actor{act3, act2, act1}

	// rigel: test cases look good, but also add reordering
	testList := []struct {
		valid  bool
		seq    uint32
		actors []basecoin.Actor
		// rigel: you forgot to sign the tx, of course the get rejected...
		signers []basecoin.Actor
	}{
		// one signer
		{false, 0, set1, set1},   // seq 0 is no good
		{false, 1, set1, set0},   // sig is required
		{true, 1, set1, set1},    // sig and seq are good
		{true, 2, set1, set1},    // increments each time
		{false, 777, set1, set1}, // seq is too high

		// independent from second signer
		{false, 1, set2, set1},  // sig must match
		{true, 1, set2, set2},   // seq of set2 independent from set1
		{true, 2, set2, set321}, // extra sigs don't change the situation

		// multisig has same requirements
		{false, 0, set12, set12}, // need valid sequence number
		{false, 1, set12, set2},  // they all must sign
		{true, 1, set12, set12},  // this is proper, independent of act1 and act2
		{true, 2, set21, set21},  // order of actors doesn't matter
		{false, 2, set12, set12}, // but can't repeat sequence
		{true, 3, set12, set321}, // no effect from extra sigs

		// tripple sigs also work
		{false, 2, set123, set123}, // must start with seq=1
		{false, 1, set123, set12},  // all must sign
		{true, 1, set123, set321},  // this works
		{true, 2, set321, set321},  // other order is the same
		{false, 2, set321, set321}, // no repetition
	}

	// rigel: don't wrap nil, it's bad, wrap a raw byte thing
	raw := stack.NewRawTx([]byte{42})
	for i, test := range testList {
		// rigel: set the permissions
		myCtx := ctx.WithPermissions(test.signers...)

		tx := NewTx(test.seq, test.actors, raw)
		nonceTx, ok := tx.Unwrap().(Tx)
		require.True(ok)

		err := nonceTx.CheckIncrementSeq(myCtx, store)
		if test.valid {
			assert.Nil(err, "%d: %+v", i, err)
		} else {
			assert.NotNil(err, "%d", i)
		}
	}
}
