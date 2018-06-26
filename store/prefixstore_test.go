package store

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tendermint/iavl"
	dbm "github.com/tendermint/tmlibs/db"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type kvpair struct {
	key   []byte
	value []byte
}

func setRandomKVPairs(t *testing.T, store KVStore) []kvpair {
	kvps := make([]kvpair, 20)

	for i := 0; i < 20; i++ {
		kvps[i].key = make([]byte, 32)
		rand.Read(kvps[i].key)
		kvps[i].value = make([]byte, 32)
		rand.Read(kvps[i].value)

		store.Set(kvps[i].key, kvps[i].value)
	}

	return kvps
}

func testPrefixStore(t *testing.T, baseStore KVStore, prefix []byte) {
	prefixStore := baseStore.Prefix(prefix)

	kvps := setRandomKVPairs(t, prefixStore)

	buf := make([]byte, 32)
	for i := 0; i < 20; i++ {
		rand.Read(buf)
		assert.False(t, prefixStore.Has(buf))
		assert.Nil(t, prefixStore.Get(buf))
		assert.False(t, baseStore.Has(append(prefix, buf...)))
		assert.Nil(t, baseStore.Get(append(prefix, buf...)))
	}

	for i := 0; i < 20; i++ {
		key := kvps[i].key
		assert.True(t, prefixStore.Has(key))
		assert.Equal(t, kvps[i].value, prefixStore.Get(key))
		assert.True(t, baseStore.Has(append(prefix, key...)))
		assert.Equal(t, kvps[i].value, baseStore.Get(append(prefix, key...)))

		prefixStore.Delete(key)
		assert.False(t, prefixStore.Has(key))
		assert.Nil(t, prefixStore.Get(key))
		assert.False(t, baseStore.Has(append(prefix, buf...)))
		assert.Nil(t, baseStore.Get(append(prefix, buf...)))
	}

}

func TestIAVLStorePrefix(t *testing.T) {
	db := dbm.NewMemDB()
	tree := iavl.NewVersionedTree(db, cacheSize)
	iavlStore := newIAVLStore(tree, numHistory)

	testPrefixStore(t, iavlStore, []byte("test"))
}

func TestCacheKVStorePrefix(t *testing.T) {
	cacheStore := newCacheKVStore()

	testPrefixStore(t, cacheStore, []byte("test"))
}

func TestGasKVStorePrefix(t *testing.T) {
	meter := sdk.NewGasMeter(100000000)
	mem := dbStoreAdapter{dbm.NewMemDB()}
	gasStore := NewGasKVStore(meter, mem)

	testPrefixStore(t, gasStore, []byte("test"))
}

func TestPrefixStoreIterate(t *testing.T) {
	db := dbm.NewMemDB()
	baseStore := dbStoreAdapter{db}
	prefix := []byte("test")
	prefixStore := baseStore.Prefix(prefix)

	setRandomKVPairs(t, prefixStore)

	bIter := sdk.KVStorePrefixIterator(baseStore, prefix)
	pIter := sdk.KVStorePrefixIterator(prefixStore, nil)

	for bIter.Valid() && pIter.Valid() {
		assert.Equal(t, bIter.Key(), append(prefix, pIter.Key()...))
		assert.Equal(t, bIter.Value(), pIter.Value())

		bIter.Next()
		pIter.Next()
	}

	assert.Equal(t, bIter.Valid(), pIter.Valid())
	bIter.Close()
	pIter.Close()
}
