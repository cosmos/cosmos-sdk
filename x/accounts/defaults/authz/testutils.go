package authz

import (
	"context"
	"errors"
	"fmt"
	"testing"

	gogoproto "github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/collections"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/header"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/defaults/authz/types"
	banktypes "cosmossdk.io/x/bank/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/authz"
)

var TestFunds = sdk.NewCoins(sdk.NewCoin("test", math.NewInt(10)))

// mock statecodec
type mockStateCodec struct {
	codec.Codec
}

var _ codec.Codec = mockStateCodec{}

func (c mockStateCodec) Marshal(m gogoproto.Message) ([]byte, error) {
	// Size() check can catch the typed nil value.
	if m == nil || gogoproto.Size(m) == 0 {
		// return empty bytes instead of nil, because nil has special meaning in places like store.Set
		return []byte{}, nil
	}

	return gogoproto.Marshal(m)
}

func (c mockStateCodec) Unmarshal(bz []byte, ptr gogoproto.Message) error {
	err := gogoproto.Unmarshal(bz, ptr)

	return err
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

func newMockContext(t *testing.T) (context.Context, store.KVStoreService) {
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

func makeMockDependencies(storeservice store.KVStoreService) accountstd.Dependencies {
	sb := collections.NewSchemaBuilder(storeservice)

	return accountstd.Dependencies{
		SchemaBuilder:    sb,
		AddressCodec:     addressCodec{},
		LegacyStateCodec: mockStateCodec{},
		Environment: appmodulev2.Environment{
			HeaderService: headerService{},
		},
	}
}

var _ types.Authorization = mockAuthorization{}

// mock grant
type mockAuthorization struct {
	*types.GenericAuthoriztion
	sendAmt sdk.Coins
	typeUrl string
}

func newMockAuthorization(typeUrl string) mockAuthorization {
	return mockAuthorization{
		GenericAuthoriztion: &types.GenericAuthoriztion{},
		sendAmt:             sdk.NewCoins(),
		typeUrl:             typeUrl,
	}
}

func (m mockAuthorization) MsgTypeURL() string {
	return m.typeUrl
}

func (m mockAuthorization) Accept(ctx context.Context, msg sdk.Msg) (authz.AcceptResponse, error) {
	msgSend, ok := msg.(*banktypes.MsgSend)
	if !ok {
		return authz.AcceptResponse{}, fmt.Errorf("invalid message")
	}

	m.sendAmt = m.sendAmt.Add(msgSend.Amount...)
	return authz.AcceptResponse{
		Accept:  true,
		Delete:  false,
		Updated: &m,
	}, nil
}

func (m mockAuthorization) ValidateBasic() error {
	return nil
}
