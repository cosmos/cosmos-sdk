package keeper

import (
	storetypes "cosmossdk.io/store/types"
	cmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/circuit/types"
	"github.com/stretchr/testify/require"
	"testing"
)

type fixture struct {
	ctx        sdk.Context
	keeper     Keeper
	mockAddr   string
	mockPerms  types.Permissions
	mockMsgURL string
}

func iniFixture(t *testing.T) *fixture {
	mockStoreKey := storetypes.NewKVStoreKey("test")
	mockAddr := "mock_address"
	keeperX := NewKeeper(mockStoreKey, mockAddr)
	mockMsgURL := "mock_url"
	mockCtx := testutil.DefaultContextWithDB(t, mockStoreKey, storetypes.NewTransientStoreKey("transient_test"))
	ctx := mockCtx.Ctx.WithBlockHeader(cmproto.Header{})
	mockPerms := types.Permissions{
		Level:         3,
		LimitTypeUrls: []string{"test"},
	}

	return &fixture{
		ctx:        ctx,
		keeper:     keeperX,
		mockAddr:   mockAddr,
		mockPerms:  mockPerms,
		mockMsgURL: mockMsgURL,
	}
}

func TestQueryServer(t *testing.T) {
	t.Parallel()
	f := iniFixture(t)

	err := f.keeper.SetPermissions(f.ctx, f.mockAddr, &f.mockPerms)
	require.NoError(t, err)

	// create a new query server
	qs := QueryServer{keeper: f.keeper}

	// test the Account method
	res, err := qs.Account(f.ctx, &types.QueryAccountRequest{Address: f.mockAddr})
	require.NoError(t, err)
	require.Equal(t, res.Permission.Level, types.Permissions_Level(3))
	require.Equal(t, res.Permission.LimitTypeUrls, []string{
		"test",
	})

	// test the Accounts method
	res1, err := qs.Accounts(f.ctx, &types.QueryAccountsRequest{
		Pagination: &query.PageRequest{Limit: 10},
	})
	require.NoError(t, err)

	//var acct *types.GenesisAccountPermissions
	for _, a := range res1.Accounts {
		require.Equal(t, f.mockAddr, string(a.Address))
		require.Equal(t, f.mockPerms, *a.Permissions)
	}

	require.NotNil(t, res1)

	// test the DisabledList method
	disabledList, err := qs.DisabledList(f.ctx, &types.QueryDisableListRequest{})
	require.NoError(t, err)
	require.Equal(t, []string{}, disabledList.DisabledList)
}
