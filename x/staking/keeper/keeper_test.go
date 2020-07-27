package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app         *simapp.SimApp
	ctx         sdk.Context
	addrs       []sdk.AccAddress
	vals        []types.Validator
	queryClient types.QueryClient
}

func (suite *KeeperTestSuite) SetupTest() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	querier := keeper.Querier{Keeper: app.StakingKeeper}

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, querier)
	queryClient := types.NewQueryClient(queryHelper)

	addrs, _, validators := createValidators(ctx, app, []int64{9, 8, 7})
	header := abci.Header{
		ChainID: "HelloChain",
		Height:  5,
	}

	hi := types.NewHistoricalInfo(header, validators)
	app.StakingKeeper.SetHistoricalInfo(ctx, 5, hi)

	suite.app, suite.ctx, suite.queryClient, suite.addrs, suite.vals = app, ctx, queryClient, addrs, validators
}
func TestParams(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	expParams := types.DefaultParams()

	//check that the empty keeper loads the default
	resParams := app.StakingKeeper.GetParams(ctx)
	require.True(t, expParams.Equal(resParams))

	//modify a params, save, and retrieve
	expParams.MaxValidators = 777
	app.StakingKeeper.SetParams(ctx, expParams)
	resParams = app.StakingKeeper.GetParams(ctx)
	require.True(t, expParams.Equal(resParams))
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
