package accounts

import (
	"context"
	"testing"

	"github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	"cosmossdk.io/collections"
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/internal/implementation"
)

func TestKeeper_Init(t *testing.T) {
	m, ctx := newKeeper(t, accountstd.AddAccount("test", NewTestAccount))
	m.queryRouter = mockQuery(func(ctx context.Context, req, resp implementation.ProtoMsg) error {
		_, ok := req.(*bankv1beta1.QueryBalanceRequest)
		require.True(t, ok)
		_, ok = resp.(*bankv1beta1.QueryBalanceResponse)
		require.True(t, ok)
		return nil
	})

	t.Run("ok", func(t *testing.T) {
		sender := []byte("sender")

		resp, addr, err := m.Init(ctx, "test", sender, &types.Empty{}, nil)
		require.NoError(t, err)
		require.Equal(t, &types.Empty{}, resp)
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
		_, _, err := m.Init(ctx, "unknown", []byte("sender"), &types.Empty{}, nil)
		require.ErrorIs(t, err, errAccountTypeNotFound)
	})
}

func TestKeeper_Execute(t *testing.T) {
	m, ctx := newKeeper(t, accountstd.AddAccount("test", NewTestAccount))
	m.queryRouter = mockQuery(func(ctx context.Context, req, resp implementation.ProtoMsg) error { return nil })

	// create account
	sender := []byte("sender")
	_, accAddr, err := m.Init(ctx, "test", sender, &types.Empty{}, nil)
	require.NoError(t, err)

	t.Run("ok", func(t *testing.T) {
		resp, err := m.Execute(ctx, accAddr, sender, &types.Empty{}, nil)
		require.NoError(t, err)
		require.Equal(t, &types.Empty{}, resp)
	})

	t.Run("unknown account", func(t *testing.T) {
		_, err := m.Execute(ctx, []byte("unknown"), sender, &types.Empty{}, nil)
		require.ErrorIs(t, err, collections.ErrNotFound)
	})

	t.Run("exec module", func(t *testing.T) {
		m.msgRouter = mockExec(func(ctx context.Context, msg, msgResp implementation.ProtoMsg) error {
			concrete, ok := msg.(*bankv1beta1.MsgSend)
			require.True(t, ok)
			require.Equal(t, concrete.ToAddress, "recipient")
			_, ok = msgResp.(*bankv1beta1.MsgSendResponse)
			require.True(t, ok)
			return nil
		})

		m.signerProvider = mockSigner(func(msg implementation.ProtoMsg) ([]byte, error) {
			require.Equal(t, msg.(*bankv1beta1.MsgSend).FromAddress, string(accAddr))
			return accAddr, nil
		})

		resp, err := m.Execute(ctx, accAddr, sender, &types.Int64Value{Value: 1000}, nil)
		require.NoError(t, err)
		require.True(t, implementation.Equal(&types.Empty{}, resp))
	})
}

func TestKeeper_Query(t *testing.T) {
	m, ctx := newKeeper(t, accountstd.AddAccount("test", NewTestAccount))
	m.queryRouter = mockQuery(func(ctx context.Context, req, resp implementation.ProtoMsg) error {
		return nil
	})

	// create account
	sender := []byte("sender")
	_, accAddr, err := m.Init(ctx, "test", sender, &types.Empty{}, nil)
	require.NoError(t, err)

	t.Run("ok", func(t *testing.T) {
		resp, err := m.Query(ctx, accAddr, &types.Empty{})
		require.NoError(t, err)
		require.Equal(t, &types.Empty{}, resp)
	})

	t.Run("unknown account", func(t *testing.T) {
		_, err := m.Query(ctx, []byte("unknown"), &types.Empty{})
		require.ErrorIs(t, err, collections.ErrNotFound)
	})

	t.Run("query module", func(t *testing.T) {
		// we inject the module query function, which accepts only a specific type of message
		// we force the response
		m.queryRouter = mockQuery(func(ctx context.Context, req, resp implementation.ProtoMsg) error {
			concrete, ok := req.(*bankv1beta1.QueryBalanceRequest)
			require.True(t, ok)
			require.Equal(t, string(accAddr), concrete.Address)
			require.Equal(t, concrete.Denom, "atom")
			copyResp := &bankv1beta1.QueryBalanceResponse{Balance: &basev1beta1.Coin{
				Denom:  "atom",
				Amount: "1000",
			}}
			proto.Merge(resp.(proto.Message), copyResp)
			return nil
		})

		resp, err := m.Query(ctx, accAddr, &types.StringValue{Value: "atom"})
		require.NoError(t, err)
		require.True(t, implementation.Equal(&types.Int64Value{Value: 1000}, resp))
	})
}
