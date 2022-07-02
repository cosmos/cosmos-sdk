package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/keeper"
	keep "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	"github.com/cosmos/cosmos-sdk/x/mint/testutil"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

type MintKeeperTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	legacyAmino *codec.LegacyAmino
	mintKeeper  keeper.Keeper
}

func (suite *MintKeeperTestSuite) SetupTest() {
	app, err := simtestutil.Setup(testutil.AppConfig,
		&suite.legacyAmino,
		&suite.mintKeeper,
	)
	suite.Require().NoError(err)

	suite.ctx = app.BaseApp.NewContext(true, tmproto.Header{})

	suite.mintKeeper.SetParams(suite.ctx, types.DefaultParams())
	suite.mintKeeper.SetMinter(suite.ctx, types.DefaultInitialMinter())
}

func (suite *MintKeeperTestSuite) TestNewQuerier(t *testing.T) {
	querier := keep.NewQuerier(suite.mintKeeper, suite.legacyAmino)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	_, err := querier(suite.ctx, []string{types.QueryParameters}, query)
	require.NoError(t, err)

	_, err = querier(suite.ctx, []string{types.QueryInflation}, query)
	require.NoError(t, err)

	_, err = querier(suite.ctx, []string{types.QueryAnnualProvisions}, query)
	require.NoError(t, err)

	_, err = querier(suite.ctx, []string{"foo"}, query)
	require.Error(t, err)
}

func (suite *MintKeeperTestSuite) TestQueryParams(t *testing.T) {
	querier := keep.NewQuerier(suite.mintKeeper, suite.legacyAmino)

	var params types.Params

	res, sdkErr := querier(suite.ctx, []string{types.QueryParameters}, abci.RequestQuery{})
	require.NoError(t, sdkErr)

	err := suite.legacyAmino.UnmarshalJSON(res, &params)
	require.NoError(t, err)

	require.Equal(t, suite.mintKeeper.GetParams(suite.ctx), params)
}

func (suite *MintKeeperTestSuite) TestQueryInflation(t *testing.T) {
	querier := keep.NewQuerier(suite.mintKeeper, suite.legacyAmino)

	var inflation sdk.Dec

	res, sdkErr := querier(suite.ctx, []string{types.QueryInflation}, abci.RequestQuery{})
	require.NoError(t, sdkErr)

	err := suite.legacyAmino.UnmarshalJSON(res, &inflation)
	require.NoError(t, err)

	require.Equal(t, suite.mintKeeper.GetMinter(suite.ctx).Inflation, inflation)
}

func (suite *MintKeeperTestSuite) TestQueryAnnualProvisions(t *testing.T) {
	querier := keep.NewQuerier(suite.mintKeeper, suite.legacyAmino)

	var annualProvisions sdk.Dec

	res, sdkErr := querier(suite.ctx, []string{types.QueryAnnualProvisions}, abci.RequestQuery{})
	require.NoError(t, sdkErr)

	err := suite.legacyAmino.UnmarshalJSON(res, &annualProvisions)
	require.NoError(t, err)

	require.Equal(t, suite.mintKeeper.GetMinter(suite.ctx).AnnualProvisions, annualProvisions)
}
