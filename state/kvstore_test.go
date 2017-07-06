package state

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKVStore(t *testing.T) {
	assert := assert.New(t)

	//stores to be tested
	ms := NewMemKVStore()
	store := NewMemKVStore()
	kvc := NewKVCache(store)

	//key value pairs to be tested within the system
	var keyvalue = []struct {
		key   string
		value string
	}{
		{"foo", "snake"},
		{"bar", "mouse"},
	}

	//set the kvc to have all the key value pairs
	setRecords := func(kv KVStore) {
		for _, n := range keyvalue {
			kv.Set([]byte(n.key), []byte(n.value))
		}
	}

	//store has all the key value pairs
	storeHasAll := func(kv KVStore) bool {
		for _, n := range keyvalue {
			if !bytes.Equal(kv.Get([]byte(n.key)), []byte(n.value)) {
				return false
			}
		}
		return true
	}

	//test read/write for MemKVStore
	setRecords(ms)
	assert.True(storeHasAll(ms), "MemKVStore doesn't retrieve after Set")

	//test read/write for KVCache
	setRecords(kvc)
	assert.True(storeHasAll(kvc), "KVCache doesn't retrieve after Set")

	//test reset
	kvc.Reset()
	assert.False(storeHasAll(kvc), "KVCache retrieving after reset")

	//test sync
	setRecords(kvc)
	assert.False(storeHasAll(store), "store retrieving before synced")
	kvc.Sync()
	assert.True(storeHasAll(store), "store isn't retrieving after synced")

	//test logging
	assert.Zero(len(kvc.GetLogLines()), "logging events existed before using SetLogging")
	kvc.SetLogging()
	setRecords(kvc)
	assert.Equal(len(kvc.GetLogLines()), 2, "incorrect number of logging events recorded")
	kvc.ClearLogLines()
	assert.Zero(len(kvc.GetLogLines()), "logging events still exists after ClearLogLines")

}
