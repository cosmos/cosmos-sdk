package accounts

import (
	"context"
	"errors"
	"testing"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/runtime/protoiface"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/testing"
	"cosmossdk.io/x/accounts/internal/implementation"
	"cosmossdk.io/x/accounts/testing/mockmodule"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ address.Codec = (*addressCodec)(nil)

type addressCodec struct{}

func (a addressCodec) StringToBytes(text string) ([]byte, error) { return []byte(text), nil }
func (a addressCodec) BytesToString(bz []byte) (string, error)   { return string(bz), nil }

type eventService struct{}

func (e eventService) Emit(event protoiface.MessageV1) error { return nil }

func (e eventService) EmitKV(eventType string, attrs ...event.Attribute) error {
	return nil
}

func (e eventService) EventManager(ctx context.Context) event.Manager { return e }

var _ CoinTransferer = (*coinTransferer)(nil)

type coinTransferer struct{}

func (c coinTransferer) MakeTransferCoinsMessage(ctx context.Context, from, to []byte, amount sdk.Coins) (implementation.ProtoMsg, implementation.ProtoMsg, error) {
	return nil, nil, errors.New("do not call")
}

func newKeeper(t *testing.T, accounts ...implementation.AccountCreatorFunc) (Keeper, context.Context) {
	t.Helper()

	addressCodec := addressCodec{}
	ir, err := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles: gogoproto.HybridResolver,
		SigningOptions: signing.Options{
			FileResolver:          gogoproto.HybridResolver,
			TypeResolver:          protoregistry.GlobalTypes,
			AddressCodec:          addressCodec,
			ValidatorAddressCodec: addressCodec,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx := coretesting.Context()
	ss := coretesting.KVStoreService(ctx, "test")
	env := appmodule.Environment{
		QueryRouterService: mockmodule.MockQueryRouter(),
		MsgRouterService:   mockmodule.MockMsgRouter(),
		KVStoreService:     ss,
		EventService:       eventService{},
	}
	env.EventService = eventService{}
	m, err := NewKeeper(codec.NewProtoCodec(ir), env, addressCodec, ir, coinTransferer{}, accounts...)
	m.getSenders = func(msg gogoproto.Message) ([][]byte, protoreflect.Message, error) {
		typedMsg := msg.(*mockmodule.MsgEcho)
		return [][]byte{[]byte(typedMsg.Sender)}, nil, nil
	}
	require.NoError(t, err)
	return m, ctx
}
