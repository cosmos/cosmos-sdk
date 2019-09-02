package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keep "github.com/cosmos/cosmos-sdk/x/mint/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/mint/internal/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

func TestNewQuerier(t *testing.T) {
	app, ctx := createTestApp(true)
	querier := keep.NewQuerier(app.MintKeeper)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	_, err := querier(ctx, []string{types.QueryParameters}, query)
	require.NoError(t, err)

	_, err = querier(ctx, []string{types.QueryInflation}, query)
	require.NoError(t, err)

	_, err = querier(ctx, []string{types.QueryAnnualProvisions}, query)
	require.NoError(t, err)

	_, err = querier(ctx, []string{"foo"}, query)
	require.Error(t, err)
}

func TestQueryParams(t *testing.T) {
	app, ctx := createTestApp(true)
	querier := keep.NewQuerier(app.MintKeeper)

	var params types.Params

	res, sdkErr := querier(ctx, []string{types.QueryParameters}, abci.RequestQuery{})
	require.NoError(t, sdkErr)

	err := app.Codec().UnmarshalJSON(res, &params)
	require.NoError(t, err)

	require.Equal(t, app.MintKeeper.GetParams(ctx), params)
}

func TestQueryInflation(t *testing.T) {
	app, ctx := createTestApp(true)
	querier := keep.NewQuerier(app.MintKeeper)

	var inflation sdk.Dec

	res, sdkErr := querier(ctx, []string{types.QueryInflation}, abci.RequestQuery{})
	require.NoError(t, sdkErr)

	err := app.Codec().UnmarshalJSON(res, &inflation)
	require.NoError(t, err)

	require.Equal(t, app.MintKeeper.GetMinter(ctx).Inflation, inflation)
}

func TestQueryAnnualProvisions(t *testing.T) {
	app, ctx := createTestApp(true)
	querier := keep.NewQuerier(app.MintKeeper)

	var annualProvisions sdk.Dec

	res, sdkErr := querier(ctx, []string{types.QueryAnnualProvisions}, abci.RequestQuery{})
	require.NoError(t, sdkErr)

	err := app.Codec().UnmarshalJSON(res, &annualProvisions)
	require.NoError(t, err)

	require.Equal(t, app.MintKeeper.GetMinter(ctx).AnnualProvisions, annualProvisions)
}
