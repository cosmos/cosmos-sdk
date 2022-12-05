package collections

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	sk, ctx := deps()
	m := NewMap(sk, NewPrefix("hi"), Uint64Key, Uint64Value)

	// test not has
	has, err := m.Has(ctx, 1)
	require.NoError(t, err)
	require.False(t, has)
	// test get error
	_, err = m.Get(ctx, 1)
	require.ErrorIs(t, err, ErrNotFound)

	// test set/get
	err = m.Set(ctx, 1, 100)
	require.NoError(t, err)
	v, err := m.Get(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, uint64(100), v)

	// test remove
	err = m.Remove(ctx, 1)
	require.NoError(t, err)
	has, err = m.Has(ctx, 1)
	require.NoError(t, err)
	require.False(t, has)
}
