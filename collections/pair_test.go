package collections

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPairRange(t *testing.T) {
	sk, ctx := deps()
	schema := NewSchema(sk)
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
	iter, err := m.Iterate(ctx, (&PairRange[string, uint64]{}).Prefix("A").StartInclusive(1).EndInclusive(2))
	require.NoError(t, err)
	keys, err := iter.Keys()
	require.NoError(t, err)
	require.Equal(t, []Pair[string, uint64]{Join("A", uint64(1)), Join("A", uint64(2))}, keys)

	// expect the whole "A" prefix
	iter, err = m.Iterate(ctx, new(PairRange[string, uint64]).Prefix("A"))
	require.NoError(t, err)
	keys, err = iter.Keys()
	require.NoError(t, err)
	require.Equal(t, []Pair[string, uint64]{Join("A", uint64(0)), Join("A", uint64(1)), Join("A", uint64(2))}, keys)

	// expect only A1
	iter, err = m.Iterate(ctx, new(PairRange[string, uint64]).Prefix("A").StartExclusive(0).EndExclusive(2))
	require.NoError(t, err)
	keys, err = iter.Keys()
	require.NoError(t, err)
	require.Equal(t, []Pair[string, uint64]{Join("A", uint64(1))}, keys)

	// expect A2, A1
	iter, err = m.Iterate(ctx, new(PairRange[string, uint64]).Descending().Prefix("A").StartExclusive(0).EndInclusive(2))
	require.NoError(t, err)
	keys, err = iter.Keys()
	require.Equal(t, []Pair[string, uint64]{Join("A", uint64(2)), Join("A", uint64(1))}, keys)
}
