package accounts

import (
	"context"
	"testing"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoregistry"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/event"
	coretesting "cosmossdk.io/core/testing"
	coretransaction "cosmossdk.io/core/transaction"
	"cosmossdk.io/x/accounts/internal/implementation"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
)

var _ address.Codec = (*addressCodec)(nil)

type addressCodec struct{}

func (a addressCodec) StringToBytes(text string) ([]byte, error) { return []byte(text), nil }
func (a addressCodec) BytesToString(bz []byte) (string, error)   { return string(bz), nil }

type eventService struct{}

func (e eventService) Emit(event gogoproto.Message) error { return nil }

func (e eventService) EmitKV(eventType string, attrs ...event.Attribute) error {
	return nil
}

func (e eventService) EventManager(ctx context.Context) event.Manager { return e }

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
	msgRouter := baseapp.NewMsgServiceRouter()
	msgRouter.SetInterfaceRegistry(ir)
	queryRouter := baseapp.NewGRPCQueryRouter()
	queryRouter.SetInterfaceRegistry(ir)

	ir.RegisterImplementations((*coretransaction.Msg)(nil),
		&bankv1beta1.MsgSend{},
		&bankv1beta1.MsgBurn{},
		&bankv1beta1.MsgSetSendEnabled{},
		&bankv1beta1.MsgMultiSend{},
		&bankv1beta1.MsgUpdateParams{},
	)
	queryRouter.RegisterService(&bankv1beta1.Query_ServiceDesc, &bankQueryServer{})
	msgRouter.RegisterService(&bankv1beta1.Msg_ServiceDesc, &bankMsgServer{})

	ctx := coretesting.Context()
	ss := coretesting.KVStoreService(ctx, "test")
	env := runtime.NewEnvironment(ss, coretesting.NewNopLogger(), runtime.EnvWithQueryRouterService(queryRouter), runtime.EnvWithMsgRouterService(msgRouter))
	env.EventService = eventService{}
	m, err := NewKeeper(codec.NewProtoCodec(ir), env, addressCodec, ir, nil, accounts...)
	require.NoError(t, err)
	return m, ctx
}

type bankQueryServer struct {
	bankv1beta1.UnimplementedQueryServer
}

type bankMsgServer struct {
	bankv1beta1.UnimplementedMsgServer
}

func (b bankQueryServer) Balance(context.Context, *bankv1beta1.QueryBalanceRequest) (*bankv1beta1.QueryBalanceResponse, error) {
	return &bankv1beta1.QueryBalanceResponse{Balance: &basev1beta1.Coin{
		Denom:  "atom",
		Amount: "1000",
	}}, nil
}

func (b bankMsgServer) Send(context.Context, *bankv1beta1.MsgSend) (*bankv1beta1.MsgSendResponse, error) {
	return &bankv1beta1.MsgSendResponse{}, nil
}
