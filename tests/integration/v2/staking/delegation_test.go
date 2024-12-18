package staking

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"cosmossdk.io/core/header"
	"cosmossdk.io/math"
	banktestutil "cosmossdk.io/x/bank/testutil"
	"cosmossdk.io/x/staking/keeper"
	"cosmossdk.io/x/staking/testutil"
	"cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestUnbondingDelegationsMaxEntries(t *testing.T) {
	t.Parallel()
	f := initFixture(t, false)

	ctx := f.ctx

	initTokens := f.stakingKeeper.TokensFromConsensusPower(ctx, int64(1000))
	assert.NilError(t, f.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens))))

	addrDel := sdk.AccAddress([]byte("addr"))
	accAmt := math.NewInt(10000)
	bondDenom, err := f.stakingKeeper.BondDenom(ctx)
	assert.NilError(t, err)

	initCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, accAmt))
	assert.NilError(t, f.bankKeeper.MintCoins(ctx, types.ModuleName, initCoins))
	assert.NilError(t, f.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addrDel, initCoins))
	addrVal := sdk.ValAddress(addrDel)

	startTokens := f.stakingKeeper.TokensFromConsensusPower(ctx, 10)

	notBondedPool := f.stakingKeeper.GetNotBondedPool(ctx)

	assert.NilError(t, banktestutil.FundModuleAccount(ctx, f.bankKeeper, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(bondDenom, startTokens))))
	f.accountKeeper.SetModuleAccount(ctx, notBondedPool)

	// create a validator and a delegator to that validator
	validator := testutil.NewValidator(t, addrVal, PKs[0])

	validator, issuedShares := validator.AddTokensFromDel(startTokens)
	assert.DeepEqual(t, startTokens, issuedShares.RoundInt())

	validator, _ = keeper.TestingUpdateValidatorV2(f.stakingKeeper, ctx, validator, true)
	assert.Assert(math.IntEq(t, startTokens, validator.BondedTokens()))
	assert.Assert(t, validator.IsBonded())

	delegation := types.NewDelegation(addrDel.String(), addrVal.String(), issuedShares)
	assert.NilError(t, f.stakingKeeper.SetDelegation(ctx, delegation))

	maxEntries, err := f.stakingKeeper.MaxEntries(ctx)
	assert.NilError(t, err)

	oldBonded := f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	oldNotBonded := f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// should all pass
	var completionTime time.Time
	totalUnbonded := math.NewInt(0)
	for i := int64(0); i < int64(maxEntries); i++ {
		var err error
		ctx = integration.SetHeaderInfo(ctx, header.Info{Height: i})
		var amount math.Int
		completionTime, amount, err = f.stakingKeeper.Undelegate(ctx, addrDel, addrVal, math.LegacyNewDec(1))
		assert.NilError(t, err)
		totalUnbonded = totalUnbonded.Add(amount)
	}

	newBonded := f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded := f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, newBonded, oldBonded.SubRaw(int64(maxEntries))))
	assert.Assert(math.IntEq(t, newNotBonded, oldNotBonded.AddRaw(int64(maxEntries))))
	assert.Assert(math.IntEq(t, totalUnbonded, oldBonded.Sub(newBonded)))
	assert.Assert(math.IntEq(t, totalUnbonded, newNotBonded.Sub(oldNotBonded)))

	oldBonded = f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	oldNotBonded = f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// an additional unbond should fail due to max entries
	_, _, err = f.stakingKeeper.Undelegate(ctx, addrDel, addrVal, math.LegacyNewDec(1))
	assert.Error(t, err, "too many unbonding delegation entries for (delegator, validator) tuple")

	newBonded = f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded = f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	assert.Assert(math.IntEq(t, newBonded, oldBonded))
	assert.Assert(math.IntEq(t, newNotBonded, oldNotBonded))

	// mature unbonding delegations
	ctx = integration.SetHeaderInfo(ctx, header.Info{Time: completionTime})
	acc := f.accountKeeper.NewAccountWithAddress(ctx, addrDel)
	f.accountKeeper.SetAccount(ctx, acc)
	_, err = f.stakingKeeper.CompleteUnbonding(ctx, addrDel, addrVal)
	assert.NilError(t, err)

	newBonded = f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded = f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, newBonded, oldBonded))
	assert.Assert(math.IntEq(t, newNotBonded, oldNotBonded.SubRaw(int64(maxEntries))))

	oldNotBonded = f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// unbonding  should work again
	_, _, err = f.stakingKeeper.Undelegate(ctx, addrDel, addrVal, math.LegacyNewDec(1))
	assert.NilError(t, err)

	newBonded = f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded = f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, newBonded, oldBonded.SubRaw(1)))
	assert.Assert(math.IntEq(t, newNotBonded, oldNotBonded.AddRaw(1)))
}
