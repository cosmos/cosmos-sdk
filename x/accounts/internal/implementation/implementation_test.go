package implementation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
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

	// schemas
	t.Run("decode init request - ok", func(t *testing.T) {
		want := &wrapperspb.StringValue{Value: "test"}
		req, err := proto.Marshal(want)
		require.NoError(t, err)

		got, err := impl.DecodeInitRequest(req)
		require.NoError(t, err)
		require.True(t, proto.Equal(want, got.(protoreflect.ProtoMessage)))
	})

	t.Run("encode init response - ok", func(t *testing.T) {
		want := &wrapperspb.StringValue{Value: "test"}

		gotBytes, err := impl.EncodeInitResponse(want)
		require.NoError(t, err)

		wantBytes, err := proto.Marshal(want)
		require.NoError(t, err)

		require.Equal(t, wantBytes, gotBytes)
	})

	t.Run("encode init response - invalid message", func(t *testing.T) {
		_, err := impl.EncodeInitResponse([]byte("invalid"))
		require.ErrorIs(t, err, errInvalidMessage)
	})

	t.Run("decode execute request - ok", func(t *testing.T) {
		wantReq := &wrapperspb.StringValue{Value: "test"}
		anyBPReq, err := anypb.New(wantReq)
		require.NoError(t, err)
		reqBytes, err := proto.Marshal(anyBPReq)
		require.NoError(t, err)
		gotReq, err := impl.DecodeExecuteRequest(reqBytes)
		require.NoError(t, err)
		require.True(t, proto.Equal(wantReq, gotReq.(protoreflect.ProtoMessage)))
	})

	t.Run("decode execute request - invalid message", func(t *testing.T) {
		req := wrapperspb.Double(1)
		anyPBReq, err := anypb.New(req)
		require.NoError(t, err)
		reqBytes, err := proto.Marshal(anyPBReq)
		require.NoError(t, err)
		_, err = impl.DecodeExecuteRequest(reqBytes)
		require.ErrorIs(t, err, errInvalidMessage)
	})

	t.Run("encode execute response - ok", func(t *testing.T) {
		resp := &wrapperspb.StringValue{Value: "test"}
		gotRespBytes, err := impl.EncodeExecuteResponse(resp)
		require.NoError(t, err)
		anyPBResp, err := anypb.New(resp)
		require.NoError(t, err)
		wantRespBytes, err := proto.Marshal(anyPBResp)
		require.NoError(t, err)
		require.Equal(t, wantRespBytes, gotRespBytes)
	})

	t.Run("encode execute response - not a protobuf message", func(t *testing.T) {
		_, err := impl.EncodeExecuteResponse("test")
		require.ErrorIs(t, err, errInvalidMessage)
		require.ErrorContains(t, err, "expected protoreflect.Message")
	})

	t.Run("encode execute response - not part of the message set", func(t *testing.T) {
		_, err := impl.EncodeExecuteResponse(&wrapperspb.DoubleValue{Value: 1})
		require.ErrorIs(t, err, errInvalidMessage)
		require.ErrorContains(t, err, "not part of message set")
	})
}
