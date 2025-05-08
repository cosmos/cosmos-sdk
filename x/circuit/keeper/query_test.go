package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/circuit/keeper"
	"github.com/cosmos/cosmos-sdk/x/circuit/types"
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

	// test invalid address
	res, err = qs.Account(f.ctx, &types.QueryAccountRequest{Address: "invalid"})
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid bech32 string")
	require.Nil(t, res)

	// test account not found
	res, err = qs.Account(f.ctx, &types.QueryAccountRequest{Address: addresses[1]})
	require.Error(t, err)
	require.ErrorContains(t, err, "not found")
	require.Nil(t, res)
}

func TestQueryAccounts(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	add, err := f.ac.StringToBytes(addresses[0])
	require.NoError(t, err)

	// create a new query server
	qs := keeper.NewQueryServer(f.keeper)

	// test the Accounts method with no accounts
	res1, err := qs.Accounts(f.ctx, &types.QueryAccountsRequest{
		Pagination: &query.PageRequest{Limit: 10},
	})
	require.NoError(t, err)
	require.Len(t, res1.Accounts, 0)

	err = f.keeper.Permissions.Set(f.ctx, add, f.mockPerms)
	require.NoError(t, err)

	// test the Accounts method
	res2, err := qs.Accounts(f.ctx, &types.QueryAccountsRequest{
		Pagination: &query.PageRequest{Limit: 10},
	})
	require.NoError(t, err)

	for _, a := range res2.Accounts {
		require.Equal(t, addresses[0], a.Address)
		require.Equal(t, f.mockPerms, *a.Permissions)
	}

	require.NotNil(t, res1)
}

func TestQueryDisabledList(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	// create a new query server
	qs := keeper.NewQueryServer(f.keeper)

	// test the DisabledList method
	disabledList, err := qs.DisabledList(f.ctx, &types.QueryDisabledListRequest{})
	require.NoError(t, err)
	require.Len(t, disabledList.DisabledList, 0)

	require.NoError(t, f.keeper.DisableList.Set(f.ctx, f.mockMsgURL))

	// test the DisabledList method
	disabledList, err = qs.DisabledList(f.ctx, &types.QueryDisabledListRequest{})
	require.NoError(t, err)
	require.Equal(t, []string{f.mockMsgURL}, disabledList.DisabledList)
}
