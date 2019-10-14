package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	keep "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestQueryAccount(t *testing.T) {
	app, ctx := createTestApp(true)
	cdc := app.Codec()

	req := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	path := []string{types.QueryAccount}
	querier := keep.NewQuerier(app.AccountKeeper)

	bz, err := querier(ctx, []string{"other"}, req)
	require.Error(t, err)
	require.Nil(t, bz)

	req = abci.RequestQuery{
		Path: fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryAccount),
		Data: []byte{},
	}
	res, err := querier(ctx, path, req)
	require.Error(t, err)
	require.Nil(t, res)

	req.Data = cdc.MustMarshalJSON(types.NewQueryAccountParams([]byte("")))
	res, err = querier(ctx, path, req)
	require.Error(t, err)
	require.Nil(t, res)

	_, _, addr := types.KeyTestPubAddr()
	req.Data = cdc.MustMarshalJSON(types.NewQueryAccountParams(addr))
	res, err = querier(ctx, path, req)
	require.Error(t, err)
	require.Nil(t, res)

	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, addr))
	res, err = querier(ctx, path, req)
	require.NoError(t, err)
	require.NotNil(t, res)

	res, err = querier(ctx, path, req)
	require.NoError(t, err)
	require.NotNil(t, res)

	var account exported.Account
	err2 := cdc.UnmarshalJSON(res, &account)
	require.Nil(t, err2)
}
