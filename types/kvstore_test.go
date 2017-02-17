package types

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemKVStore(t *testing.T) {

	ms := NewMemKVStore()
	ms.Set([]byte("foo"), []byte("snake"))
	ms.Set([]byte("bar"), []byte("mouse"))
	assert.True(t, bytes.Equal(ms.Get([]byte("foo")), []byte("snake")), "MemKVStore doesn't retrieve after Set")
	assert.True(t, bytes.Equal(ms.Get([]byte("bar")), []byte("mouse")), "MemKVStore doesn't retrieve after Set")
}

func TestKVCache(t *testing.T) {

	store := NewMemKVStore()
	kvc := NewKVCache(store)

	setRecords := func() {
		kvc.Set([]byte("foo"), []byte("snake"))
		kvc.Set([]byte("bar"), []byte("mouse"))
	}

	//test read/write
	setRecords()
	assert.True(t, bytes.Equal(kvc.Get([]byte("foo")), []byte("snake")), "KVCache doesn't retrieve after Set")
	assert.True(t, bytes.Equal(kvc.Get([]byte("bar")), []byte("mouse")), "KVCache doesn't retrieve after Set")

	//test reset
	kvc.Reset()
	assert.True(t, !bytes.Equal(kvc.Get([]byte("foo")), []byte("snake")), "KVCache retrieving after reset")
	assert.True(t, !bytes.Equal(kvc.Get([]byte("bar")), []byte("mouse")), "KVCache retrieving after reset")

	//test sync
	setRecords()
	assert.True(t, !bytes.Equal(store.Get([]byte("foo")), []byte("snake")), "store retrieving before synced")
	assert.True(t, !bytes.Equal(store.Get([]byte("bar")), []byte("mouse")), "store retrieving before synced")
	kvc.Sync()
	assert.True(t, bytes.Equal(store.Get([]byte("foo")), []byte("snake")), "store isn't retrieving after synced")
	assert.True(t, bytes.Equal(store.Get([]byte("bar")), []byte("mouse")), "store isn't retrieving after synced")

	//test logging
	assert.True(t, len(kvc.GetLogLines()) == 0, "logging events existed before using SetLogging")
	fmt.Println(len(kvc.GetLogLines()))

	kvc.SetLogging()
	setRecords()
	assert.True(t, len(kvc.GetLogLines()) == 2, "incorrect number of logging events recorded")

	kvc.ClearLogLines()
	assert.True(t, len(kvc.GetLogLines()) == 0, "logging events still exists after ClearLogLines")

}
