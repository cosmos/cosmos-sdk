package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/poa/internal/types"
)

func TestParams(t *testing.T) {
	ctx, _, keeper, _ := CreateTestInput(t, false)
	expParams := types.DefaultParams()

	//check that the empty keeper loads the default
	resParams := keeper.GetParams(ctx)
	require.True(t, expParams.Equal(resParams))

	//modify a params, save, and retrieve
	expParams.MaxValidators = 777
	keeper.SetParams(ctx, expParams)
	resParams = keeper.GetParams(ctx)
	require.True(t, expParams.Equal(resParams))

	expParams.AcceptAllValidators = false
	expParams.IncreaseWeight = false
	keeper.SetParams(ctx, expParams)
	resParams = keeper.GetParams(ctx)
	require.True(t, expParams.Equal(resParams))
}
