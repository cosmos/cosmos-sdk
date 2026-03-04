package cache

import (
	"testing"

	"cosmossdk.io/store/types"
)

func freshMgr() *KVStoreCacheManager {
	return &KVStoreCacheManager{
		caches: map[string]types.KVStore{
			"a1":           nil,
			"alalalalalal": nil,
		},
	}
}

func populate(mgr *KVStoreCacheManager) {
	mgr.caches["this one"] = (types.KVStore)(nil)
	mgr.caches["those ones are the ones"] = (types.KVStore)(nil)
	mgr.caches["very huge key right here and there are we going to ones are the ones"] = (types.KVStore)(nil)
}

func BenchmarkReset(b *testing.B) {
	b.ReportAllocs()
	mgr := freshMgr()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		mgr.Reset()
		if len(mgr.caches) != 0 {
			b.Fatal("Reset failed")
		}
		populate(mgr)
		if len(mgr.caches) == 0 {
			b.Fatal("populate failed")
		}
		mgr.Reset()
		if len(mgr.caches) != 0 {
			b.Fatal("Reset failed")
		}
	}

	if mgr == nil {
		b.Fatal("Impossible condition")
	}
}
