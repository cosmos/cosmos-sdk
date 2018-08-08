package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

func TestParams(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)

	//check that the empty keeper loads the default
	resParams := keeper.GetParams(ctx)

	//modify a params, save, and retrieve
	resParams.MaxValidators = 777
	keeper.SetParams(ctx, resParams)
	resParams = keeper.GetParams(ctx)
	require.True(t, resParams.Equal(resParams))
}

func TestPool(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)
	expPool := types.InitialPool()

	//check that the empty keeper loads the default
	resPool := keeper.GetPool(ctx)
	require.True(t, expPool.Equal(resPool))

	//modify a params, save, and retrieve
	expPool.BondedTokens = sdk.NewRat(777)
	keeper.SetPool(ctx, expPool)
	resPool = keeper.GetPool(ctx)
	require.True(t, expPool.Equal(resPool))
}
