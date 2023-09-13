package types_test

import (
	"testing"

	"cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/testutil"
)

func BenchmarkContext_KVStore(b *testing.B) {
	key := types.NewKVStoreKey(b.Name() + "_TestCacheContext")

	ctx := testutil.DefaultContext(key, types.NewTransientStoreKey("transient_"+b.Name()))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ctx.KVStore(key)
	}
}

func BenchmarkContext_TransientStore(b *testing.B) {
	key := types.NewKVStoreKey(b.Name() + "_TestCacheContext")

	ctx := testutil.DefaultContext(key, types.NewTransientStoreKey("transient_"+b.Name()))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ctx.TransientStore(key)
	}
}

func BenchmarkContext_CacheContext(b *testing.B) {
	key := types.NewKVStoreKey(b.Name() + "_TestCacheContext")

	ctx := testutil.DefaultContext(key, types.NewTransientStoreKey("transient_"+b.Name()))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ctx.CacheContext()
	}
}
