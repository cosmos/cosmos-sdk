package accounts

import (
	"context"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/internal/implementation"
	banktypes "cosmossdk.io/x/bank/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestKeeper_Init(t *testing.T) {
	m, ctx := newKeeper(t, accountstd.AddAccount("test", NewTestAccount))
	m.queryRouter = mockQuery(func(ctx context.Context, req, resp implementation.ProtoMsg) error {
		_, ok := req.(*banktypes.QueryBalanceRequest)
		require.True(t, ok)
		_, ok = resp.(*banktypes.QueryBalanceResponse)
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
			concrete, ok := msg.(*banktypes.MsgSend)
			require.True(t, ok)
			require.Equal(t, concrete.ToAddress, "recipient")
			_, ok = msgResp.(*banktypes.MsgSendResponse)
			require.True(t, ok)
			return nil
		})

		m.signerProvider = mockSigner(func(msg implementation.ProtoMsg) ([]byte, error) {
			require.Equal(t, msg.(*banktypes.MsgSend).FromAddress, string(accAddr))
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
			concrete, ok := req.(*banktypes.QueryBalanceRequest)
			require.True(t, ok)
			require.Equal(t, string(accAddr), concrete.Address)
			require.Equal(t, concrete.Denom, "atom")
			copyResp := &banktypes.QueryBalanceResponse{Balance: &sdk.Coin{
				Denom:  "atom",
				Amount: math.NewInt(1000),
			}}
			proto.Merge(resp, copyResp)
			return nil
		})

		resp, err := m.Query(ctx, accAddr, &types.StringValue{Value: "atom"})
		require.NoError(t, err)
		require.True(t, implementation.Equal(&types.Int64Value{Value: 1000}, resp))
	})
}
