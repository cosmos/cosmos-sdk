package keeper_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/math"
	"cosmossdk.io/x/distribution/types"
	stakingtestutil "cosmossdk.io/x/staking/testutil"
	stakingtypes "cosmossdk.io/x/staking/types"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	PKS = simtestutil.CreateTestPubKeys(3)

	valConsPk0 = PKS[0]
)

func setupValidatorWithCommission(t *testing.T, f *fixture, valAddr sdk.ValAddress, initialStake int64) {
	t.Helper()
	initTokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, int64(1000))
	assert.NilError(t, f.bankKeeper.MintCoins(f.sdkCtx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens))))
	assert.NilError(t, f.stakingKeeper.Params.Set(f.sdkCtx, stakingtypes.DefaultParams()))

	funds := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, int64(1000))
	assert.NilError(t, f.bankKeeper.SendCoinsFromModuleToAccount(f.sdkCtx, types.ModuleName, sdk.AccAddress(valAddr), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, funds))))
	f.accountKeeper.SetAccount(f.sdkCtx, f.accountKeeper.NewAccountWithAddress(f.sdkCtx, sdk.AccAddress(valAddr)))

	tstaking := stakingtestutil.NewHelper(t, f.sdkCtx, f.stakingKeeper)
	tstaking.Commission = stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))
	tstaking.CreateValidator(valAddr, valConsPk0, math.NewInt(initialStake), true)
}
