package blockstm

import (
	"fmt"
	"testing"
)

func BenchmarkCacheMergeIterator(b *testing.B) {
	for _, keyCount := range []int{100, 1_000, 10_000} {
		parent, cache := benchmarkMergeIteratorStores(b, keyCount)

		for _, ascending := range []bool{true, false} {
			for _, readValue := range []bool{false, true} {
				name := "ascending/key"
				if !ascending {
					name = "descending/key"
				}
				if readValue {
					name += "+value"
				}

				b.Run(fmt.Sprintf("keys-%d/%s", keyCount, name), func(b *testing.B) {
					b.ReportAllocs()
					b.SetBytes(int64(keyCount))
					for b.Loop() {
						var parentIter, cacheIter = parent.Iterator(nil, nil), cache.Iterator(nil, nil)
						if !ascending {
							parentIter = parent.ReverseIterator(nil, nil)
							cacheIter = cache.ReverseIterator(nil, nil)
						}

						iter := NewCacheMergeIterator(parentIter, cacheIter, ascending, nil, bytesIsZero)
						for ; iter.Valid(); iter.Next() {
							_ = iter.Key()
							if readValue {
								_ = iter.Value()
							}
						}
						if err := iter.Close(); err != nil {
							b.Fatal(err)
						}
					}
				})
			}
		}
	}
}

func benchmarkMergeIteratorStores(b *testing.B, keyCount int) (*MemDB, *MemDB) {
	b.Helper()
	parent := NewMemDB()
	cache := NewWriteSet(bytesIsZero, bytesValueLen)
	for i := range keyCount {
		key := []byte(fmt.Sprintf("key-%08d", i))
		parent.Set(key, []byte("parent"))

		switch i % 3 {
		case 0:
			cache.OverlaySet(key, nil) // cache deletion
		case 1:
			cache.OverlaySet(key, []byte("cache")) // cache shadow
		}
	}
	return parent, cache
}
