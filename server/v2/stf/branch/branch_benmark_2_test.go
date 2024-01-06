package branch

import (
	"testing"
)

var sink interface{}

const defaultValueSizeBz = 1 << 12

// This benchmark measures the time of iterator.Next() when the parent store is blank
func benchmarkBlankParentIteratorNext(b *testing.B, keysize int) {
	b.Helper()
	mem := newMemState()
	kvstore := NewStore(mem)
	// Use a singleton for value, to not waste time computing it
	value := randSlice(defaultValueSizeBz)
	// Use simple values for keys, pick a random start,
	// and take next b.N keys sequentially after.]
	startKey := randSlice(32)

	// Add 1 to avoid issues when b.N = 1
	keys := generateSequentialKeys(startKey, b.N+1)
	for _, k := range keys {
		kvstore.Set(k, value)
	}

	b.ReportAllocs()
	b.ResetTimer()

	iter, err := kvstore.Iterator(keys[0], keys[b.N])
	if err != nil {
		b.Fatal(err)
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		_ = iter.Key()
		// deadcode elimination stub
		sink = iter
	}
}

// Benchmark setting New keys to a store, where the new keys are in sequence.
func benchmarkBlankParentAppend(b *testing.B, keysize int) {
	b.Helper()
	mem := newMemState()
	kvstore := NewStore(mem)

	// Use a singleton for value, to not waste time computing it
	value := randSlice(32)
	// Use simple values for keys, pick a random start,
	// and take next b.N keys sequentially after.
	startKey := randSlice(32)

	keys := generateSequentialKeys(startKey, b.N)

	b.ReportAllocs()
	b.ResetTimer()

	for _, k := range keys {
		kvstore.Set(k, value)
	}
}

// Benchmark setting New keys to a store, where the new keys are random.
// the speed of this function does not depend on the values in the parent store
func benchmarkRandomSet(b *testing.B, keysize int) {
	b.Helper()
	mem := newMemState()
	kvstore := NewStore(mem)

	// Use a singleton for value, to not waste time computing it
	value := randSlice(defaultValueSizeBz)
	// Add 1 to avoid issues when b.N = 1
	keys := generateRandomKeys(keysize, b.N+1)

	b.ReportAllocs()
	b.ResetTimer()

	for _, k := range keys {
		_ = kvstore.Set(k, value)
	}

	iter, err := kvstore.Iterator(keys[0], keys[b.N])
	if err != nil {
		b.Fatal(err)
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		_ = iter.Key()
		// deadcode elimination stub
		sink = iter
	}
}

/*
// Benchmark creating an iterator on a parent with D entries,
// that are all deleted in the cacheKV store.
// We essentially are benchmarking the cacheKV iterator creation & iteration times
// with the number of entries deleted in the parent.
func benchmarkIteratorOnParentWithManyDeletes(b *testing.B, numDeletes int) {
	b.Helper()
	mem := dbadapter.Store{DB: dbm.NewMemDB()}

	// Use a singleton for value, to not waste time computing it
	value := randSlice(32)
	// Use simple values for keys, pick a random start,
	// and take next D keys sequentially after.
	startKey := randSlice(32)
	// Add 1 to avoid issues when numDeletes = 1
	keys := generateSequentialKeys(startKey, numDeletes+1)
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

	iter := kvstore.Iterator(keys[0], keys[numDeletes])
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		_ = iter.Key()
		// deadcode elimination stub
		sink = iter
	}
}

*/

// BenchmarkBlankParentIteratorNextKeySize32-14    	64378976	        42.42 ns/op	       0 B/op	       0 allocs/op
func BenchmarkBlankParentIteratorNextKeySize32(b *testing.B) {
	benchmarkBlankParentIteratorNext(b, 32)
}

// BenchmarkBlankParentAppendKeySize32-14    	 6852801	       198.9 ns/op	     101 B/op	       0 allocs/op
func BenchmarkBlankParentAppendKeySize32(b *testing.B) {
	benchmarkBlankParentAppend(b, 32)
}

// BenchmarkSetKeySize32-14    	 1888374	       832.9 ns/op	     138 B/op	       0 allocs/op
func BenchmarkSetKeySize32(b *testing.B) {
	benchmarkRandomSet(b, 32)
}

/*
func BenchmarkIteratorOnParentWith1MDeletes(b *testing.B) {
	benchmarkIteratorOnParentWithManyDeletes(b, 1_000_000)
}
*/
