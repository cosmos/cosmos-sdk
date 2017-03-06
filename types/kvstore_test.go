package types

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKVStore(t *testing.T) {

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

	//define the test list
	var testList = []struct {
		testPass func() bool
		errMsg   string
	}{
		//test read/write for MemKVStore
		{func() bool { setRecords(ms); return storeHasAll(ms) },
			"MemKVStore doesn't retrieve after Set"},

		//test read/write for KVCache
		{func() bool { setRecords(kvc); return storeHasAll(kvc) },
			"KVCache doesn't retrieve after Set"},

		//test reset
		{func() bool { kvc.Reset(); return !storeHasAll(kvc) },
			"KVCache retrieving after reset"},

		//test sync
		{func() bool { setRecords(kvc); return !storeHasAll(store) },
			"store retrieving before synced"},
		{func() bool { kvc.Sync(); return storeHasAll(store) },
			"store isn't retrieving after synced"},

		//test logging
		{func() bool { return len(kvc.GetLogLines()) == 0 },
			"logging events existed before using SetLogging"},
		{func() bool { kvc.SetLogging(); setRecords(kvc); return len(kvc.GetLogLines()) == 2 },
			"incorrect number of logging events recorded"},
		{func() bool { kvc.ClearLogLines(); return len(kvc.GetLogLines()) == 0 },
			"logging events still exists after ClearLogLines"},
	}

	//execute the tests
	for _, tl := range testList {
		assert.True(t, tl.testPass(), tl.errMsg)
	}
}
