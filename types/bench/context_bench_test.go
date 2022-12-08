package bench_test

import (
	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"testing"
)

func BenchmarkContext_KVStore(b *testing.B) {
	key := types.NewKVStoreKey(b.Name() + "_TestCacheContext")

	ctx := testutil.DefaultContext(key, types.NewTransientStoreKey("transient_"+b.Name()))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ctx.KVStore(key)
	}
}
