package accounts

import (
	"testing"

	"github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/internal/implementation"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestKeeper_Init(t *testing.T) {
	m, ctx := newKeeper(t, accountstd.AddAccount("test", NewTestAccount))

	t.Run("ok", func(t *testing.T) {
		sender := []byte("sender")

		resp, addr, err := m.Init(ctx, "test", sender, &types.Empty{}, nil, nil)
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
		_, _, err := m.Init(ctx, "unknown", []byte("sender"), &types.Empty{}, nil, nil)
		require.ErrorIs(t, err, errAccountTypeNotFound)
	})
}

func TestKeeper_Execute(t *testing.T) {
	m, ctx := newKeeper(t, accountstd.AddAccount("test", NewTestAccount))

	// create account
	sender := []byte("sender")
	_, accAddr, err := m.Init(ctx, "test", sender, &types.Empty{}, nil, nil)
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
		resp, err := m.Execute(ctx, accAddr, sender, &types.Int64Value{Value: 1000}, nil)
		require.NoError(t, err)
		require.True(t, implementation.Equal(&types.Empty{}, resp))
	})
}

func TestKeeper_Query(t *testing.T) {
	m, ctx := newKeeper(t, accountstd.AddAccount("test", NewTestAccount))

	// create account
	sender := []byte("sender")
	_, accAddr, err := m.Init(ctx, "test", sender, &types.Empty{}, nil, nil)
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
		resp, err := m.Query(ctx, accAddr, &types.StringValue{Value: "atom"})
		require.NoError(t, err)
		require.True(t, implementation.Equal(&types.Int64Value{Value: 1000}, resp))
	})
}

func TestKeeper_NextAccountNumber(t *testing.T) {
	m, ctx := newKeeper(t, accountstd.AddAccount("test", NewTestAccount))
	num := uint64(0)
	err := m.AccountNumber.Set(ctx, num)
	require.NoError(t, err)

	n, err := m.NextAccountNumber(ctx)
	require.NoError(t, err)
	require.Equal(t, num, n)

	err = m.AccountNumber.Set(ctx, 0)
	require.NoError(t, err)

	store := m.KVStoreService.OpenKVStore(ctx)
	num = uint64(10)
	val := &types.UInt64Value{
		Value: num,
	}
	data, err := val.Marshal()
	require.NoError(t, err)
	err = store.Set(authtypes.LegacyGlobalAccountNumberKey, data)
	require.NoError(t, err)

	n, err = m.NextAccountNumber(ctx)
	require.NoError(t, err)
	require.Equal(t, num, n)
}
