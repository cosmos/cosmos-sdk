package accounts

import (
	"context"
	"testing"

	gogoproto "github.com/cosmos/gogoproto/proto"
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

var _ InterfaceRegistry = (*interfaceRegistry)(nil)

type interfaceRegistry struct{}

func (i interfaceRegistry) RegisterInterface(string, any, ...gogoproto.Message) {}

func (i interfaceRegistry) RegisterImplementations(any, ...gogoproto.Message) {}

func newKeeper(t *testing.T, accounts ...implementation.AccountCreatorFunc) (Keeper, context.Context) {
	t.Helper()
	ss, ctx := colltest.MockStore()
	m, err := NewKeeper(ss, eventService{}, nil, addressCodec{}, nil, nil, nil, interfaceRegistry{}, accounts...)
	require.NoError(t, err)
	return m, ctx
}

var _ QueryRouter = (*mockQuery)(nil)

type mockQuery func(ctx context.Context, req, resp implementation.ProtoMsg) error

func (m mockQuery) HybridHandlerByRequestName(_ string) []func(ctx context.Context, req, resp implementation.ProtoMsg) error {
	return []func(ctx context.Context, req, resp protoiface.MessageV1) error{func(ctx context.Context, req, resp protoiface.MessageV1) error {
		return m(ctx, req, resp)
	}}
}

var _ SignerProvider = (*mockSigner)(nil)

type mockSigner func(msg implementation.ProtoMsg) ([]byte, error)

func (m mockSigner) GetMsgV1Signers(msg gogoproto.Message) ([][]byte, proto.Message, error) {
	s, err := m(msg)
	if err != nil {
		return nil, nil, err
	}
	return [][]byte{s}, nil, nil
}

var _ MsgRouter = (*mockExec)(nil)

type mockExec func(ctx context.Context, msg, msgResp implementation.ProtoMsg) error

func (m mockExec) HybridHandlerByMsgName(_ string) func(ctx context.Context, req, resp protoiface.MessageV1) error {
	return func(ctx context.Context, req, resp protoiface.MessageV1) error {
		return m(ctx, req, resp)
	}
}

func (m mockExec) ResponseNameByRequestName(name string) string {
	return name + "Response"
}
