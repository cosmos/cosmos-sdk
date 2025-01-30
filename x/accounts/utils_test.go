package accounts

import (
	"context"
	"testing"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoregistry"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	"cosmossdk.io/core/address"
	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/core/testing/msgrouter"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/internal/implementation"
	banktypes "cosmossdk.io/x/bank/types"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ address.Codec = (*addressCodec)(nil)

type addressCodec struct{}

func (a addressCodec) StringToBytes(text string) ([]byte, error) { return []byte(text), nil }
func (a addressCodec) BytesToString(bz []byte) (string, error)   { return string(bz), nil }

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

	ir.RegisterImplementations((*transaction.Msg)(nil),
		&bankv1beta1.MsgSend{},
		&bankv1beta1.MsgBurn{},
		&bankv1beta1.MsgSetSendEnabled{},
		&bankv1beta1.MsgMultiSend{},
		&bankv1beta1.MsgUpdateParams{},
	)

	msgRouter := msgrouter.NewRouterService()
	msgRouter.RegisterHandler(Send, gogoproto.MessageName(&banktypes.MsgSend{}))

	queryRouter := msgrouter.NewRouterService()
	queryRouter.RegisterHandler(Balance, gogoproto.MessageName(&banktypes.QueryBalanceRequest{}))

	ctx, env := coretesting.NewTestEnvironment(coretesting.TestEnvironmentConfig{
		ModuleName:  "test",
		Logger:      coretesting.NewNopLogger(),
		MsgRouter:   msgRouter,
		QueryRouter: queryRouter,
	})

	m, err := NewKeeper(codec.NewProtoCodec(ir), env.Environment, addressCodec, ir, nil, accounts...)
	require.NoError(t, err)
	return m, ctx
}

func Balance(context.Context, transaction.Msg) (transaction.Msg, error) {
	return &banktypes.QueryBalanceResponse{Balance: &sdk.Coin{
		Denom:  "atom",
		Amount: math.NewInt(1000),
	}}, nil
}

func Send(context.Context, transaction.Msg) (transaction.Msg, error) {
	return &bankv1beta1.MsgSendResponse{}, nil
}
