package cachekv

import (
	"strconv"
	"testing"

	"cosmossdk.io/store/internal/btree"
)

func BenchmarkLargeUnsortedMisses(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		store := generateStore()
		b.StartTimer()

		for k := 0; k < 10000; k++ {
			// cache has A + Z values
			// these are within range, but match nothing
			store.dirtyItems([]byte("B1"), []byte("B2"))
		}
	}
}

func generateStore() *Store {
	cache := map[string]*cValue[[]byte]{}
	unsorted := map[string]struct{}{}
	for i := 0; i < 5000; i++ {
		key := "A" + strconv.Itoa(i)
		unsorted[key] = struct{}{}
		cache[key] = &cValue[[]byte]{}
	}

	for i := 0; i < 5000; i++ {
		key := "Z" + strconv.Itoa(i)
		unsorted[key] = struct{}{}
		cache[key] = &cValue[[]byte]{}
	}

	return &GStore[[]byte]{
		cache:         cache,
		unsortedCache: unsorted,
		sortedCache:   btree.NewBTree[[]byte](),
	}
}
