package prefix

import (
	"crypto/rand"
	"testing"

	"github.com/cosmos/cosmos-sdk/store/dbadapter"
	"github.com/cosmos/cosmos-sdk/store/gaskv"
	"github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"

	tiavl "github.com/tendermint/iavl"
	dbm "github.com/tendermint/tendermint/libs/db"
)

// copied from iavl/store_test.go
var (
	cacheSize        = 100
	numRecent  int64 = 5
	storeEvery int64 = 3
)

func bz(s string) []byte { return []byte(s) }

type kvpair struct {
	key   []byte
	value []byte
}

func genRandomKVPairs(t *testing.T) []kvpair {
	kvps := make([]kvpair, 20)

	for i := 0; i < 20; i++ {
		kvps[i].key = make([]byte, 32)
		rand.Read(kvps[i].key)
		kvps[i].value = make([]byte, 32)
		rand.Read(kvps[i].value)
	}

	return kvps
}

func setRandomKVPairs(t *testing.T, store types.KVStore) []kvpair {
	kvps := genRandomKVPairs(t)
	for _, kvp := range kvps {
		store.Set(kvp.key, kvp.value)
	}
	return kvps
}

func testPrefixStore(t *testing.T, baseStore types.KVStore, prefix []byte) {
	prefixStore := NewStore(baseStore, prefix)
	prefixPrefixStore := NewStore(prefixStore, []byte("prefix"))

	require.Panics(t, func() { prefixStore.Get(nil) })
	require.Panics(t, func() { prefixStore.Set(nil, []byte{}) })

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
	tree := tiavl.NewMutableTree(db, cacheSize)
	iavlStore := iavl.UnsafeNewStore(tree, numRecent, storeEvery)

	testPrefixStore(t, iavlStore, []byte("test"))
}

func TestPrefixKVStoreNoNilSet(t *testing.T) {
	meter := types.NewGasMeter(100000000)
	mem := dbadapter.Store{dbm.NewMemDB()}
	gasStore := gaskv.NewStore(mem, meter, types.KVGasConfig())
	require.Panics(t, func() { gasStore.Set([]byte("key"), nil) }, "setting a nil value should panic")
}

func TestPrefixStoreIterate(t *testing.T) {
	db := dbm.NewMemDB()
	baseStore := dbadapter.Store{db}
	prefix := []byte("test")
	prefixStore := NewStore(baseStore, prefix)

	setRandomKVPairs(t, prefixStore)

	bIter := types.KVStorePrefixIterator(baseStore, prefix)
	pIter := types.KVStorePrefixIterator(prefixStore, nil)

	for bIter.Valid() && pIter.Valid() {
		require.Equal(t, bIter.Key(), append(prefix, pIter.Key()...))
		require.Equal(t, bIter.Value(), pIter.Value())

		bIter.Next()
		pIter.Next()
	}

	bIter.Close()
	pIter.Close()
}

func incFirstByte(bz []byte) {
	bz[0]++
}

func TestCloneAppend(t *testing.T) {
	kvps := genRandomKVPairs(t)
	for _, kvp := range kvps {
		bz := cloneAppend(kvp.key, kvp.value)
		require.Equal(t, bz, append(kvp.key, kvp.value...))

		incFirstByte(bz)
		require.NotEqual(t, bz, append(kvp.key, kvp.value...))

		bz = cloneAppend(kvp.key, kvp.value)
		incFirstByte(kvp.key)
		require.NotEqual(t, bz, append(kvp.key, kvp.value...))

		bz = cloneAppend(kvp.key, kvp.value)
		incFirstByte(kvp.value)
		require.NotEqual(t, bz, append(kvp.key, kvp.value...))
	}
}

func TestPrefixStoreIteratorEdgeCase(t *testing.T) {
	db := dbm.NewMemDB()
	baseStore := dbadapter.Store{db}

	// overflow in cpIncr
	prefix := []byte{0xAA, 0xFF, 0xFF}
	prefixStore := NewStore(baseStore, prefix)

	// ascending order
	baseStore.Set([]byte{0xAA, 0xFF, 0xFE}, []byte{})
	baseStore.Set([]byte{0xAA, 0xFF, 0xFE, 0x00}, []byte{})
	baseStore.Set([]byte{0xAA, 0xFF, 0xFF}, []byte{})
	baseStore.Set([]byte{0xAA, 0xFF, 0xFF, 0x00}, []byte{})
	baseStore.Set([]byte{0xAB}, []byte{})
	baseStore.Set([]byte{0xAB, 0x00}, []byte{})
	baseStore.Set([]byte{0xAB, 0x00, 0x00}, []byte{})

	iter := prefixStore.Iterator(nil, nil)

	checkDomain(t, iter, nil, nil)
	checkItem(t, iter, []byte{}, bz(""))
	checkNext(t, iter, true)
	checkItem(t, iter, []byte{0x00}, bz(""))
	checkNext(t, iter, false)

	checkInvalid(t, iter)

	iter.Close()
}

func TestPrefixStoreReverseIteratorEdgeCase(t *testing.T) {
	db := dbm.NewMemDB()
	baseStore := dbadapter.Store{db}

	// overflow in cpIncr
	prefix := []byte{0xAA, 0xFF, 0xFF}
	prefixStore := NewStore(baseStore, prefix)

	// descending order
	baseStore.Set([]byte{0xAB, 0x00, 0x00}, []byte{})
	baseStore.Set([]byte{0xAB, 0x00}, []byte{})
	baseStore.Set([]byte{0xAB}, []byte{})
	baseStore.Set([]byte{0xAA, 0xFF, 0xFF, 0x00}, []byte{})
	baseStore.Set([]byte{0xAA, 0xFF, 0xFF}, []byte{})
	baseStore.Set([]byte{0xAA, 0xFF, 0xFE, 0x00}, []byte{})
	baseStore.Set([]byte{0xAA, 0xFF, 0xFE}, []byte{})

	iter := prefixStore.ReverseIterator(nil, nil)

	checkDomain(t, iter, nil, nil)
	checkItem(t, iter, []byte{0x00}, bz(""))
	checkNext(t, iter, true)
	checkItem(t, iter, []byte{}, bz(""))
	checkNext(t, iter, false)

	checkInvalid(t, iter)

	iter.Close()

	db = dbm.NewMemDB()
	baseStore = dbadapter.Store{db}

	// underflow in cpDecr
	prefix = []byte{0xAA, 0x00, 0x00}
	prefixStore = NewStore(baseStore, prefix)

	baseStore.Set([]byte{0xAB, 0x00, 0x01, 0x00, 0x00}, []byte{})
	baseStore.Set([]byte{0xAB, 0x00, 0x01, 0x00}, []byte{})
	baseStore.Set([]byte{0xAB, 0x00, 0x01}, []byte{})
	baseStore.Set([]byte{0xAA, 0x00, 0x00, 0x00}, []byte{})
	baseStore.Set([]byte{0xAA, 0x00, 0x00}, []byte{})
	baseStore.Set([]byte{0xA9, 0xFF, 0xFF, 0x00}, []byte{})
	baseStore.Set([]byte{0xA9, 0xFF, 0xFF}, []byte{})

	iter = prefixStore.ReverseIterator(nil, nil)

	checkDomain(t, iter, nil, nil)
	checkItem(t, iter, []byte{0x00}, bz(""))
	checkNext(t, iter, true)
	checkItem(t, iter, []byte{}, bz(""))
	checkNext(t, iter, false)

	checkInvalid(t, iter)

	iter.Close()
}

// Tests below are ported from https://github.com/tendermint/tendermint/blob/master/libs/db/prefix_db_test.go

func mockStoreWithStuff() types.KVStore {
	db := dbm.NewMemDB()
	store := dbadapter.Store{db}
	// Under "key" prefix
	store.Set(bz("key"), bz("value"))
	store.Set(bz("key1"), bz("value1"))
	store.Set(bz("key2"), bz("value2"))
	store.Set(bz("key3"), bz("value3"))
	store.Set(bz("something"), bz("else"))
	store.Set(bz(""), bz(""))
	store.Set(bz("k"), bz(sdk.PrefixValidator))
	store.Set(bz("ke"), bz("valu"))
	store.Set(bz("kee"), bz("valuu"))
	return store
}

func checkValue(t *testing.T, store types.KVStore, key []byte, expected []byte) {
	bz := store.Get(key)
	require.Equal(t, expected, bz)
}

func checkValid(t *testing.T, itr types.Iterator, expected bool) {
	valid := itr.Valid()
	require.Equal(t, expected, valid)
}

func checkNext(t *testing.T, itr types.Iterator, expected bool) {
	itr.Next()
	valid := itr.Valid()
	require.Equal(t, expected, valid)
}

func checkDomain(t *testing.T, itr types.Iterator, start, end []byte) {
	ds, de := itr.Domain()
	require.Equal(t, start, ds)
	require.Equal(t, end, de)
}

func checkItem(t *testing.T, itr types.Iterator, key, value []byte) {
	require.Exactly(t, key, itr.Key())
	require.Exactly(t, value, itr.Value())
}

func checkInvalid(t *testing.T, itr types.Iterator) {
	checkValid(t, itr, false)
	checkKeyPanics(t, itr)
	checkValuePanics(t, itr)
	checkNextPanics(t, itr)
}

func checkKeyPanics(t *testing.T, itr types.Iterator) {
	require.Panics(t, func() { itr.Key() })
}

func checkValuePanics(t *testing.T, itr types.Iterator) {
	require.Panics(t, func() { itr.Value() })
}

func checkNextPanics(t *testing.T, itr types.Iterator) {
	require.Panics(t, func() { itr.Next() })
}

func TestPrefixDBSimple(t *testing.T) {
	store := mockStoreWithStuff()
	pstore := NewStore(store, bz("key"))

	checkValue(t, pstore, bz("key"), nil)
	checkValue(t, pstore, bz(""), bz("value"))
	checkValue(t, pstore, bz("key1"), nil)
	checkValue(t, pstore, bz("1"), bz("value1"))
	checkValue(t, pstore, bz("key2"), nil)
	checkValue(t, pstore, bz("2"), bz("value2"))
	checkValue(t, pstore, bz("key3"), nil)
	checkValue(t, pstore, bz("3"), bz("value3"))
	checkValue(t, pstore, bz("something"), nil)
	checkValue(t, pstore, bz("k"), nil)
	checkValue(t, pstore, bz("ke"), nil)
	checkValue(t, pstore, bz("kee"), nil)
}

func TestPrefixDBIterator1(t *testing.T) {
	store := mockStoreWithStuff()
	pstore := NewStore(store, bz("key"))

	itr := pstore.Iterator(nil, nil)
	checkDomain(t, itr, nil, nil)
	checkItem(t, itr, bz(""), bz("value"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("1"), bz("value1"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("2"), bz("value2"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("3"), bz("value3"))
	checkNext(t, itr, false)
	checkInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBIterator2(t *testing.T) {
	store := mockStoreWithStuff()
	pstore := NewStore(store, bz("key"))

	itr := pstore.Iterator(nil, bz(""))
	checkDomain(t, itr, nil, bz(""))
	checkInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBIterator3(t *testing.T) {
	store := mockStoreWithStuff()
	pstore := NewStore(store, bz("key"))

	itr := pstore.Iterator(bz(""), nil)
	checkDomain(t, itr, bz(""), nil)
	checkItem(t, itr, bz(""), bz("value"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("1"), bz("value1"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("2"), bz("value2"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("3"), bz("value3"))
	checkNext(t, itr, false)
	checkInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBIterator4(t *testing.T) {
	store := mockStoreWithStuff()
	pstore := NewStore(store, bz("key"))

	itr := pstore.Iterator(bz(""), bz(""))
	checkDomain(t, itr, bz(""), bz(""))
	checkInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBReverseIterator1(t *testing.T) {
	store := mockStoreWithStuff()
	pstore := NewStore(store, bz("key"))

	itr := pstore.ReverseIterator(nil, nil)
	checkDomain(t, itr, nil, nil)
	checkItem(t, itr, bz("3"), bz("value3"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("2"), bz("value2"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("1"), bz("value1"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz(""), bz("value"))
	checkNext(t, itr, false)
	checkInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBReverseIterator2(t *testing.T) {
	store := mockStoreWithStuff()
	pstore := NewStore(store, bz("key"))

	itr := pstore.ReverseIterator(bz(""), nil)
	checkDomain(t, itr, bz(""), nil)
	checkItem(t, itr, bz("3"), bz("value3"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("2"), bz("value2"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("1"), bz("value1"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz(""), bz("value"))
	checkNext(t, itr, false)
	checkInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBReverseIterator3(t *testing.T) {
	store := mockStoreWithStuff()
	pstore := NewStore(store, bz("key"))

	itr := pstore.ReverseIterator(nil, bz(""))
	checkDomain(t, itr, nil, bz(""))
	checkInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBReverseIterator4(t *testing.T) {
	store := mockStoreWithStuff()
	pstore := NewStore(store, bz("key"))

	itr := pstore.ReverseIterator(bz(""), bz(""))
	checkInvalid(t, itr)
	itr.Close()
}
