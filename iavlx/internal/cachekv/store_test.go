package cachekv_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math/unsafe"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/store/v2/dbadapter"
	"github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/iavlx/internal/cachekv"
)

func newStoreWithParent() (*cachekv.Store, dbadapter.Store) {
	parent := dbadapter.Store{DB: dbm.NewMemDB()}
	return cachekv.NewStore(parent), parent
}

func TestGet_FromParentWhenClean(t *testing.T) {
	st, parent := newStoreWithParent()
	parent.Set([]byte("key"), []byte("parent_value"))

	got := st.Get([]byte("key"))
	require.Equal(t, []byte("parent_value"), got)
}

func TestGet_FromWriteMap(t *testing.T) {
	st, parent := newStoreWithParent()
	parent.Set([]byte("key"), []byte("parent_value"))

	st.Set([]byte("key"), []byte("cached_value"))

	got := st.Get([]byte("key"))
	require.Equal(t, []byte("cached_value"), got)
}

func TestGet_DeletedKeyReturnsNil(t *testing.T) {
	st, parent := newStoreWithParent()
	parent.Set([]byte("key"), []byte("parent_value"))

	st.Delete([]byte("key"))

	got := st.Get([]byte("key"))
	require.Nil(t, got)
}

func TestHas_CachedValue(t *testing.T) {
	st, _ := newStoreWithParent()
	st.Set([]byte("key"), []byte("value"))

	require.True(t, st.Has([]byte("key")))
}

func TestHas_DeletedKey(t *testing.T) {
	st, parent := newStoreWithParent()
	parent.Set([]byte("key"), []byte("value"))

	st.Delete([]byte("key"))

	require.False(t, st.Has([]byte("key")))
}

func TestSet_MarksStoreDirty(t *testing.T) {
	st, _ := newStoreWithParent()

	_, count := st.Updates()
	require.Equal(t, 0, count)

	st.Set([]byte("key"), []byte("value"))

	_, count = st.Updates()
	require.Equal(t, 1, count)
}

func TestDelete_MarksStoreDirty(t *testing.T) {
	st, _ := newStoreWithParent()

	_, count := st.Updates()
	require.Equal(t, 0, count)

	st.Delete([]byte("key"))

	_, count = st.Updates()
	require.Equal(t, 1, count)
}

// --- Input Validation ---

func TestSet_PanicsOnInvalidInput(t *testing.T) {
	st, _ := newStoreWithParent()

	require.Panics(t, func() { st.Set(nil, []byte("value")) }, "nil key")
	require.Panics(t, func() { st.Set([]byte(""), []byte("value")) }, "empty key")
	require.Panics(t, func() { st.Set([]byte("key"), nil) }, "nil value")
}

func TestGet_PanicsOnInvalidKey(t *testing.T) {
	st, _ := newStoreWithParent()

	require.Panics(t, func() { st.Get(nil) })
	require.Panics(t, func() { st.Get([]byte("")) })
}

func TestHas_PanicsOnInvalidKey(t *testing.T) {
	st, _ := newStoreWithParent()

	require.Panics(t, func() { st.Has(nil) })
	require.Panics(t, func() { st.Has([]byte("")) })
}

func TestDelete_PanicsOnInvalidKey(t *testing.T) {
	st, _ := newStoreWithParent()

	require.Panics(t, func() { st.Delete(nil) })
	require.Panics(t, func() { st.Delete([]byte("")) })
}

// --- Write Method ---

func TestWrite_FlushesToParent(t *testing.T) {
	st, parent := newStoreWithParent()

	st.Set([]byte("key"), []byte("value"))
	require.False(t, parent.Has([]byte("key")))

	st.Write()

	require.Equal(t, []byte("value"), parent.Get([]byte("key")))
}

func TestWrite_DeletesPropagated(t *testing.T) {
	st, parent := newStoreWithParent()
	parent.Set([]byte("key"), []byte("value"))

	st.Delete([]byte("key"))
	require.True(t, parent.Has([]byte("key")))

	st.Write()

	require.False(t, parent.Has([]byte("key")))
}

func TestWrite_ClearsWriteMap(t *testing.T) {
	st, _ := newStoreWithParent()

	st.Set([]byte("key"), []byte("value"))
	_, count := st.Updates()
	require.Equal(t, 1, count)

	st.Write()

	// after write, store is clean again (dirty=false path in Updates)
	_, count = st.Updates()
	require.Equal(t, 0, count)
}

func TestWrite_Overwriting(t *testing.T) {
	st, parent := newStoreWithParent()

	st.Set([]byte("key"), []byte("first"))
	st.Set([]byte("key"), []byte("second"))
	st.Set([]byte("key"), []byte("third"))

	st.Write()

	require.Equal(t, []byte("third"), parent.Get([]byte("key")))
}

func TestWrite_SetThenDelete(t *testing.T) {
	st, parent := newStoreWithParent()

	st.Set([]byte("key"), []byte("value"))
	st.Delete([]byte("key"))

	st.Write()

	require.False(t, parent.Has([]byte("key")))
}

func TestWrite_OverwritesParent(t *testing.T) {
	st, parent := newStoreWithParent()
	parent.Set([]byte("key"), []byte("original"))

	st.Set([]byte("key"), []byte("new_value"))

	st.Write()

	require.Equal(t, []byte("new_value"), parent.Get([]byte("key")))
}

// --- Updates() Method ---

func TestUpdates_EmptyWhenClean(t *testing.T) {
	st, _ := newStoreWithParent()

	updates, count := st.Updates()
	require.Equal(t, 0, count)

	collected := 0
	updates(func(cachekv.Update[[]byte]) bool {
		collected++
		return true
	})
	require.Equal(t, 0, collected)
}

func TestUpdates_ReturnsAllChanges(t *testing.T) {
	st, _ := newStoreWithParent()

	st.Set([]byte("a"), []byte("1"))
	st.Set([]byte("b"), []byte("2"))
	st.Delete([]byte("c"))

	updates, count := st.Updates()
	require.Equal(t, 3, count)

	result := make(map[string][]byte)
	updates(func(update cachekv.Update[[]byte]) bool {
		result[string(update.Key)] = update.Value
		return true
	})

	require.Equal(t, []byte("1"), result["a"])
	require.Equal(t, []byte("2"), result["b"])
	require.Nil(t, result["c"]) // deletion marked with nil
}

func TestUpdates_CorrectCount(t *testing.T) {
	st, _ := newStoreWithParent()

	st.Set([]byte("a"), []byte("1"))
	st.Set([]byte("b"), []byte("2"))
	st.Set([]byte("a"), []byte("updated")) // same key, should not increase count

	_, count := st.Updates()
	require.Equal(t, 2, count)
}

// --- CacheWrap Hierarchy ---

func TestCacheWrap_ReturnsNewStore(t *testing.T) {
	st, _ := newStoreWithParent()

	wrapped := st.CacheWrap()

	require.IsType(t, &cachekv.Store{}, wrapped)
	require.NotSame(t, st, wrapped)
}

func TestCacheWrap_IsolatesWrites(t *testing.T) {
	st, _ := newStoreWithParent()
	st.Set([]byte("key"), []byte("original"))

	child := st.CacheWrap().(*cachekv.Store)
	child.Set([]byte("key"), []byte("modified"))
	child.Set([]byte("new"), []byte("value"))

	// parent should not see child changes
	require.Equal(t, []byte("original"), st.Get([]byte("key")))
	require.False(t, st.Has([]byte("new")))

	// child sees its own changes
	require.Equal(t, []byte("modified"), child.Get([]byte("key")))
	require.True(t, child.Has([]byte("new")))
}

func keyFmt(i int) []byte { return bz(fmt.Sprintf("key%0.8d", i)) }
func valFmt(i int) []byte { return bz(fmt.Sprintf("value%0.8d", i)) }

func TestCacheWrap_NestedWrite(t *testing.T) {
	st, parent := newStoreWithParent()

	child := st.CacheWrap().(*cachekv.Store)
	k := keyFmt(0)
	v := valFmt(0)
	child.Set(k, v)

	// write child to parent cache
	child.Write()
	require.Equal(t, v, st.Get(k))
	require.False(t, parent.Has(k))

	// write parent cache to underlying
	st.Write()
	assertIterateDomain(t, st, 1)

	// delete the other key in cache and asserts its empty
	st.Delete(k)
	assertIterateDomain(t, st, 0)
}

func newCacheKVStore() *cachekv.Store {
	st, _ := newStoreWithParent()
	return st
}

func TestCacheKVMergeIteratorDeleteLast(t *testing.T) {
	st := newCacheKVStore()

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

	// delete them all
	for i := 0; i < nItems*2; i++ {
		last := nItems*2 - 1 - i
		st.Delete(keyFmt(last))
		assertIterateDomain(t, st, last)
	}
}

func TestCacheKVMergeIteratorDeletes(t *testing.T) {
	st := newCacheKVStore()
	truth := dbm.NewMemDB()

	// set some items and write them
	nItems := 10
	for i := 0; i < nItems; i++ {
		doOp(t, st, truth, opSet, i)
	}
	st.Write()

	// delete every other item, starting from 0
	for i := 0; i < nItems; i += 2 {
		doOp(t, st, truth, opDel, i)
		assertIterateDomainCompare(t, st, truth)
	}

	// reset
	st = newCacheKVStore()
	truth = dbm.NewMemDB()

	// set some items and write them
	for i := 0; i < nItems; i++ {
		doOp(t, st, truth, opSet, i)
	}
	st.Write()

	// delete every other item, starting from 1
	for i := 1; i < nItems; i += 2 {
		doOp(t, st, truth, opDel, i)
		assertIterateDomainCompare(t, st, truth)
	}
}

func TestCacheKVMergeIteratorChunks(t *testing.T) {
	st := newCacheKVStore()

	// Use the truth to check values on the merge iterator
	truth := dbm.NewMemDB()

	// sets to the parent
	setRange(t, st, truth, 0, 20)
	setRange(t, st, truth, 40, 60)
	st.Write()

	// sets to the cache
	setRange(t, st, truth, 20, 40)
	setRange(t, st, truth, 60, 80)
	assertIterateDomainCheck(t, st, truth, []keyRange{{0, 80}})

	// remove some parents and some cache
	deleteRange(t, st, truth, 15, 25)
	assertIterateDomainCheck(t, st, truth, []keyRange{{0, 15}, {25, 80}})

	// remove some parents and some cache
	deleteRange(t, st, truth, 35, 45)
	assertIterateDomainCheck(t, st, truth, []keyRange{{0, 15}, {25, 35}, {45, 80}})

	// write, add more to the cache, and delete some cache
	st.Write()
	setRange(t, st, truth, 38, 42)
	deleteRange(t, st, truth, 40, 43)
	assertIterateDomainCheck(t, st, truth, []keyRange{{0, 15}, {25, 35}, {38, 40}, {45, 80}})
}

func TestCacheKVMergeIteratorDomain(t *testing.T) {
	st := newCacheKVStore()

	itr := st.Iterator(nil, nil)
	start, end := itr.Domain()
	require.Equal(t, start, end)
	require.NoError(t, itr.Close())

	itr = st.Iterator(keyFmt(40), keyFmt(60))
	start, end = itr.Domain()
	require.Equal(t, keyFmt(40), start)
	require.Equal(t, keyFmt(60), end)
	require.NoError(t, itr.Close())

	itr = st.ReverseIterator(keyFmt(0), keyFmt(80))
	start, end = itr.Domain()
	require.Equal(t, keyFmt(0), start)
	require.Equal(t, keyFmt(80), end)
	require.NoError(t, itr.Close())
}

func TestCacheKVMergeIteratorRandom(t *testing.T) {
	st := newCacheKVStore()
	truth := dbm.NewMemDB()

	start, end := 25, 975
	max := 1000
	setRange(t, st, truth, start, end)

	// do an op, test the iterator
	for i := 0; i < 2000; i++ {
		doRandomOp(t, st, truth, max)
		assertIterateDomainCompare(t, st, truth)
	}
}

func TestNilEndIterator(t *testing.T) {
	const SIZE = 3000

	tests := []struct {
		name       string
		write      bool
		startIndex int
		end        []byte
	}{
		{name: "write=false, end=nil", write: false, end: nil, startIndex: 1000},
		{name: "write=false, end=nil; full key scan", write: false, end: nil, startIndex: 2000},
		{name: "write=true, end=nil", write: true, end: nil, startIndex: 1000},
		{name: "write=false, end=non-nil", write: false, end: keyFmt(3000), startIndex: 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := newCacheKVStore()

			for i := 0; i < SIZE; i++ {
				kstr := keyFmt(i)
				st.Set(kstr, valFmt(i))
			}

			if tt.write {
				st.Write()
			}

			itr := st.Iterator(keyFmt(tt.startIndex), tt.end)
			i := tt.startIndex
			j := 0
			for itr.Valid() {
				require.Equal(t, keyFmt(i), itr.Key())
				require.Equal(t, valFmt(i), itr.Value())
				itr.Next()
				i++
				j++
			}

			require.Equal(t, SIZE-tt.startIndex, j)
			require.NoError(t, itr.Close())
		})
	}
}

// --- Post-Write behavior (exercises !dirty fast paths) ---

func TestGet_AfterWrite_ReadsFromParent(t *testing.T) {
	st, parent := newStoreWithParent()

	st.Set([]byte("key"), []byte("value"))
	st.Write()

	// after Write, store is clean — Get should delegate to parent
	require.Equal(t, []byte("value"), st.Get([]byte("key")))
	require.Equal(t, []byte("value"), parent.Get([]byte("key")))

	// modifying parent directly should be visible through clean cache
	parent.Set([]byte("key"), []byte("updated"))
	require.Equal(t, []byte("updated"), st.Get([]byte("key")))
}

func TestHas_AfterWrite_ReadsFromParent(t *testing.T) {
	st, _ := newStoreWithParent()

	st.Set([]byte("key"), []byte("value"))
	st.Write()

	require.True(t, st.Has([]byte("key")))
	require.False(t, st.Has([]byte("nonexistent")))
}

func TestWrite_DoubleWriteIsNoop(t *testing.T) {
	st, parent := newStoreWithParent()

	st.Set([]byte("key"), []byte("value"))
	st.Write()
	require.Equal(t, []byte("value"), parent.Get([]byte("key")))

	// second Write with no new changes should be a no-op
	st.Write()
	require.Equal(t, []byte("value"), parent.Get([]byte("key")))
}

// --- Iterator: cache overrides parent values ---

func TestIterator_CacheOverridesParentValues(t *testing.T) {
	st, parent := newStoreWithParent()

	// set keys in parent
	for i := 0; i < 5; i++ {
		parent.Set(keyFmt(i), valFmt(i))
	}

	// override some in cache with different values
	st.Set(keyFmt(1), []byte("override1"))
	st.Set(keyFmt(3), []byte("override3"))

	itr := st.Iterator(nil, nil)
	defer itr.Close()

	expected := []struct {
		key []byte
		val []byte
	}{
		{keyFmt(0), valFmt(0)},
		{keyFmt(1), []byte("override1")},
		{keyFmt(2), valFmt(2)},
		{keyFmt(3), []byte("override3")},
		{keyFmt(4), valFmt(4)},
	}

	for i, exp := range expected {
		require.True(t, itr.Valid(), "expected entry %d", i)
		require.Equal(t, exp.key, itr.Key())
		require.Equal(t, exp.val, itr.Value())
		itr.Next()
	}
	require.False(t, itr.Valid())
}

// --- ReverseIterator correctness ---

func TestReverseIterator_CorrectOrder(t *testing.T) {
	st := newCacheKVStore()

	nItems := 10
	for i := 0; i < nItems; i++ {
		st.Set(keyFmt(i), valFmt(i))
	}

	itr := st.ReverseIterator(nil, nil)
	defer itr.Close()

	i := nItems - 1
	for ; itr.Valid(); itr.Next() {
		require.Equal(t, keyFmt(i), itr.Key())
		require.Equal(t, valFmt(i), itr.Value())
		i--
	}
	require.Equal(t, -1, i)
}

func TestReverseIterator_MixedParentAndCache(t *testing.T) {
	st, parent := newStoreWithParent()

	// put some in parent
	for i := 0; i < 5; i++ {
		parent.Set(keyFmt(i), valFmt(i))
	}
	// put some in cache only
	for i := 5; i < 10; i++ {
		st.Set(keyFmt(i), valFmt(i))
	}

	itr := st.ReverseIterator(nil, nil)
	defer itr.Close()

	i := 9
	for ; itr.Valid(); itr.Next() {
		require.Equal(t, keyFmt(i), itr.Key())
		require.Equal(t, valFmt(i), itr.Value())
		i--
	}
	require.Equal(t, -1, i)
}

func TestReverseIterator_WithDeletes(t *testing.T) {
	st, parent := newStoreWithParent()

	for i := 0; i < 5; i++ {
		parent.Set(keyFmt(i), valFmt(i))
	}

	// delete keys 1 and 3 in cache
	st.Delete(keyFmt(1))
	st.Delete(keyFmt(3))

	itr := st.ReverseIterator(nil, nil)
	defer itr.Close()

	expected := []int{4, 2, 0}
	for _, idx := range expected {
		require.True(t, itr.Valid())
		require.Equal(t, keyFmt(idx), itr.Key())
		require.Equal(t, valFmt(idx), itr.Value())
		itr.Next()
	}
	require.False(t, itr.Valid())
}

// --- Bounded iteration with mixed parent/cache ---

func TestIterator_BoundedWithMixedParentAndCache(t *testing.T) {
	st, parent := newStoreWithParent()

	// keys 0-4 in parent, 5-9 in cache
	for i := 0; i < 5; i++ {
		parent.Set(keyFmt(i), valFmt(i))
	}
	for i := 5; i < 10; i++ {
		st.Set(keyFmt(i), valFmt(i))
	}

	// iterate over range spanning both parent and cache
	itr := st.Iterator(keyFmt(3), keyFmt(8))
	defer itr.Close()

	i := 3
	for ; itr.Valid(); itr.Next() {
		require.Equal(t, keyFmt(i), itr.Key())
		require.Equal(t, valFmt(i), itr.Value())
		i++
	}
	require.Equal(t, 8, i) // end is exclusive
}

func TestReverseIterator_Bounded(t *testing.T) {
	st := newCacheKVStore()

	for i := 0; i < 10; i++ {
		st.Set(keyFmt(i), valFmt(i))
	}

	itr := st.ReverseIterator(keyFmt(3), keyFmt(7))
	defer itr.Close()

	i := 6 // end is exclusive
	for ; itr.Valid(); itr.Next() {
		require.Equal(t, keyFmt(i), itr.Key())
		require.Equal(t, valFmt(i), itr.Value())
		i--
	}
	require.Equal(t, 2, i) // stopped before start
}

// --- Iterator: delete cache-only key ---

func TestIterator_DeleteCacheOnlyKey(t *testing.T) {
	st, parent := newStoreWithParent()

	parent.Set(keyFmt(0), valFmt(0))
	parent.Set(keyFmt(2), valFmt(2))

	// set a cache-only key then delete it
	st.Set(keyFmt(1), valFmt(1))
	st.Delete(keyFmt(1))

	itr := st.Iterator(nil, nil)
	defer itr.Close()

	// should only see parent keys, the cache-only key was deleted
	require.True(t, itr.Valid())
	require.Equal(t, keyFmt(0), itr.Key())
	itr.Next()
	require.True(t, itr.Valid())
	require.Equal(t, keyFmt(2), itr.Key())
	itr.Next()
	require.False(t, itr.Valid())
}

// --- Updates() early stop ---

func TestUpdates_EarlyStop(t *testing.T) {
	st, _ := newStoreWithParent()

	st.Set([]byte("a"), []byte("1"))
	st.Set([]byte("b"), []byte("2"))
	st.Set([]byte("c"), []byte("3"))

	updates, count := st.Updates()
	require.Equal(t, 3, count)

	collected := 0
	updates(func(cachekv.Update[[]byte]) bool {
		collected++
		return false // stop after first
	})
	require.Equal(t, 1, collected)
}

// --- Iterator after Write (clean store, delegates to parent) ---

func TestIterator_AfterWrite_DelegatesToParent(t *testing.T) {
	st, _ := newStoreWithParent()

	for i := 0; i < 5; i++ {
		st.Set(keyFmt(i), valFmt(i))
	}
	st.Write()

	// store is clean now, iterator should delegate to parent
	itr := st.Iterator(nil, nil)
	defer itr.Close()

	i := 0
	for ; itr.Valid(); itr.Next() {
		require.Equal(t, keyFmt(i), itr.Key())
		require.Equal(t, valFmt(i), itr.Value())
		i++
	}
	require.Equal(t, 5, i)
}

// TestIteratorDeadlock demonstrate the deadlock issue in cache store.
func TestIteratorDeadlock(t *testing.T) {
	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	store := cachekv.NewStore(mem)
	// the channel buffer is 64 and received once, so put at least 66 elements.
	for i := 0; i < 66; i++ {
		store.Set([]byte(fmt.Sprintf("key%d", i)), []byte{1})
	}
	it := store.Iterator(nil, nil)
	defer it.Close()
	store.Set([]byte("key20"), []byte{1})
	// it'll be blocked here with previous version, or enable lock on btree.
	it2 := store.Iterator(nil, nil)
	defer it2.Close()
}

//-------------------------------------------------------------------------------------------
// do some random ops

const (
	opSet      = 0
	opSetRange = 1
	opDel      = 2
	opDelRange = 3
	opWrite    = 4

	totalOps = 5 // number of possible operations
)

func randInt(n int) int {
	return unsafe.NewRand().Int() % n
}

// useful for replaying an error case if we find one
func doOp(t *testing.T, st types.CacheKVStore, truth dbm.DB, op int, args ...int) {
	t.Helper()
	switch op {
	case opSet:
		k := args[0]
		st.Set(keyFmt(k), valFmt(k))
		err := truth.Set(keyFmt(k), valFmt(k))
		require.NoError(t, err)
	case opSetRange:
		require.True(t, len(args) > 1)
		start := args[0]
		end := args[1]
		setRange(t, st, truth, start, end)
	case opDel:
		k := args[0]
		st.Delete(keyFmt(k))
		err := truth.Delete(keyFmt(k))
		require.NoError(t, err)
	case opDelRange:
		require.True(t, len(args) > 1)
		start := args[0]
		end := args[1]
		deleteRange(t, st, truth, start, end)
	case opWrite:
		st.Write()
	}
}

func doRandomOp(t *testing.T, st types.CacheKVStore, truth dbm.DB, maxKey int) {
	t.Helper()
	r := randInt(totalOps)
	switch r {
	case opSet:
		k := randInt(maxKey)
		st.Set(keyFmt(k), valFmt(k))
		err := truth.Set(keyFmt(k), valFmt(k))
		require.NoError(t, err)
	case opSetRange:
		start := randInt(maxKey - 2)
		end := randInt(maxKey-start) + start
		setRange(t, st, truth, start, end)
	case opDel:
		k := randInt(maxKey)
		st.Delete(keyFmt(k))
		err := truth.Delete(keyFmt(k))
		require.NoError(t, err)
	case opDelRange:
		start := randInt(maxKey - 2)
		end := randInt(maxKey-start) + start
		deleteRange(t, st, truth, start, end)
	case opWrite:
		st.Write()
	}
}

//-------------------------------------------------------------------------------------------

// iterate over whole domain
func assertIterateDomain(t *testing.T, st types.KVStore, expectedN int) {
	t.Helper()
	itr := st.Iterator(nil, nil)
	i := 0
	for ; itr.Valid(); itr.Next() {
		k, v := itr.Key(), itr.Value()
		require.Equal(t, keyFmt(i), k)
		require.Equal(t, valFmt(i), v)
		i++
	}
	require.Equal(t, expectedN, i)
	require.NoError(t, itr.Close())
}

func assertIterateDomainCheck(t *testing.T, st types.KVStore, mem dbm.DB, r []keyRange) {
	t.Helper()
	// iterate over each and check they match the other
	itr := st.Iterator(nil, nil)
	itr2, err := mem.Iterator(nil, nil) // ground truth
	require.NoError(t, err)

	krc := newKeyRangeCounter(r)

	for ; krc.valid(); krc.next() {
		require.True(t, itr.Valid())
		require.True(t, itr2.Valid())

		// check the key/val matches the ground truth
		k, v := itr.Key(), itr.Value()
		k2, v2 := itr2.Key(), itr2.Value()
		require.Equal(t, k, k2)
		require.Equal(t, v, v2)

		// check they match the counter
		require.Equal(t, k, keyFmt(krc.key()))

		itr.Next()
		itr2.Next()
	}

	require.False(t, itr.Valid())
	require.False(t, itr2.Valid())
	require.NoError(t, itr.Close())
	require.NoError(t, itr2.Close())
}

func assertIterateDomainCompare(t *testing.T, st types.KVStore, mem dbm.DB) {
	t.Helper()
	itr := st.Iterator(nil, nil)
	itr2, err := mem.Iterator(nil, nil) // ground truth
	require.NoError(t, err)

	for ; itr.Valid(); itr.Next() {
		require.True(t, itr2.Valid(), "store has more entries than ground truth")
		require.Equal(t, itr.Key(), itr2.Key())
		require.Equal(t, itr.Value(), itr2.Value())
		itr2.Next()
	}
	require.False(t, itr2.Valid(), "ground truth has more entries than store")

	require.NoError(t, itr.Close())
	require.NoError(t, itr2.Close())
}

//--------------------------------------------------------

func setRange(t *testing.T, st types.KVStore, mem dbm.DB, start, end int) {
	t.Helper()
	for i := start; i < end; i++ {
		st.Set(keyFmt(i), valFmt(i))
		err := mem.Set(keyFmt(i), valFmt(i))
		require.NoError(t, err)
	}
}

func deleteRange(t *testing.T, st types.KVStore, mem dbm.DB, start, end int) {
	t.Helper()
	for i := start; i < end; i++ {
		st.Delete(keyFmt(i))
		err := mem.Delete(keyFmt(i))
		require.NoError(t, err)
	}
}

//--------------------------------------------------------

type keyRange struct {
	start int
	end   int
}

func (kr keyRange) len() int {
	return kr.end - kr.start
}

func newKeyRangeCounter(kr []keyRange) *keyRangeCounter {
	return &keyRangeCounter{keyRanges: kr}
}

// we can iterate over this and make sure our real iterators have all the right keys
type keyRangeCounter struct {
	rangeIdx  int
	idx       int
	keyRanges []keyRange
}

func (krc *keyRangeCounter) valid() bool {
	if krc.rangeIdx >= len(krc.keyRanges) {
		return false
	}
	return krc.idx < krc.keyRanges[krc.rangeIdx].len()
}

func (krc *keyRangeCounter) next() {
	thisKeyRange := krc.keyRanges[krc.rangeIdx]
	if krc.idx == thisKeyRange.len()-1 {
		krc.rangeIdx++
		krc.idx = 0
	} else {
		krc.idx++
	}
}

func (krc *keyRangeCounter) key() int {
	thisKeyRange := krc.keyRanges[krc.rangeIdx]
	return thisKeyRange.start + krc.idx
}

//--------------------------------------------------------

func bz(s string) []byte { return []byte(s) }

func BenchmarkCacheKVStoreGetNoKeyFound(b *testing.B) {
	b.ReportAllocs()
	st := newCacheKVStore()
	b.ResetTimer()
	// assumes b.N < 2**24
	idx := 0
	for b.Loop() {
		st.Get([]byte{byte((idx & 0xFF0000) >> 16), byte((idx & 0xFF00) >> 8), byte(idx & 0xFF)})
		idx++
	}
}

func BenchmarkCacheKVStoreGetKeyFound(b *testing.B) {
	b.ReportAllocs()
	st := newCacheKVStore()
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		st.Set(arr, arr)
	}
	b.ResetTimer()
	// assumes b.N < 2**24
	idx := 0
	for b.Loop() {
		st.Get([]byte{byte((idx & 0xFF0000) >> 16), byte((idx & 0xFF00) >> 8), byte(idx & 0xFF)})
		idx++
	}
}
