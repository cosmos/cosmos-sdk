package collections

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestItem(t *testing.T) {
	sk, ctx := deps()
	item := NewItem(sk, NewPrefix("item"), Uint64Value)
	// set
	item.Set(ctx, 1000)
	// get
	i, err := item.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(1000), i)
	// has
	require.True(t, item.Has(ctx))
	// remove
	item.Remove(ctx)
	require.False(t, item.Has(ctx))
}
