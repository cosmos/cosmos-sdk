package keeper_test

import (
	gocontext "context"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/require"
)

func TestGRPCQueryParams(t *testing.T) {
	// TODO Setup tests - create queryClient

	// No current plan
	res, err := queryClient.Parameters(gocontext.Background(), &types.QueryCurrentPlanRequest{})
	require.NoError(t, err)
	require.Nil(t, res.Plan)

	// TODO Create upgrade plan. Does it need to go through gov & upgrade proposal?
	res, err := queryClient.Parameters(gocontext.Background(), &types.QueryCurrentPlanRequest{})
	require.NoError(t, err)
	require.Equal(t, res.Plan, plan)

	// TODO Wait for plan to be applied. Possible to advance N blocks programmatically in test so that plan gets applied?
	res, err := queryClient.Parameters(gocontext.Background(), &types.QueryCurrentPlanRequest{})
	require.NoError(t, err)
	require.Nil(t, res.Plan)

	res, err := queryClient.Parameters(gocontext.Background(), &types.QueryAppliedPlanRequest{Name: "test-plan"})
	require.NoError(t, err)
	require.Equal(t, res.Height, 1234)
}
