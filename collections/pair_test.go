package collections

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPair(t *testing.T) {
	keyCodec := PairKeyCodec(StringKey, StringKey)
	t.Run("stringify", func(t *testing.T) {
		s := keyCodec.Stringify(Join("a", "b"))
		require.Equal(t, `("a", "b")`, s)
		s = keyCodec.Stringify(PairPrefix[string, string]("a"))
		require.Equal(t, `("a", <nil>)`, s)
		s = keyCodec.Stringify(Pair[string, string]{})
		require.Equal(t, `(<nil>, <nil>)`, s)
	})

	t.Run("json", func(t *testing.T) {
		b, err := keyCodec.EncodeJSON(Join("k1", "k2"))
		require.NoError(t, err)
		require.Equal(t, []byte(`["k1","k2"]`), b)
	})
}

func TestPairRange(t *testing.T) {
	sk, ctx := deps()
	schema := NewSchemaBuilder(sk)
	pc := PairKeyCodec(StringKey, Uint64Key)
	m := NewMap(schema, NewPrefix(0), "pair", pc, Uint64Value)

	require.NoError(t, m.Set(ctx, Join("A", uint64(0)), 1))
	require.NoError(t, m.Set(ctx, Join("A", uint64(1)), 0))
	require.NoError(t, m.Set(ctx, Join("A", uint64(2)), 0))
	require.NoError(t, m.Set(ctx, Join("B", uint64(3)), 0))

	v, err := m.Get(ctx, Join("A", uint64(0)))
	require.NoError(t, err)
	require.Equal(t, uint64(1), v)

	// EXPECT only A1,2
	iter, err := m.Iterate(ctx, NewPrefixedPairRange[string, uint64]("A").StartInclusive(1).EndInclusive(2))
	require.NoError(t, err)
	keys, err := iter.Keys()
	require.NoError(t, err)
	require.Equal(t, []Pair[string, uint64]{Join("A", uint64(1)), Join("A", uint64(2))}, keys)

	// expect the whole "A" prefix
	iter, err = m.Iterate(ctx, NewPrefixedPairRange[string, uint64]("A"))
	require.NoError(t, err)
	keys, err = iter.Keys()
	require.NoError(t, err)
	require.Equal(t, []Pair[string, uint64]{Join("A", uint64(0)), Join("A", uint64(1)), Join("A", uint64(2))}, keys)

	// expect only A1
	iter, err = m.Iterate(ctx, NewPrefixedPairRange[string, uint64]("A").StartExclusive(0).EndExclusive(2))
	require.NoError(t, err)
	keys, err = iter.Keys()
	require.NoError(t, err)
	require.Equal(t, []Pair[string, uint64]{Join("A", uint64(1))}, keys)

	// expect A2, A1
	iter, err = m.Iterate(ctx, NewPrefixedPairRange[string, uint64]("A").Descending().StartExclusive(0).EndInclusive(2))
	require.NoError(t, err)
	keys, err = iter.Keys()
	require.NoError(t, err)
	require.Equal(t, []Pair[string, uint64]{Join("A", uint64(2)), Join("A", uint64(1))}, keys)
}

func TestPairDecodeConsumesAllBytes(t *testing.T) {
    // Prepare a buffer with a non-terminal first key and a terminal second key.
    // This ensures decoding must advance by the cumulative readTotal before decoding K2.
    pc := PairKeyCodec(StringKey, Uint64Key)

    // Encode using the codec to avoid manual mistakes
    buf := make([]byte, pc.Size(Join("A", uint64(1))))
    n, err := pc.Encode(buf, Join("A", uint64(1)))
    require.NoError(t, err)
    require.Equal(t, len(buf), n)

    // Iterator.Key enforces full consumption; use the decode directly as well
    read, key, err := pc.Decode(buf)
    require.NoError(t, err)
    require.Equal(t, len(buf), read)
    require.Equal(t, Join("A", uint64(1)), key)
}

func TestPairRange_ErrorPropagation(t *testing.T) {
    // Demonstrate that a stored error on PairRange is propagated by RangeValues
    rng := &PairRange[string, uint64]{}
    sentinel := fmt.Errorf("sentinel range error")
    rng.err = sentinel

    start, end, order, err := rng.RangeValues()
    require.Nil(t, start)
    require.Nil(t, end)
    require.Equal(t, Order(0), order)
    require.Equal(t, sentinel, err)
}
