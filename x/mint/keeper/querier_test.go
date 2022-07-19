package keeper_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/mint/keeper"
	keep "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttestutil "github.com/cosmos/cosmos-sdk/x/mint/testutil"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

type MintKeeperTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	legacyAmino *codec.LegacyAmino
	mintKeeper  keeper.Keeper
}

func (suite *MintKeeperTestSuite) SetupTest() {
	encCfg := moduletestutil.MakeTestEncodingConfig(mint.AppModuleBasic{})
	key := sdk.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, sdk.NewTransientStoreKey("transient_test"))
	suite.ctx = testCtx.Ctx

	// gomock initializations
	ctrl := gomock.NewController(suite.T())
	accountKeeper := minttestutil.NewMockAccountKeeper(ctrl)
	bankKeeper := minttestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := minttestutil.NewMockStakingKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return(sdk.AccAddress{})

	suite.mintKeeper = keeper.NewKeeper(
		encCfg.Codec,
		key,
		stakingKeeper,
		accountKeeper,
		bankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	err := suite.mintKeeper.SetParams(suite.ctx, types.DefaultParams())
	suite.Require().NoError(err)
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
