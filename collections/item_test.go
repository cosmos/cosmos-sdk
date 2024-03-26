package collections

import (
	"testing"

	"cosmossdk.io/server/v2/stf"
	"github.com/stretchr/testify/require"
)

func TestItem(t *testing.T) {
	sk, ctx := deps()
	schemaBuilder := NewSchemaBuilder(sk)
	item := NewItem(schemaBuilder, NewPrefix("item"), "item", Uint64Value)
	_, err := schemaBuilder.Build()
	require.NoError(t, err)

	// set
	err = item.Set(ctx, 1000)
	require.NoError(t, err)

	// get
	i, err := item.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(1000), i)

	// has
	has, err := item.Has(ctx)
	require.NoError(t, err)
	require.True(t, has)

	// remove
	err = item.Remove(ctx)
	require.NoError(t, err)
	has, err = item.Has(ctx)
	require.NoError(t, err)
	require.False(t, has)
}

func TestCacheCtx(t *testing.T) {
	storeService := stf.NewStoreService("test")
	ctx := stf.NewExecutionContext()
	schemaBuilder := NewSchemaBuilder(storeService)
	item := NewItem(schemaBuilder, NewPrefix("item"), "item", Uint64Value)
	_, err := schemaBuilder.Build()
	require.NoError(t, err)

	// set
	err = item.Set(ctx, 1000)
	require.NoError(t, err)
	// Check if item was cached
	cacheContainer := ctx.Cache.GetContainer([]byte("test"))
	v, ok := cacheContainer.Get([]byte("item"))
	require.True(t, ok)
	require.Equal(t, uint64(1000), v.(uint64))

	// get
	i, err := item.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(1000), i)

	// has
	has, err := item.Has(ctx)
	require.NoError(t, err)
	require.True(t, has)

	// remove
	err = item.Remove(ctx)
	require.NoError(t, err)
	has, err = item.Has(ctx)
	require.NoError(t, err)
	require.False(t, has)
	// Check if item was removed from cache
	cacheContainer = ctx.Cache.GetContainer([]byte("test"))
	v, ok = cacheContainer.Get([]byte("item"))
	require.False(t, ok)

}
