package coretesting

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKVStoreService(t *testing.T) {
	ctx := Context()
	svc1 := KVStoreService(ctx, "bank")

	// must panic
	t.Run("must panic on invalid ctx", func(t *testing.T) {
		require.Panics(t, func() {
			svc1.OpenKVStore(context.Background())
		})
	})

	t.Run("success", func(t *testing.T) {
		kv := svc1.OpenKVStore(ctx)
		require.NoError(t, kv.Set([]byte("key"), []byte("value")))

		value, err := kv.Get([]byte("key"))
		require.NoError(t, err)
		require.Equal(t, []byte("value"), value)
	})

	t.Run("contains module name", func(t *testing.T) {
		KVStoreService(ctx, "auth")
		_, ok := unwrap(ctx).stores["auth"]
		require.True(t, ok)
	})

}
