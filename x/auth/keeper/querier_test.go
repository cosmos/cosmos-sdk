package keeper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestQueryAccount(t *testing.T) {
	input := SetupTestInput()

	req := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	querier := NewQuerier(input.AccountKeeper)

	bz, err := querier(input.Ctx, []string{"other"}, req)
	require.Error(t, err)
	require.Nil(t, bz)

	req = abci.RequestQuery{
		Path: fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryAccount),
		Data: []byte{},
	}
	res, err := queryAccount(input.Ctx, req, input.AccountKeeper)
	require.Error(t, err)
	require.Nil(t, res)

	req.Data = input.cdc.MustMarshalJSON(types.NewQueryAccountParams([]byte("")))
	res, err = queryAccount(input.Ctx, req, input.AccountKeeper)
	require.Error(t, err)
	require.Nil(t, res)

	_, _, addr := types.KeyTestPubAddr()
	req.Data = input.cdc.MustMarshalJSON(types.NewQueryAccountParams(addr))
	res, err = queryAccount(input.Ctx, req, input.AccountKeeper)
	require.Error(t, err)
	require.Nil(t, res)

	input.AccountKeeper.SetAccount(input.Ctx, input.AccountKeeper.NewAccountWithAddress(input.Ctx, addr))
	res, err = queryAccount(input.Ctx, req, input.AccountKeeper)
	require.NoError(t, err)
	require.NotNil(t, res)

	res, err = querier(input.Ctx, []string{types.QueryAccount}, req)
	require.NoError(t, err)
	require.NotNil(t, res)

	var account exported.Account
	err2 := input.cdc.UnmarshalJSON(res, &account)
	require.Nil(t, err2)
}
