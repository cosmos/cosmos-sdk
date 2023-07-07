package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/x/circuit/keeper"
	"cosmossdk.io/x/circuit/types"

	"github.com/cosmos/cosmos-sdk/types/query"
)

func TestQueryAccount(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	add, err := f.ac.StringToBytes(addresses[0])
	require.NoError(t, err)

	err = f.keeper.Permissions.Set(f.ctx, add, f.mockPerms)
	require.NoError(t, err)

	// create a new query server
	qs := keeper.NewQueryServer(f.keeper)

	// test the Account method
	res, err := qs.Account(f.ctx, &types.QueryAccountRequest{Address: addresses[0]})
	require.NoError(t, err)
	require.Equal(t, res.Permission.Level, types.Permissions_Level(3))
	require.Equal(t, res.Permission.LimitTypeUrls, []string{
		"test",
	})
}

func TestQueryAccounts(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	add, err := f.ac.StringToBytes(addresses[0])
	require.NoError(t, err)

	err = f.keeper.Permissions.Set(f.ctx, add, f.mockPerms)
	require.NoError(t, err)

	// create a new query server
	qs := keeper.NewQueryServer(f.keeper)

	// test the Accounts method
	res1, err := qs.Accounts(f.ctx, &types.QueryAccountsRequest{
		Pagination: &query.PageRequest{Limit: 10},
	})
	require.NoError(t, err)

	for _, a := range res1.Accounts {
		require.Equal(t, addresses[0], a.Address)
		require.Equal(t, f.mockPerms, *a.Permissions)
	}

	require.NotNil(t, res1)
}

func TestQueryDisabledList(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	require.NoError(t, f.keeper.DisableList.Set(f.ctx, f.mockMsgURL))

	// create a new query server
	qs := keeper.NewQueryServer(f.keeper)

	// test the DisabledList method
	disabledList, err := qs.DisabledList(f.ctx, &types.QueryDisabledListRequest{})
	require.NoError(t, err)
	require.Equal(t, []string{f.mockMsgURL}, disabledList.DisabledList)
}
