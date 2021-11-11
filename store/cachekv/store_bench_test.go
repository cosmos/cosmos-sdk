package cachekv_test

import (
	"testing"

	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/store/cachekv"
	"github.com/cosmos/cosmos-sdk/store/dbadapter"
)

var sink interface{}

const defaultValueSizeBz = 1 << 12

// This benchmark measures the time of iterator.Next() when the parent store is blank
func benchmarkBlankParentIteratorNext(b *testing.B, keysize int) {
	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	kvstore := cachekv.NewStore(mem)
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

	iter := kvstore.Iterator(keys[0], keys[b.N])
	defer iter.Close()

	for _ = iter.Key(); iter.Valid(); iter.Next() {
		// deadcode elimination stub
		sink = iter
	}
}

// Benchmark setting New keys to a store, where the new keys are in sequence.
func benchmarkBlankParentAppend(b *testing.B, keysize int) {
	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	kvstore := cachekv.NewStore(mem)

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
