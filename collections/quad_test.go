package collections_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/colltest"
	coretesting "cosmossdk.io/core/testing"
)

func TestQuad(t *testing.T) {
	kc := collections.QuadKeyCodec(collections.Uint64Key, collections.StringKey, collections.BytesKey, collections.BoolKey)

	t.Run("conformance", func(t *testing.T) {
		colltest.TestKeyCodec(t, kc, collections.Join4(uint64(1), "2", []byte("3"), true))
	})
}

func TestQuadRange(t *testing.T) {
	ctx := coretesting.Context()
	sk := coretesting.KVStoreService(ctx, "test")
	schema := collections.NewSchemaBuilder(sk)
	// this is a key composed of 4 parts: uint64, string, []byte, bool
	kc := collections.QuadKeyCodec(collections.Uint64Key, collections.StringKey, collections.BytesKey, collections.BoolKey)

	keySet := collections.NewKeySet(schema, collections.NewPrefix(0), "Quad", kc)

	keys := []collections.Quad[uint64, string, []byte, bool]{
		collections.Join4(uint64(1), "A", []byte("1"), true),
		collections.Join4(uint64(1), "A", []byte("2"), true),
		collections.Join4(uint64(1), "B", []byte("3"), false),
		collections.Join4(uint64(2), "B", []byte("4"), false),
	}

	for _, k := range keys {
		require.NoError(t, keySet.Set(ctx, k))
	}

	// we prefix over (1) we expect 3 results
	iter, err := keySet.Iterate(ctx, collections.NewPrefixedQuadRange[uint64, string, []byte, bool](uint64(1)))
	require.NoError(t, err)
	gotKeys, err := iter.Keys()
	require.NoError(t, err)
	require.Equal(t, keys[:3], gotKeys)

	// we super prefix over Join(1, "A") we expect 2 results
	iter, err = keySet.Iterate(ctx, collections.NewSuperPrefixedQuadRange[uint64, string, []byte, bool](1, "A"))
	require.NoError(t, err)
	gotKeys, err = iter.Keys()
	require.NoError(t, err)
	require.Equal(t, keys[:2], gotKeys)

	// we super prefix 3 over Join(1, "A", []byte("1")) we expect 1 result
	iter, err = keySet.Iterate(ctx, collections.NewSuperPrefixedQuadRange3[uint64, string, []byte, bool](1, "A", []byte("1")))
	require.NoError(t, err)
	gotKeys, err = iter.Keys()
	require.NoError(t, err)
	require.Equal(t, keys[:1], gotKeys)
}
