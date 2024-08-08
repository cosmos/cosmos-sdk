package types_test

import (
	"testing"

	"cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/testutil"
)

const (
	tcc       = "_TestCacheContext"
	transient = "transient_"
)

func BenchmarkContext_KVStore(b *testing.B) {
	key := types.NewKVStoreKey(b.Name() + tcc)

	ctx := testutil.DefaultContext(key.Name())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ctx.KVStore(key)
	}
}

func BenchmarkContext_CacheContext(b *testing.B) {
	key := types.NewKVStoreKey(b.Name() + tcc)

	ctx := testutil.DefaultContext(key.Name())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ctx.CacheContext()
	}
}
