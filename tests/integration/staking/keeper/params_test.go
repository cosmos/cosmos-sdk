package keeper_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestParams(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	expParams := types.DefaultParams()

	// check that the empty keeper loads the default
	resParams := f.stakingKeeper.GetParams(f.sdkCtx)
	assert.Assert(t, expParams.Equal(resParams))

	// modify a params, save, and retrieve
	expParams.MaxValidators = 777
	f.stakingKeeper.SetParams(f.sdkCtx, expParams)
	resParams = f.stakingKeeper.GetParams(f.sdkCtx)
	assert.Assert(t, expParams.Equal(resParams))
}
