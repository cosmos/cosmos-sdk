package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/simapp"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// bootstrapSlashTest creates 3 validators and bootstrap the app.
func bootstrapSlashTest(t *testing.T, power int64) (*simapp.SimApp, sdk.Context, []sdk.AccAddress, []sdk.ValAddress) {
	_, app, ctx := createTestInput(t)

	addrDels, addrVals := generateAddresses(app, ctx, 100)

	amt := app.StakingKeeper.TokensFromConsensusPower(ctx, power)
	totalSupply := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), amt.MulRaw(int64(len(addrDels)))))

	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
	assert.NilError(t, banktestutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), totalSupply))

	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	numVals := int64(3)
	bondedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), amt.MulRaw(numVals)))
	bondedPool := app.StakingKeeper.GetBondedPool(ctx)

	// set bonded pool balance
	app.AccountKeeper.SetModuleAccount(ctx, bondedPool)
	assert.NilError(t, banktestutil.FundModuleAccount(app.BankKeeper, ctx, bondedPool.GetName(), bondedCoins))

	for i := int64(0); i < numVals; i++ {
		validator := testutil.NewValidator(t, addrVals[i], PKs[i])
		validator, _ = validator.AddTokensFromDel(amt)
		validator = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validator, true)
		app.StakingKeeper.SetValidatorByConsAddr(ctx, validator)
	}

	return app, ctx, addrDels, addrVals
}

// tests slashUnbondingDelegation
func TestSlashUnbondingDelegation(t *testing.T) {
	app, ctx, addrDels, addrVals := bootstrapSlashTest(t, 10)

	fraction := sdk.NewDecWithPrec(5, 1)

	// set an unbonding delegation with expiration timestamp (beyond which the
	// unbonding delegation shouldn't be slashed)
	ubd := types.NewUnbondingDelegation(addrDels[0], addrVals[0], 0,
		time.Unix(5, 0), sdk.NewInt(10), 0)

	app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)

	// unbonding started prior to the infraction height, stakw didn't contribute
	slashAmount := app.StakingKeeper.SlashUnbondingDelegation(ctx, ubd, 1, fraction)
	assert.Assert(t, slashAmount.Equal(sdk.NewInt(0)))

	// after the expiration time, no longer eligible for slashing
	ctx = ctx.WithBlockHeader(cmtproto.Header{Time: time.Unix(10, 0)})
	app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)
	slashAmount = app.StakingKeeper.SlashUnbondingDelegation(ctx, ubd, 0, fraction)
	assert.Assert(t, slashAmount.Equal(sdk.NewInt(0)))

	// test valid slash, before expiration timestamp and to which stake contributed
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
	oldUnbondedPoolBalances := app.BankKeeper.GetAllBalances(ctx, notBondedPool.GetAddress())
	ctx = ctx.WithBlockHeader(cmtproto.Header{Time: time.Unix(0, 0)})
	app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)
	slashAmount = app.StakingKeeper.SlashUnbondingDelegation(ctx, ubd, 0, fraction)
	assert.Assert(t, slashAmount.Equal(sdk.NewInt(5)))
	ubd, found := app.StakingKeeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	assert.Assert(t, found)
	assert.Assert(t, len(ubd.Entries) == 1)

	// initial balance unchanged
	assert.DeepEqual(t, sdk.NewInt(10), ubd.Entries[0].InitialBalance)

	// balance decreased
	assert.DeepEqual(t, sdk.NewInt(5), ubd.Entries[0].Balance)
	newUnbondedPoolBalances := app.BankKeeper.GetAllBalances(ctx, notBondedPool.GetAddress())
	diffTokens := oldUnbondedPoolBalances.Sub(newUnbondedPoolBalances...)
	assert.Assert(t, diffTokens.AmountOf(app.StakingKeeper.BondDenom(ctx)).Equal(sdk.NewInt(5)))
}

// tests slashRedelegation
func TestSlashRedelegation(t *testing.T) {
	app, ctx, addrDels, addrVals := bootstrapSlashTest(t, 10)
	fraction := sdk.NewDecWithPrec(5, 1)

	// add bonded tokens to pool for (re)delegations
	startCoins := sdk.NewCoins(sdk.NewInt64Coin(app.StakingKeeper.BondDenom(ctx), 15))
	bondedPool := app.StakingKeeper.GetBondedPool(ctx)
	_ = app.BankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())

	assert.NilError(t, banktestutil.FundModuleAccount(app.BankKeeper, ctx, bondedPool.GetName(), startCoins))
	app.AccountKeeper.SetModuleAccount(ctx, bondedPool)

	// set a redelegation with an expiration timestamp beyond which the
	// redelegation shouldn't be slashed
	rd := types.NewRedelegation(addrDels[0], addrVals[0], addrVals[1], 0,
		time.Unix(5, 0), sdk.NewInt(10), math.LegacyNewDec(10), 0)

	app.StakingKeeper.SetRedelegation(ctx, rd)

	// set the associated delegation
	del := types.NewDelegation(addrDels[0], addrVals[1], math.LegacyNewDec(10))
	app.StakingKeeper.SetDelegation(ctx, del)

	// started redelegating prior to the current height, stake didn't contribute to infraction
	validator, found := app.StakingKeeper.GetValidator(ctx, addrVals[1])
	assert.Assert(t, found)
	slashAmount := app.StakingKeeper.SlashRedelegation(ctx, validator, rd, 1, fraction)
	assert.Assert(t, slashAmount.Equal(sdk.NewInt(0)))

	// after the expiration time, no longer eligible for slashing
	ctx = ctx.WithBlockHeader(cmtproto.Header{Time: time.Unix(10, 0)})
	app.StakingKeeper.SetRedelegation(ctx, rd)
	validator, found = app.StakingKeeper.GetValidator(ctx, addrVals[1])
	assert.Assert(t, found)
	slashAmount = app.StakingKeeper.SlashRedelegation(ctx, validator, rd, 0, fraction)
	assert.Assert(t, slashAmount.Equal(sdk.NewInt(0)))

	balances := app.BankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())

	// test valid slash, before expiration timestamp and to which stake contributed
	ctx = ctx.WithBlockHeader(cmtproto.Header{Time: time.Unix(0, 0)})
	app.StakingKeeper.SetRedelegation(ctx, rd)
	validator, found = app.StakingKeeper.GetValidator(ctx, addrVals[1])
	assert.Assert(t, found)
	slashAmount = app.StakingKeeper.SlashRedelegation(ctx, validator, rd, 0, fraction)
	assert.Assert(t, slashAmount.Equal(sdk.NewInt(5)))
	rd, found = app.StakingKeeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	assert.Assert(t, found)
	assert.Assert(t, len(rd.Entries) == 1)

	// end block
	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 1)

	// initialbalance unchanged
	assert.DeepEqual(t, sdk.NewInt(10), rd.Entries[0].InitialBalance)

	// shares decreased
	del, found = app.StakingKeeper.GetDelegation(ctx, addrDels[0], addrVals[1])
	assert.Assert(t, found)
	assert.Equal(t, int64(5), del.Shares.RoundInt64())

	// pool bonded tokens should decrease
	burnedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), slashAmount))
	assert.DeepEqual(t, balances.Sub(burnedCoins...), app.BankKeeper.GetAllBalances(ctx, bondedPool.GetAddress()))
}

// test slash at a negative height
// this just represents pre-genesis and should have the same effect as slashing at height 0
func TestSlashAtNegativeHeight(t *testing.T) {
	app, ctx, _, _ := bootstrapSlashTest(t, 10)
	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := sdk.NewDecWithPrec(5, 1)

	bondedPool := app.StakingKeeper.GetBondedPool(ctx)
	oldBondedPoolBalances := app.BankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())

	_, found := app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	assert.Assert(t, found)
	app.StakingKeeper.Slash(ctx, consAddr, -2, 10, fraction)

	// read updated state
	validator, found := app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	assert.Assert(t, found)

	// end block
	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 1)

	validator, found = app.StakingKeeper.GetValidator(ctx, validator.GetOperator())
	assert.Assert(t, found)
	// power decreased
	assert.Equal(t, int64(5), validator.GetConsensusPower(app.StakingKeeper.PowerReduction(ctx)))

	// pool bonded shares decreased
	newBondedPoolBalances := app.BankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())
	diffTokens := oldBondedPoolBalances.Sub(newBondedPoolBalances...).AmountOf(app.StakingKeeper.BondDenom(ctx))
	assert.DeepEqual(t, app.StakingKeeper.TokensFromConsensusPower(ctx, 5).String(), diffTokens.String())
}

// tests Slash at the current height
func TestSlashValidatorAtCurrentHeight(t *testing.T) {
	app, ctx, _, _ := bootstrapSlashTest(t, 10)
	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := sdk.NewDecWithPrec(5, 1)

	bondedPool := app.StakingKeeper.GetBondedPool(ctx)
	oldBondedPoolBalances := app.BankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())

	_, found := app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	assert.Assert(t, found)
	app.StakingKeeper.Slash(ctx, consAddr, ctx.BlockHeight(), 10, fraction)

	// read updated state
	validator, found := app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	assert.Assert(t, found)

	// end block
	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 1)

	validator, found = app.StakingKeeper.GetValidator(ctx, validator.GetOperator())
	assert.Assert(t, found)
	// power decreased
	assert.Equal(t, int64(5), validator.GetConsensusPower(app.StakingKeeper.PowerReduction(ctx)))

	// pool bonded shares decreased
	newBondedPoolBalances := app.BankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())
	diffTokens := oldBondedPoolBalances.Sub(newBondedPoolBalances...).AmountOf(app.StakingKeeper.BondDenom(ctx))
	assert.DeepEqual(t, app.StakingKeeper.TokensFromConsensusPower(ctx, 5).String(), diffTokens.String())
}

// tests Slash at a previous height with an unbonding delegation
func TestSlashWithUnbondingDelegation(t *testing.T) {
	app, ctx, addrDels, addrVals := bootstrapSlashTest(t, 10)

	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := sdk.NewDecWithPrec(5, 1)

	// set an unbonding delegation with expiration timestamp beyond which the
	// unbonding delegation shouldn't be slashed
	ubdTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 4)
	ubd := types.NewUnbondingDelegation(addrDels[0], addrVals[0], 11, time.Unix(0, 0), ubdTokens, 0)
	app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)

	// slash validator for the first time
	ctx = ctx.WithBlockHeight(12)
	bondedPool := app.StakingKeeper.GetBondedPool(ctx)
	oldBondedPoolBalances := app.BankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())

	_, found := app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	assert.Assert(t, found)
	app.StakingKeeper.Slash(ctx, consAddr, 10, 10, fraction)

	// end block
	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 1)

	// read updating unbonding delegation
	ubd, found = app.StakingKeeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	assert.Assert(t, found)
	assert.Assert(t, len(ubd.Entries) == 1)

	// balance decreased
	assert.DeepEqual(t, app.StakingKeeper.TokensFromConsensusPower(ctx, 2), ubd.Entries[0].Balance)

	// bonded tokens burned
	newBondedPoolBalances := app.BankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())
	diffTokens := oldBondedPoolBalances.Sub(newBondedPoolBalances...).AmountOf(app.StakingKeeper.BondDenom(ctx))
	assert.DeepEqual(t, app.StakingKeeper.TokensFromConsensusPower(ctx, 3), diffTokens)

	// read updated validator
	validator, found := app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	assert.Assert(t, found)

	// power decreased by 3 - 6 stake originally bonded at the time of infraction
	// was still bonded at the time of discovery and was slashed by half, 4 stake
	// bonded at the time of discovery hadn't been bonded at the time of infraction
	// and wasn't slashed
	assert.Equal(t, int64(7), validator.GetConsensusPower(app.StakingKeeper.PowerReduction(ctx)))

	// slash validator again
	ctx = ctx.WithBlockHeight(13)
	app.StakingKeeper.Slash(ctx, consAddr, 9, 10, fraction)

	ubd, found = app.StakingKeeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	assert.Assert(t, found)
	assert.Assert(t, len(ubd.Entries) == 1)

	// balance decreased again
	assert.DeepEqual(t, sdk.NewInt(0), ubd.Entries[0].Balance)

	// bonded tokens burned again
	newBondedPoolBalances = app.BankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())
	diffTokens = oldBondedPoolBalances.Sub(newBondedPoolBalances...).AmountOf(app.StakingKeeper.BondDenom(ctx))
	assert.DeepEqual(t, app.StakingKeeper.TokensFromConsensusPower(ctx, 6), diffTokens)

	// read updated validator
	validator, found = app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	assert.Assert(t, found)

	// power decreased by 3 again
	assert.Equal(t, int64(4), validator.GetConsensusPower(app.StakingKeeper.PowerReduction(ctx)))

	// slash validator again
	// all originally bonded stake has been slashed, so this will have no effect
	// on the unbonding delegation, but it will slash stake bonded since the infraction
	// this may not be the desirable behavior, ref https://github.com/cosmos/cosmos-sdk/issues/1440
	ctx = ctx.WithBlockHeight(13)
	app.StakingKeeper.Slash(ctx, consAddr, 9, 10, fraction)

	ubd, found = app.StakingKeeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	assert.Assert(t, found)
	assert.Assert(t, len(ubd.Entries) == 1)

	// balance unchanged
	assert.DeepEqual(t, sdk.NewInt(0), ubd.Entries[0].Balance)

	// bonded tokens burned again
	newBondedPoolBalances = app.BankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())
	diffTokens = oldBondedPoolBalances.Sub(newBondedPoolBalances...).AmountOf(app.StakingKeeper.BondDenom(ctx))
	assert.DeepEqual(t, app.StakingKeeper.TokensFromConsensusPower(ctx, 9), diffTokens)

	// read updated validator
	validator, found = app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	assert.Assert(t, found)

	// power decreased by 3 again
	assert.Equal(t, int64(1), validator.GetConsensusPower(app.StakingKeeper.PowerReduction(ctx)))

	// slash validator again
	// all originally bonded stake has been slashed, so this will have no effect
	// on the unbonding delegation, but it will slash stake bonded since the infraction
	// this may not be the desirable behavior, ref https://github.com/cosmos/cosmos-sdk/issues/1440
	ctx = ctx.WithBlockHeight(13)
	app.StakingKeeper.Slash(ctx, consAddr, 9, 10, fraction)

	ubd, found = app.StakingKeeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	assert.Assert(t, found)
	assert.Assert(t, len(ubd.Entries) == 1)

	// balance unchanged
	assert.DeepEqual(t, sdk.NewInt(0), ubd.Entries[0].Balance)

	// just 1 bonded token burned again since that's all the validator now has
	newBondedPoolBalances = app.BankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())
	diffTokens = oldBondedPoolBalances.Sub(newBondedPoolBalances...).AmountOf(app.StakingKeeper.BondDenom(ctx))
	assert.DeepEqual(t, app.StakingKeeper.TokensFromConsensusPower(ctx, 10), diffTokens)

	// apply TM updates
	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, -1)

	// read updated validator
	// power decreased by 1 again, validator is out of stake
	// validator should be in unbonding period
	validator, _ = app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	assert.Equal(t, validator.GetStatus(), types.Unbonding)
}

// tests Slash at a previous height with a redelegation
func TestSlashWithRedelegation(t *testing.T) {
	app, ctx, addrDels, addrVals := bootstrapSlashTest(t, 10)
	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := sdk.NewDecWithPrec(5, 1)
	bondDenom := app.StakingKeeper.BondDenom(ctx)

	// set a redelegation
	rdTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 6)
	rd := types.NewRedelegation(addrDels[0], addrVals[0], addrVals[1], 11, time.Unix(0, 0), rdTokens, sdk.NewDecFromInt(rdTokens), 0)
	app.StakingKeeper.SetRedelegation(ctx, rd)

	// set the associated delegation
	del := types.NewDelegation(addrDels[0], addrVals[1], sdk.NewDecFromInt(rdTokens))
	app.StakingKeeper.SetDelegation(ctx, del)

	// update bonded tokens
	bondedPool := app.StakingKeeper.GetBondedPool(ctx)
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
	rdCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, rdTokens.MulRaw(2)))

	assert.NilError(t, banktestutil.FundModuleAccount(app.BankKeeper, ctx, bondedPool.GetName(), rdCoins))

	app.AccountKeeper.SetModuleAccount(ctx, bondedPool)

	oldBonded := app.BankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount
	oldNotBonded := app.BankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), bondDenom).Amount

	// slash validator
	ctx = ctx.WithBlockHeight(12)
	_, found := app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	assert.Assert(t, found)

	require.NotPanics(t, func() {
		app.StakingKeeper.Slash(ctx, consAddr, 10, 10, fraction)
	})
	burnAmount := sdk.NewDecFromInt(app.StakingKeeper.TokensFromConsensusPower(ctx, 10)).Mul(fraction).TruncateInt()

	bondedPool = app.StakingKeeper.GetBondedPool(ctx)
	notBondedPool = app.StakingKeeper.GetNotBondedPool(ctx)

	// burn bonded tokens from only from delegations
	bondedPoolBalance := app.BankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldBonded.Sub(burnAmount), bondedPoolBalance))

	notBondedPoolBalance := app.BankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldNotBonded, notBondedPoolBalance))
	oldBonded = app.BankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount

	// read updating redelegation
	rd, found = app.StakingKeeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	assert.Assert(t, found)
	assert.Assert(t, len(rd.Entries) == 1)
	// read updated validator
	validator, found := app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	assert.Assert(t, found)
	// power decreased by 2 - 4 stake originally bonded at the time of infraction
	// was still bonded at the time of discovery and was slashed by half, 4 stake
	// bonded at the time of discovery hadn't been bonded at the time of infraction
	// and wasn't slashed
	assert.Equal(t, int64(8), validator.GetConsensusPower(app.StakingKeeper.PowerReduction(ctx)))

	// slash the validator again
	_, found = app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	assert.Assert(t, found)

	require.NotPanics(t, func() {
		app.StakingKeeper.Slash(ctx, consAddr, 10, 10, math.LegacyOneDec())
	})
	burnAmount = app.StakingKeeper.TokensFromConsensusPower(ctx, 7)

	// read updated pool
	bondedPool = app.StakingKeeper.GetBondedPool(ctx)
	notBondedPool = app.StakingKeeper.GetNotBondedPool(ctx)

	// seven bonded tokens burned
	bondedPoolBalance = app.BankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldBonded.Sub(burnAmount), bondedPoolBalance))
	assert.Assert(math.IntEq(t, oldNotBonded, notBondedPoolBalance))

	bondedPoolBalance = app.BankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldBonded.Sub(burnAmount), bondedPoolBalance))

	notBondedPoolBalance = app.BankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldNotBonded, notBondedPoolBalance))
	oldBonded = app.BankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount

	// read updating redelegation
	rd, found = app.StakingKeeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	assert.Assert(t, found)
	assert.Assert(t, len(rd.Entries) == 1)
	// read updated validator
	validator, found = app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	assert.Assert(t, found)
	// power decreased by 4
	assert.Equal(t, int64(4), validator.GetConsensusPower(app.StakingKeeper.PowerReduction(ctx)))

	// slash the validator again, by 100%
	ctx = ctx.WithBlockHeight(12)
	_, found = app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	assert.Assert(t, found)

	require.NotPanics(t, func() {
		app.StakingKeeper.Slash(ctx, consAddr, 10, 10, math.LegacyOneDec())
	})

	burnAmount = sdk.NewDecFromInt(app.StakingKeeper.TokensFromConsensusPower(ctx, 10)).Mul(math.LegacyOneDec()).TruncateInt()
	burnAmount = burnAmount.Sub(math.LegacyOneDec().MulInt(rdTokens).TruncateInt())

	// read updated pool
	bondedPool = app.StakingKeeper.GetBondedPool(ctx)
	notBondedPool = app.StakingKeeper.GetNotBondedPool(ctx)

	bondedPoolBalance = app.BankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldBonded.Sub(burnAmount), bondedPoolBalance))
	notBondedPoolBalance = app.BankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldNotBonded, notBondedPoolBalance))
	oldBonded = app.BankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount

	// read updating redelegation
	rd, found = app.StakingKeeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	assert.Assert(t, found)
	assert.Assert(t, len(rd.Entries) == 1)
	// apply TM updates
	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, -1)
	// read updated validator
	// validator decreased to zero power, should be in unbonding period
	validator, _ = app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	assert.Equal(t, validator.GetStatus(), types.Unbonding)

	// slash the validator again, by 100%
	// no stake remains to be slashed
	ctx = ctx.WithBlockHeight(12)
	// validator still in unbonding period
	validator, _ = app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	assert.Equal(t, validator.GetStatus(), types.Unbonding)

	require.NotPanics(t, func() {
		app.StakingKeeper.Slash(ctx, consAddr, 10, 10, math.LegacyOneDec())
	})

	// read updated pool
	bondedPool = app.StakingKeeper.GetBondedPool(ctx)
	notBondedPool = app.StakingKeeper.GetNotBondedPool(ctx)

	bondedPoolBalance = app.BankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldBonded, bondedPoolBalance))
	notBondedPoolBalance = app.BankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldNotBonded, notBondedPoolBalance))

	// read updating redelegation
	rd, found = app.StakingKeeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	assert.Assert(t, found)
	assert.Assert(t, len(rd.Entries) == 1)
	// read updated validator
	// power still zero, still in unbonding period
	validator, _ = app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	assert.Equal(t, validator.GetStatus(), types.Unbonding)
}

// tests Slash at a previous height with both an unbonding delegation and a redelegation
func TestSlashBoth(t *testing.T) {
	app, ctx, addrDels, addrVals := bootstrapSlashTest(t, 10)
	fraction := sdk.NewDecWithPrec(5, 1)
	bondDenom := app.StakingKeeper.BondDenom(ctx)

	// set a redelegation with expiration timestamp beyond which the
	// redelegation shouldn't be slashed
	rdATokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 6)
	rdA := types.NewRedelegation(addrDels[0], addrVals[0], addrVals[1], 11, time.Unix(0, 0), rdATokens, sdk.NewDecFromInt(rdATokens), 0)
	app.StakingKeeper.SetRedelegation(ctx, rdA)

	// set the associated delegation
	delA := types.NewDelegation(addrDels[0], addrVals[1], sdk.NewDecFromInt(rdATokens))
	app.StakingKeeper.SetDelegation(ctx, delA)

	// set an unbonding delegation with expiration timestamp (beyond which the
	// unbonding delegation shouldn't be slashed)
	ubdATokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 4)
	ubdA := types.NewUnbondingDelegation(addrDels[0], addrVals[0], 11,
		time.Unix(0, 0), ubdATokens, 0)
	app.StakingKeeper.SetUnbondingDelegation(ctx, ubdA)

	bondedCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, rdATokens.MulRaw(2)))
	notBondedCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, ubdATokens))

	// update bonded tokens
	bondedPool := app.StakingKeeper.GetBondedPool(ctx)
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)

	assert.NilError(t, banktestutil.FundModuleAccount(app.BankKeeper, ctx, bondedPool.GetName(), bondedCoins))
	assert.NilError(t, banktestutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), notBondedCoins))

	app.AccountKeeper.SetModuleAccount(ctx, bondedPool)
	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	oldBonded := app.BankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount
	oldNotBonded := app.BankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), bondDenom).Amount
	// slash validator
	ctx = ctx.WithBlockHeight(12)
	_, found := app.StakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(PKs[0]))
	assert.Assert(t, found)
	consAddr0 := sdk.ConsAddress(PKs[0].Address())
	app.StakingKeeper.Slash(ctx, consAddr0, 10, 10, fraction)

	burnedNotBondedAmount := fraction.MulInt(ubdATokens).TruncateInt()
	burnedBondAmount := sdk.NewDecFromInt(app.StakingKeeper.TokensFromConsensusPower(ctx, 10)).Mul(fraction).TruncateInt()
	burnedBondAmount = burnedBondAmount.Sub(burnedNotBondedAmount)

	// read updated pool
	bondedPool = app.StakingKeeper.GetBondedPool(ctx)
	notBondedPool = app.StakingKeeper.GetNotBondedPool(ctx)

	bondedPoolBalance := app.BankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldBonded.Sub(burnedBondAmount), bondedPoolBalance))

	notBondedPoolBalance := app.BankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldNotBonded.Sub(burnedNotBondedAmount), notBondedPoolBalance))

	// read updating redelegation
	rdA, found = app.StakingKeeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	assert.Assert(t, found)
	assert.Assert(t, len(rdA.Entries) == 1)
	// read updated validator
	validator, found := app.StakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(PKs[0]))
	assert.Assert(t, found)
	// power not decreased, all stake was bonded since
	assert.Equal(t, int64(10), validator.GetConsensusPower(app.StakingKeeper.PowerReduction(ctx)))
}

func TestSlashAmount(t *testing.T) {
	app, ctx, _, _ := bootstrapSlashTest(t, 10)
	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := sdk.NewDecWithPrec(5, 1)
	burnedCoins := app.StakingKeeper.Slash(ctx, consAddr, ctx.BlockHeight(), 10, fraction)
	assert.Assert(t, burnedCoins.GT(math.ZeroInt()))

	// test the case where the validator was not found, which should return no coins
	_, addrVals := generateAddresses(app, ctx, 100)
	noBurned := app.StakingKeeper.Slash(ctx, sdk.ConsAddress(addrVals[0]), ctx.BlockHeight(), 10, fraction)
	assert.Assert(t, sdk.NewInt(0).Equal(noBurned))
}
