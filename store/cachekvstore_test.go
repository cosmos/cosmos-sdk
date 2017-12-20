package store

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
)

func keyFmt(i int) []byte { return bz(cmn.Fmt("key%d", i)) }
func valFmt(i int) []byte { return bz(cmn.Fmt("value%d", i)) }

func TestCacheKVStore(t *testing.T) {
	mem := dbm.NewMemDB()
	st := NewCacheKVStore(mem)

	require.Empty(t, st.Get(keyFmt(1)), "Expected `key1` to be empty")

	// put something in mem and in cache
	mem.Set(keyFmt(1), valFmt(1))
	st.Set(keyFmt(1), valFmt(1))
	require.Equal(t, valFmt(1), st.Get(keyFmt(1)))

	// update it in cache, shoudn't change mem
	st.Set(keyFmt(1), valFmt(2))
	require.Equal(t, valFmt(2), st.Get(keyFmt(1)))
	require.Equal(t, valFmt(1), mem.Get(keyFmt(1)))

	// write it. should change mem
	st.Write()
	require.Equal(t, valFmt(2), mem.Get(keyFmt(1)))
	require.Equal(t, valFmt(2), st.Get(keyFmt(1)))

	// more writes and checks
	st.Write()
	st.Write()
	require.Equal(t, valFmt(2), mem.Get(keyFmt(1)))
	require.Equal(t, valFmt(2), st.Get(keyFmt(1)))

	// make a new one, check it
	st = NewCacheKVStore(mem)
	require.Equal(t, valFmt(2), st.Get(keyFmt(1)))

	// make a new one and delete - should not be removed from mem
	st = NewCacheKVStore(mem)
	st.Delete(keyFmt(1))
	require.Empty(t, st.Get(keyFmt(1)))
	require.Equal(t, mem.Get(keyFmt(1)), valFmt(2))

	// Write. should now be removed from both
	st.Write()
	require.Empty(t, st.Get(keyFmt(1)), "Expected `key1` to be empty")
	require.Empty(t, mem.Get(keyFmt(1)), "Expected `key1` to be empty")
}

func TestCacheKVStoreNested(t *testing.T) {
	mem := dbm.NewMemDB()
	st := NewCacheKVStore(mem)

	// set. check its there on st and not on mem.
	st.Set(keyFmt(1), valFmt(1))
	require.Empty(t, mem.Get(keyFmt(1)))
	require.Equal(t, valFmt(1), st.Get(keyFmt(1)))

	// make a new from st and check
	st2 := NewCacheKVStore(st)
	require.Equal(t, valFmt(1), st2.Get(keyFmt(1)))

	// update the value on st2, check it only effects st2
	st2.Set(keyFmt(1), valFmt(3))
	require.Equal(t, []byte(nil), mem.Get(keyFmt(1)))
	require.Equal(t, valFmt(1), st.Get(keyFmt(1)))
	require.Equal(t, valFmt(3), st2.Get(keyFmt(1)))

	// st2 writes to its parent, st. doesnt effect mem
	st2.Write()
	require.Equal(t, []byte(nil), mem.Get(keyFmt(1)))
	require.Equal(t, valFmt(3), st.Get(keyFmt(1)))

	// updates mem
	st.Write()
	require.Equal(t, valFmt(3), mem.Get(keyFmt(1)))
}

func TestCacheKVIteratorBounds(t *testing.T) {
	mem := dbm.NewMemDB()
	st := NewCacheKVStore(mem)

	// set some items
	nItems := 5
	for i := 0; i < nItems; i++ {
		st.Set(keyFmt(i), valFmt(i))
	}

	// iterate over all of them
	itr := st.Iterator(nil, nil)
	var i = 0
	for ; itr.Valid(); itr.Next() {
		k, v := itr.Key(), itr.Value()
		assert.Equal(t, keyFmt(i), k)
		assert.Equal(t, valFmt(i), v)
		i += 1
	}
	assert.Equal(t, nItems, i)

	// iterate over none
	itr = st.Iterator(bz("money"), nil)
	i = 0
	for ; itr.Valid(); itr.Next() {
		fmt.Println(string(itr.Key()))
		i += 1
	}
	assert.Equal(t, 0, i)

	// iterate over lower
	itr = st.Iterator(keyFmt(0), keyFmt(3))
	i = 0
	for ; itr.Valid(); itr.Next() {
		k, v := itr.Key(), itr.Value()
		assert.Equal(t, keyFmt(i), k)
		assert.Equal(t, valFmt(i), v)
		i += 1
	}
	assert.Equal(t, 3, i)

	// iterate over upper
	itr = st.Iterator(keyFmt(2), keyFmt(4))
	i = 2
	for ; itr.Valid(); itr.Next() {
		k, v := itr.Key(), itr.Value()
		assert.Equal(t, keyFmt(i), k)
		assert.Equal(t, valFmt(i), v)
		i += 1
	}
	assert.Equal(t, 4, i)
}

func TestCacheKVMergeIteratorBasics(t *testing.T) {
	mem := dbm.NewMemDB()
	st := NewCacheKVStore(mem)

	// set an item in the cache, iterator should be empty
	k, v := keyFmt(0), valFmt(0)
	st.Set(k, v)
	st.Delete(k)
	assertIterateDomain(t, st, 0)

	// now set it and assert its there
	st.Set(k, v)
	assertIterateDomain(t, st, 1)

	// write it and assert its there
	st.Write()
	assertIterateDomain(t, st, 1)

	// remove it in cache and assert its not
	st.Delete(k)
	assertIterateDomain(t, st, 0)

	// write the delete and assert its not there
	st.Write()
	assertIterateDomain(t, st, 0)

	// add two keys and assert theyre there
	k1, v1 := keyFmt(1), valFmt(1)
	st.Set(k, v)
	st.Set(k1, v1)
	assertIterateDomain(t, st, 2)

	// write it and assert theyre there
	st.Write()
	assertIterateDomain(t, st, 2)

	// remove one in cache and assert its not
	st.Delete(k1)
	assertIterateDomain(t, st, 1)

	// write the delete and assert its not there
	st.Write()
	assertIterateDomain(t, st, 1)

	// delete the other key in cache and asserts its empty
	st.Delete(k)
	assertIterateDomain(t, st, 0)
}

func TestCacheKVMergeIterator(t *testing.T) {
	mem := dbm.NewMemDB()
	st := NewCacheKVStore(mem)

	// set some items and write them
	nItems := 5
	for i := 0; i < nItems; i++ {
		st.Set(keyFmt(i), valFmt(i))
	}
	st.Write()

	// set some more items and leave dirty
	for i := nItems; i < nItems*2; i++ {
		st.Set(keyFmt(i), valFmt(i))
	}

	// iterate over all of them
	assertIterateDomain(t, st, nItems*2)

	// delete the last dirty item, ensure we dont see it
	last := nItems*2 - 1
	st.Delete(keyFmt(last))
	assertIterateDomain(t, st, nItems*2-1)

	// write and check again
	st.Write()
	assertIterateDomain(t, st, nItems*2-1)

	// delete the next last one, ensure we dont see it
	last = last - 1
	st.Delete(keyFmt(last))
	assertIterateDomain(t, st, nItems*2-2)
}

// iterate over whole domain
func assertIterateDomain(t *testing.T, st KVStore, expectedN int) {
	itr := st.Iterator(nil, nil)
	var i = 0
	for ; itr.Valid(); itr.Next() {
		k, v := itr.Key(), itr.Value()
		assert.Equal(t, keyFmt(i), k)
		assert.Equal(t, valFmt(i), v)
		i += 1
	}
	assert.Equal(t, expectedN, i)
}

func bz(s string) []byte { return []byte(s) }
