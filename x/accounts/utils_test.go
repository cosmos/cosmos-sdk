package accounts

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoiface"

	"cosmossdk.io/collections/colltest"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/event"
	"cosmossdk.io/x/accounts/internal/implementation"
)

var _ address.Codec = (*addressCodec)(nil)

type addressCodec struct{}

func (a addressCodec) StringToBytes(text string) ([]byte, error) { return []byte(text), nil }
func (a addressCodec) BytesToString(bz []byte) (string, error)   { return string(bz), nil }

type eventService struct{}

func (e eventService) Emit(ctx context.Context, event protoiface.MessageV1) error { return nil }

func (e eventService) EmitKV(ctx context.Context, eventType string, attrs ...event.Attribute) error {
	return nil
}

func (e eventService) EmitNonConsensus(ctx context.Context, event protoiface.MessageV1) error {
	return nil
}

func (e eventService) EventManager(ctx context.Context) event.Manager { return e }

func newKeeper(t *testing.T, accounts ...implementation.AccountCreatorFunc) (Keeper, context.Context) {
	t.Helper()
	ss, ctx := colltest.MockStore()
	m, err := NewKeeper(ss, eventService{}, nil, addressCodec{}, nil, nil, nil, accounts...)
	require.NoError(t, err)
	return m, ctx
}

var _ QueryRouter = (*mockQuery)(nil)

type mockQuery func(ctx context.Context, req, resp proto.Message) error

func (m mockQuery) HybridHandlerByRequestName(_ string) []func(ctx context.Context, req, resp protoiface.MessageV1) error {
	return []func(ctx context.Context, req, resp protoiface.MessageV1) error{func(ctx context.Context, req, resp protoiface.MessageV1) error {
		return m(ctx, req.(proto.Message), resp.(proto.Message))
	}}
}

var _ SignerProvider = (*mockSigner)(nil)

type mockSigner func(msg proto.Message) ([]byte, error)

func (m mockSigner) GetSigners(msg proto.Message) ([][]byte, error) {
	s, err := m(msg)
	if err != nil {
		return nil, err
	}
	return [][]byte{s}, nil
}

var _ MsgRouter = (*mockExec)(nil)

type mockExec func(ctx context.Context, msg, msgResp proto.Message) error

func (m mockExec) HybridHandlerByMsgName(_ string) func(ctx context.Context, req, resp protoiface.MessageV1) error {
	return func(ctx context.Context, req, resp protoiface.MessageV1) error {
		return m(ctx, req.(proto.Message), resp.(proto.Message))
	}
}

func (m mockExec) ResponseNameByRequestName(name string) string {
	return name + "Response"
}
