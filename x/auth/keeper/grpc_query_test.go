package keeper_test

import (
	gocontext "context"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
)

func TestGRPCQueryAccount(t *testing.T) {
	app, ctx := createTestApp(true)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.AccountKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	res, err := queryClient.Account(gocontext.Background(), &types.QueryAccountRequest{Address: []byte{}})
	require.Error(t, err)
	require.Nil(t, res)

	res, err = queryClient.Account(gocontext.Background(), &types.QueryAccountRequest{Address: []byte("")})
	require.Error(t, err)
	require.Nil(t, res)

	_, _, addr := types.KeyTestPubAddr()
	res, err = queryClient.Account(gocontext.Background(), &types.QueryAccountRequest{Address: addr})
	require.Error(t, err)
	require.Nil(t, res)

	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, addr))
	res, err = queryClient.Account(gocontext.Background(), &types.QueryAccountRequest{Address: addr})
	t.Log("Response....", res)
	require.NoError(t, err)
	require.NotNil(t, res)
	account := keeper.GetAccount(res.Account)
	require.NotNil(t, account)
	require.True(t, addr.Equals(account.GetAddress()))
}

func TestGRPCQueryParameters(t *testing.T) {
	app, ctx := createTestApp(true)
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.AccountKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	resp, err := queryClient.Parameters(gocontext.Background(), &types.QueryParametersRequest{})
	require.NoError(t, err)
	require.Equal(t, app.AccountKeeper.GetParams(ctx), resp.Params)
}
