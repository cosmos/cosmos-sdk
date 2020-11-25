package keeper_test

import (
	gocontext "context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

type UpgradeTestSuite struct {
	suite.Suite

	app         *simapp.SimApp
	ctx         sdk.Context
	queryClient types.QueryClient
}

func (suite *UpgradeTestSuite) SetupTest() {
	suite.app = simapp.Setup(false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.UpgradeKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
}

func (suite *UpgradeTestSuite) TestQueryCurrentPlan() {
	var (
		req         *types.QueryCurrentPlanRequest
		expResponse types.QueryCurrentPlanResponse
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"without current upgrade plan",
			func() {
				req = &types.QueryCurrentPlanRequest{}
				expResponse = types.QueryCurrentPlanResponse{}
			},
			true,
		},
		{
			"with current upgrade plan",
			func() {
				plan := types.Plan{Name: "test-plan", Height: 5}
				suite.app.UpgradeKeeper.ScheduleUpgrade(suite.ctx, plan)

				req = &types.QueryCurrentPlanRequest{}
				expResponse = types.QueryCurrentPlanResponse{Plan: &plan}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			res, err := suite.queryClient.CurrentPlan(gocontext.Background(), req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(&expResponse, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *UpgradeTestSuite) TestAppliedCurrentPlan() {
	var (
		req       *types.QueryAppliedPlanRequest
		expHeight int64
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"with non-existent upgrade plan",
			func() {
				req = &types.QueryAppliedPlanRequest{Name: "foo"}
			},
			true,
		},
		{
			"with applied upgrade plan",
			func() {
				expHeight = 5

				planName := "test-plan"
				plan := types.Plan{Name: planName, Height: expHeight}
				suite.app.UpgradeKeeper.ScheduleUpgrade(suite.ctx, plan)

				suite.ctx = suite.ctx.WithBlockHeight(expHeight)
				suite.app.UpgradeKeeper.SetUpgradeHandler(planName, func(ctx sdk.Context, plan types.Plan) {})
				suite.app.UpgradeKeeper.ApplyUpgrade(suite.ctx, plan)

				req = &types.QueryAppliedPlanRequest{Name: planName}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			res, err := suite.queryClient.AppliedPlan(gocontext.Background(), req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(expHeight, res.Height)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}
