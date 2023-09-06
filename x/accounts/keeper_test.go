package accounts

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/colltest"
	"cosmossdk.io/x/accounts/internal/implementation"
)

func newKeeper(t *testing.T, accounts map[string]implementation.Account) (Keeper, context.Context) {
	t.Helper()
	ss, ctx := colltest.MockStore()
	m, err := NewKeeper(ss, accounts)
	require.NoError(t, err)
	return m, ctx
}

func TestKeeper_Create(t *testing.T) {
	m, ctx := newKeeper(t, map[string]implementation.Account{
		"test": TestAccount{},
	})

	t.Run("ok", func(t *testing.T) {
		sender := []byte("sender")

		resp, addr, err := m.Create(ctx, "test", sender, &emptypb.Empty{})
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
		_, _, err := m.Create(ctx, "unknown", []byte("sender"), &emptypb.Empty{})
		require.ErrorIs(t, err, errAccountTypeNotFound)
	})
}

func TestKeeper_Execute(t *testing.T) {
	m, ctx := newKeeper(t, map[string]implementation.Account{
		"test": TestAccount{},
	})

	// create account
	sender := []byte("sender")
	_, accAddr, err := m.Create(ctx, "test", sender, &emptypb.Empty{})
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
}

func TestKeeper_Query(t *testing.T) {
	m, ctx := newKeeper(t, map[string]implementation.Account{
		"test": TestAccount{},
	})

	// create account
	sender := []byte("sender")
	_, accAddr, err := m.Create(ctx, "test", sender, &emptypb.Empty{})
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
}
