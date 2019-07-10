package keeper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply/internal/types"
)

func TestNewQuerier(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 1000, 2)

	supplyCoins := sdk.NewCoins(
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),
		sdk.NewCoin("photon", sdk.NewInt(50)),
		sdk.NewCoin("atom", sdk.NewInt(2000)),
		sdk.NewCoin("btc", sdk.NewInt(21000000)),
	)

	keeper.SetSupply(ctx, types.NewSupply(supplyCoins))

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	querier := NewQuerier(keeper)

	bz, err := querier(ctx, []string{"other"}, query)
	require.Error(t, err)
	require.Nil(t, bz)

	queryTotalSupplyParams := types.NewQueryTotalSupplyParams(1, 20)
	bz, errRes := keeper.cdc.MarshalJSON(queryTotalSupplyParams)
	require.Nil(t, errRes)

	query.Path = fmt.Sprintf("/custom/supply/%s", types.QueryTotalSupply)
	query.Data = bz

	_, err = querier(ctx, []string{types.QueryTotalSupply}, query)
	require.Nil(t, err)

	querySupplyParams := types.NewQuerySupplyOfParams(sdk.DefaultBondDenom)
	bz, errRes = keeper.cdc.MarshalJSON(querySupplyParams)
	require.Nil(t, errRes)

	query.Path = fmt.Sprintf("/custom/supply/%s", types.QuerySupplyOf)
	query.Data = bz

	_, err = querier(ctx, []string{types.QuerySupplyOf}, query)
	require.Nil(t, err)
}

func TestQuerySupply(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 1000, 2)

	supplyCoins := sdk.NewCoins(
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),
		sdk.NewCoin("photon", sdk.NewInt(50)),
		sdk.NewCoin("atom", sdk.NewInt(2000)),
		sdk.NewCoin("btc", sdk.NewInt(21000000)),
	)

	keeper.SetSupply(ctx, types.NewSupply(supplyCoins))

	queryTotalSupplyParams := types.NewQueryTotalSupplyParams(1, 10)
	bz, errRes := keeper.cdc.MarshalJSON(queryTotalSupplyParams)
	require.Nil(t, errRes)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	query.Path = fmt.Sprintf("/custom/supply/%s", types.QueryTotalSupply)
	query.Data = bz

	res, err := queryTotalSupply(ctx, query, keeper)
	require.Nil(t, err)

	var totalCoins sdk.Coins
	errRes = keeper.cdc.UnmarshalJSON(res, &totalCoins)
	require.Nil(t, errRes)
	require.Equal(t, supplyCoins, totalCoins)

	querySupplyParams := types.NewQuerySupplyOfParams(sdk.DefaultBondDenom)
	bz, errRes = keeper.cdc.MarshalJSON(querySupplyParams)
	require.Nil(t, errRes)

	query.Path = fmt.Sprintf("/custom/supply/%s", types.QuerySupplyOf)
	query.Data = bz

	res, err = querySupplyOf(ctx, query, keeper)
	require.Nil(t, err)

	var supply sdk.Int
	errRes = supply.UnmarshalJSON(res)
	require.Nil(t, errRes)
	require.True(sdk.IntEq(t, sdk.NewInt(100), supply))

}
