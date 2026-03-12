package cachekv_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math/unsafe"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/store/cachekv"
	"cosmossdk.io/store/dbadapter"
	"cosmossdk.io/store/types"
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

	start, end = st.ReverseIterator(keyFmt(0), keyFmt(80)).Domain()
	require.Equal(t, keyFmt(0), start)
	require.Equal(t, keyFmt(80), end)
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
	// iterate over each and check they match the other
	itr := st.Iterator(nil, nil)
	itr2, err := mem.Iterator(nil, nil) // ground truth
	require.NoError(t, err)
	checkIterators(t, itr, itr2)
	checkIterators(t, itr2, itr)
	require.NoError(t, itr.Close())
	require.NoError(t, itr2.Close())
}

func checkIterators(t *testing.T, itr, itr2 types.Iterator) {
	t.Helper()
	for ; itr.Valid(); itr.Next() {
		require.True(t, itr2.Valid())
		k, v := itr.Key(), itr.Value()
		k2, v2 := itr2.Key(), itr2.Value()
		require.Equal(t, k, k2)
		require.Equal(t, v, v2)
		itr2.Next()
	}
	require.False(t, itr.Valid())
	require.False(t, itr2.Valid())
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
	maxRangeIdx := len(krc.keyRanges) - 1
	maxRange := krc.keyRanges[maxRangeIdx]

	// if we're not in the max range, we're valid
	if krc.rangeIdx <= maxRangeIdx &&
		krc.idx < maxRange.len() {
		return true
	}

	return false
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
