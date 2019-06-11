package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/internal/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

func TestNewQuerier(t *testing.T) {
	input := newTestInput(t)
	querier := NewQuerier(input.mintKeeper)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	_, err := querier(input.ctx, []string{types.QueryParameters}, query)
	require.NoError(t, err)

	_, err = querier(input.ctx, []string{types.QueryInflation}, query)
	require.NoError(t, err)

	_, err = querier(input.ctx, []string{types.QueryAnnualProvisions}, query)
	require.NoError(t, err)

	_, err = querier(input.ctx, []string{"foo"}, query)
	require.Error(t, err)
}

func TestQueryParams(t *testing.T) {
	input := newTestInput(t)

	var params types.Params

	res, sdkErr := queryParams(input.ctx, input.mintKeeper)
	require.NoError(t, sdkErr)

	err := input.cdc.UnmarshalJSON(res, &params)
	require.NoError(t, err)

	require.Equal(t, input.mintKeeper.GetParams(input.ctx), params)
}

func TestQueryInflation(t *testing.T) {
	input := newTestInput(t)

	var inflation sdk.Dec

	res, sdkErr := queryInflation(input.ctx, input.mintKeeper)
	require.NoError(t, sdkErr)

	err := input.cdc.UnmarshalJSON(res, &inflation)
	require.NoError(t, err)

	require.Equal(t, input.mintKeeper.GetMinter(input.ctx).Inflation, inflation)
}

func TestQueryAnnualProvisions(t *testing.T) {
	input := newTestInput(t)

	var annualProvisions sdk.Dec

	res, sdkErr := queryAnnualProvisions(input.ctx, input.mintKeeper)
	require.NoError(t, sdkErr)

	err := input.cdc.UnmarshalJSON(res, &annualProvisions)
	require.NoError(t, err)

	require.Equal(t, input.mintKeeper.GetMinter(input.ctx).AnnualProvisions, annualProvisions)
}
