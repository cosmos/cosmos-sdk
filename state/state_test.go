package state

import (
	"bytes"
	"testing"

	"github.com/tendermint/basecoin/types"
	eyes "github.com/tendermint/merkleeyes/client"

	"github.com/stretchr/testify/assert"
)

func TestState(t *testing.T) {

	//States and Stores for tests
	store := types.NewMemKVStore()
	state := NewState(store)
	cache := state.CacheWrap()
	eyesCli := eyes.NewLocalClient("", 0)

	//Account and address for tests
	dumAddr := []byte("dummyAddress")

	acc := &types.Account{
		PubKey:   nil,
		Sequence: 1,
		Balance:  nil,
	}

	//reset the store/state/cache
	reset := func() {
		store = types.NewMemKVStore()
		state = NewState(store)
		cache = state.CacheWrap()
	}

	//set the state to using the eyesCli instead of MemKVStore
	useEyesCli := func() {
		state = NewState(eyesCli)
		cache = state.CacheWrap()
	}

	//key value pairs to be tested within the system
	keyvalue := []struct {
		key   string
		value string
	}{
		{"foo", "snake"},
		{"bar", "mouse"},
	}

	//set the kvc to have all the key value pairs
	setRecords := func(kv types.KVStore) {
		for _, n := range keyvalue {
			kv.Set([]byte(n.key), []byte(n.value))
		}
	}

	//store has all the key value pairs
	storeHasAll := func(kv types.KVStore) bool {
		for _, n := range keyvalue {
			if !bytes.Equal(kv.Get([]byte(n.key)), []byte(n.value)) {
				return false
			}
		}
		return true
	}

	//define the test list
	testList := []struct {
		testPass func() bool
		errMsg   string
	}{
		//test chainID
		{func() bool { state.SetChainID("testchain"); return state.GetChainID() == "testchain" },
			"ChainID is improperly stored"},

		//test basic retrieve
		{func() bool { setRecords(state); return storeHasAll(state) },
			"state doesn't retrieve after Set"},

		// Test account retrieve
		{func() bool { state.SetAccount(dumAddr, acc); return state.GetAccount(dumAddr).Sequence == 1 },
			"GetAccount not retrieving"},

		//Test CacheWrap with local mem store
		{func() bool { reset(); setRecords(cache); return !storeHasAll(store) },
			"store retrieving before CacheSync"},
		{func() bool { cache.CacheSync(); return storeHasAll(store) },
			"store doesn't retrieve after CacheSync"},

		//Test Commit on state with non-merkle store
		{func() bool { return !state.Commit().IsOK() },
			"Commit shouldn't work with non-merkle store"},

		//Test CacheWrap with merkleeyes client store
		{func() bool { useEyesCli(); setRecords(cache); return !storeHasAll(eyesCli) },
			"eyesCli retrieving before Commit"},
		{func() bool { cache.CacheSync(); return state.Commit().IsOK() },
			"Bad Commit"},
		{func() bool { return storeHasAll(eyesCli) },
			"eyesCli doesn't retrieve after Commit"},
	}

	//execute the tests
	for _, tl := range testList {
		assert.True(t, tl.testPass(), tl.errMsg)
	}
}
