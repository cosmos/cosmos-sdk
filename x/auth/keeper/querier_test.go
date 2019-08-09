package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestQueryAccount(t *testing.T) {
	ctx, app := newTestApp(t)

	req := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	path := []string{types.QueryAccount}
	querier := NewQuerier(app.AccountKeeper)

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

	req.Data = input.cdc.MustMarshalJSON(types.NewQueryAccountParams([]byte("")))
	res, err = querier(ctx, path, req)
	require.Error(t, err)
	require.Nil(t, res)

	_, _, addr := types.KeyTestPubAddr()
	req.Data = input.cdc.MustMarshalJSON(types.NewQueryAccountParams(addr))
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
	err2 := input.cdc.UnmarshalJSON(res, &account)
	require.Nil(t, err2)
}
