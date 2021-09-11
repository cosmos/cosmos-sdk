package cachekv_test

import (
	"testing"

	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/store/cachekv"
	"github.com/cosmos/cosmos-sdk/store/dbadapter"
)

// This benchmark measures the time of iterator.Next() when the parent store is blank
func benchmarkBlankParentIteratorNext(b *testing.B, keysize int) {
	b.StopTimer()

	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	kvstore := cachekv.NewStore(mem)
	// Use a singleton for value, to not waste time computing it
	value := randSlice(32)
	// Use simple values for keys, pick a random start,
	// and take next b.N keys sequentially after.]
	startKey := randSlice(32)

	// Add 1 to avoid issues when b.N = 1
	keys := generateSequentialKeys(startKey, b.N+1)
	for _, k := range keys {
		kvstore.Set(k, value)
	}

	b.ReportAllocs()
	b.StartTimer()

	iter := kvstore.Iterator(keys[0], keys[b.N])

	for _ = iter.Key(); iter.Valid(); iter.Next() {
		// TODO: Check if we need to ensure this isn't getting deadcode eliminated
	}

	iter.Close()
}

// Benchmark setting New keys to a store, where the new keys are in sequence.
func benchmarkBlankParentAppend(b *testing.B, keysize int) {
	b.StopTimer()

	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	kvstore := cachekv.NewStore(mem)

	// Use a singleton for value, to not waste time computing it
	value := randSlice(32)
	// Use simple values for keys, pick a random start,
	// and take next b.N keys sequentially after.
	startKey := randSlice(32)

	keys := generateSequentialKeys(startKey, b.N)

	b.ReportAllocs()
	b.StartTimer()

	for _, k := range keys {
		kvstore.Set(k, value)
	}
}

// Benchmark setting New keys to a store, where the new keys are random.
// the speed of this function does not depend on the values in the parent store
func benchmarkRandomSet(b *testing.B, keysize int) {
	b.StopTimer()

	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	kvstore := cachekv.NewStore(mem)

	// Use a singleton for value, to not waste time computing it
	value := randSlice(32)
	keys := generateRandomKeys(keysize, b.N)

	b.ReportAllocs()
	b.StartTimer()

	for _, k := range keys {
		kvstore.Set(k, value)
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
