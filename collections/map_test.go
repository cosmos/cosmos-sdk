package collections

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	sk, ctx := deps()
	m := NewMap(sk, NewPrefix("hi"), Uint64Key, Uint64Value)

	// test not has
	require.False(t, m.Has(ctx, 1))
	// test get error
	_, err := m.Get(ctx, 1)
	require.ErrorIs(t, err, ErrNotFound{
		HumanizedKey: "1",
		RawKey:       m.encodeKey(1),
		ValueType:    m.vc.ValueType(),
	})

	// test set/get
	m.Set(ctx, 1, 100)
	v, err := m.Get(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, uint64(100), v)

	// test remove
	m.Remove(ctx, 1)
	require.False(t, m.Has(ctx, 1))

	// test get or
}
