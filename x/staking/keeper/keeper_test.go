package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestParams(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)
	expParams := types.DefaultParams()

	//check that the empty keeper loads the default
	resParams := keeper.GetParams(ctx)
	require.True(t, expParams.Equal(resParams))

	//modify a params, save, and retrieve
	expParams.MaxValidators = 777
	keeper.SetParams(ctx, expParams)
	resParams = keeper.GetParams(ctx)
	require.True(t, expParams.Equal(resParams))
}

func TestPool(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)
	expPool := types.InitialPool()

	//check that the empty keeper loads the default
	resPool := keeper.GetPool(ctx)
	require.True(t, expPool.Equal(resPool))

	//modify a params, save, and retrieve
	expPool.BondedTokens = sdk.NewInt(777)
	keeper.SetPool(ctx, expPool)
	resPool = keeper.GetPool(ctx)
	require.True(t, expPool.Equal(resPool))
}
