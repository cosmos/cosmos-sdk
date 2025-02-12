package lockup

import (
	"context"
	"errors"
	"testing"
	"time"

	gogoproto "github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/collections"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/header"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	banktypes "cosmossdk.io/x/bank/types"
	distrtypes "cosmossdk.io/x/distribution/types"
	stakingtypes "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
		0, []byte("lockup_account"), []byte("owner"), TestFunds,
		func(ctx context.Context, sender []byte, msg transaction.Msg) (transaction.Msg, error) {
			typeUrl := sdk.MsgTypeURL(msg)
			switch typeUrl {
			case "/cosmos.staking.v1beta1.MsgDelegate":
				return &stakingtypes.MsgDelegateResponse{}, nil
			case "/cosmos.staking.v1beta1.MsgUndelegate":
				return &stakingtypes.MsgUndelegateResponse{
					Amount: sdk.NewCoin("test", math.NewInt(1)),
				}, nil
			case "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward":
				return &distrtypes.MsgWithdrawDelegatorRewardResponse{}, nil
			case "/cosmos.bank.v1beta1.MsgSend":
				return &banktypes.MsgSendResponse{}, nil
			default:
				return nil, errors.New("unrecognized request type")
			}
		}, func(ctx context.Context, req transaction.Msg) (transaction.Msg, error) {
			typeUrl := sdk.MsgTypeURL(req)
			switch typeUrl {
			case "/cosmos.staking.v1beta1.QueryParamsRequest":
				return &stakingtypes.QueryParamsResponse{
					Params: stakingtypes.Params{
						BondDenom: "test",
					},
				}, nil
			case "/cosmos.staking.v1beta1.QueryUnbondingDelegationRequest":
				return &stakingtypes.QueryUnbondingDelegationResponse{
					Unbond: stakingtypes.UnbondingDelegation{
						DelegatorAddress: "sender",
						ValidatorAddress: valAddress,
						Entries: []stakingtypes.UnbondingDelegationEntry{
							{
								CreationHeight: 1,
								CompletionTime: time.Now(),
								Balance:        math.NewInt(1),
							},
							{
								CreationHeight: 1,
								CompletionTime: time.Now().Add(time.Hour),
								Balance:        math.NewInt(1),
							},
						},
					},
				}, nil
			case "/cosmos.bank.v1beta1.QueryBalanceRequest":
				return &banktypes.QueryBalanceResponse{
					Balance: &(sdk.Coin{
						Denom:  "test",
						Amount: TestFunds.AmountOf("test"),
					}),
				}, nil
			default:
				return nil, errors.New("unrecognized request type")
			}
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
