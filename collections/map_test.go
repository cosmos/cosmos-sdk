package collections

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	sk, ctx := deps()
	schemaBuilder := NewSchemaBuilder(sk)
	m := NewMap(schemaBuilder, NewPrefix("hi"), "m", Uint64Key, Uint64Value)
	_, err := schemaBuilder.Build()
	require.NoError(t, err)

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

func TestMap_IterateRaw(t *testing.T) {
	sk, ctx := deps()
	// safety check to ensure prefix boundaries are not crossed
	sk.OpenKVStore(ctx).Set([]byte{0x0, 0x0}, []byte("before prefix"))
	sk.OpenKVStore(ctx).Set([]byte{0x2, 0x0}, []byte("after prefix"))

	sb := NewSchemaBuilder(sk)

	m := NewMap(sb, NewPrefix(1), "m", Uint64Key, Uint64Value)
	require.NoError(t, m.Set(ctx, 0, 0))
	require.NoError(t, m.Set(ctx, 1, 1))
	require.NoError(t, m.Set(ctx, 2, 2))

	// test non nil end in ascending order
	twoBigEndian, err := encodeKeyWithPrefix(nil, Uint64Key, 2)
	require.NoError(t, err)
	iter, err := m.IterateRaw(ctx, nil, twoBigEndian, OrderAscending)
	require.NoError(t, err)
	defer iter.Close()

	keys, err := iter.Keys()
	require.NoError(t, err)

	require.Equal(t, []uint64{0, 1}, keys)

	// test nil end in reverse
	iter, err = m.IterateRaw(ctx, nil, nil, OrderDescending)
	require.NoError(t, err)
	defer iter.Close()

	keys, err = iter.Keys()
	require.NoError(t, err)
	require.Equal(t, []uint64{2, 1, 0}, keys)
}

func Test_encodeKey(t *testing.T) {
	prefix := "prefix"
	number := []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	expectedKey := append([]byte(prefix), number...)

	gotKey, err := encodeKeyWithPrefix(NewPrefix(prefix).Bytes(), Uint64Key, 0)
	require.NoError(t, err)
	require.Equal(t, expectedKey, gotKey)
}
