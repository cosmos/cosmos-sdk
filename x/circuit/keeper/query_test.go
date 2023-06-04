package keeper

import (
	"testing"

	"cosmossdk.io/x/circuit/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types/query"
)

func TestQueryAccount(t *testing.T) {
	t.Parallel()
	f := setupFixture(t)

	add, err := f.Keeper.addressCodec.StringToBytes(addresses[0])
	require.NoError(t, err)

	err = f.Keeper.SetPermissions(f.Ctx, add, &f.MockPerms)
	require.NoError(t, err)

	// create a new query server
	qs := QueryServer{keeper: f.Keeper}

	// test the Account method
	res, err := qs.Account(f.Ctx, &types.QueryAccountRequest{Address: addresses[0]})
	require.NoError(t, err)
	require.Equal(t, res.Permission.Level, types.Permissions_Level(3))
	require.Equal(t, res.Permission.LimitTypeUrls, []string{
		"test",
	})
}

func TestQueryAccounts(t *testing.T) {
	t.Parallel()
	f := setupFixture(t)

	add, err := f.Keeper.addressCodec.StringToBytes(addresses[0])
	require.NoError(t, err)

	err = f.Keeper.SetPermissions(f.Ctx, add, &f.MockPerms)
	require.NoError(t, err)

	// create a new query server
	qs := QueryServer{keeper: f.Keeper}

	// test the Accounts method
	res1, err := qs.Accounts(f.Ctx, &types.QueryAccountsRequest{
		Pagination: &query.PageRequest{Limit: 10},
	})
	require.NoError(t, err)

	for _, a := range res1.Accounts {
		require.Equal(t, addresses[0], a.Address)
		require.Equal(t, f.MockPerms, *a.Permissions)
	}

	require.NotNil(t, res1)
}

func TestQueryDisabledList(t *testing.T) {
	t.Parallel()
	f := setupFixture(t)

	f.Keeper.DisableMsg(f.Ctx, f.MockMsgURL)

	// create a new query server
	qs := QueryServer{keeper: f.Keeper}

	// test the DisabledList method
	disabledList, err := qs.DisabledList(f.Ctx, &types.QueryDisabledListRequest{})
	require.NoError(t, err)
	require.Equal(t, []string{f.MockMsgURL}, disabledList.DisabledList)
}
