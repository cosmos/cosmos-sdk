package staking

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/header"
	"cosmossdk.io/math"
	_ "cosmossdk.io/x/accounts"
	banktestutil "cosmossdk.io/x/bank/testutil"
	_ "cosmossdk.io/x/consensus"
	_ "cosmossdk.io/x/distribution"
	_ "cosmossdk.io/x/protocolpool"
	_ "cosmossdk.io/x/slashing"
	"cosmossdk.io/x/staking/keeper"
	"cosmossdk.io/x/staking/testutil"
	"cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
)

// bootstrapSlashTest creates 3 validators and bootstrap the app.
func bootstrapSlashTest(t *testing.T, power int64) (*fixture, []sdk.AccAddress, []sdk.ValAddress) {
	t.Helper()
	t.Parallel()
	f := initFixture(t, false)

	addrDels, addrVals := generateAddresses(f, 100)

	amt := f.stakingKeeper.TokensFromConsensusPower(f.ctx, power)
	bondDenom, err := f.stakingKeeper.BondDenom(f.ctx)
	require.NoError(t, err)
	totalSupply := sdk.NewCoins(sdk.NewCoin(bondDenom, amt.MulRaw(int64(len(addrDels)))))

	notBondedPool := f.stakingKeeper.GetNotBondedPool(f.ctx)
	assert.NilError(t, banktestutil.FundModuleAccount(f.ctx, f.bankKeeper, notBondedPool.GetName(), totalSupply))

	f.accountKeeper.SetModuleAccount(f.ctx, notBondedPool)

	numVals := int64(3)
	bondedCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, amt.MulRaw(numVals)))
	bondedPool := f.stakingKeeper.GetBondedPool(f.ctx)

	// set bonded pool balance
	f.accountKeeper.SetModuleAccount(f.ctx, bondedPool)
	assert.NilError(t, banktestutil.FundModuleAccount(f.ctx, f.bankKeeper, bondedPool.GetName(), bondedCoins))

	for i := int64(0); i < numVals; i++ {
		validator := testutil.NewValidator(t, addrVals[i], PKs[i])
		validator, _ = validator.AddTokensFromDel(amt)
		validator, _ = keeper.TestingUpdateValidatorV2(f.stakingKeeper, f.ctx, validator, true)
		assert.NilError(t, f.stakingKeeper.SetValidatorByConsAddr(f.ctx, validator))
	}

	return f, addrDels, addrVals
}

// tests slashUnbondingDelegation
func TestSlashUnbondingDelegation(t *testing.T) {
	f, addrDels, addrVals := bootstrapSlashTest(t, 10)

	fraction := math.LegacyNewDecWithPrec(5, 1)

	// set an unbonding delegation with expiration timestamp (beyond which the
	// unbonding delegation shouldn't be slashed)
	ubd := types.NewUnbondingDelegation(addrDels[0], addrVals[0], 0,
		time.Unix(5, 0), math.NewInt(10), address.NewBech32Codec("cosmosvaloper"), address.NewBech32Codec("cosmos"))

	assert.NilError(t, f.stakingKeeper.SetUnbondingDelegation(f.ctx, ubd))

	// unbonding started prior to the infraction height, stakw didn't contribute
	slashAmount, err := f.stakingKeeper.SlashUnbondingDelegation(f.ctx, ubd, 1, fraction)
	assert.NilError(t, err)
	assert.Assert(t, slashAmount.Equal(math.NewInt(0)))

	// after the expiration time, no longer eligible for slashing
	f.ctx = integration.SetHeaderInfo(f.ctx, header.Info{Time: time.Unix(10, 0)})
	assert.NilError(t, f.stakingKeeper.SetUnbondingDelegation(f.ctx, ubd))
	slashAmount, err = f.stakingKeeper.SlashUnbondingDelegation(f.ctx, ubd, 0, fraction)
	assert.NilError(t, err)
	assert.Assert(t, slashAmount.Equal(math.NewInt(0)))

	// test valid slash, before expiration timestamp and to which stake contributed
	notBondedPool := f.stakingKeeper.GetNotBondedPool(f.ctx)
	oldUnbondedPoolBalances := f.bankKeeper.GetAllBalances(f.ctx, notBondedPool.GetAddress())
	f.ctx = integration.SetHeaderInfo(f.ctx, header.Info{Time: time.Unix(0, 0)})
	assert.NilError(t, f.stakingKeeper.SetUnbondingDelegation(f.ctx, ubd))
	slashAmount, err = f.stakingKeeper.SlashUnbondingDelegation(f.ctx, ubd, 0, fraction)
	assert.NilError(t, err)
	assert.Assert(t, slashAmount.Equal(math.NewInt(5)))
	ubd, found := f.stakingKeeper.GetUnbondingDelegation(f.ctx, addrDels[0], addrVals[0])
	assert.Assert(t, found)
	assert.Assert(t, len(ubd.Entries) == 1)

	// initial balance unchanged
	assert.DeepEqual(t, math.NewInt(10), ubd.Entries[0].InitialBalance)

	// balance decreased
	assert.DeepEqual(t, math.NewInt(5), ubd.Entries[0].Balance)
	newUnbondedPoolBalances := f.bankKeeper.GetAllBalances(f.ctx, notBondedPool.GetAddress())
	diffTokens := oldUnbondedPoolBalances.Sub(newUnbondedPoolBalances...)
	bondDenom, err := f.stakingKeeper.BondDenom(f.ctx)
	assert.NilError(t, err)
	assert.Assert(t, diffTokens.AmountOf(bondDenom).Equal(math.NewInt(5)))
}

// tests slashRedelegation
func TestSlashRedelegation(t *testing.T) {
	f, addrDels, addrVals := bootstrapSlashTest(t, 10)
	fraction := math.LegacyNewDecWithPrec(5, 1)

	bondDenom, err := f.stakingKeeper.BondDenom(f.ctx)
	assert.NilError(t, err)

	// add bonded tokens to pool for (re)delegations
	startCoins := sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 15))
	bondedPool := f.stakingKeeper.GetBondedPool(f.ctx)
	_ = f.bankKeeper.GetAllBalances(f.ctx, bondedPool.GetAddress())

	assert.NilError(t, banktestutil.FundModuleAccount(f.ctx, f.bankKeeper, bondedPool.GetName(), startCoins))
	f.accountKeeper.SetModuleAccount(f.ctx, bondedPool)

	// set a redelegation with an expiration timestamp beyond which the
	// redelegation shouldn't be slashed
	rd := types.NewRedelegation(addrDels[0], addrVals[0], addrVals[1], 0,
		time.Unix(5, 0), math.NewInt(10), math.LegacyNewDec(10), address.NewBech32Codec("cosmosvaloper"), address.NewBech32Codec("cosmos"))

	assert.NilError(t, f.stakingKeeper.SetRedelegation(f.ctx, rd))

	// set the associated delegation
	del := types.NewDelegation(addrDels[0].String(), addrVals[1].String(), math.LegacyNewDec(10))
	assert.NilError(t, f.stakingKeeper.SetDelegation(f.ctx, del))

	// started redelegating prior to the current height, stake didn't contribute to infraction
	validator, found := f.stakingKeeper.GetValidator(f.ctx, addrVals[1])
	assert.Assert(t, found)
	slashAmount, err := f.stakingKeeper.SlashRedelegation(f.ctx, validator, rd, 1, fraction)
	assert.NilError(t, err)
	assert.Assert(t, slashAmount.Equal(math.NewInt(0)))

	// after the expiration time, no longer eligible for slashing
	f.ctx = integration.SetHeaderInfo(f.ctx, header.Info{Time: time.Unix(10, 0)})
	assert.NilError(t, f.stakingKeeper.SetRedelegation(f.ctx, rd))
	validator, found = f.stakingKeeper.GetValidator(f.ctx, addrVals[1])
	assert.Assert(t, found)
	slashAmount, err = f.stakingKeeper.SlashRedelegation(f.ctx, validator, rd, 0, fraction)
	assert.NilError(t, err)
	assert.Assert(t, slashAmount.Equal(math.NewInt(0)))

	balances := f.bankKeeper.GetAllBalances(f.ctx, bondedPool.GetAddress())

	// test valid slash, before expiration timestamp and to which stake contributed
	f.ctx = integration.SetHeaderInfo(f.ctx, header.Info{Time: time.Unix(0, 0)})
	assert.NilError(t, f.stakingKeeper.SetRedelegation(f.ctx, rd))
	validator, found = f.stakingKeeper.GetValidator(f.ctx, addrVals[1])
	assert.Assert(t, found)
	slashAmount, err = f.stakingKeeper.SlashRedelegation(f.ctx, validator, rd, 0, fraction)
	assert.NilError(t, err)
	assert.Assert(t, slashAmount.Equal(math.NewInt(5)))
	rd, found = f.stakingKeeper.Redelegations.Get(f.ctx, collections.Join3(addrDels[0].Bytes(), addrVals[0].Bytes(), addrVals[1].Bytes()))
	assert.Assert(t, found)
	assert.Assert(t, len(rd.Entries) == 1)

	// end block
	applyValidatorSetUpdates(t, f.ctx, f.stakingKeeper, 1)

	// initialbalance unchanged
	assert.DeepEqual(t, math.NewInt(10), rd.Entries[0].InitialBalance)

	// shares decreased
	del, found = f.stakingKeeper.Delegations.Get(f.ctx, collections.Join(addrDels[0], addrVals[1]))
	assert.Assert(t, found)
	assert.Equal(t, int64(5), del.Shares.RoundInt64())

	// pool bonded tokens should decrease
	burnedCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, slashAmount))
	assert.DeepEqual(t, balances.Sub(burnedCoins...), f.bankKeeper.GetAllBalances(f.ctx, bondedPool.GetAddress()))
}

// test slash at a negative height
// this just represents pre-genesis and should have the same effect as slashing at height 0
func TestSlashAtNegativeHeight(t *testing.T) {
	f, _, _ := bootstrapSlashTest(t, 10)
	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := math.LegacyNewDecWithPrec(5, 1)

	bondedPool := f.stakingKeeper.GetBondedPool(f.ctx)
	oldBondedPoolBalances := f.bankKeeper.GetAllBalances(f.ctx, bondedPool.GetAddress())

	_, found := f.stakingKeeper.GetValidatorByConsAddr(f.ctx, consAddr)
	assert.Assert(t, found)
	_, err := f.stakingKeeper.Slash(f.ctx, consAddr, -2, 10, fraction)
	assert.NilError(t, err)

	// read updated state
	validator, found := f.stakingKeeper.GetValidatorByConsAddr(f.ctx, consAddr)
	assert.Assert(t, found)

	// end block
	applyValidatorSetUpdates(t, f.ctx, f.stakingKeeper, 1)

	valbz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
	assert.NilError(t, err)

	validator, found = f.stakingKeeper.GetValidator(f.ctx, valbz)
	assert.Assert(t, found)
	// power decreased
	assert.Equal(t, int64(5), validator.GetConsensusPower(f.stakingKeeper.PowerReduction(f.ctx)))

	bondDenom, err := f.stakingKeeper.BondDenom(f.ctx)
	assert.NilError(t, err)

	// pool bonded shares decreased
	newBondedPoolBalances := f.bankKeeper.GetAllBalances(f.ctx, bondedPool.GetAddress())
	diffTokens := oldBondedPoolBalances.Sub(newBondedPoolBalances...).AmountOf(bondDenom)
	assert.DeepEqual(t, f.stakingKeeper.TokensFromConsensusPower(f.ctx, 5).String(), diffTokens.String())
}

// tests Slash at the current height
func TestSlashValidatorAtCurrentHeight(t *testing.T) {
	f, _, _ := bootstrapSlashTest(t, 10)
	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := math.LegacyNewDecWithPrec(5, 1)

	bondDenom, err := f.stakingKeeper.BondDenom(f.ctx)
	assert.NilError(t, err)

	bondedPool := f.stakingKeeper.GetBondedPool(f.ctx)
	oldBondedPoolBalances := f.bankKeeper.GetAllBalances(f.ctx, bondedPool.GetAddress())

	_, found := f.stakingKeeper.GetValidatorByConsAddr(f.ctx, consAddr)
	assert.Assert(t, found)
	_, err = f.stakingKeeper.Slash(f.ctx, consAddr, int64(f.app.LastBlockHeight()), 10, fraction)
	assert.NilError(t, err)

	// read updated state
	validator, found := f.stakingKeeper.GetValidatorByConsAddr(f.ctx, consAddr)
	assert.Assert(t, found)

	// end block
	applyValidatorSetUpdates(t, f.ctx, f.stakingKeeper, 1)

	valbz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
	assert.NilError(t, err)

	validator, found = f.stakingKeeper.GetValidator(f.ctx, valbz)
	assert.Assert(t, found)
	// power decreased
	assert.Equal(t, int64(5), validator.GetConsensusPower(f.stakingKeeper.PowerReduction(f.ctx)))

	// pool bonded shares decreased
	newBondedPoolBalances := f.bankKeeper.GetAllBalances(f.ctx, bondedPool.GetAddress())
	diffTokens := oldBondedPoolBalances.Sub(newBondedPoolBalances...).AmountOf(bondDenom)
	assert.DeepEqual(t, f.stakingKeeper.TokensFromConsensusPower(f.ctx, 5).String(), diffTokens.String())
}

// TestSlashWithUnbondingDelegation tests the slashing of a validator with an unbonding delegation.
// It sets up an environment with a validator and an unbonding delegation, and then performs slashing
// operations on the validator. The test verifies that the slashing correctly affects the unbonding
// delegation and the validator's power.
func TestSlashWithUnbondingDelegation(t *testing.T) {
	f, addrDels, addrVals := bootstrapSlashTest(t, 10)

	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := math.LegacyNewDecWithPrec(5, 1)

	bondDenom, err := f.stakingKeeper.BondDenom(f.ctx)
	assert.NilError(t, err)

	// set an unbonding delegation with expiration timestamp beyond which the
	// unbonding delegation shouldn't be slashed
	ubdTokens := f.stakingKeeper.TokensFromConsensusPower(f.ctx, 4)
	ubd := types.NewUnbondingDelegation(addrDels[0], addrVals[0], 11, time.Unix(0, 0), ubdTokens, address.NewBech32Codec("cosmosvaloper"), address.NewBech32Codec("cosmos"))
	assert.NilError(t, f.stakingKeeper.SetUnbondingDelegation(f.ctx, ubd))

	// slash validator for the first time
	f.ctx = integration.SetHeaderInfo(f.ctx, header.Info{Height: 12})
	bondedPool := f.stakingKeeper.GetBondedPool(f.ctx)
	oldBondedPoolBalances := f.bankKeeper.GetAllBalances(f.ctx, bondedPool.GetAddress())

	_, found := f.stakingKeeper.GetValidatorByConsAddr(f.ctx, consAddr)
	assert.Assert(t, found)
	_, err = f.stakingKeeper.Slash(f.ctx, consAddr, 10, 10, fraction)
	assert.NilError(t, err)

	// end block
	applyValidatorSetUpdates(t, f.ctx, f.stakingKeeper, 1)

	// read updating unbonding delegation
	ubd, found = f.stakingKeeper.GetUnbondingDelegation(f.ctx, addrDels[0], addrVals[0])
	assert.Assert(t, found)
	assert.Assert(t, len(ubd.Entries) == 1)

	// balance decreased
	assert.DeepEqual(t, f.stakingKeeper.TokensFromConsensusPower(f.ctx, 2), ubd.Entries[0].Balance)

	// bonded tokens burned
	newBondedPoolBalances := f.bankKeeper.GetAllBalances(f.ctx, bondedPool.GetAddress())
	diffTokens := oldBondedPoolBalances.Sub(newBondedPoolBalances...).AmountOf(bondDenom)
	assert.DeepEqual(t, f.stakingKeeper.TokensFromConsensusPower(f.ctx, 3), diffTokens)

	// read updated validator
	validator, found := f.stakingKeeper.GetValidatorByConsAddr(f.ctx, consAddr)
	assert.Assert(t, found)

	// power decreased by 3 - 6 stake originally bonded at the time of infraction
	// was still bonded at the time of discovery and was slashed by half, 4 stake
	// bonded at the time of discovery hadn't been bonded at the time of infraction
	// and wasn't slashed
	assert.Equal(t, int64(7), validator.GetConsensusPower(f.stakingKeeper.PowerReduction(f.ctx)))

	// slash validator again
	f.ctx = integration.SetHeaderInfo(f.ctx, header.Info{Height: 13})
	_, err = f.stakingKeeper.Slash(f.ctx, consAddr, 9, 10, fraction)
	assert.NilError(t, err)

	ubd, found = f.stakingKeeper.GetUnbondingDelegation(f.ctx, addrDels[0], addrVals[0])
	assert.Assert(t, found)
	assert.Assert(t, len(ubd.Entries) == 1)

	// balance decreased again
	assert.DeepEqual(t, math.NewInt(0), ubd.Entries[0].Balance)

	// bonded tokens burned again
	newBondedPoolBalances = f.bankKeeper.GetAllBalances(f.ctx, bondedPool.GetAddress())
	diffTokens = oldBondedPoolBalances.Sub(newBondedPoolBalances...).AmountOf(bondDenom)
	assert.DeepEqual(t, f.stakingKeeper.TokensFromConsensusPower(f.ctx, 6), diffTokens)

	// read updated validator
	validator, found = f.stakingKeeper.GetValidatorByConsAddr(f.ctx, consAddr)
	assert.Assert(t, found)

	// power decreased by 3 again
	assert.Equal(t, int64(4), validator.GetConsensusPower(f.stakingKeeper.PowerReduction(f.ctx)))

	// slash validator again
	// all originally bonded stake has been slashed, so this will have no effect
	// on the unbonding delegation, but it will slash stake bonded since the infraction
	// this may not be the desirable behavior, ref https://github.com/cosmos/cosmos-sdk/issues/1440
	f.ctx = integration.SetHeaderInfo(f.ctx, header.Info{Height: 13})
	_, err = f.stakingKeeper.Slash(f.ctx, consAddr, 9, 10, fraction)
	assert.NilError(t, err)

	ubd, found = f.stakingKeeper.GetUnbondingDelegation(f.ctx, addrDels[0], addrVals[0])
	assert.Assert(t, found)
	assert.Assert(t, len(ubd.Entries) == 1)

	// balance unchanged
	assert.DeepEqual(t, math.NewInt(0), ubd.Entries[0].Balance)

	// bonded tokens burned again
	newBondedPoolBalances = f.bankKeeper.GetAllBalances(f.ctx, bondedPool.GetAddress())
	diffTokens = oldBondedPoolBalances.Sub(newBondedPoolBalances...).AmountOf(bondDenom)
	assert.DeepEqual(t, f.stakingKeeper.TokensFromConsensusPower(f.ctx, 9), diffTokens)

	// read updated validator
	validator, found = f.stakingKeeper.GetValidatorByConsAddr(f.ctx, consAddr)
	assert.Assert(t, found)

	// power decreased by 3 again
	assert.Equal(t, int64(1), validator.GetConsensusPower(f.stakingKeeper.PowerReduction(f.ctx)))

	// slash validator again
	// all originally bonded stake has been slashed, so this will have no effect
	// on the unbonding delegation, but it will slash stake bonded since the infraction
	// this may not be the desirable behavior, ref https://github.com/cosmos/cosmos-sdk/issues/1440
	f.ctx = integration.SetHeaderInfo(f.ctx, header.Info{Height: 13})
	_, err = f.stakingKeeper.Slash(f.ctx, consAddr, 9, 10, fraction)
	assert.NilError(t, err)

	ubd, found = f.stakingKeeper.GetUnbondingDelegation(f.ctx, addrDels[0], addrVals[0])
	assert.Assert(t, found)
	assert.Assert(t, len(ubd.Entries) == 1)

	// balance unchanged
	assert.DeepEqual(t, math.NewInt(0), ubd.Entries[0].Balance)

	// just 1 bonded token burned again since that's all the validator now has
	newBondedPoolBalances = f.bankKeeper.GetAllBalances(f.ctx, bondedPool.GetAddress())
	diffTokens = oldBondedPoolBalances.Sub(newBondedPoolBalances...).AmountOf(bondDenom)
	assert.DeepEqual(t, f.stakingKeeper.TokensFromConsensusPower(f.ctx, 10), diffTokens)

	// apply TM updates
	applyValidatorSetUpdates(t, f.ctx, f.stakingKeeper, -1)

	// read updated validator
	// power decreased by 1 again, validator is out of stake
	// validator should be in unbonding period
	validator, _ = f.stakingKeeper.GetValidatorByConsAddr(f.ctx, consAddr)
	assert.Equal(t, validator.GetStatus(), sdk.Unbonding)
}

// tests Slash at a previous height with a redelegation
func TestSlashWithRedelegation(t *testing.T) {
	f, addrDels, addrVals := bootstrapSlashTest(t, 10)
	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := math.LegacyNewDecWithPrec(5, 1)
	bondDenom, err := f.stakingKeeper.BondDenom(f.ctx)
	assert.NilError(t, err)

	// set a redelegation
	rdTokens := f.stakingKeeper.TokensFromConsensusPower(f.ctx, 6)
	rd := types.NewRedelegation(addrDels[0], addrVals[0], addrVals[1], 11, time.Unix(0, 0), rdTokens, math.LegacyNewDecFromInt(rdTokens), address.NewBech32Codec("cosmosvaloper"), address.NewBech32Codec("cosmos"))
	assert.NilError(t, f.stakingKeeper.SetRedelegation(f.ctx, rd))

	// set the associated delegation
	del := types.NewDelegation(addrDels[0].String(), addrVals[1].String(), math.LegacyNewDecFromInt(rdTokens))
	assert.NilError(t, f.stakingKeeper.SetDelegation(f.ctx, del))

	// update bonded tokens
	bondedPool := f.stakingKeeper.GetBondedPool(f.ctx)
	notBondedPool := f.stakingKeeper.GetNotBondedPool(f.ctx)
	rdCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, rdTokens.MulRaw(2)))

	assert.NilError(t, banktestutil.FundModuleAccount(f.ctx, f.bankKeeper, bondedPool.GetName(), rdCoins))

	f.accountKeeper.SetModuleAccount(f.ctx, bondedPool)

	oldBonded := f.bankKeeper.GetBalance(f.ctx, bondedPool.GetAddress(), bondDenom).Amount
	oldNotBonded := f.bankKeeper.GetBalance(f.ctx, notBondedPool.GetAddress(), bondDenom).Amount

	// slash validator
	f.ctx = integration.SetHeaderInfo(f.ctx, header.Info{Height: 12})
	_, found := f.stakingKeeper.GetValidatorByConsAddr(f.ctx, consAddr)
	assert.Assert(t, found)

	_, err = f.stakingKeeper.Slash(f.ctx, consAddr, 10, 10, fraction)
	assert.NilError(t, err)

	burnAmount := math.LegacyNewDecFromInt(f.stakingKeeper.TokensFromConsensusPower(f.ctx, 10)).Mul(fraction).TruncateInt()

	bondedPool = f.stakingKeeper.GetBondedPool(f.ctx)
	notBondedPool = f.stakingKeeper.GetNotBondedPool(f.ctx)

	// burn bonded tokens from only from delegations
	bondedPoolBalance := f.bankKeeper.GetBalance(f.ctx, bondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldBonded.Sub(burnAmount), bondedPoolBalance))

	notBondedPoolBalance := f.bankKeeper.GetBalance(f.ctx, notBondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldNotBonded, notBondedPoolBalance))
	oldBonded = f.bankKeeper.GetBalance(f.ctx, bondedPool.GetAddress(), bondDenom).Amount

	// read updating redelegation
	rd, found = f.stakingKeeper.Redelegations.Get(f.ctx, collections.Join3(addrDels[0].Bytes(), addrVals[0].Bytes(), addrVals[1].Bytes()))
	assert.Assert(t, found)
	assert.Assert(t, len(rd.Entries) == 1)
	// read updated validator
	validator, found := f.stakingKeeper.GetValidatorByConsAddr(f.ctx, consAddr)
	assert.Assert(t, found)
	// power decreased by 2 - 4 stake originally bonded at the time of infraction
	// was still bonded at the time of discovery and was slashed by half, 4 stake
	// bonded at the time of discovery hadn't been bonded at the time of infraction
	// and wasn't slashed
	assert.Equal(t, int64(8), validator.GetConsensusPower(f.stakingKeeper.PowerReduction(f.ctx)))

	// slash the validator again
	_, found = f.stakingKeeper.GetValidatorByConsAddr(f.ctx, consAddr)
	assert.Assert(t, found)

	_, err = f.stakingKeeper.Slash(f.ctx, consAddr, 10, 10, math.LegacyOneDec())
	assert.NilError(t, err)
	burnAmount = f.stakingKeeper.TokensFromConsensusPower(f.ctx, 7)

	// read updated pool
	bondedPool = f.stakingKeeper.GetBondedPool(f.ctx)
	notBondedPool = f.stakingKeeper.GetNotBondedPool(f.ctx)

	// seven bonded tokens burned
	bondedPoolBalance = f.bankKeeper.GetBalance(f.ctx, bondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldBonded.Sub(burnAmount), bondedPoolBalance))
	assert.Assert(math.IntEq(t, oldNotBonded, notBondedPoolBalance))

	bondedPoolBalance = f.bankKeeper.GetBalance(f.ctx, bondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldBonded.Sub(burnAmount), bondedPoolBalance))

	notBondedPoolBalance = f.bankKeeper.GetBalance(f.ctx, notBondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldNotBonded, notBondedPoolBalance))
	oldBonded = f.bankKeeper.GetBalance(f.ctx, bondedPool.GetAddress(), bondDenom).Amount

	// read updating redelegation
	rd, found = f.stakingKeeper.Redelegations.Get(f.ctx, collections.Join3(addrDels[0].Bytes(), addrVals[0].Bytes(), addrVals[1].Bytes()))
	assert.Assert(t, found)
	assert.Assert(t, len(rd.Entries) == 1)
	// read updated validator
	validator, found = f.stakingKeeper.GetValidatorByConsAddr(f.ctx, consAddr)
	assert.Assert(t, found)
	// power decreased by 4
	assert.Equal(t, int64(4), validator.GetConsensusPower(f.stakingKeeper.PowerReduction(f.ctx)))

	// slash the validator again, by 100%
	f.ctx = integration.SetHeaderInfo(f.ctx, header.Info{Height: 12})
	_, found = f.stakingKeeper.GetValidatorByConsAddr(f.ctx, consAddr)
	assert.Assert(t, found)

	_, err = f.stakingKeeper.Slash(f.ctx, consAddr, 10, 10, math.LegacyOneDec())
	assert.NilError(t, err)

	burnAmount = math.LegacyNewDecFromInt(f.stakingKeeper.TokensFromConsensusPower(f.ctx, 10)).Mul(math.LegacyOneDec()).TruncateInt()
	burnAmount = burnAmount.Sub(math.LegacyOneDec().MulInt(rdTokens).TruncateInt())

	// read updated pool
	bondedPool = f.stakingKeeper.GetBondedPool(f.ctx)
	notBondedPool = f.stakingKeeper.GetNotBondedPool(f.ctx)

	bondedPoolBalance = f.bankKeeper.GetBalance(f.ctx, bondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldBonded.Sub(burnAmount), bondedPoolBalance))
	notBondedPoolBalance = f.bankKeeper.GetBalance(f.ctx, notBondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldNotBonded, notBondedPoolBalance))
	oldBonded = f.bankKeeper.GetBalance(f.ctx, bondedPool.GetAddress(), bondDenom).Amount

	// read updating redelegation
	rd, found = f.stakingKeeper.Redelegations.Get(f.ctx, collections.Join3(addrDels[0].Bytes(), addrVals[0].Bytes(), addrVals[1].Bytes()))
	assert.Assert(t, found)
	assert.Assert(t, len(rd.Entries) == 1)
	// apply TM updates
	applyValidatorSetUpdates(t, f.ctx, f.stakingKeeper, -1)
	// read updated validator
	// validator decreased to zero power, should be in unbonding period
	validator, _ = f.stakingKeeper.GetValidatorByConsAddr(f.ctx, consAddr)
	assert.Equal(t, validator.GetStatus(), sdk.Unbonding)

	// slash the validator again, by 100%
	// no stake remains to be slashed
	f.ctx = integration.SetHeaderInfo(f.ctx, header.Info{Height: 12})
	// validator still in unbonding period
	validator, _ = f.stakingKeeper.GetValidatorByConsAddr(f.ctx, consAddr)
	assert.Equal(t, validator.GetStatus(), sdk.Unbonding)

	_, err = f.stakingKeeper.Slash(f.ctx, consAddr, 10, 10, math.LegacyOneDec())
	assert.NilError(t, err)

	// read updated pool
	bondedPool = f.stakingKeeper.GetBondedPool(f.ctx)
	notBondedPool = f.stakingKeeper.GetNotBondedPool(f.ctx)

	bondedPoolBalance = f.bankKeeper.GetBalance(f.ctx, bondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldBonded, bondedPoolBalance))
	notBondedPoolBalance = f.bankKeeper.GetBalance(f.ctx, notBondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldNotBonded, notBondedPoolBalance))

	// read updating redelegation
	rd, found = f.stakingKeeper.Redelegations.Get(f.ctx, collections.Join3(addrDels[0].Bytes(), addrVals[0].Bytes(), addrVals[1].Bytes()))
	assert.Assert(t, found)
	assert.Assert(t, len(rd.Entries) == 1)
	// read updated validator
	// power still zero, still in unbonding period
	validator, _ = f.stakingKeeper.GetValidatorByConsAddr(f.ctx, consAddr)
	assert.Equal(t, validator.GetStatus(), sdk.Unbonding)
}

// tests Slash at a previous height with both an unbonding delegation and a redelegation
func TestSlashBoth(t *testing.T) {
	f, addrDels, addrVals := bootstrapSlashTest(t, 10)
	fraction := math.LegacyNewDecWithPrec(5, 1)
	bondDenom, err := f.stakingKeeper.BondDenom(f.ctx)
	assert.NilError(t, err)

	// set a redelegation with expiration timestamp beyond which the
	// redelegation shouldn't be slashed
	rdATokens := f.stakingKeeper.TokensFromConsensusPower(f.ctx, 6)
	rdA := types.NewRedelegation(addrDels[0], addrVals[0], addrVals[1], 11, time.Unix(0, 0), rdATokens, math.LegacyNewDecFromInt(rdATokens), address.NewBech32Codec("cosmosvaloper"), address.NewBech32Codec("cosmos"))
	assert.NilError(t, f.stakingKeeper.SetRedelegation(f.ctx, rdA))

	// set the associated delegation
	delA := types.NewDelegation(addrDels[0].String(), addrVals[1].String(), math.LegacyNewDecFromInt(rdATokens))
	assert.NilError(t, f.stakingKeeper.SetDelegation(f.ctx, delA))

	// set an unbonding delegation with expiration timestamp (beyond which the
	// unbonding delegation shouldn't be slashed)
	ubdATokens := f.stakingKeeper.TokensFromConsensusPower(f.ctx, 4)
	ubdA := types.NewUnbondingDelegation(addrDels[0], addrVals[0], 11,
		time.Unix(0, 0), ubdATokens, address.NewBech32Codec("cosmosvaloper"), address.NewBech32Codec("cosmos"))
	assert.NilError(t, f.stakingKeeper.SetUnbondingDelegation(f.ctx, ubdA))

	bondedCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, rdATokens.MulRaw(2)))
	notBondedCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, ubdATokens))

	// update bonded tokens
	bondedPool := f.stakingKeeper.GetBondedPool(f.ctx)
	notBondedPool := f.stakingKeeper.GetNotBondedPool(f.ctx)

	assert.NilError(t, banktestutil.FundModuleAccount(f.ctx, f.bankKeeper, bondedPool.GetName(), bondedCoins))
	assert.NilError(t, banktestutil.FundModuleAccount(f.ctx, f.bankKeeper, notBondedPool.GetName(), notBondedCoins))

	f.accountKeeper.SetModuleAccount(f.ctx, bondedPool)
	f.accountKeeper.SetModuleAccount(f.ctx, notBondedPool)

	oldBonded := f.bankKeeper.GetBalance(f.ctx, bondedPool.GetAddress(), bondDenom).Amount
	oldNotBonded := f.bankKeeper.GetBalance(f.ctx, notBondedPool.GetAddress(), bondDenom).Amount
	// slash validator
	f.ctx = integration.SetHeaderInfo(f.ctx, header.Info{Height: 12})
	_, found := f.stakingKeeper.GetValidatorByConsAddr(f.ctx, sdk.GetConsAddress(PKs[0]))
	assert.Assert(t, found)
	consAddr0 := sdk.ConsAddress(PKs[0].Address())
	_, err = f.stakingKeeper.Slash(f.ctx, consAddr0, 10, 10, fraction)
	assert.NilError(t, err)

	burnedNotBondedAmount := fraction.MulInt(ubdATokens).TruncateInt()
	burnedBondAmount := math.LegacyNewDecFromInt(f.stakingKeeper.TokensFromConsensusPower(f.ctx, 10)).Mul(fraction).TruncateInt()
	burnedBondAmount = burnedBondAmount.Sub(burnedNotBondedAmount)

	// read updated pool
	bondedPool = f.stakingKeeper.GetBondedPool(f.ctx)
	notBondedPool = f.stakingKeeper.GetNotBondedPool(f.ctx)

	bondedPoolBalance := f.bankKeeper.GetBalance(f.ctx, bondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldBonded.Sub(burnedBondAmount), bondedPoolBalance))

	notBondedPoolBalance := f.bankKeeper.GetBalance(f.ctx, notBondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldNotBonded.Sub(burnedNotBondedAmount), notBondedPoolBalance))

	// read updating redelegation
	rdA, found = f.stakingKeeper.Redelegations.Get(f.ctx, collections.Join3(addrDels[0].Bytes(), addrVals[0].Bytes(), addrVals[1].Bytes()))
	assert.Assert(t, found)
	assert.Assert(t, len(rdA.Entries) == 1)
	// read updated validator
	validator, found := f.stakingKeeper.GetValidatorByConsAddr(f.ctx, sdk.GetConsAddress(PKs[0]))
	assert.Assert(t, found)
	// power not decreased, all stake was bonded since
	assert.Equal(t, int64(10), validator.GetConsensusPower(f.stakingKeeper.PowerReduction(f.ctx)))
}

func TestSlashAmount(t *testing.T) {
	f, _, _ := bootstrapSlashTest(t, 10)
	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := math.LegacyNewDecWithPrec(5, 1)
	burnedCoins, err := f.stakingKeeper.Slash(f.ctx, consAddr, int64(f.app.LastBlockHeight()), 10, fraction)
	assert.NilError(t, err)
	assert.Assert(t, burnedCoins.GT(math.ZeroInt()))

	// test the case where the validator was not found, which should return no coins
	_, addrVals := generateAddresses(f, 100)
	noBurned, err := f.stakingKeeper.Slash(f.ctx, sdk.ConsAddress(addrVals[0]), int64(f.app.LastBlockHeight())+1, 10, fraction)
	assert.NilError(t, err)
	assert.Assert(t, math.NewInt(0).Equal(noBurned))
}

// TestFixAvoidFullSlashPenalty fixes the following issue: https://github.com/cosmos/cosmos-sdk/issues/20641
func TestFixAvoidFullSlashPenalty(t *testing.T) {
	// setup
	f := initFixture(t, false)
	ctx := f.ctx

	stakingMsgServer := keeper.NewMsgServerImpl(f.stakingKeeper)
	bondDenom, err := f.stakingKeeper.BondDenom(ctx)
	require.NoError(t, err)
	// create 2 evil validators, controlled by attacker
	evilValPubKey := secp256k1.GenPrivKey().PubKey()
	evilValPubKey2 := secp256k1.GenPrivKey().PubKey()
	// attacker user account
	badtestAcc := sdk.AccAddress("addr1_______________")
	// normal users who stakes on evilValAddr1
	testAcc1 := sdk.AccAddress("addr2_______________")
	testAcc2 := sdk.AccAddress("addr3_______________")
	createAccount(t, ctx, f.accountKeeper, badtestAcc)
	createAccount(t, ctx, f.accountKeeper, testAcc1)
	createAccount(t, ctx, f.accountKeeper, testAcc2)
	// fund all accounts
	testCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, f.stakingKeeper.TokensFromConsensusPower(ctx, 1)))
	require.NoError(t, banktestutil.FundAccount(ctx, f.bankKeeper, badtestAcc, testCoins))
	require.NoError(t, banktestutil.FundAccount(ctx, f.bankKeeper, testAcc1, testCoins))
	require.NoError(t, banktestutil.FundAccount(ctx, f.bankKeeper, testAcc2, testCoins))
	// create evilValAddr1 for normal staking operations
	evilValAddr1 := sdk.ValAddress(evilValPubKey.Address())
	createAccount(t, ctx, f.accountKeeper, evilValAddr1.Bytes())
	require.NoError(t, banktestutil.FundAccount(ctx, f.bankKeeper, sdk.AccAddress(evilValAddr1), testCoins))
	createValMsg1, _ := types.NewMsgCreateValidator(
		evilValAddr1.String(), evilValPubKey, testCoins[0], types.Description{Details: "test"}, types.NewCommissionRates(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0)), math.OneInt())
	_, err = stakingMsgServer.CreateValidator(ctx, createValMsg1)
	require.NoError(t, err)
	// very small amount coin for evilValAddr2
	smallCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewInt(1)))
	// create evilValAddr2 to circumvent slashing
	evilValAddr2 := sdk.ValAddress(evilValPubKey2.Address())
	require.NoError(t, banktestutil.FundAccount(ctx, f.bankKeeper, sdk.AccAddress(evilValAddr2), smallCoins))
	createValMsg3, _ := types.NewMsgCreateValidator(
		evilValAddr2.String(), evilValPubKey2, smallCoins[0], types.Description{Details: "test"}, types.NewCommissionRates(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0)), math.OneInt())
	createAccount(t, ctx, f.accountKeeper, evilValAddr2.Bytes())
	_, err = stakingMsgServer.CreateValidator(ctx, createValMsg3)
	require.NoError(t, err)
	// next block
	ctx = integration.SetHeaderInfo(ctx, header.Info{Height: int64(f.app.LastBlockHeight()) + 1})
	_, state := f.app.Deliver(t, ctx, nil)
	_, err = f.app.Commit(state)
	require.NoError(t, err)
	// all accs delegate to evilValAddr1
	delMsg := types.NewMsgDelegate(badtestAcc.String(), evilValAddr1.String(), testCoins[0])
	_, err = stakingMsgServer.Delegate(ctx, delMsg)
	require.NoError(t, err)
	delMsg = types.NewMsgDelegate(testAcc1.String(), evilValAddr1.String(), testCoins[0])
	_, err = stakingMsgServer.Delegate(ctx, delMsg)
	require.NoError(t, err)
	delMsg = types.NewMsgDelegate(testAcc2.String(), evilValAddr1.String(), testCoins[0])
	_, err = stakingMsgServer.Delegate(ctx, delMsg)
	require.NoError(t, err)
	// next block
	_, state = f.app.Deliver(t, ctx, nil)
	_, err = f.app.Commit(state)
	require.NoError(t, err)
	// 1. badtestAcc redelegates from evilValAddr1 to evilValAddr2
	redelMsg := types.NewMsgBeginRedelegate(badtestAcc.String(), evilValAddr1.String(), evilValAddr2.String(), smallCoins[0])
	_, err = stakingMsgServer.BeginRedelegate(ctx, redelMsg)
	require.NoError(t, err)
	// 2. evilValAddr2 undelegates its self-delegation and jail themselves
	undelMsg := types.NewMsgUndelegate(sdk.AccAddress(evilValAddr2).String(), evilValAddr2.String(), smallCoins[0])
	_, err = stakingMsgServer.Undelegate(ctx, undelMsg)
	require.NoError(t, err)
	// assert evilValAddr2 is jailed
	evilVal2, err := f.stakingKeeper.GetValidator(ctx, evilValAddr2)
	require.NoError(t, err)
	require.True(t, evilVal2.Jailed)
	// next block
	_, state = f.app.Deliver(t, ctx, nil)
	_, err = f.app.Commit(state)
	require.NoError(t, err)
	// evilValAddr1 is bad!
	// lets slash evilValAddr1 with a 100% penalty
	evilVal, err := f.stakingKeeper.GetValidator(ctx, evilValAddr1)
	require.NoError(t, err)
	evilValConsAddr, err := evilVal.GetConsAddr()
	require.NoError(t, err)
	evilPower := f.stakingKeeper.TokensToConsensusPower(ctx, evilVal.Tokens)
	err = f.slashKeeper.Slash(ctx, evilValConsAddr, math.LegacyMustNewDecFromStr("1.0"), evilPower, 3)
	require.NoError(t, err)
}

func createAccount(t *testing.T, ctx context.Context, k authkeeper.AccountKeeperI, addr sdk.AccAddress) {
	t.Helper()
	acc := k.NewAccountWithAddress(ctx, addr)
	k.SetAccount(ctx, acc)
}
