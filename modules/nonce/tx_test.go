package nonce

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"

	"github.com/tendermint/tmlibs/log"
)

func TestNonce(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// generic args here...
	chainID := "my-chain"
	height := uint64(100)
	ctx := stack.NewContext(chainID, height, log.NewNopLogger())
	store := state.NewMemKVStore()

	act1 := basecoin.Actor{ChainID: chainID, App: "fooz", Address: []byte{1, 2, 3, 4}}
	act2 := basecoin.Actor{ChainID: chainID, App: "fooz", Address: []byte{1, 1, 1, 1}}
	act3 := basecoin.Actor{ChainID: chainID, App: "fooz", Address: []byte{3, 3, 3, 3}}

	testList := []struct {
		valid  bool
		seq    uint32
		actors []basecoin.Actor
	}{
		{false, 0, []basecoin.Actor{act1}},
		{true, 1, []basecoin.Actor{act1}},
		{false, 777, []basecoin.Actor{act1}},
		{true, 2, []basecoin.Actor{act1}},
		{false, 0, []basecoin.Actor{act1, act2}},
		{true, 1, []basecoin.Actor{act1, act2}},
		{true, 2, []basecoin.Actor{act1, act2}},
		{true, 3, []basecoin.Actor{act1, act2}},
		{false, 2, []basecoin.Actor{act1, act2}},
		{false, 2, []basecoin.Actor{act1, act2, act3}},
		{true, 1, []basecoin.Actor{act1, act2, act3}},
	}

	for _, test := range testList {

		tx := NewTx(test.seq, test.actors, basecoin.Tx{})
		nonceTx, ok := tx.Unwrap().(Tx)
		require.True(ok)
		err := nonceTx.CheckIncrementSeq(ctx, store)
		if test.valid {
			assert.Nil(err, "%v,%v: %+v", test.seq, test.actors, err)
		} else {
			assert.NotNil(err, "%v,%v", test.seq, test.actors)
		}
	}
}
