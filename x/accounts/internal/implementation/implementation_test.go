package implementation

import (
	"context"
	"testing"

	"github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
)

func TestImplementation(t *testing.T) {
	impl, err := newImplementation(collections.NewSchemaBuilderFromAccessor(OpenKVStore), TestAccount{})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("execute ok", func(t *testing.T) {
		resp, err := impl.Execute(ctx, &types.StringValue{Value: "test"})
		require.NoError(t, err)
		require.Equal(t, "testexecute-echo", resp.(*types.StringValue).Value)

		resp, err = impl.Execute(ctx, &types.BytesValue{Value: []byte("test")})
		require.NoError(t, err)
		require.Equal(t, "testbytes-execute-echo", string(resp.(*types.BytesValue).Value))
	})

	t.Run("execute - unknown message", func(t *testing.T) {
		_, err := impl.Execute(ctx, &types.Int32Value{Value: 1})
		require.ErrorIs(t, err, errInvalidMessage)
	})

	t.Run("init ok", func(t *testing.T) {
		resp, err := impl.Init(ctx, &types.StringValue{Value: "test"})
		require.NoError(t, err)
		require.Equal(t, "testinit-echo", resp.(*types.StringValue).Value)
	})

	t.Run("init - unknown message", func(t *testing.T) {
		_, err := impl.Init(ctx, &types.Int32Value{Value: 1})
		require.ErrorIs(t, err, errInvalidMessage)
	})

	t.Run("query ok", func(t *testing.T) {
		resp, err := impl.Query(ctx, &types.StringValue{Value: "test"})
		require.NoError(t, err)
		require.Equal(t, "testquery-echo", resp.(*types.StringValue).Value)

		resp, err = impl.Query(ctx, &types.BytesValue{Value: []byte("test")})
		require.NoError(t, err)
		require.Equal(t, "testbytes-query-echo", string(resp.(*types.BytesValue).Value))
	})

	t.Run("query - unknown message", func(t *testing.T) {
		_, err := impl.Query(ctx, &types.Int32Value{Value: 1})
		require.ErrorIs(t, err, errInvalidMessage)
	})
}
