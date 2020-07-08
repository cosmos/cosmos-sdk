package keeper_test

import (
	gocontext "context"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestGRPCQueryUpgrade(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.UpgradeKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	t.Log("Verify that the scheduled upgrade plan can be queried")
	plan := types.Plan{Name: "test-plan", Height: 5}
	app.UpgradeKeeper.ScheduleUpgrade(ctx, plan)

	res, err := queryClient.CurrentPlan(gocontext.Background(), &types.QueryCurrentPlanRequest{})
	require.NoError(t, err)
	require.Equal(t, res.Plan, &plan)

	t.Log("Verify that the upgrade plan can be successfully applied and queried")
	ctx = ctx.WithBlockHeight(5)
	app.UpgradeKeeper.SetUpgradeHandler("test-plan", func(ctx sdk.Context, plan types.Plan) {})
	app.UpgradeKeeper.ApplyUpgrade(ctx, plan)

	res, err = queryClient.CurrentPlan(gocontext.Background(), &types.QueryCurrentPlanRequest{})
	require.NoError(t, err)
	require.Nil(t, res.Plan)

	appliedRes, appliedErr := queryClient.AppliedPlan(gocontext.Background(), &types.QueryAppliedPlanRequest{Name: "test-plan"})
	require.NoError(t, appliedErr)
	require.Equal(t, int64(5), appliedRes.Height)
}
