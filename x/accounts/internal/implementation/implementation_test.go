package implementation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestImplementation(t *testing.T) {
	impl, err := NewImplementation(TestAccount{})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("execute ok", func(t *testing.T) {
		resp, err := impl.Execute(ctx, &wrapperspb.StringValue{Value: "test"})
		require.NoError(t, err)
		require.Equal(t, "testexecute-echo", resp.(*wrapperspb.StringValue).Value)

		resp, err = impl.Execute(ctx, &wrapperspb.BytesValue{Value: []byte("test")})
		require.NoError(t, err)
		require.Equal(t, "testbytes-execute-echo", string(resp.(*wrapperspb.BytesValue).Value))
	})

	t.Run("execute - unknown message", func(t *testing.T) {
		_, err := impl.Execute(ctx, &wrapperspb.Int32Value{Value: 1})
		require.ErrorIs(t, err, errInvalidMessage)
	})

	t.Run("init ok", func(t *testing.T) {
		resp, err := impl.Init(ctx, &wrapperspb.StringValue{Value: "test"})
		require.NoError(t, err)
		require.Equal(t, "testinit-echo", resp.(*wrapperspb.StringValue).Value)
	})

	t.Run("init - unknown message", func(t *testing.T) {
		_, err := impl.Init(ctx, &wrapperspb.Int32Value{Value: 1})
		require.ErrorIs(t, err, errInvalidMessage)
	})

	t.Run("query ok", func(t *testing.T) {
		resp, err := impl.Query(ctx, &wrapperspb.StringValue{Value: "test"})
		require.NoError(t, err)
		require.Equal(t, "testquery-echo", resp.(*wrapperspb.StringValue).Value)

		resp, err = impl.Query(ctx, &wrapperspb.BytesValue{Value: []byte("test")})
		require.NoError(t, err)
		require.Equal(t, "testbytes-query-echo", string(resp.(*wrapperspb.BytesValue).Value))
	})

	t.Run("query - unknown message", func(t *testing.T) {
		_, err := impl.Query(ctx, &wrapperspb.Int32Value{Value: 1})
		require.ErrorIs(t, err, errInvalidMessage)
	})

	t.Run("all - not a protobuf message", func(t *testing.T) {
		_, err := impl.Execute(ctx, "test")
		require.ErrorIs(t, err, errInvalidMessage)
		_, err = impl.Query(ctx, "test")
		require.ErrorIs(t, err, errInvalidMessage)
		_, err = impl.Init(ctx, "test")
		require.ErrorIs(t, err, errInvalidMessage)
	})
}
