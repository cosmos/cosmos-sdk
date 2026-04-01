package cache_test

import (
	"math/rand"
	"testing"

	"github.com/cosmos/iavl/cache"
)

func BenchmarkAdd(b *testing.B) {
	b.ReportAllocs()
	testcases := map[string]struct {
		cacheMax int
		keySize  int
	}{
		"small - max: 10K, key size - 10b": {
			cacheMax: 10000,
			keySize:  10,
		},
		"med - max: 100K, key size 20b": {
			cacheMax: 100000,
			keySize:  20,
		},
		"large - max: 1M, key size 30b": {
			cacheMax: 1000000,
			keySize:  30,
		},
	}

	for name, tc := range testcases {
		cache := cache.New(tc.cacheMax)
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				key := randBytes(tc.keySize)
				b.StartTimer()

				_ = cache.Add(&testNode{
					key: key,
				})
			}
		})
	}
}

func BenchmarkRemove(b *testing.B) {
	b.ReportAllocs()

	cache := cache.New(1000)
	existentKeyMirror := [][]byte{}
	// Populate cache
	for i := 0; i < 50; i++ {
		key := randBytes(1000)

		existentKeyMirror = append(existentKeyMirror, key)

		cache.Add(&testNode{
			key: key,
		})
	}

	randSeed := 498727689 // For deterministic tests
	r := rand.New(rand.NewSource(int64(randSeed)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := existentKeyMirror[r.Intn(len(existentKeyMirror))]
		_ = cache.Remove(key)
	}
}
