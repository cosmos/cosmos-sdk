package slashing

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

func TestNewQuerier(t *testing.T) {
	cdc := codec.New()
	ctx, _, _, _, keeper := createTestInput(t, keeperTestParams())
	querier := NewQuerier(keeper, cdc)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	_, err := querier(ctx, []string{"parameters"}, query)
	require.NoError(t, err)
}

func TestQueryParams(t *testing.T) {
	cdc := codec.New()
	ctx, _, _, _, keeper := createTestInput(t, keeperTestParams())

	var params Params

	res, errRes := queryParams(ctx, cdc, keeper)
	require.NoError(t, errRes)

	err := cdc.UnmarshalJSON(res, &params)
	require.NoError(t, err)
	require.Equal(t, keeper.GetParams(ctx), params)
}
