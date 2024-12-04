package distribution

import (
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/math"
	"cosmossdk.io/x/distribution/types"
	stakingtestutil "cosmossdk.io/x/staking/testutil"
	stakingtypes "cosmossdk.io/x/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func setupValidatorWithCommission(t *testing.T, f *fixture, valAddr sdk.ValAddress, initialStake int64) {
	t.Helper()
	initTokens := f.stakingKeeper.TokensFromConsensusPower(f.ctx, int64(1000))
	assert.NilError(t, f.bankKeeper.MintCoins(f.ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens))))
	assert.NilError(t, f.stakingKeeper.Params.Set(f.ctx, stakingtypes.DefaultParams()))

	funds := f.stakingKeeper.TokensFromConsensusPower(f.ctx, int64(1000))
	assert.NilError(t, f.bankKeeper.SendCoinsFromModuleToAccount(f.ctx, types.ModuleName, sdk.AccAddress(valAddr), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, funds))))
	f.authKeeper.SetAccount(f.ctx, f.authKeeper.NewAccountWithAddress(f.ctx, sdk.AccAddress(valAddr)))

	tstaking := stakingtestutil.NewHelper(t, f.ctx, f.stakingKeeper)
	tstaking.Commission = stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))
	tstaking.CreateValidator(valAddr, valConsPk0, math.NewInt(initialStake), true)
}
