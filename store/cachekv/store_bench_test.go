package cachekv_test

import (
	"crypto/rand"
	"sort"
	"testing"

	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/store/cachekv"
	"github.com/cosmos/cosmos-sdk/store/dbadapter"
)

func benchmarkCacheKVStoreIterator(numKVs int, b *testing.B) {
	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	cstore := cachekv.NewStore(mem)
	keys := make([]string, numKVs, numKVs)

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

func BenchmarkCacheKVStoreIterator500(b *testing.B)    { benchmarkCacheKVStoreIterator(500, b) }
func BenchmarkCacheKVStoreIterator1000(b *testing.B)   { benchmarkCacheKVStoreIterator(1000, b) }
func BenchmarkCacheKVStoreIterator10000(b *testing.B)  { benchmarkCacheKVStoreIterator(10000, b) }
func BenchmarkCacheKVStoreIterator50000(b *testing.B)  { benchmarkCacheKVStoreIterator(50000, b) }
func BenchmarkCacheKVStoreIterator100000(b *testing.B) { benchmarkCacheKVStoreIterator(100000, b) }
