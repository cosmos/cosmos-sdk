package store

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/iavl"
	dbm "github.com/tendermint/tendermint/libs/db"

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
	prefixPrefixStore := prefixStore.Prefix([]byte("prefix"))

	kvps := setRandomKVPairs(t, prefixPrefixStore)

	for i := 0; i < 20; i++ {
		key := kvps[i].key
		value := kvps[i].value
		require.True(t, prefixPrefixStore.Has(key))
		require.Equal(t, value, prefixPrefixStore.Get(key))

		key = append([]byte("prefix"), key...)
		require.True(t, prefixStore.Has(key))
		require.Equal(t, value, prefixStore.Get(key))
		key = append(prefix, key...)
		require.True(t, baseStore.Has(key))
		require.Equal(t, value, baseStore.Get(key))

		key = kvps[i].key
		prefixPrefixStore.Delete(key)
		require.False(t, prefixPrefixStore.Has(key))
		require.Nil(t, prefixPrefixStore.Get(key))
		key = append([]byte("prefix"), key...)
		require.False(t, prefixStore.Has(key))
		require.Nil(t, prefixStore.Get(key))
		key = append(prefix, key...)
		require.False(t, baseStore.Has(key))
		require.Nil(t, baseStore.Get(key))
	}
}

func TestIAVLStorePrefix(t *testing.T) {
	db := dbm.NewMemDB()
	tree := iavl.NewVersionedTree(db, cacheSize)
	iavlStore := newIAVLStore(tree, numRecent, storeEvery)

	testPrefixStore(t, iavlStore, []byte("test"))
}

func TestCacheKVStorePrefix(t *testing.T) {
	cacheStore := newCacheKVStore()

	testPrefixStore(t, cacheStore, []byte("test"))
}

func TestGasKVStorePrefix(t *testing.T) {
	meter := sdk.NewGasMeter(100000000)
	mem := dbStoreAdapter{dbm.NewMemDB()}
	gasStore := NewGasKVStore(meter, sdk.DefaultGasConfig(), mem)

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
		require.Equal(t, bIter.Key(), append(prefix, pIter.Key()...))
		require.Equal(t, bIter.Value(), pIter.Value())

		bIter.Next()
		pIter.Next()
	}

	require.Equal(t, bIter.Valid(), pIter.Valid())
	bIter.Close()
	pIter.Close()
}
