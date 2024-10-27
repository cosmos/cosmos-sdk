package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/simapp"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	app         *simapp.SimApp
	ctx         sdk.Context
	addrs       []sdk.AccAddress
	vals        []types.Validator
	queryClient types.QueryClient
	msgServer   types.MsgServer
}

func (suite *IntegrationTestSuite) SetupTest() {
	app := simapp.Setup(suite.T(), false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	querier := keeper.Querier{Keeper: app.StakingKeeper}

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, querier)
	queryClient := types.NewQueryClient(queryHelper)

	suite.msgServer = keeper.NewMsgServerImpl(app.StakingKeeper)

	addrs, _, validators := createValidators(suite.T(), ctx, app, []int64{9, 8, 7})
	header := tmproto.Header{
		ChainID: "HelloChain",
		Height:  5,
	}

	// sort a copy of the validators, so that original validators does not
	// have its order changed
	sortedVals := make([]types.Validator, len(validators))
	copy(sortedVals, validators)
	hi := types.NewHistoricalInfo(header, sortedVals, app.StakingKeeper.PowerReduction(ctx))
	app.StakingKeeper.SetHistoricalInfo(ctx, 5, &hi)

	suite.app, suite.ctx, suite.queryClient, suite.addrs, suite.vals = app, ctx, queryClient, addrs, validators
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
