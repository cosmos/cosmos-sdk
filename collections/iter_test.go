package collections

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIteratorBasic(t *testing.T) {
	sk, ctx := deps()
	m := NewMap(sk, NewPrefix("some super amazing prefix"), StringKey, Uint64Value)

	for i := uint64(1); i <= 2; i++ {
		require.NoError(t, m.Set(ctx, fmt.Sprintf("%d", i), i))
	}

	iter, err := m.Iterate(ctx, nil)
	require.NoError(t, err)
	defer iter.Close()

	// key codec
	key, err := iter.Key()
	require.NoError(t, err)
	require.Equal(t, "1", key)

	// value codec
	value, err := iter.Value()
	require.NoError(t, err)
	require.Equal(t, uint64(1), value)

	// assert expected prefixing on iter
	require.Equal(t, m.prefix, iter.iter.Key()[:len(m.prefix)])

	// advance iter
	iter.Next()
	require.True(t, iter.Valid())

	// key 2
	key, err = iter.Key()
	require.NoError(t, err)
	require.Equal(t, "2", key)

	// value 2
	value, err = iter.Value()
	require.NoError(t, err)
	require.Equal(t, uint64(2), value)

	// call next, invalid
	iter.Next()
	require.False(t, iter.Valid())
	// close no errors
	require.NoError(t, iter.Close())
}

func TestIteratorKeyValues(t *testing.T) {
	sk, ctx := deps()
	m := NewMap(sk, NewPrefix("some super amazing prefix"), StringKey, Uint64Value)

	for i := uint64(0); i <= 5; i++ {
		require.NoError(t, m.Set(ctx, fmt.Sprintf("%d", i), i))
	}

	// test keys
	iter, err := m.Iterate(ctx, nil)
	require.NoError(t, err)
	keys, err := iter.Keys()
	require.NoError(t, err)

	for i, key := range keys {
		require.Equal(t, fmt.Sprintf("%d", i), key)
	}
	require.NoError(t, iter.Close())
	require.False(t, iter.Valid())

	// test values
	iter, err = m.Iterate(ctx, nil)
	require.NoError(t, err)
	values, err := iter.Values()
	require.NoError(t, err)

	for i, value := range values {
		require.Equal(t, uint64(i), value)
	}
	require.NoError(t, iter.Close())
	require.False(t, iter.Valid())

	// test key value pairings
	iter, err = m.Iterate(ctx, nil)
	require.NoError(t, err)
	kvs, err := iter.KeyValues()
	require.NoError(t, err)

	for i, kv := range kvs {
		require.Equal(t, fmt.Sprintf("%d", i), kv.Key)
		require.Equal(t, uint64(i), kv.Value)
	}
	require.NoError(t, iter.Close())
	require.False(t, iter.Valid())
}

func TestIteratorRanging(t *testing.T) {
	sk, ctx := deps()
	m := NewMap(sk, NewPrefix("cool"), Uint64Key, Uint64Value)

	for i := uint64(0); i <= 7; i++ {
		require.NoError(t, m.Set(ctx, i, i))
	}

	// let's range (1-5]; expected: 2..5
	iter, err := m.Iterate(ctx, (&Range[uint64]{}).StartExclusive(1).EndInclusive(5))
	require.NoError(t, err)
	result, err := iter.Keys()
	require.NoError(t, err)
	require.Equal(t, []uint64{2, 3, 4, 5}, result)

	// let's range [1-5); expected 1..4
	iter, err = m.Iterate(ctx, (&Range[uint64]{}).StartInclusive(1).EndExclusive(5))
	require.NoError(t, err)
	result, err = iter.Keys()
	require.NoError(t, err)
	require.Equal(t, []uint64{1, 2, 3, 4}, result)

	// let's range [1-5) descending; expected 4..1
	iter, err = m.Iterate(ctx, (&Range[uint64]{}).StartInclusive(1).EndExclusive(5).Descending())
	require.NoError(t, err)
	result, err = iter.Keys()
	require.NoError(t, err)
	require.Equal(t, []uint64{4, 3, 2, 1}, result)
}

func TestIteratorPrefixRanging(t *testing.T) {
	sk, ctx := deps()
	m := NewMap(sk, NewPrefix("cool"), StringKey, Uint64Value)
	require.NoError(t, m.Set(ctx, "AA1", 1))
	require.NoError(t, m.Set(ctx, "AA2", 2))
	require.NoError(t, m.Set(ctx, "AA3", 3))
	require.NoError(t, m.Set(ctx, "AB1", 4))

	rng := new(Range[string]).
		Prefix("AA").
		StartExclusive("1").
		EndInclusive("3")
	// expected AA2,AA3
	iter, err := m.Iterate(ctx, rng)
	require.NoError(t, err)
	keys, err := iter.Keys()
	require.NoError(t, err)
	require.Equal(t, []string{"AA2", "AA3"}, keys)
}
