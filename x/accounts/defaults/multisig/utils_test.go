package multisig

import (
	"context"
	"testing"
	"time"

	gogoproto "github.com/cosmos/gogoproto/proto"
	types "github.com/cosmos/gogoproto/types/any"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/collections"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/header"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	banktypes "cosmossdk.io/x/bank/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var TestFunds = sdk.NewCoins(sdk.NewCoin("test", math.NewInt(10)))

// mock statecodec
type mockStateCodec struct {
	codec.Codec
}

// GetMsgAnySigners implements codec.Codec.
func (mockStateCodec) GetMsgAnySigners(msg *types.Any) ([][]byte, protoreflect.Message, error) {
	panic("unimplemented")
}

// GetMsgSigners implements codec.Codec.
func (mockStateCodec) GetMsgSigners(msg gogoproto.Message) ([][]byte, protoreflect.Message, error) {
	panic("unimplemented")
}

// GetReflectMsgSigners implements codec.Codec.
func (mockStateCodec) GetReflectMsgSigners(msg protoreflect.Message) ([][]byte, error) {
	panic("unimplemented")
}

// InterfaceRegistry implements codec.Codec.
func (mockStateCodec) InterfaceRegistry() codectypes.InterfaceRegistry {
	panic("unimplemented")
}

// MarshalInterface implements codec.Codec.
func (mockStateCodec) MarshalInterface(i gogoproto.Message) ([]byte, error) {
	panic("unimplemented")
}

// MarshalInterfaceJSON implements codec.Codec.
func (mockStateCodec) MarshalInterfaceJSON(i gogoproto.Message) ([]byte, error) {
	panic("unimplemented")
}

// MarshalJSON implements codec.Codec.
func (mockStateCodec) MarshalJSON(o gogoproto.Message) ([]byte, error) {
	panic("unimplemented")
}

// MarshalLengthPrefixed implements codec.Codec.
func (mockStateCodec) MarshalLengthPrefixed(o gogoproto.Message) ([]byte, error) {
	panic("unimplemented")
}

// MustMarshal implements codec.Codec.
func (mockStateCodec) MustMarshal(o gogoproto.Message) []byte {
	panic("unimplemented")
}

// MustMarshalJSON implements codec.Codec.
func (mockStateCodec) MustMarshalJSON(o gogoproto.Message) []byte {
	panic("unimplemented")
}

// MustMarshalLengthPrefixed implements codec.Codec.
func (mockStateCodec) MustMarshalLengthPrefixed(o gogoproto.Message) []byte {
	panic("unimplemented")
}

// MustUnmarshal implements codec.Codec.
func (mockStateCodec) MustUnmarshal(bz []byte, ptr gogoproto.Message) {
	panic("unimplemented")
}

// MustUnmarshalJSON implements codec.Codec.
func (mockStateCodec) MustUnmarshalJSON(bz []byte, ptr gogoproto.Message) {
	panic("unimplemented")
}

// MustUnmarshalLengthPrefixed implements codec.Codec.
func (mockStateCodec) MustUnmarshalLengthPrefixed(bz []byte, ptr gogoproto.Message) {
	panic("unimplemented")
}

// UnmarshalInterface implements codec.Codec.
func (mockStateCodec) UnmarshalInterface(bz []byte, ptr interface{}) error {
	panic("unimplemented")
}

// UnmarshalInterfaceJSON implements codec.Codec.
func (mockStateCodec) UnmarshalInterfaceJSON(bz []byte, ptr interface{}) error {
	panic("unimplemented")
}

// UnmarshalJSON implements codec.Codec.
func (mockStateCodec) UnmarshalJSON(bz []byte, ptr gogoproto.Message) error {
	panic("unimplemented")
}

// UnmarshalLengthPrefixed implements codec.Codec.
func (mockStateCodec) UnmarshalLengthPrefixed(bz []byte, ptr gogoproto.Message) error {
	panic("unimplemented")
}

// UnpackAny implements codec.Codec.
func (mockStateCodec) UnpackAny(any *types.Any, iface interface{}) error {
	panic("unimplemented")
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

func newMockContext(t *testing.T) (context.Context, store.KVStoreService) {
	t.Helper()
	return accountstd.NewMockContext(
		0, []byte("mock_multisig_account"), []byte("sender"), TestFunds, func(ctx context.Context, sender []byte, msg transaction.Msg) (transaction.Msg, error) {
			return nil, nil
		}, func(ctx context.Context, req transaction.Msg) (transaction.Msg, error) {
			var resp transaction.Msg
			_, ok := req.(*banktypes.QueryBalanceRequest)
			require.True(t, ok)
			gogoproto.Merge(resp, &banktypes.QueryBalanceResponse{
				Balance: &sdk.Coin{
					Denom:  "test",
					Amount: math.NewInt(5),
				},
			})
			return resp, nil
		},
	)
}

func makeMockDependencies(storeservice store.KVStoreService, timefn func() time.Time) accountstd.Dependencies {
	sb := collections.NewSchemaBuilder(storeservice)

	return accountstd.Dependencies{
		SchemaBuilder:    sb,
		AddressCodec:     addressCodec{},
		LegacyStateCodec: mockStateCodec{},
		Environment: appmodulev2.Environment{
			HeaderService: headerService{timefn},
			EventService:  eventService{},
		},
	}
}

type headerService struct {
	timefn func() time.Time
}

func (h headerService) HeaderInfo(context.Context) header.Info {
	return header.Info{
		Time: h.timefn(),
	}
}

type eventService struct{}

// EventManager implements event.Service.
func (eventService) EventManager(context.Context) event.Manager {
	return runtime.EventService{Events: runtime.Events{EventManagerI: sdk.NewEventManager()}}
}
