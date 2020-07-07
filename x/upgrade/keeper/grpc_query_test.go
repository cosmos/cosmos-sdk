package keeper_test

import (
	gocontext "context"
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

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
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.UpgradeKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.app = app
	suite.ctx = ctx
	suite.queryClient = queryClient
}

func (suite *UpgradeTestSuite) TestGRPCQuery() {
	queryClient := suite.queryClient

	suite.T().Log("Verify that the scheduled upgrade plan can be queried")
	plan := types.Plan{Name: "test-plan", Height: 5}
	suite.app.UpgradeKeeper.ScheduleUpgrade(suite.ctx, plan)

	res, err := queryClient.CurrentPlan(gocontext.Background(), &types.QueryCurrentPlanRequest{})
	suite.NoError(err)
	suite.Equal(res.Plan, &plan)

	suite.T().Log("Verify that the upgrade plan can be successfully applied and queried")
	suite.ctx = suite.ctx.WithBlockHeight(5)
	suite.app.UpgradeKeeper.SetUpgradeHandler("test-plan", func(ctx sdk.Context, plan types.Plan) {})
	suite.app.UpgradeKeeper.ApplyUpgrade(suite.ctx, plan)

	res, err = queryClient.CurrentPlan(gocontext.Background(), &types.QueryCurrentPlanRequest{})
	suite.NoError(err)
	suite.Nil(res.Plan)

	appliedRes, appliedErr := queryClient.AppliedPlan(gocontext.Background(), &types.QueryAppliedPlanRequest{Name: "test-plan"})
	suite.NoError(appliedErr)
	suite.Equal(int64(5), appliedRes.Height)
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}
