package keeper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/uniswap/internal/types"
)

func TestNewQuerier(t *testing.T) {
	cdc := codec.New()
	ctx, keeper := createNewInput()

	req := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	querier := NewQuerier(keeper)

	// query for non existent address should return an error
	req.Path = fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryBalance)
	req.Data = keeper.cdc.MustMarshalJSON(addr)
	res, err := querier(ctx, []string{"balance"}, req)
	require.NotNil(t, err)
	require.Nil(res)

	// query for non existent reserve pool should return an error
	req.Path = fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryLiquidity)
	req.Data = keeper.cdc.MustMarshalJSON("btc")
	res, err = querier(ctx, []string{"liquidity"}, req)
	require.NotNil(t, err)
	require.Nil(res)

	// query for set fee params
	var feeParams types.FeeParams
	req.Path = fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, types.QueryParameters, types.ParamFee)
	req.Data = []byte{}
	res, err = querier(ctx, []string{types.QueryParameters, types.ParamFee}, req)
	keeper.cdc.UnmarshalJSON(res, &feeParams)
	require.Nil(t, err)
	require.Equal(t, feeParams, types.DefaultParams())

	// query for set native asset param
	req.Path = fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, types.QueryParameters, types.ParamNativeAsset)
	res, err = querier(ctx, []string{types.QueryParameters, types.ParamNativeAsset}, req)
	require.Nil(t, err)
}

// TODO: Add tests for valid UNI balance queries and valid liquidity queries
