package collections

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeySet(t *testing.T) {
	sk, ctx := deps()
	schema := NewSchemaBuilder(sk)
	ks := NewKeySet(schema, NewPrefix("keyset"), "keyset", StringKey)

	// set
	require.NoError(t, ks.Set(ctx, "A"))
	require.NoError(t, ks.Set(ctx, "B"))
	require.NoError(t, ks.Set(ctx, "C"))
	require.NoError(t, ks.Set(ctx, "CC"))

	// exists
	exists, err := ks.Has(ctx, "CC")
	require.NoError(t, err)
	require.True(t, exists)

	// remove
	err = ks.Remove(ctx, "A")
	require.NoError(t, err)

	// non exists
	exists, err = ks.Has(ctx, "A")
	require.NoError(t, err)
	require.False(t, exists)

	// iter
	iter, err := ks.Iterate(ctx, nil)
	require.NoError(t, err)

	// iter next
	iter.Next()

	// iter key
	key, err := iter.Key()
	require.NoError(t, err)
	require.Equal(t, "C", key)

	// iter keys
	keys, err := iter.Keys()
	require.NoError(t, err)
	require.Equal(t, []string{"C", "CC"}, keys)

	// validity
	require.False(t, iter.Valid())
}

func Test_noValue(t *testing.T) {
	require.Equal(t, noValueValueType, noValueCodec.ValueType())
	require.Equal(t, noValueValueType, noValueCodec.Stringify(NoValue{}))

	b, err := noValueCodec.Encode(NoValue{})
	require.NoError(t, err)
	require.Equal(t, []byte{}, b)

	nv, err := noValueCodec.Decode(b)
	require.NoError(t, err)
	require.Equal(t, NoValue{}, nv)

	_, err = noValueCodec.Decode([]byte("bad"))
	require.ErrorIs(t, err, ErrEncoding)
}
