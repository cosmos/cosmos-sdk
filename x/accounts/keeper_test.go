package accounts

import (
	"context"
	"testing"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/colltest"
	"cosmossdk.io/core/address"
	"cosmossdk.io/x/accounts/internal/implementation"
)

var _ address.Codec = (*addressCodec)(nil)

type addressCodec struct{}

func (a addressCodec) StringToBytes(text string) ([]byte, error) { return []byte(text), nil }
func (a addressCodec) BytesToString(bz []byte) (string, error)   { return string(bz), nil }

func newKeeper(t *testing.T, accounts map[string]implementation.Account) (Keeper, context.Context) {
	t.Helper()
	ss, ctx := colltest.MockStore()
	m, err := NewKeeper(ss, addressCodec{}, nil, nil, nil, accounts)
	require.NoError(t, err)
	return m, ctx
}

func TestKeeper_Init(t *testing.T) {
	m, ctx := newKeeper(t, map[string]implementation.Account{
		"test": TestAccount{},
	})
	m.queryModuleFunc = func(ctx context.Context, msg proto.Message) (proto.Message, error) {
		_, ok := msg.(*bankv1beta1.QueryBalanceRequest)
		require.True(t, ok)
		return &bankv1beta1.QueryBalanceResponse{}, nil
	}

	t.Run("ok", func(t *testing.T) {
		sender := []byte("sender")

		resp, addr, err := m.Init(ctx, "test", sender, &emptypb.Empty{})
		require.NoError(t, err)
		require.Equal(t, &emptypb.Empty{}, resp)
		require.NotNil(t, addr)

		// ensure acc number was increased.
		num, err := m.AccountNumber.Peek(ctx)
		require.NoError(t, err)
		require.Equal(t, uint64(1), num)

		// ensure account mapping
		accType, err := m.AccountsByType.Get(ctx, addr)
		require.NoError(t, err)
		require.Equal(t, "test", accType)
	})

	t.Run("unknown account type", func(t *testing.T) {
		_, _, err := m.Init(ctx, "unknown", []byte("sender"), &emptypb.Empty{})
		require.ErrorIs(t, err, errAccountTypeNotFound)
	})
}

func TestKeeper_Execute(t *testing.T) {
	m, ctx := newKeeper(t, map[string]implementation.Account{
		"test": TestAccount{},
	})

	// create account
	sender := []byte("sender")
	_, accAddr, err := m.Init(ctx, "test", sender, &emptypb.Empty{})
	require.NoError(t, err)

	t.Run("ok", func(t *testing.T) {
		resp, err := m.Execute(ctx, accAddr, sender, &emptypb.Empty{})
		require.NoError(t, err)
		require.Equal(t, &emptypb.Empty{}, resp)
	})

	t.Run("unknown account", func(t *testing.T) {
		_, err := m.Execute(ctx, []byte("unknown"), sender, &emptypb.Empty{})
		require.ErrorIs(t, err, collections.ErrNotFound)
	})

	t.Run("exec module", func(t *testing.T) {
		m.execModuleFunc = func(ctx context.Context, msg proto.Message) (proto.Message, error) {
			concrete, ok := msg.(*bankv1beta1.MsgSend)
			require.True(t, ok)
			require.Equal(t, concrete.ToAddress, "recipient")
			return &bankv1beta1.MsgSendResponse{}, nil
		}

		m.getSenderFunc = func(msg proto.Message) ([]byte, error) {
			require.Equal(t, msg.(*bankv1beta1.MsgSend).FromAddress, string(accAddr))
			return accAddr, nil
		}

		resp, err := m.Execute(ctx, accAddr, sender, &wrapperspb.Int64Value{Value: 1000})
		require.NoError(t, err)
		require.True(t, proto.Equal(&emptypb.Empty{}, resp.(proto.Message)))
	})
}

func TestKeeper_Query(t *testing.T) {
	m, ctx := newKeeper(t, map[string]implementation.Account{
		"test": TestAccount{},
	})

	// create account
	sender := []byte("sender")
	_, accAddr, err := m.Init(ctx, "test", sender, &emptypb.Empty{})
	require.NoError(t, err)

	t.Run("ok", func(t *testing.T) {
		resp, err := m.Query(ctx, accAddr, &emptypb.Empty{})
		require.NoError(t, err)
		require.Equal(t, &emptypb.Empty{}, resp)
	})

	t.Run("unknown account", func(t *testing.T) {
		_, err := m.Query(ctx, []byte("unknown"), &emptypb.Empty{})
		require.ErrorIs(t, err, collections.ErrNotFound)
	})

	t.Run("query module", func(t *testing.T) {
		// we inject the module query function, which accepts only a specific type of message
		// we force the response
		m.queryModuleFunc = func(ctx context.Context, msg proto.Message) (proto.Message, error) {
			concrete, ok := msg.(*bankv1beta1.QueryBalanceRequest)
			require.True(t, ok)
			require.Equal(t, string(accAddr), concrete.Address)
			require.Equal(t, concrete.Denom, "atom")
			return &bankv1beta1.QueryBalanceResponse{Balance: &basev1beta1.Coin{
				Denom:  "atom",
				Amount: "1000",
			}}, nil
		}

		resp, err := m.Query(ctx, accAddr, wrapperspb.String("atom"))
		require.NoError(t, err)
		require.True(t, proto.Equal(wrapperspb.Int64(1000), resp.(proto.Message)))
	})
}
