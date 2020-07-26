package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/KiraCore/cosmos-sdk/codec"
	"github.com/KiraCore/cosmos-sdk/simapp"
	"github.com/KiraCore/cosmos-sdk/x/slashing/keeper"
	"github.com/KiraCore/cosmos-sdk/x/slashing/types"
)

func TestNewQuerier(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})
	app.SlashingKeeper.SetParams(ctx, keeper.TestParams())
	legacyQuerierCdc := codec.NewAminoCodec(app.Codec())
	querier := keeper.NewQuerier(app.SlashingKeeper, legacyQuerierCdc)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	_, err := querier(ctx, []string{types.QueryParameters}, query)
	require.NoError(t, err)
}

func TestQueryParams(t *testing.T) {
	cdc := codec.New()
	legacyQuerierCdc := codec.NewAminoCodec(cdc)
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})
	app.SlashingKeeper.SetParams(ctx, keeper.TestParams())

	querier := keeper.NewQuerier(app.SlashingKeeper, legacyQuerierCdc)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	var params types.Params

	res, err := querier(ctx, []string{types.QueryParameters}, query)
	require.NoError(t, err)

	err = cdc.UnmarshalJSON(res, &params)
	require.NoError(t, err)
	require.Equal(t, app.SlashingKeeper.GetParams(ctx), params)
}
