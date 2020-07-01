package keeper_test

import (
	gocontext "context"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
)

func TestQueryAccount(t *testing.T) {
	app, ctx := createTestApp(true)
	cdc := app.Codec()

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, app.AccountKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	res, err := queryClient.Account(gocontext.Background(), []byte{})
	require.Error(t, err)
	require.Nil(t, res)

	req := cdc.MustMarshalJSON(types.NewQueryAccountParams([]byte("")))
	res, err = queryClient.Account(gocontext.Background(), req)
	require.Error(t, err)
	require.Nil(t, res)

	_, _, addr := types.KeyTestPubAddr()
	req = cdc.MustMarshalJSON(types.NewQueryAccountParams(addr))
	res, err = queryClient.Account(gocontext.Background(), req)
	require.Error(t, err)
	require.Nil(t, res)

	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, addr))
	res, err = queryClient.Account(gocontext.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, res)

	var account types.AccountI
	err2 := cdc.UnmarshalJSON(res, &account)
	require.Nil(t, err2)
}

func TestQueryParameters(t *testing.T) {
	app, ctx := createTestApp(true)
	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, app.AccountKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	resp, err := queryClient.Parameters(gocontext.Background(), &types.QueryParametersRequest{})
	require.NoError(t, err)
	require.Equal(t, app.AccountKeeper.GetParams(ctx), resp.Params)
}
