package authz

import (
	"context"
	"errors"
	"testing"

	gogoproto "github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/collections"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/header"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	banktypes "cosmossdk.io/x/bank/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var TestFunds = sdk.NewCoins(sdk.NewCoin("test", math.NewInt(10)))

// mock statecodec
type mockStateCodec struct {
	codec.Codec
}

func (c mockStateCodec) MarshalInterface(i gogoproto.Message) ([]byte, error) {
	any, err := types.NewAnyWithValue(i)
	if err != nil {
		return nil, err
	}

	return c.Marshal(any)
}

type (
	ModuleExecUntypedFunc = func(ctx context.Context, sender []byte, msg transaction.Msg) (transaction.Msg, error)
	ModuleExecFunc        = func(ctx context.Context, sender []byte, msg, msgResp transaction.Msg) error
	ModuleQueryFunc       = func(ctx context.Context, queryReq, queryResp transaction.Msg) error
)

// mock address codec
type addressCodec struct{}

func (a addressCodec) StringToBytes(text string) ([]byte, error) { return []byte(text), nil }
func (a addressCodec) BytesToString(bz []byte) (string, error)   { return string(bz), nil }

// mock header service
type headerService struct{}

func (h headerService) HeaderInfo(ctx context.Context) header.Info {
	return sdk.UnwrapSDKContext(ctx).HeaderInfo()
}

func NewMockContext(t *testing.T) (context.Context, store.KVStoreService) {
	t.Helper()
	return accountstd.NewMockContext(
		0, []byte("authz"), []byte("sender"), TestFunds,
		func(ctx context.Context, sender []byte, msg transaction.Msg) (transaction.Msg, error) {
			typeUrl := sdk.MsgTypeURL(msg)
			switch typeUrl {
			case "/cosmos.bank.v1beta1.MsgSend":
				return &banktypes.MsgSendResponse{}, nil
			default:
				return nil, errors.New("unrecognized request type")
			}
		}, func(ctx context.Context, req transaction.Msg) (transaction.Msg, error) {
			return nil, nil
		},
	)
}

func MakeMockDependencies(storeservice store.KVStoreService, codec codec.Codec) accountstd.Dependencies {
	sb := collections.NewSchemaBuilder(storeservice)

	return accountstd.Dependencies{
		SchemaBuilder: sb,
		AddressCodec:  addressCodec{},
		LegacyStateCodec: mockStateCodec{
			Codec: codec,
		},
		Environment: appmodulev2.Environment{
			HeaderService: headerService{},
		},
	}
}
