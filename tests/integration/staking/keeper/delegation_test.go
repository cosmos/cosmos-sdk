package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"gotest.tools/v3/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestUnbondingDelegationsMaxEntries(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx

	initTokens := f.stakingKeeper.TokensFromConsensusPower(ctx, int64(1000))
	f.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens)))

	addrDel := sdk.AccAddress([]byte("addr"))
	accAmt := sdk.NewInt(10000)
	initCoins := sdk.NewCoins(sdk.NewCoin(f.stakingKeeper.BondDenom(ctx), accAmt))
	if err := f.bankKeeper.MintCoins(ctx, types.ModuleName, initCoins); err != nil {
		panic(err)
	}

	if err := f.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addrDel, initCoins); err != nil {
		panic(err)
	}
	addrVal := sdk.ValAddress(addrDel)

	startTokens := f.stakingKeeper.TokensFromConsensusPower(ctx, 10)

	bondDenom := f.stakingKeeper.BondDenom(ctx)
	notBondedPool := f.stakingKeeper.GetNotBondedPool(ctx)

	assert.NilError(t, banktestutil.FundModuleAccount(ctx, f.bankKeeper, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(bondDenom, startTokens))))
	f.accountKeeper.SetModuleAccount(ctx, notBondedPool)

	// create a validator and a delegator to that validator
	validator := testutil.NewValidator(t, addrVal, PKs[0])

	validator, issuedShares := validator.AddTokensFromDel(startTokens)
	assert.DeepEqual(t, startTokens, issuedShares.RoundInt())

	validator = keeper.TestingUpdateValidator(f.stakingKeeper, ctx, validator, true)
	assert.Assert(math.IntEq(t, startTokens, validator.BondedTokens()))
	assert.Assert(t, validator.IsBonded())

	delegation := types.NewDelegation(addrDel, addrVal, issuedShares)
	f.stakingKeeper.SetDelegation(ctx, delegation)

	maxEntries := f.stakingKeeper.MaxEntries(ctx)

	oldBonded := f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	oldNotBonded := f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// should all pass
	var completionTime time.Time
	totalUnbonded := math.NewInt(0)
	for i := int64(0); i < int64(maxEntries); i++ {
		var err error
		ctx = ctx.WithBlockHeight(i)
		var amount math.Int
		completionTime, amount, err = f.stakingKeeper.Undelegate(ctx, addrDel, addrVal, math.LegacyNewDec(1))
		totalUnbonded = totalUnbonded.Add(amount)

		assert.NilError(t, err)
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
	_, _, err := f.stakingKeeper.Undelegate(ctx, addrDel, addrVal, math.LegacyNewDec(1))
	assert.Error(t, err, "too many unbonding delegation entries for (delegator, validator) tuple")

	newBonded = f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded = f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	assert.Assert(math.IntEq(t, newBonded, oldBonded))
	assert.Assert(math.IntEq(t, newNotBonded, oldNotBonded))

	// mature unbonding delegations
	ctx = ctx.WithBlockTime(completionTime)
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
