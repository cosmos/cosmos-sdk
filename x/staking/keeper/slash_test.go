package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// bootstrapSlashTest creates 3 validators and bootstrap the app.
func bootstrapSlashTest(t *testing.T, power int64) (*simapp.SimApp, sdk.Context, []sdk.AccAddress, []sdk.ValAddress) {
	_, app, ctx := createTestInput()

	addrDels, addrVals := generateAddresses(app, ctx, 100)

	amt := app.StakingKeeper.TokensFromConsensusPower(ctx, power)
	totalSupply := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), amt.MulRaw(int64(len(addrDels)))))

	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
	require.NoError(t, simapp.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), totalSupply))

	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	numVals := int64(3)
	bondedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), amt.MulRaw(numVals)))
	bondedPool := app.StakingKeeper.GetBondedPool(ctx)

	// set bonded pool balance
	app.AccountKeeper.SetModuleAccount(ctx, bondedPool)
	require.NoError(t, simapp.FundModuleAccount(app.BankKeeper, ctx, bondedPool.GetName(), bondedCoins))

	for i := int64(0); i < numVals; i++ {
		validator := teststaking.NewValidator(t, addrVals[i], PKs[i])
		validator, _ = validator.AddTokensFromDel(amt)
		validator = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validator, true)
		app.StakingKeeper.SetValidatorByConsAddr(ctx, validator)
	}

	return app, ctx, addrDels, addrVals
}

// tests Jail, Unjail
func TestRevocation(t *testing.T) {
	app, ctx, _, addrVals := bootstrapSlashTest(t, 5)

	consAddr := sdk.ConsAddress(PKs[0].Address())

	// initial state
	val, found := app.StakingKeeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.False(t, val.IsJailed())

	// test jail
	app.StakingKeeper.Jail(ctx, consAddr)
	val, found = app.StakingKeeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.True(t, val.IsJailed())

	// test unjail
	app.StakingKeeper.Unjail(ctx, consAddr)
	val, found = app.StakingKeeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.False(t, val.IsJailed())
}

// tests slashUnbondingDelegation
func TestSlashUnbondingDelegation(t *testing.T) {
	app, ctx, addrDels, addrVals := bootstrapSlashTest(t, 10)

	fraction := sdk.NewDecWithPrec(5, 1)

	// set an unbonding delegation with expiration timestamp (beyond which the
	// unbonding delegation shouldn't be slashed)
	ubd := types.NewUnbondingDelegation(addrDels[0], addrVals[0], 0,
		time.Unix(5, 0), sdk.NewInt(10))

	app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)

	// unbonding started prior to the infraction height, stakw didn't contribute
	slashAmount := app.StakingKeeper.SlashUnbondingDelegation(ctx, ubd, 1, fraction)
	require.True(t, slashAmount.Equal(sdk.NewInt(0)))

	// after the expiration time, no longer eligible for slashing
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: time.Unix(10, 0)})
	app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)
	slashAmount = app.StakingKeeper.SlashUnbondingDelegation(ctx, ubd, 0, fraction)
	require.True(t, slashAmount.Equal(sdk.NewInt(0)))

	// test valid slash, before expiration timestamp and to which stake contributed
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
	oldUnbondedPoolBalances := app.BankKeeper.GetAllBalances(ctx, notBondedPool.GetAddress())
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: time.Unix(0, 0)})
	app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)
	slashAmount = app.StakingKeeper.SlashUnbondingDelegation(ctx, ubd, 0, fraction)
	require.True(t, slashAmount.Equal(sdk.NewInt(5)))
	ubd, found := app.StakingKeeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)

	// initial balance unchanged
	require.Equal(t, sdk.NewInt(10), ubd.Entries[0].InitialBalance)

	// balance decreased
	require.Equal(t, sdk.NewInt(5), ubd.Entries[0].Balance)
	newUnbondedPoolBalances := app.BankKeeper.GetAllBalances(ctx, notBondedPool.GetAddress())
	diffTokens := oldUnbondedPoolBalances.Sub(newUnbondedPoolBalances)
	require.True(t, diffTokens.AmountOf(app.StakingKeeper.BondDenom(ctx)).Equal(sdk.NewInt(5)))
}

// tests Slash at a future height (must panic)
func TestSlashAtFutureHeight(t *testing.T) {
	app, ctx, _, _ := bootstrapSlashTest(t, 10)

	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := sdk.NewDecWithPrec(5, 1)
	require.Panics(t, func() { app.StakingKeeper.Slash(ctx, consAddr, 1, 10, fraction) })
}

// test slash at a negative height
// this just represents pre-genesis and should have the same effect as slashing at height 0
func TestSlashAtNegativeHeight(t *testing.T) {
	app, ctx, _, _ := bootstrapSlashTest(t, 10)
	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := sdk.NewDecWithPrec(5, 1)

	bondedPool := app.StakingKeeper.GetBondedPool(ctx)
	oldBondedPoolBalances := app.BankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())

	validator, found := app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	require.True(t, found)
	app.StakingKeeper.Slash(ctx, consAddr, -2, 10, fraction)

	// read updated state
	validator, found = app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	require.True(t, found)

	// end block
	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 1)

	validator, found = app.StakingKeeper.GetValidator(ctx, validator.GetOperator())
	require.True(t, found)
	// power decreased
	require.Equal(t, int64(5), validator.GetConsensusPower(app.StakingKeeper.PowerReduction(ctx)))

	// pool bonded shares decreased
	newBondedPoolBalances := app.BankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())
	diffTokens := oldBondedPoolBalances.Sub(newBondedPoolBalances).AmountOf(app.StakingKeeper.BondDenom(ctx))
	require.Equal(t, app.StakingKeeper.TokensFromConsensusPower(ctx, 5).String(), diffTokens.String())
}

// tests Slash at the current height
func TestSlashValidatorAtCurrentHeight(t *testing.T) {
	app, ctx, _, _ := bootstrapSlashTest(t, 10)
	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := sdk.NewDecWithPrec(5, 1)

	bondedPool := app.StakingKeeper.GetBondedPool(ctx)
	oldBondedPoolBalances := app.BankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())

	validator, found := app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	require.True(t, found)
	app.StakingKeeper.Slash(ctx, consAddr, ctx.BlockHeight(), 10, fraction)

	// read updated state
	validator, found = app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	require.True(t, found)

	// end block
	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 1)

	validator, found = app.StakingKeeper.GetValidator(ctx, validator.GetOperator())
	assert.True(t, found)
	// power decreased
	require.Equal(t, int64(5), validator.GetConsensusPower(app.StakingKeeper.PowerReduction(ctx)))

	// pool bonded shares decreased
	newBondedPoolBalances := app.BankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())
	diffTokens := oldBondedPoolBalances.Sub(newBondedPoolBalances).AmountOf(app.StakingKeeper.BondDenom(ctx))
	require.Equal(t, app.StakingKeeper.TokensFromConsensusPower(ctx, 5).String(), diffTokens.String())
}

// tests Slash at a previous height with an unbonding delegation
func TestSlashWithUnbondingDelegation(t *testing.T) {
	app, ctx, addrDels, addrVals := bootstrapSlashTest(t, 10)

	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := sdk.NewDecWithPrec(5, 1)

	// set an unbonding delegation with expiration timestamp beyond which the
	// unbonding delegation shouldn't be slashed
	ubdTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 4)
	ubd := types.NewUnbondingDelegation(addrDels[0], addrVals[0], 11, time.Unix(0, 0), ubdTokens)
	app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)

	// slash validator for the first time
	ctx = ctx.WithBlockHeight(12)
	bondedPool := app.StakingKeeper.GetBondedPool(ctx)
	oldBondedPoolBalances := app.BankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())

	validator, found := app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	require.True(t, found)
	app.StakingKeeper.Slash(ctx, consAddr, 10, 10, fraction)

	// end block
	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 1)

	// read updating unbonding delegation
	ubd, found = app.StakingKeeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)

	// balance decreased
	require.Equal(t, app.StakingKeeper.TokensFromConsensusPower(ctx, 2), ubd.Entries[0].Balance)

	// bonded tokens burned
	newBondedPoolBalances := app.BankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())
	diffTokens := oldBondedPoolBalances.Sub(newBondedPoolBalances).AmountOf(app.StakingKeeper.BondDenom(ctx))
	require.Equal(t, app.StakingKeeper.TokensFromConsensusPower(ctx, 3), diffTokens)

	// read updated validator
	validator, found = app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	require.True(t, found)

	// power decreased by 3 - 6 stake originally bonded at the time of infraction
	// was still bonded at the time of discovery and was slashed by half, 4 stake
	// bonded at the time of discovery hadn't been bonded at the time of infraction
	// and wasn't slashed
	require.Equal(t, int64(7), validator.GetConsensusPower(app.StakingKeeper.PowerReduction(ctx)))

	// slash validator again
	ctx = ctx.WithBlockHeight(13)
	app.StakingKeeper.Slash(ctx, consAddr, 9, 10, fraction)

	ubd, found = app.StakingKeeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)

	// balance decreased again
	require.Equal(t, sdk.NewInt(0), ubd.Entries[0].Balance)

	// bonded tokens burned again
	newBondedPoolBalances = app.BankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())
	diffTokens = oldBondedPoolBalances.Sub(newBondedPoolBalances).AmountOf(app.StakingKeeper.BondDenom(ctx))
	require.Equal(t, app.StakingKeeper.TokensFromConsensusPower(ctx, 6), diffTokens)

	// read updated validator
	validator, found = app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	require.True(t, found)

	// power decreased by 3 again
	require.Equal(t, int64(4), validator.GetConsensusPower(app.StakingKeeper.PowerReduction(ctx)))

	// slash validator again
	// all originally bonded stake has been slashed, so this will have no effect
	// on the unbonding delegation, but it will slash stake bonded since the infraction
	// this may not be the desirable behaviour, ref https://github.com/cosmos/cosmos-sdk/issues/1440
	ctx = ctx.WithBlockHeight(13)
	app.StakingKeeper.Slash(ctx, consAddr, 9, 10, fraction)

	ubd, found = app.StakingKeeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)

	// balance unchanged
	require.Equal(t, sdk.NewInt(0), ubd.Entries[0].Balance)

	// bonded tokens burned again
	newBondedPoolBalances = app.BankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())
	diffTokens = oldBondedPoolBalances.Sub(newBondedPoolBalances).AmountOf(app.StakingKeeper.BondDenom(ctx))
	require.Equal(t, app.StakingKeeper.TokensFromConsensusPower(ctx, 9), diffTokens)

	// read updated validator
	validator, found = app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	require.True(t, found)

	// power decreased by 3 again
	require.Equal(t, int64(1), validator.GetConsensusPower(app.StakingKeeper.PowerReduction(ctx)))

	// slash validator again
	// all originally bonded stake has been slashed, so this will have no effect
	// on the unbonding delegation, but it will slash stake bonded since the infraction
	// this may not be the desirable behaviour, ref https://github.com/cosmos/cosmos-sdk/issues/1440
	ctx = ctx.WithBlockHeight(13)
	app.StakingKeeper.Slash(ctx, consAddr, 9, 10, fraction)

	ubd, found = app.StakingKeeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)

	// balance unchanged
	require.Equal(t, sdk.NewInt(0), ubd.Entries[0].Balance)

	// just 1 bonded token burned again since that's all the validator now has
	newBondedPoolBalances = app.BankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())
	diffTokens = oldBondedPoolBalances.Sub(newBondedPoolBalances).AmountOf(app.StakingKeeper.BondDenom(ctx))
	require.Equal(t, app.StakingKeeper.TokensFromConsensusPower(ctx, 10), diffTokens)

	// apply TM updates
	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, -1)

	// read updated validator
	// power decreased by 1 again, validator is out of stake
	// validator should be in unbonding period
	validator, _ = app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	require.Equal(t, validator.GetStatus(), types.Unbonding)
}
