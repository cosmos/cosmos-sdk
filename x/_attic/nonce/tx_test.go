package nonce

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/cosmos-sdk/state"
)

func TestNonce(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// generic args here...
	chainID := "my-chain"
	chain2ID := "woohoo"

	height := uint64(100)
	ctx := stack.MockContext(chainID, height)
	store := state.NewMemKVStore()

	appName1 := "fooz"
	appName2 := "foot"

	//root actors for the tests
	act1 := sdk.Actor{ChainID: chainID, App: appName1, Address: []byte{1, 2, 3, 4}}
	act2 := sdk.Actor{ChainID: chainID, App: appName1, Address: []byte{1, 1, 1, 1}}
	act3 := sdk.Actor{ChainID: chainID, App: appName1, Address: []byte{3, 3, 3, 3}}
	act1DiffChain := sdk.Actor{ChainID: chain2ID, App: appName1, Address: []byte{1, 2, 3, 4}}
	act2DiffChain := sdk.Actor{ChainID: chain2ID, App: appName1, Address: []byte{1, 1, 1, 1}}
	act3DiffChain := sdk.Actor{ChainID: chain2ID, App: appName1, Address: []byte{3, 3, 3, 3}}
	act1DiffApp := sdk.Actor{ChainID: chainID, App: appName2, Address: []byte{1, 2, 3, 4}}
	act2DiffApp := sdk.Actor{ChainID: chainID, App: appName2, Address: []byte{1, 1, 1, 1}}
	act3DiffApp := sdk.Actor{ChainID: chainID, App: appName2, Address: []byte{3, 3, 3, 3}}

	// let's construct some tests to make the table a bit less verbose
	set0 := []sdk.Actor{}
	set1 := []sdk.Actor{act1}
	set2 := []sdk.Actor{act2}
	set12 := []sdk.Actor{act1, act2}
	set21 := []sdk.Actor{act2, act1}
	set123 := []sdk.Actor{act1, act2, act3}
	set321 := []sdk.Actor{act3, act2, act1}

	//some more test cases for different chains and apps for each actor
	set123Chain2 := []sdk.Actor{act1DiffChain, act2DiffChain, act3DiffChain}
	set123App2 := []sdk.Actor{act1DiffApp, act2DiffApp, act3DiffApp}
	set123MixedChains := []sdk.Actor{act1, act2DiffChain, act3}
	set123MixedApps := []sdk.Actor{act1, act2DiffApp, act3}

	testList := []struct {
		valid   bool
		seq     uint32
		actors  []sdk.Actor
		signers []sdk.Actor
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

		// triple sigs also work
		{false, 2, set123, set123}, // must start with seq=1
		{false, 1, set123, set12},  // all must sign
		{true, 1, set123, set321},  // this works
		{true, 2, set321, set321},  // other order is the same
		{false, 2, set321, set321}, // no repetition

		// signers with different chain-IDs and apps from actors
		{false, 3, set123, set123Chain2},      // sign with different chain actors
		{false, 3, set123, set123App2},        // sign with different app actors
		{false, 3, set123, set123MixedChains}, // sign with mixed chain actor
		{false, 3, set123, set123MixedApps},   // sign with mixed app actors

		// signers from different chain-IDs and apps, working
		{true, 1, set123Chain2, set123Chain2},
		{true, 1, set123App2, set123App2},
		{true, 1, set123MixedChains, set123MixedChains},
		{true, 1, set123MixedApps, set123MixedApps},
	}

	raw := stack.NewRawTx([]byte{42})
	for i, test := range testList {

		// set the permissions
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
