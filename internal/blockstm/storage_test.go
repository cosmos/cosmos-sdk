package blockstm

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
)

// TestGCachedStorage covers hit/miss memoization for both V=[]byte and V=any,
// the ObjKV case is a regression guard against the nil-interface assertion panic.
func TestGCachedStorage(t *testing.T) {
	t.Run("KV", func(t *testing.T) {
		parent := NewMemDB()
		parent.Set([]byte("k"), []byte("v"))
		assertCache(t, parent, storetypes.BytesIsZero, []byte("v"), nil)
	})

	t.Run("ObjKV", func(t *testing.T) {
		parent := NewObjMemDB()
		parent.Set([]byte("k"), "v")
		assertCache[any](t, parent, storetypes.AnyIsZero, "v", nil)
	})
}

func assertCache[V any](t *testing.T, parent GStorage[V], isZero func(V) bool, hitValue, missValue V) {
	t.Helper()
	counted := &countingStorage[V]{GStorage: parent}
	cached := NewGCachedStorage(counted, isZero)

	for i := 0; i < 3; i++ {
		require.Equal(t, hitValue, cached.Get([]byte("k")))
		require.Equal(t, missValue, cached.Get([]byte("missing")))
	}
	require.EqualValues(t, 2, counted.gets.Load(), "each distinct key reads parent exactly once")

	// Has uses the populated cache; no extra parent probes.
	require.True(t, cached.Has([]byte("k")))
	require.False(t, cached.Has([]byte("missing")))
	require.EqualValues(t, 2, counted.gets.Load()+counted.hasOps.Load(),
		"Has should reuse the Get cache without extra parent probes")
}

type countingStorage[V any] struct {
	GStorage[V]
	gets   atomic.Int64
	hasOps atomic.Int64
}

func (c *countingStorage[V]) Get(key []byte) V {
	c.gets.Add(1)
	return c.GStorage.Get(key)
}

func (c *countingStorage[V]) Has(key []byte) bool {
	c.hasOps.Add(1)
	return c.GStorage.Has(key)
}
