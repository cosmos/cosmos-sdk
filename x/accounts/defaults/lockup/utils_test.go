package lockup

import (
	"context"
	"testing"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/runtime/protoiface"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	banktypes "cosmossdk.io/x/bank/types"
	stakingtypes "cosmossdk.io/x/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ProtoMsg = protoiface.MessageV1

var TestFunds = sdk.NewCoins(sdk.NewCoin("test", math.NewInt(10)))

// mock statecodec
type mockStateCodec struct{}

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
	ModuleExecUntypedFunc = func(ctx context.Context, sender []byte, msg ProtoMsg) (ProtoMsg, error)
	ModuleExecFunc        = func(ctx context.Context, sender []byte, msg, msgResp ProtoMsg) error
	ModuleQueryFunc       = func(ctx context.Context, queryReq, queryResp ProtoMsg) error
)

// mock address codec
type addressCodec struct{}

func (a addressCodec) StringToBytes(text string) ([]byte, error) { return []byte(text), nil }
func (a addressCodec) BytesToString(bz []byte) (string, error)   { return string(bz), nil }

func newMockContext(t *testing.T) (context.Context, store.KVStoreService) {
	t.Helper()
	return accountstd.NewMockContext(
		0, []byte("lockup_account"), []byte("sender"), TestFunds, func(ctx context.Context, sender []byte, msg, msgResp ProtoMsg) error {
			return nil
		}, func(ctx context.Context, sender []byte, msg ProtoMsg) (ProtoMsg, error) {
			return nil, nil
		}, func(ctx context.Context, req, resp ProtoMsg) error {
			_, ok := req.(*banktypes.QueryBalanceRequest)
			if !ok {
				_, ok = req.(*stakingtypes.QueryParamsRequest)
				require.True(t, ok)
				gogoproto.Merge(resp.(gogoproto.Message), &stakingtypes.QueryParamsResponse{
					Params: stakingtypes.Params{
						BondDenom: "test",
					},
				})
				return nil
			}
			gogoproto.Merge(resp.(gogoproto.Message), &banktypes.QueryBalanceResponse{
				Balance: &sdk.Coin{
					Denom:  "test",
					Amount: math.NewInt(5),
				},
			})

			return nil
		},
	)
}

func makeMockDependencies(storeservice store.KVStoreService) accountstd.Dependencies {
	sb := collections.NewSchemaBuilder(storeservice)

	return accountstd.Dependencies{
		SchemaBuilder:    sb,
		AddressCodec:     addressCodec{},
		LegacyStateCodec: mockStateCodec{},
	}
}
