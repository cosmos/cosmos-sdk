package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/slashing/internal/types"
)

func TestNewQuerier(t *testing.T) {
	ctx, _, _, _, keeper := CreateTestInput(t, TestParams())
	querier := NewQuerier(keeper)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	_, err := querier(ctx, []string{"parameters"}, query)
	require.NoError(t, err)
}

func TestQueryParams(t *testing.T) {
	cdc := codec.New()
	ctx, _, _, _, keeper := CreateTestInput(t, TestParams())

	var params types.Params

	res, errRes := queryParams(ctx, keeper)
	require.NoError(t, errRes)

	err := cdc.UnmarshalJSON(res, &params)
	require.NoError(t, err)
	require.Equal(t, keeper.GetParams(ctx), params)
}
