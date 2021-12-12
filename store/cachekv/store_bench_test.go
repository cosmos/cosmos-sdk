package cachekv_test

import (
	"crypto/rand"
	"sort"
	"testing"

	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/store/cachekv"
	"github.com/cosmos/cosmos-sdk/store/dbadapter"
)

func benchmarkCacheKVStoreIterator(numKVs int, b *testing.B) {
	b.ReportAllocs()
	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	cstore := cachekv.NewStore(mem)
	keys := make([]string, numKVs)

	for i := 0; i < numKVs; i++ {
		key := make([]byte, 32)
		value := make([]byte, 32)

		_, _ = rand.Read(key)
		_, _ = rand.Read(value)

		keys[i] = string(key)
		cstore.Set(key, value)
	}

	sort.Strings(keys)

	for n := 0; n < b.N; n++ {
		iter := cstore.Iterator([]byte(keys[0]), []byte(keys[numKVs-1]))

		for _ = iter.Key(); iter.Valid(); iter.Next() {
		}

		iter.Close()
	}
}

<<<<<<< HEAD
func BenchmarkCacheKVStoreIterator500(b *testing.B)    { benchmarkCacheKVStoreIterator(500, b) }
func BenchmarkCacheKVStoreIterator1000(b *testing.B)   { benchmarkCacheKVStoreIterator(1000, b) }
func BenchmarkCacheKVStoreIterator10000(b *testing.B)  { benchmarkCacheKVStoreIterator(10000, b) }
func BenchmarkCacheKVStoreIterator50000(b *testing.B)  { benchmarkCacheKVStoreIterator(50000, b) }
func BenchmarkCacheKVStoreIterator100000(b *testing.B) { benchmarkCacheKVStoreIterator(100000, b) }
=======
// Benchmark setting New keys to a store, where the new keys are random.
// the speed of this function does not depend on the values in the parent store
func benchmarkRandomSet(b *testing.B, keysize int) {
	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	kvstore := cachekv.NewStore(mem)

	// Use a singleton for value, to not waste time computing it
	value := randSlice(defaultValueSizeBz)
	keys := generateRandomKeys(keysize, b.N)

	b.ReportAllocs()
	b.ResetTimer()

	for _, k := range keys {
		kvstore.Set(k, value)
	}

	iter := kvstore.Iterator(keys[0], keys[b.N])
	defer iter.Close()

	for _ = iter.Key(); iter.Valid(); iter.Next() {
		// deadcode elimination stub
		sink = iter
	}
}

// Benchmark creating an iterator on a parent with D entries,
// that are all deleted in the cacheKV store.
// We essentially are benchmarking the cacheKV iterator creation & iteration times
// with the number of entries deleted in the parent.
func benchmarkIteratorOnParentWithManyDeletes(b *testing.B, numDeletes int) {
	mem := dbadapter.Store{DB: dbm.NewMemDB()}

	// Use a singleton for value, to not waste time computing it
	value := randSlice(32)
	// Use simple values for keys, pick a random start,
	// and take next D keys sequentially after.
	startKey := randSlice(32)
	keys := generateSequentialKeys(startKey, numDeletes)
	// setup parent db with D keys.
	for _, k := range keys {
		mem.Set(k, value)
	}
	kvstore := cachekv.NewStore(mem)
	// Delete all keys from the cache KV store.
	// The keys[1:] is to keep at least one entry in parent, due to a bug in the SDK iterator design.
	// Essentially the iterator will never be valid, in that it should never run.
	// However, this is incompatible with the for loop structure the SDK uses, hence
	// causes a panic. Thus we do keys[1:].
	for _, k := range keys[1:] {
		kvstore.Delete(k)
	}

	b.ReportAllocs()
	b.ResetTimer()

	iter := kvstore.Iterator(keys[0], keys[b.N])
	defer iter.Close()

	for _ = iter.Key(); iter.Valid(); iter.Next() {
		// deadcode elimination stub
		sink = iter
	}
}

func BenchmarkBlankParentIteratorNextKeySize32(b *testing.B) {
	benchmarkBlankParentIteratorNext(b, 32)
}

func BenchmarkBlankParentAppendKeySize32(b *testing.B) {
	benchmarkBlankParentAppend(b, 32)
}

func BenchmarkSetKeySize32(b *testing.B) {
	benchmarkRandomSet(b, 32)
}

func BenchmarkIteratorOnParentWith1MDeletes(b *testing.B) {
	benchmarkIteratorOnParentWithManyDeletes(b, 1_000_000)
}
>>>>>>> 314e1d52c (perf: Speedup cachekv iterator on large deletions & IBC v2 upgrade logic (#10741))
