package collections

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestItem(t *testing.T) {
	sk, ctx := deps()
	item := NewItem(sk, NewPrefix("item"), Uint64Value)
	// set
	err := item.Set(ctx, 1000)
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
