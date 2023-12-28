package tree

import (
	"context"
	"testing"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/colltest"
	"github.com/stretchr/testify/require"
)

func TestMap_Higher(t *testing.T) {
	ctx, tree := newTree(t, 10, 20, 30, 40, 50)
	assert := func(val uint64, want uint64) {
		got, err := tree.Higher(ctx, val)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, int(want), int(*got))
	}

	notFound := func(val uint64) {
		got, err := tree.Higher(ctx, val)
		require.NoError(t, err)
		require.Nil(t, got)
	}

	assert(5, 10)
	assert(10, 20)
	assert(11, 20)
	assert(20, 30)
	assert(49, 50)
	notFound(50)
	notFound(51)
}

func TestInsertion_LL(t *testing.T) {
	ctx, tree := newTree(t)
	//	require.Equal(t, 0, height(t, ctx, tree))

	require.NoError(t, tree.Set(ctx, 3, 3))
	require.Equal(t, 1, height(t, ctx, tree))

	require.NoError(t, tree.Set(ctx, 2, 2))
	require.Equal(t, 2, height(t, ctx, tree))

	require.NoError(t, tree.Set(ctx, 1, 1))
	require.Equal(t, 2, height(t, ctx, tree))
}

func TestInsertion_RR(t *testing.T) {
	ctx, tree := newTree(t)
	//	require.Equal(t, 0, height(t, ctx, tree))

	require.NoError(t, tree.Set(ctx, 1, 1))
	require.Equal(t, 1, height(t, ctx, tree))

	require.NoError(t, tree.Set(ctx, 2, 2))
	require.Equal(t, 2, height(t, ctx, tree))

	require.NoError(t, tree.Set(ctx, 3, 3))
	require.Equal(t, 2, height(t, ctx, tree))
}

func TestIteration(t *testing.T) {
	ctx, tree := newTree(t, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	err := tree.Walk(ctx, nil, func(key, _ uint64) (stop bool, err error) {
		t.Logf("%d", key)
		return false, nil
	})
	require.NoError(t, err)
}

func newTree(t *testing.T, vals ...uint64) (context.Context, Map[uint64, uint64]) {
	ss, ctx := colltest.MockStore()
	sb := collections.NewSchemaBuilder(ss)
	tree := NewMap(sb, collections.NewPrefix(0), "test", collections.Uint64Key, collections.Uint64Value)
	_, err := sb.Build()
	require.NoError(t, err)
	for _, v := range vals {
		require.NoError(t, tree.Set(ctx, v, v))
	}
	return ctx, tree
}

func debugTree[K, V any](t *testing.T, ctx context.Context, tree Map[K, V]) {
	err := tree.tree.Walk(ctx, nil, func(index uint64, elem Node[K]) (stop bool, err error) {
		t.Logf("index: %d, value: %#v", index, tree.tree.ValueCodec().Stringify(elem))
		return false, err
	})
	require.NoError(t, err)
}

func height[K, V any](t *testing.T, ctx context.Context, tree Map[K, V]) int {
	root, err := tree.getRoot(ctx)
	require.NoError(t, err)

	node, err := tree.getNode(ctx, root)
	require.NoError(t, err)

	return int(node.Height)
}
