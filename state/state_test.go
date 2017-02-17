package state

import (
	"bytes"
	"testing"

	"github.com/tendermint/basecoin/types"
	eyes "github.com/tendermint/merkleeyes/client"

	"github.com/stretchr/testify/assert"
)

func TestState(t *testing.T) {

	s := NewState(types.NewMemKVStore())

	s.SetChainID("testchain")
	assert.True(t, s.GetChainID() == "testchain", "ChainID is improperly stored")

	setRecords := func(kv types.KVStore) {
		kv.Set([]byte("foo"), []byte("snake"))
		kv.Set([]byte("bar"), []byte("mouse"))
	}

	setRecords(s)
	assert.True(t, bytes.Equal(s.Get([]byte("foo")), []byte("snake")), "state doesn't retrieve after Set")
	assert.True(t, bytes.Equal(s.Get([]byte("bar")), []byte("mouse")), "state doesn't retrieve after Set")

	// Test account retrieve
	dumAddr := []byte("dummyAddress")

	acc := &types.Account{
		PubKey:   nil,
		Sequence: 1,
		Balance:  nil,
	}

	s.SetAccount(dumAddr, acc)
	assert.True(t, s.GetAccount(dumAddr).Sequence == 1, "GetAccount not retrieving")

	//Test CacheWrap with local mem store
	store := types.NewMemKVStore()
	s = NewState(store)
	cache := s.CacheWrap()
	setRecords(cache)
	assert.True(t, !bytes.Equal(store.Get([]byte("foo")), []byte("snake")), "store retrieving before Commit")
	assert.True(t, !bytes.Equal(store.Get([]byte("bar")), []byte("mouse")), "store retrieving before Commit")
	cache.CacheSync()
	assert.True(t, bytes.Equal(store.Get([]byte("foo")), []byte("snake")), "store doesn't retrieve after Commit")
	assert.True(t, bytes.Equal(store.Get([]byte("bar")), []byte("mouse")), "store doesn't retrieve after Commit")

	//Test Commit on state with non-merkle store
	assert.True(t, !s.Commit().IsOK(), "Commit shouldn't work with non-merkle store")

	//Test CacheWrap with merkleeyes client store
	eyesCli := eyes.NewLocalClient("", 0)
	s = NewState(eyesCli)

	cache = s.CacheWrap()
	setRecords(cache)
	assert.True(t, !bytes.Equal(eyesCli.Get([]byte("foo")), []byte("snake")), "store retrieving before Commit")
	assert.True(t, !bytes.Equal(eyesCli.Get([]byte("bar")), []byte("mouse")), "store retrieving before Commit")
	cache.CacheSync()
	assert.True(t, s.Commit().IsOK(), "Bad Commit")
	assert.True(t, bytes.Equal(eyesCli.Get([]byte("foo")), []byte("snake")), "store doesn't retrieve after Commit")
	assert.True(t, bytes.Equal(eyesCli.Get([]byte("bar")), []byte("mouse")), "store doesn't retrieve after Commit")

}
