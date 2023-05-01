package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"gotest.tools/v3/assert"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestUnbondingDelegationsMaxEntries(t *testing.T) {
	_, app, ctx := createTestInput(t)

	addrDels := simtestutil.AddTestAddrsIncremental(app.BankKeeper, app.StakingKeeper, ctx, 1, sdk.NewInt(10000))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	startTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)

	bondDenom := app.StakingKeeper.BondDenom(ctx)
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)

	assert.NilError(t, banktestutil.FundModuleAccount(ctx, app.BankKeeper, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(bondDenom, startTokens))))
	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	// create a validator and a delegator to that validator
	validator := testutil.NewValidator(t, addrVals[0], PKs[0])

	validator, issuedShares := validator.AddTokensFromDel(startTokens)
	assert.DeepEqual(t, startTokens, issuedShares.RoundInt())

	validator = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validator, true)
	assert.Assert(math.IntEq(t, startTokens, validator.BondedTokens()))
	assert.Assert(t, validator.IsBonded())

	delegation := types.NewDelegation(addrDels[0], addrVals[0], issuedShares)
	app.StakingKeeper.SetDelegation(ctx, delegation)

	maxEntries := app.StakingKeeper.MaxEntries(ctx)

	oldBonded := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	oldNotBonded := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// should all pass
	var completionTime time.Time
	totalUnbonded := math.NewInt(0)
	for i := int64(0); i < int64(maxEntries); i++ {
		var err error
		ctx = ctx.WithBlockHeight(i)
		var amount math.Int
		completionTime, amount, err = app.StakingKeeper.Undelegate(ctx, addrDels[0], addrVals[0], math.LegacyNewDec(1))
		totalUnbonded = totalUnbonded.Add(amount)
		assert.NilError(t, err)
	}

	newBonded := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, newBonded, oldBonded.SubRaw(int64(maxEntries))))
	assert.Assert(math.IntEq(t, newNotBonded, oldNotBonded.AddRaw(int64(maxEntries))))
	assert.Assert(math.IntEq(t, totalUnbonded, oldBonded.Sub(newBonded)))
	assert.Assert(math.IntEq(t, totalUnbonded, newNotBonded.Sub(oldNotBonded)))

	oldBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	oldNotBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// an additional unbond should fail due to max entries
	_, _, err := app.StakingKeeper.Undelegate(ctx, addrDels[0], addrVals[0], math.LegacyNewDec(1))
	assert.Error(t, err, "too many unbonding delegation entries for (delegator, validator) tuple")

	newBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	assert.Assert(math.IntEq(t, newBonded, oldBonded))
	assert.Assert(math.IntEq(t, newNotBonded, oldNotBonded))

	// mature unbonding delegations
	ctx = ctx.WithBlockTime(completionTime)
	_, err = app.StakingKeeper.CompleteUnbonding(ctx, addrDels[0], addrVals[0])
	assert.NilError(t, err)

	newBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, newBonded, oldBonded))
	assert.Assert(math.IntEq(t, newNotBonded, oldNotBonded.SubRaw(int64(maxEntries))))

	oldNotBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// unbonding  should work again
	_, _, err = app.StakingKeeper.Undelegate(ctx, addrDels[0], addrVals[0], math.LegacyNewDec(1))
	assert.NilError(t, err)

	newBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, newBonded, oldBonded.SubRaw(1)))
	assert.Assert(math.IntEq(t, newNotBonded, oldNotBonded.AddRaw(1)))
}
