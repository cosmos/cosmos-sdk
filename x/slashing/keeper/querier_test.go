package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/testslashing"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

func TestNewQuerier(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	app.SlashingKeeper.SetParams(ctx, testslashing.TestParams())
	legacyQuerierCdc := codec.NewAminoCodec(app.LegacyAmino())
	querier := keeper.NewQuerier(app.SlashingKeeper, legacyQuerierCdc.LegacyAmino)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	_, err := querier(ctx, []string{types.QueryParameters}, query)
	require.NoError(t, err)
}

func TestQueryParams(t *testing.T) {
	cdc := codec.NewLegacyAmino()
	legacyQuerierCdc := codec.NewAminoCodec(cdc)
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	app.SlashingKeeper.SetParams(ctx, testslashing.TestParams())

	querier := keeper.NewQuerier(app.SlashingKeeper, legacyQuerierCdc.LegacyAmino)

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
