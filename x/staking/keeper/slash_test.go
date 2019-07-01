package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TODO integrate with test_common.go helper (CreateTestInput)
// setup helper function - creates two validators
func setupHelper(t *testing.T, power int64) (sdk.Context, Keeper, types.Params) {

	// setup
	ctx, _, keeper, _ := CreateTestInput(t, false, power)
	params := keeper.GetParams(ctx)
	numVals := int64(3)
	amt := sdk.TokensFromConsensusPower(power)
	bondedCoins := sdk.NewCoins(sdk.NewCoin(keeper.BondDenom(ctx), amt.MulRaw(numVals)))

	bondedPool := keeper.GetBondedPool(ctx)
	err := bondedPool.SetCoins(bondedCoins)
	require.NoError(t, err)
	keeper.supplyKeeper.SetModuleAccount(ctx, bondedPool)

	// add numVals validators
	for i := int64(0); i < numVals; i++ {
		validator := types.NewValidator(addrVals[i], PKs[i], types.Description{})
		validator, _ = validator.AddTokensFromDel(amt)
		validator = TestingUpdateValidator(keeper, ctx, validator, true)
		keeper.SetValidatorByConsAddr(ctx, validator)
	}

	return ctx, keeper, params
}

//_________________________________________________________________________________

// tests Jail, Unjail
func TestRevocation(t *testing.T) {

	// setup
	ctx, keeper, _ := setupHelper(t, 10)
	addr := addrVals[0]
	consAddr := sdk.ConsAddress(PKs[0].Address())

	// initial state
	val, found := keeper.GetValidator(ctx, addr)
	require.True(t, found)
	require.False(t, val.IsJailed())

	// test jail
	keeper.Jail(ctx, consAddr)
	val, found = keeper.GetValidator(ctx, addr)
	require.True(t, found)
	require.True(t, val.IsJailed())

	// test unjail
	keeper.Unjail(ctx, consAddr)
	val, found = keeper.GetValidator(ctx, addr)
	require.True(t, found)
	require.False(t, val.IsJailed())
}

// tests slashUnbondingDelegation
func TestSlashUnbondingDelegation(t *testing.T) {
	ctx, keeper, _ := setupHelper(t, 10)
	fraction := sdk.NewDecWithPrec(5, 1)

	// set an unbonding delegation with expiration timestamp (beyond which the
	// unbonding delegation shouldn't be slashed)
	ubd := types.NewUnbondingDelegation(addrDels[0], addrVals[0], 0,
		time.Unix(5, 0), sdk.NewInt(10))

	keeper.SetUnbondingDelegation(ctx, ubd)

	// unbonding started prior to the infraction height, stakw didn't contribute
	slashAmount := keeper.slashUnbondingDelegation(ctx, ubd, 1, fraction)
	require.Equal(t, int64(0), slashAmount.Int64())

	// after the expiration time, no longer eligible for slashing
	ctx = ctx.WithBlockHeader(abci.Header{Time: time.Unix(10, 0)})
	keeper.SetUnbondingDelegation(ctx, ubd)
	slashAmount = keeper.slashUnbondingDelegation(ctx, ubd, 0, fraction)
	require.Equal(t, int64(0), slashAmount.Int64())

	// test valid slash, before expiration timestamp and to which stake contributed
	oldUnbondedPool := keeper.GetNotBondedPool(ctx)
	ctx = ctx.WithBlockHeader(abci.Header{Time: time.Unix(0, 0)})
	keeper.SetUnbondingDelegation(ctx, ubd)
	slashAmount = keeper.slashUnbondingDelegation(ctx, ubd, 0, fraction)
	require.Equal(t, int64(5), slashAmount.Int64())
	ubd, found := keeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)

	// initial balance unchanged
	require.Equal(t, sdk.NewInt(10), ubd.Entries[0].InitialBalance)

	// balance decreased
	require.Equal(t, sdk.NewInt(5), ubd.Entries[0].Balance)
	newUnbondedPool := keeper.GetNotBondedPool(ctx)
	diffTokens := oldUnbondedPool.GetCoins().Sub(newUnbondedPool.GetCoins()).AmountOf(keeper.BondDenom(ctx))
	require.Equal(t, int64(5), diffTokens.Int64())
}

// tests slashRedelegation
func TestSlashRedelegation(t *testing.T) {
	ctx, keeper, _ := setupHelper(t, 10)
	fraction := sdk.NewDecWithPrec(5, 1)

	// add bonded tokens to pool for (re)delegations
	startCoins := sdk.NewCoins(sdk.NewInt64Coin(keeper.BondDenom(ctx), 15))
	bondedPool := keeper.GetBondedPool(ctx)
	err := bondedPool.SetCoins(bondedPool.GetCoins().Add(startCoins))
	require.NoError(t, err)
	keeper.supplyKeeper.SetModuleAccount(ctx, bondedPool)

	// set a redelegation with an expiration timestamp beyond which the
	// redelegation shouldn't be slashed
	rd := types.NewRedelegation(addrDels[0], addrVals[0], addrVals[1], 0,
		time.Unix(5, 0), sdk.NewInt(10), sdk.NewDec(10))

	keeper.SetRedelegation(ctx, rd)

	// set the associated delegation
	del := types.NewDelegation(addrDels[0], addrVals[1], sdk.NewDec(10))
	keeper.SetDelegation(ctx, del)

	// started redelegating prior to the current height, stake didn't contribute to infraction
	validator, found := keeper.GetValidator(ctx, addrVals[1])
	require.True(t, found)
	slashAmount := keeper.slashRedelegation(ctx, validator, rd, 1, fraction)
	require.Equal(t, int64(0), slashAmount.Int64())

	// after the expiration time, no longer eligible for slashing
	ctx = ctx.WithBlockHeader(abci.Header{Time: time.Unix(10, 0)})
	keeper.SetRedelegation(ctx, rd)
	validator, found = keeper.GetValidator(ctx, addrVals[1])
	require.True(t, found)
	slashAmount = keeper.slashRedelegation(ctx, validator, rd, 0, fraction)
	require.Equal(t, int64(0), slashAmount.Int64())

	// test valid slash, before expiration timestamp and to which stake contributed
	ctx = ctx.WithBlockHeader(abci.Header{Time: time.Unix(0, 0)})
	keeper.SetRedelegation(ctx, rd)
	validator, found = keeper.GetValidator(ctx, addrVals[1])
	require.True(t, found)
	slashAmount = keeper.slashRedelegation(ctx, validator, rd, 0, fraction)
	require.Equal(t, int64(5), slashAmount.Int64())
	rd, found = keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.True(t, found)
	require.Len(t, rd.Entries, 1)

	// end block
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(updates))

	// initialbalance unchanged
	require.Equal(t, sdk.NewInt(10), rd.Entries[0].InitialBalance)

	// shares decreased
	del, found = keeper.GetDelegation(ctx, addrDels[0], addrVals[1])
	require.True(t, found)
	require.Equal(t, int64(5), del.Shares.RoundInt64())

	// pool bonded tokens should decrease
	burnedCoins := sdk.NewCoins(sdk.NewCoin(keeper.BondDenom(ctx), slashAmount))
	newBondedPool := keeper.GetBondedPool(ctx)
	require.Equal(t, bondedPool.GetCoins().Sub(burnedCoins), newBondedPool.GetCoins())
}

// tests Slash at a future height (must panic)
func TestSlashAtFutureHeight(t *testing.T) {
	ctx, keeper, _ := setupHelper(t, 10)
	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := sdk.NewDecWithPrec(5, 1)
	require.Panics(t, func() { keeper.Slash(ctx, consAddr, 1, 10, fraction) })
}

// test slash at a negative height
// this just represents pre-genesis and should have the same effect as slashing at height 0
func TestSlashAtNegativeHeight(t *testing.T) {
	ctx, keeper, _ := setupHelper(t, 10)
	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := sdk.NewDecWithPrec(5, 1)

	oldBondedPool := keeper.GetBondedPool(ctx)
	validator, found := keeper.GetValidatorByConsAddr(ctx, consAddr)
	require.True(t, found)
	keeper.Slash(ctx, consAddr, -2, 10, fraction)

	// read updated state
	validator, found = keeper.GetValidatorByConsAddr(ctx, consAddr)
	require.True(t, found)
	newBondedPool := keeper.GetBondedPool(ctx)

	// end block
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(updates), "cons addr: %v, updates: %v", []byte(consAddr), updates)

	validator = keeper.mustGetValidator(ctx, validator.OperatorAddress)
	// power decreased
	require.Equal(t, int64(5), validator.GetConsensusPower())
	// pool bonded shares decreased
	diffTokens := oldBondedPool.GetCoins().Sub(newBondedPool.GetCoins()).AmountOf(keeper.BondDenom(ctx))
	require.Equal(t, sdk.TokensFromConsensusPower(5), diffTokens)
}

// tests Slash at the current height
func TestSlashValidatorAtCurrentHeight(t *testing.T) {
	ctx, keeper, _ := setupHelper(t, 10)
	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := sdk.NewDecWithPrec(5, 1)

	oldBondedPool := keeper.GetBondedPool(ctx)
	validator, found := keeper.GetValidatorByConsAddr(ctx, consAddr)
	require.True(t, found)
	keeper.Slash(ctx, consAddr, ctx.BlockHeight(), 10, fraction)

	// read updated state
	validator, found = keeper.GetValidatorByConsAddr(ctx, consAddr)
	require.True(t, found)
	newBondedPool := keeper.GetBondedPool(ctx)

	// end block
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(updates), "cons addr: %v, updates: %v", []byte(consAddr), updates)

	validator = keeper.mustGetValidator(ctx, validator.OperatorAddress)
	// power decreased
	require.Equal(t, int64(5), validator.GetConsensusPower())
	// pool bonded shares decreased
	diffTokens := oldBondedPool.GetCoins().Sub(newBondedPool.GetCoins()).AmountOf(keeper.BondDenom(ctx))
	require.Equal(t, sdk.TokensFromConsensusPower(5), diffTokens)
}

// tests Slash at a previous height with an unbonding delegation
func TestSlashWithUnbondingDelegation(t *testing.T) {
	ctx, keeper, _ := setupHelper(t, 10)
	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := sdk.NewDecWithPrec(5, 1)

	// set an unbonding delegation with expiration timestamp beyond which the
	// unbonding delegation shouldn't be slashed
	ubdTokens := sdk.TokensFromConsensusPower(4)
	ubd := types.NewUnbondingDelegation(addrDels[0], addrVals[0], 11,
		time.Unix(0, 0), ubdTokens)
	keeper.SetUnbondingDelegation(ctx, ubd)

	// slash validator for the first time
	ctx = ctx.WithBlockHeight(12)
	oldBondedPool := keeper.GetBondedPool(ctx)
	validator, found := keeper.GetValidatorByConsAddr(ctx, consAddr)
	require.True(t, found)
	keeper.Slash(ctx, consAddr, 10, 10, fraction)

	// end block
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(updates))

	// read updating unbonding delegation
	ubd, found = keeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)
	// balance decreased
	require.Equal(t, sdk.TokensFromConsensusPower(2), ubd.Entries[0].Balance)
	// read updated pool
	newBondedPool := keeper.GetBondedPool(ctx)
	// bonded tokens burned
	diffTokens := oldBondedPool.GetCoins().Sub(newBondedPool.GetCoins()).AmountOf(keeper.BondDenom(ctx))
	require.Equal(t, sdk.TokensFromConsensusPower(3), diffTokens)
	// read updated validator
	validator, found = keeper.GetValidatorByConsAddr(ctx, consAddr)
	require.True(t, found)
	// power decreased by 3 - 6 stake originally bonded at the time of infraction
	// was still bonded at the time of discovery and was slashed by half, 4 stake
	// bonded at the time of discovery hadn't been bonded at the time of infraction
	// and wasn't slashed
	require.Equal(t, int64(7), validator.GetConsensusPower())

	// slash validator again
	ctx = ctx.WithBlockHeight(13)
	keeper.Slash(ctx, consAddr, 9, 10, fraction)
	ubd, found = keeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)
	// balance decreased again
	require.Equal(t, sdk.NewInt(0), ubd.Entries[0].Balance)
	// read updated pool
	newBondedPool = keeper.GetBondedPool(ctx)
	// bonded tokens burned again
	diffTokens = oldBondedPool.GetCoins().Sub(newBondedPool.GetCoins()).AmountOf(keeper.BondDenom(ctx))
	require.Equal(t, sdk.TokensFromConsensusPower(6), diffTokens)
	// read updated validator
	validator, found = keeper.GetValidatorByConsAddr(ctx, consAddr)
	require.True(t, found)
	// power decreased by 3 again
	require.Equal(t, int64(4), validator.GetConsensusPower())

	// slash validator again
	// all originally bonded stake has been slashed, so this will have no effect
	// on the unbonding delegation, but it will slash stake bonded since the infraction
	// this may not be the desirable behaviour, ref https://github.com/cosmos/cosmos-sdk/issues/1440
	ctx = ctx.WithBlockHeight(13)
	keeper.Slash(ctx, consAddr, 9, 10, fraction)
	ubd, found = keeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)
	// balance unchanged
	require.Equal(t, sdk.NewInt(0), ubd.Entries[0].Balance)
	// read updated pool
	newBondedPool = keeper.GetBondedPool(ctx)
	// bonded tokens burned again
	diffTokens = oldBondedPool.GetCoins().Sub(newBondedPool.GetCoins()).AmountOf(keeper.BondDenom(ctx))
	require.Equal(t, sdk.TokensFromConsensusPower(9), diffTokens)
	// read updated validator
	validator, found = keeper.GetValidatorByConsAddr(ctx, consAddr)
	require.True(t, found)
	// power decreased by 3 again
	require.Equal(t, int64(1), validator.GetConsensusPower())

	// slash validator again
	// all originally bonded stake has been slashed, so this will have no effect
	// on the unbonding delegation, but it will slash stake bonded since the infraction
	// this may not be the desirable behaviour, ref https://github.com/cosmos/cosmos-sdk/issues/1440
	ctx = ctx.WithBlockHeight(13)
	keeper.Slash(ctx, consAddr, 9, 10, fraction)
	ubd, found = keeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)
	// balance unchanged
	require.Equal(t, sdk.NewInt(0), ubd.Entries[0].Balance)
	// read updated pool
	newBondedPool = keeper.GetBondedPool(ctx)
	// just 1 bonded token burned again since that's all the validator now has
	diffTokens = oldBondedPool.GetCoins().Sub(newBondedPool.GetCoins()).AmountOf(keeper.BondDenom(ctx))
	require.Equal(t, sdk.TokensFromConsensusPower(10), diffTokens)
	// apply TM updates
	keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	// read updated validator
	// power decreased by 1 again, validator is out of stake
	// validator should be in unbonding period
	validator, _ = keeper.GetValidatorByConsAddr(ctx, consAddr)
	require.Equal(t, validator.GetStatus(), sdk.Unbonding)
}

// tests Slash at a previous height with a redelegation
func TestSlashWithRedelegation(t *testing.T) {
	ctx, keeper, _ := setupHelper(t, 10)
	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := sdk.NewDecWithPrec(5, 1)
	bondDenom := keeper.BondDenom(ctx)

	// set a redelegation
	rdTokens := sdk.TokensFromConsensusPower(6)
	rd := types.NewRedelegation(addrDels[0], addrVals[0], addrVals[1], 11,
		time.Unix(0, 0), rdTokens, rdTokens.ToDec())
	keeper.SetRedelegation(ctx, rd)

	// set the associated delegation
	del := types.NewDelegation(addrDels[0], addrVals[1], rdTokens.ToDec())
	keeper.SetDelegation(ctx, del)

	// update bonded tokens
	bondedPool := keeper.GetBondedPool(ctx)
	notBondedPool := keeper.GetNotBondedPool(ctx)
	rdCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, rdTokens.MulRaw(2)))
	err := bondedPool.SetCoins(bondedPool.GetCoins().Add(rdCoins))
	require.NoError(t, err)
	keeper.supplyKeeper.SetModuleAccount(ctx, bondedPool)

	oldBonded := bondedPool.GetCoins().AmountOf(bondDenom)
	oldNotBonded := notBondedPool.GetCoins().AmountOf(bondDenom)

	// slash validator
	ctx = ctx.WithBlockHeight(12)
	validator, found := keeper.GetValidatorByConsAddr(ctx, consAddr)
	require.True(t, found)

	require.NotPanics(t, func() { keeper.Slash(ctx, consAddr, 10, 10, fraction) })
	burnAmount := sdk.TokensFromConsensusPower(10).ToDec().Mul(fraction).TruncateInt()

	bondedPool = keeper.GetBondedPool(ctx)
	notBondedPool = keeper.GetNotBondedPool(ctx)
	// burn bonded tokens from only from delegations
	require.True(sdk.IntEq(t, oldBonded.Sub(burnAmount), bondedPool.GetCoins().AmountOf(bondDenom)))
	require.True(sdk.IntEq(t, oldNotBonded, notBondedPool.GetCoins().AmountOf(bondDenom)))
	oldBonded = bondedPool.GetCoins().AmountOf(bondDenom)

	// read updating redelegation
	rd, found = keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.True(t, found)
	require.Len(t, rd.Entries, 1)
	// read updated validator
	validator, found = keeper.GetValidatorByConsAddr(ctx, consAddr)
	require.True(t, found)
	// power decreased by 2 - 4 stake originally bonded at the time of infraction
	// was still bonded at the time of discovery and was slashed by half, 4 stake
	// bonded at the time of discovery hadn't been bonded at the time of infraction
	// and wasn't slashed
	require.Equal(t, int64(8), validator.GetConsensusPower())

	// slash the validator again
	validator, found = keeper.GetValidatorByConsAddr(ctx, consAddr)
	require.True(t, found)

	require.NotPanics(t, func() { keeper.Slash(ctx, consAddr, 10, 10, sdk.OneDec()) })
	burnAmount = sdk.TokensFromConsensusPower(7)

	// read updated pool
	bondedPool = keeper.GetBondedPool(ctx)
	notBondedPool = keeper.GetNotBondedPool(ctx)
	// seven bonded tokens burned
	require.True(sdk.IntEq(t, oldBonded.Sub(burnAmount), bondedPool.GetCoins().AmountOf(bondDenom)))
	require.True(sdk.IntEq(t, oldNotBonded, notBondedPool.GetCoins().AmountOf(bondDenom)))
	oldBonded = bondedPool.GetCoins().AmountOf(bondDenom)

	// read updating redelegation
	rd, found = keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.True(t, found)
	require.Len(t, rd.Entries, 1)
	// read updated validator
	validator, found = keeper.GetValidatorByConsAddr(ctx, consAddr)
	require.True(t, found)
	// power decreased by 4
	require.Equal(t, int64(4), validator.GetConsensusPower())

	// slash the validator again, by 100%
	ctx = ctx.WithBlockHeight(12)
	validator, found = keeper.GetValidatorByConsAddr(ctx, consAddr)
	require.True(t, found)

	require.NotPanics(t, func() { keeper.Slash(ctx, consAddr, 10, 10, sdk.OneDec()) })

	burnAmount = sdk.TokensFromConsensusPower(10).ToDec().Mul(sdk.OneDec()).TruncateInt()
	burnAmount = burnAmount.Sub(sdk.OneDec().MulInt(rdTokens).TruncateInt())

	// read updated pool
	bondedPool = keeper.GetBondedPool(ctx)
	notBondedPool = keeper.GetNotBondedPool(ctx)
	require.True(sdk.IntEq(t, oldBonded.Sub(burnAmount), bondedPool.GetCoins().AmountOf(bondDenom)))
	require.True(sdk.IntEq(t, oldNotBonded, notBondedPool.GetCoins().AmountOf(bondDenom)))
	oldBonded = bondedPool.GetCoins().AmountOf(bondDenom)

	// read updating redelegation
	rd, found = keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.True(t, found)
	require.Len(t, rd.Entries, 1)
	// apply TM updates
	keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	// read updated validator
	// validator decreased to zero power, should be in unbonding period
	validator, _ = keeper.GetValidatorByConsAddr(ctx, consAddr)
	require.Equal(t, validator.GetStatus(), sdk.Unbonding)

	// slash the validator again, by 100%
	// no stake remains to be slashed
	ctx = ctx.WithBlockHeight(12)
	// validator still in unbonding period
	validator, _ = keeper.GetValidatorByConsAddr(ctx, consAddr)
	require.Equal(t, validator.GetStatus(), sdk.Unbonding)

	require.NotPanics(t, func() { keeper.Slash(ctx, consAddr, 10, 10, sdk.OneDec()) })

	// read updated pool
	bondedPool = keeper.GetBondedPool(ctx)
	notBondedPool = keeper.GetNotBondedPool(ctx)
	require.True(sdk.IntEq(t, oldBonded, bondedPool.GetCoins().AmountOf(bondDenom)))
	require.True(sdk.IntEq(t, oldNotBonded, notBondedPool.GetCoins().AmountOf(bondDenom)))

	// read updating redelegation
	rd, found = keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.True(t, found)
	require.Len(t, rd.Entries, 1)
	// read updated validator
	// power still zero, still in unbonding period
	validator, _ = keeper.GetValidatorByConsAddr(ctx, consAddr)
	require.Equal(t, validator.GetStatus(), sdk.Unbonding)
}

// tests Slash at a previous height with both an unbonding delegation and a redelegation
func TestSlashBoth(t *testing.T) {
	ctx, keeper, _ := setupHelper(t, 10)
	fraction := sdk.NewDecWithPrec(5, 1)
	bondDenom := keeper.BondDenom(ctx)

	// set a redelegation with expiration timestamp beyond which the
	// redelegation shouldn't be slashed
	rdATokens := sdk.TokensFromConsensusPower(6)
	rdA := types.NewRedelegation(addrDels[0], addrVals[0], addrVals[1], 11,
		time.Unix(0, 0), rdATokens,
		rdATokens.ToDec())
	keeper.SetRedelegation(ctx, rdA)

	// set the associated delegation
	delA := types.NewDelegation(addrDels[0], addrVals[1], rdATokens.ToDec())
	keeper.SetDelegation(ctx, delA)

	// set an unbonding delegation with expiration timestamp (beyond which the
	// unbonding delegation shouldn't be slashed)
	ubdATokens := sdk.TokensFromConsensusPower(4)
	ubdA := types.NewUnbondingDelegation(addrDels[0], addrVals[0], 11,
		time.Unix(0, 0), ubdATokens)
	keeper.SetUnbondingDelegation(ctx, ubdA)

	bondedCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, rdATokens.MulRaw(2)))
	notBondedCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, ubdATokens))

	// update bonded tokens
	bondedPool := keeper.GetBondedPool(ctx)
	notBondedPool := keeper.GetNotBondedPool(ctx)
	require.NoError(t, bondedPool.SetCoins(bondedPool.GetCoins().Add(bondedCoins)))
	require.NoError(t, bondedPool.SetCoins(notBondedPool.GetCoins().Add(notBondedCoins)))
	keeper.supplyKeeper.SetModuleAccount(ctx, bondedPool)
	keeper.supplyKeeper.SetModuleAccount(ctx, notBondedPool)

	oldBonded := bondedPool.GetCoins().AmountOf(bondDenom)
	oldNotBonded := notBondedPool.GetCoins().AmountOf(bondDenom)

	// slash validator
	ctx = ctx.WithBlockHeight(12)
	validator, found := keeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(PKs[0]))
	require.True(t, found)
	consAddr0 := sdk.ConsAddress(PKs[0].Address())
	keeper.Slash(ctx, consAddr0, 10, 10, fraction)

	burnedNotBondedAmount := fraction.MulInt(ubdATokens).TruncateInt()
	burnedBondAmount := sdk.TokensFromConsensusPower(10).ToDec().Mul(fraction).TruncateInt()
	burnedBondAmount = burnedBondAmount.Sub(burnedNotBondedAmount)

	// read updated pool
	bondedPool = keeper.GetBondedPool(ctx)
	notBondedPool = keeper.GetNotBondedPool(ctx)
	require.True(sdk.IntEq(t, oldBonded.Sub(burnedBondAmount), bondedPool.GetCoins().AmountOf(bondDenom)))
	require.True(sdk.IntEq(t, oldNotBonded.Sub(burnedNotBondedAmount), notBondedPool.GetCoins().AmountOf(bondDenom)))

	// read updating redelegation
	rdA, found = keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.True(t, found)
	require.Len(t, rdA.Entries, 1)
	// read updated validator
	validator, found = keeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(PKs[0]))
	require.True(t, found)
	// power not decreased, all stake was bonded since
	require.Equal(t, int64(10), validator.GetConsensusPower())
}
