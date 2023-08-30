package keeper_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// bootstrapSlashTest creates 3 validators and bootstrap the app.
func bootstrapSlashTest(t *testing.T, power int64) (*fixture, []sdk.AccAddress, []sdk.ValAddress) {
	t.Helper()
	t.Parallel()
	f := initFixture(t)

	addrDels, addrVals := generateAddresses(f, 100)

	amt := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, power)
	bondDenom, err := f.stakingKeeper.BondDenom(f.sdkCtx)
	require.NoError(t, err)
	totalSupply := sdk.NewCoins(sdk.NewCoin(bondDenom, amt.MulRaw(int64(len(addrDels)))))

	notBondedPool := f.stakingKeeper.GetNotBondedPool(f.sdkCtx)
	assert.NilError(t, banktestutil.FundModuleAccount(f.sdkCtx, f.bankKeeper, notBondedPool.GetName(), totalSupply))

	f.accountKeeper.SetModuleAccount(f.sdkCtx, notBondedPool)

	numVals := int64(3)
	bondedCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, amt.MulRaw(numVals)))
	bondedPool := f.stakingKeeper.GetBondedPool(f.sdkCtx)

	// set bonded pool balance
	f.accountKeeper.SetModuleAccount(f.sdkCtx, bondedPool)
	assert.NilError(t, banktestutil.FundModuleAccount(f.sdkCtx, f.bankKeeper, bondedPool.GetName(), bondedCoins))

	for i := int64(0); i < numVals; i++ {
		validator := testutil.NewValidator(t, addrVals[i], PKs[i])
		validator, _ = validator.AddTokensFromDel(amt)
		validator = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validator, true)
		assert.NilError(t, f.stakingKeeper.SetValidatorByConsAddr(f.sdkCtx, validator))
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
		time.Unix(5, 0), math.NewInt(10), 0, address.NewBech32Codec("cosmosvaloper"), address.NewBech32Codec("cosmos"))

	assert.NilError(t, f.stakingKeeper.SetUnbondingDelegation(f.sdkCtx, ubd))

	// unbonding started prior to the infraction height, stakw didn't contribute
	slashAmount, err := f.stakingKeeper.SlashUnbondingDelegation(f.sdkCtx, ubd, 1, fraction)
	assert.NilError(t, err)
	assert.Assert(t, slashAmount.Equal(math.NewInt(0)))

	// after the expiration time, no longer eligible for slashing
	f.sdkCtx = f.sdkCtx.WithBlockHeader(cmtproto.Header{Time: time.Unix(10, 0)})
	assert.NilError(t, f.stakingKeeper.SetUnbondingDelegation(f.sdkCtx, ubd))
	slashAmount, err = f.stakingKeeper.SlashUnbondingDelegation(f.sdkCtx, ubd, 0, fraction)
	assert.NilError(t, err)
	assert.Assert(t, slashAmount.Equal(math.NewInt(0)))

	// test valid slash, before expiration timestamp and to which stake contributed
	notBondedPool := f.stakingKeeper.GetNotBondedPool(f.sdkCtx)
	oldUnbondedPoolBalances := f.bankKeeper.GetAllBalances(f.sdkCtx, notBondedPool.GetAddress())
	f.sdkCtx = f.sdkCtx.WithBlockHeader(cmtproto.Header{Time: time.Unix(0, 0)})
	assert.NilError(t, f.stakingKeeper.SetUnbondingDelegation(f.sdkCtx, ubd))
	slashAmount, err = f.stakingKeeper.SlashUnbondingDelegation(f.sdkCtx, ubd, 0, fraction)
	assert.NilError(t, err)
	assert.Assert(t, slashAmount.Equal(math.NewInt(5)))
	ubd, found := f.stakingKeeper.GetUnbondingDelegation(f.sdkCtx, addrDels[0], addrVals[0])
	assert.Assert(t, found)
	assert.Assert(t, len(ubd.Entries) == 1)

	// initial balance unchanged
	assert.DeepEqual(t, math.NewInt(10), ubd.Entries[0].InitialBalance)

	// balance decreased
	assert.DeepEqual(t, math.NewInt(5), ubd.Entries[0].Balance)
	newUnbondedPoolBalances := f.bankKeeper.GetAllBalances(f.sdkCtx, notBondedPool.GetAddress())
	diffTokens := oldUnbondedPoolBalances.Sub(newUnbondedPoolBalances...)
	bondDenom, err := f.stakingKeeper.BondDenom(f.sdkCtx)
	assert.NilError(t, err)
	assert.Assert(t, diffTokens.AmountOf(bondDenom).Equal(math.NewInt(5)))
}

// tests slashRedelegation
func TestSlashRedelegation(t *testing.T) {
	f, addrDels, addrVals := bootstrapSlashTest(t, 10)
	fraction := math.LegacyNewDecWithPrec(5, 1)

	bondDenom, err := f.stakingKeeper.BondDenom(f.sdkCtx)
	assert.NilError(t, err)

	// add bonded tokens to pool for (re)delegations
	startCoins := sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 15))
	bondedPool := f.stakingKeeper.GetBondedPool(f.sdkCtx)
	_ = f.bankKeeper.GetAllBalances(f.sdkCtx, bondedPool.GetAddress())

	assert.NilError(t, banktestutil.FundModuleAccount(f.sdkCtx, f.bankKeeper, bondedPool.GetName(), startCoins))
	f.accountKeeper.SetModuleAccount(f.sdkCtx, bondedPool)

	// set a redelegation with an expiration timestamp beyond which the
	// redelegation shouldn't be slashed
	rd := types.NewRedelegation(addrDels[0], addrVals[0], addrVals[1], 0,
		time.Unix(5, 0), math.NewInt(10), math.LegacyNewDec(10), 0, address.NewBech32Codec("cosmosvaloper"), address.NewBech32Codec("cosmos"))

	assert.NilError(t, f.stakingKeeper.SetRedelegation(f.sdkCtx, rd))

	// set the associated delegation
	del := types.NewDelegation(addrDels[0].String(), addrVals[1].String(), math.LegacyNewDec(10))
	assert.NilError(t, f.stakingKeeper.SetDelegation(f.sdkCtx, del))

	// started redelegating prior to the current height, stake didn't contribute to infraction
	validator, found := f.stakingKeeper.GetValidator(f.sdkCtx, addrVals[1])
	assert.Assert(t, found)
	slashAmount, err := f.stakingKeeper.SlashRedelegation(f.sdkCtx, validator, rd, 1, fraction)
	assert.NilError(t, err)
	assert.Assert(t, slashAmount.Equal(math.NewInt(0)))

	// after the expiration time, no longer eligible for slashing
	f.sdkCtx = f.sdkCtx.WithBlockHeader(cmtproto.Header{Time: time.Unix(10, 0)})
	assert.NilError(t, f.stakingKeeper.SetRedelegation(f.sdkCtx, rd))
	validator, found = f.stakingKeeper.GetValidator(f.sdkCtx, addrVals[1])
	assert.Assert(t, found)
	slashAmount, err = f.stakingKeeper.SlashRedelegation(f.sdkCtx, validator, rd, 0, fraction)
	assert.NilError(t, err)
	assert.Assert(t, slashAmount.Equal(math.NewInt(0)))

	balances := f.bankKeeper.GetAllBalances(f.sdkCtx, bondedPool.GetAddress())

	// test valid slash, before expiration timestamp and to which stake contributed
	f.sdkCtx = f.sdkCtx.WithBlockHeader(cmtproto.Header{Time: time.Unix(0, 0)})
	assert.NilError(t, f.stakingKeeper.SetRedelegation(f.sdkCtx, rd))
	validator, found = f.stakingKeeper.GetValidator(f.sdkCtx, addrVals[1])
	assert.Assert(t, found)
	slashAmount, err = f.stakingKeeper.SlashRedelegation(f.sdkCtx, validator, rd, 0, fraction)
	assert.NilError(t, err)
	assert.Assert(t, slashAmount.Equal(math.NewInt(5)))
	rd, found = f.stakingKeeper.Redelegations.Get(f.sdkCtx, collections.Join3(addrDels[0].Bytes(), addrVals[0].Bytes(), addrVals[1].Bytes()))
	assert.Assert(t, found)
	assert.Assert(t, len(rd.Entries) == 1)

	// end block
	applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 1)

	// initialbalance unchanged
	assert.DeepEqual(t, math.NewInt(10), rd.Entries[0].InitialBalance)

	// shares decreased
	del, found = f.stakingKeeper.Delegations.Get(f.sdkCtx, collections.Join(addrDels[0], addrVals[1]))
	assert.Assert(t, found)
	assert.Equal(t, int64(5), del.Shares.RoundInt64())

	// pool bonded tokens should decrease
	burnedCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, slashAmount))
	assert.DeepEqual(t, balances.Sub(burnedCoins...), f.bankKeeper.GetAllBalances(f.sdkCtx, bondedPool.GetAddress()))
}

// test slash at a negative height
// this just represents pre-genesis and should have the same effect as slashing at height 0
func TestSlashAtNegativeHeight(t *testing.T) {
	f, _, _ := bootstrapSlashTest(t, 10)
	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := math.LegacyNewDecWithPrec(5, 1)

	bondedPool := f.stakingKeeper.GetBondedPool(f.sdkCtx)
	oldBondedPoolBalances := f.bankKeeper.GetAllBalances(f.sdkCtx, bondedPool.GetAddress())

	_, found := f.stakingKeeper.GetValidatorByConsAddr(f.sdkCtx, consAddr)
	assert.Assert(t, found)
	_, err := f.stakingKeeper.Slash(f.sdkCtx, consAddr, -2, 10, fraction)
	assert.NilError(t, err)

	// read updated state
	validator, found := f.stakingKeeper.GetValidatorByConsAddr(f.sdkCtx, consAddr)
	assert.Assert(t, found)

	// end block
	applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 1)

	valbz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
	assert.NilError(t, err)

	validator, found = f.stakingKeeper.GetValidator(f.sdkCtx, valbz)
	assert.Assert(t, found)
	// power decreased
	assert.Equal(t, int64(5), validator.GetConsensusPower(f.stakingKeeper.PowerReduction(f.sdkCtx)))

	bondDenom, err := f.stakingKeeper.BondDenom(f.sdkCtx)
	assert.NilError(t, err)

	// pool bonded shares decreased
	newBondedPoolBalances := f.bankKeeper.GetAllBalances(f.sdkCtx, bondedPool.GetAddress())
	diffTokens := oldBondedPoolBalances.Sub(newBondedPoolBalances...).AmountOf(bondDenom)
	assert.DeepEqual(t, f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 5).String(), diffTokens.String())
}

// tests Slash at the current height
func TestSlashValidatorAtCurrentHeight(t *testing.T) {
	f, _, _ := bootstrapSlashTest(t, 10)
	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := math.LegacyNewDecWithPrec(5, 1)

	bondDenom, err := f.stakingKeeper.BondDenom(f.sdkCtx)
	assert.NilError(t, err)

	bondedPool := f.stakingKeeper.GetBondedPool(f.sdkCtx)
	oldBondedPoolBalances := f.bankKeeper.GetAllBalances(f.sdkCtx, bondedPool.GetAddress())

	_, found := f.stakingKeeper.GetValidatorByConsAddr(f.sdkCtx, consAddr)
	assert.Assert(t, found)
	_, err = f.stakingKeeper.Slash(f.sdkCtx, consAddr, f.sdkCtx.BlockHeight(), 10, fraction)
	assert.NilError(t, err)

	// read updated state
	validator, found := f.stakingKeeper.GetValidatorByConsAddr(f.sdkCtx, consAddr)
	assert.Assert(t, found)

	// end block
	applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 1)

	valbz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
	assert.NilError(t, err)

	validator, found = f.stakingKeeper.GetValidator(f.sdkCtx, valbz)
	assert.Assert(t, found)
	// power decreased
	assert.Equal(t, int64(5), validator.GetConsensusPower(f.stakingKeeper.PowerReduction(f.sdkCtx)))

	// pool bonded shares decreased
	newBondedPoolBalances := f.bankKeeper.GetAllBalances(f.sdkCtx, bondedPool.GetAddress())
	diffTokens := oldBondedPoolBalances.Sub(newBondedPoolBalances...).AmountOf(bondDenom)
	assert.DeepEqual(t, f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 5).String(), diffTokens.String())
}

// tests Slash at a previous height with an unbonding delegation
func TestSlashWithUnbondingDelegation(t *testing.T) {
	f, addrDels, addrVals := bootstrapSlashTest(t, 10)

	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := math.LegacyNewDecWithPrec(5, 1)

	bondDenom, err := f.stakingKeeper.BondDenom(f.sdkCtx)
	assert.NilError(t, err)

	// set an unbonding delegation with expiration timestamp beyond which the
	// unbonding delegation shouldn't be slashed
	ubdTokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 4)
	ubd := types.NewUnbondingDelegation(addrDels[0], addrVals[0], 11, time.Unix(0, 0), ubdTokens, 0, address.NewBech32Codec("cosmosvaloper"), address.NewBech32Codec("cosmos"))
	assert.NilError(t, f.stakingKeeper.SetUnbondingDelegation(f.sdkCtx, ubd))

	// slash validator for the first time
	f.sdkCtx = f.sdkCtx.WithBlockHeight(12)
	bondedPool := f.stakingKeeper.GetBondedPool(f.sdkCtx)
	oldBondedPoolBalances := f.bankKeeper.GetAllBalances(f.sdkCtx, bondedPool.GetAddress())

	_, found := f.stakingKeeper.GetValidatorByConsAddr(f.sdkCtx, consAddr)
	assert.Assert(t, found)
	_, err = f.stakingKeeper.Slash(f.sdkCtx, consAddr, 10, 10, fraction)
	assert.NilError(t, err)

	// end block
	applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 1)

	// read updating unbonding delegation
	ubd, found = f.stakingKeeper.GetUnbondingDelegation(f.sdkCtx, addrDels[0], addrVals[0])
	assert.Assert(t, found)
	assert.Assert(t, len(ubd.Entries) == 1)

	// balance decreased
	assert.DeepEqual(t, f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 2), ubd.Entries[0].Balance)

	// bonded tokens burned
	newBondedPoolBalances := f.bankKeeper.GetAllBalances(f.sdkCtx, bondedPool.GetAddress())
	diffTokens := oldBondedPoolBalances.Sub(newBondedPoolBalances...).AmountOf(bondDenom)
	assert.DeepEqual(t, f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 3), diffTokens)

	// read updated validator
	validator, found := f.stakingKeeper.GetValidatorByConsAddr(f.sdkCtx, consAddr)
	assert.Assert(t, found)

	// power decreased by 3 - 6 stake originally bonded at the time of infraction
	// was still bonded at the time of discovery and was slashed by half, 4 stake
	// bonded at the time of discovery hadn't been bonded at the time of infraction
	// and wasn't slashed
	assert.Equal(t, int64(7), validator.GetConsensusPower(f.stakingKeeper.PowerReduction(f.sdkCtx)))

	// slash validator again
	f.sdkCtx = f.sdkCtx.WithBlockHeight(13)
	_, err = f.stakingKeeper.Slash(f.sdkCtx, consAddr, 9, 10, fraction)
	assert.NilError(t, err)

	ubd, found = f.stakingKeeper.GetUnbondingDelegation(f.sdkCtx, addrDels[0], addrVals[0])
	assert.Assert(t, found)
	assert.Assert(t, len(ubd.Entries) == 1)

	// balance decreased again
	assert.DeepEqual(t, math.NewInt(0), ubd.Entries[0].Balance)

	// bonded tokens burned again
	newBondedPoolBalances = f.bankKeeper.GetAllBalances(f.sdkCtx, bondedPool.GetAddress())
	diffTokens = oldBondedPoolBalances.Sub(newBondedPoolBalances...).AmountOf(bondDenom)
	assert.DeepEqual(t, f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 6), diffTokens)

	// read updated validator
	validator, found = f.stakingKeeper.GetValidatorByConsAddr(f.sdkCtx, consAddr)
	assert.Assert(t, found)

	// power decreased by 3 again
	assert.Equal(t, int64(4), validator.GetConsensusPower(f.stakingKeeper.PowerReduction(f.sdkCtx)))

	// slash validator again
	// all originally bonded stake has been slashed, so this will have no effect
	// on the unbonding delegation, but it will slash stake bonded since the infraction
	// this may not be the desirable behavior, ref https://github.com/cosmos/cosmos-sdk/issues/1440
	f.sdkCtx = f.sdkCtx.WithBlockHeight(13)
	_, err = f.stakingKeeper.Slash(f.sdkCtx, consAddr, 9, 10, fraction)
	assert.NilError(t, err)

	ubd, found = f.stakingKeeper.GetUnbondingDelegation(f.sdkCtx, addrDels[0], addrVals[0])
	assert.Assert(t, found)
	assert.Assert(t, len(ubd.Entries) == 1)

	// balance unchanged
	assert.DeepEqual(t, math.NewInt(0), ubd.Entries[0].Balance)

	// bonded tokens burned again
	newBondedPoolBalances = f.bankKeeper.GetAllBalances(f.sdkCtx, bondedPool.GetAddress())
	diffTokens = oldBondedPoolBalances.Sub(newBondedPoolBalances...).AmountOf(bondDenom)
	assert.DeepEqual(t, f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 9), diffTokens)

	// read updated validator
	validator, found = f.stakingKeeper.GetValidatorByConsAddr(f.sdkCtx, consAddr)
	assert.Assert(t, found)

	// power decreased by 3 again
	assert.Equal(t, int64(1), validator.GetConsensusPower(f.stakingKeeper.PowerReduction(f.sdkCtx)))

	// slash validator again
	// all originally bonded stake has been slashed, so this will have no effect
	// on the unbonding delegation, but it will slash stake bonded since the infraction
	// this may not be the desirable behavior, ref https://github.com/cosmos/cosmos-sdk/issues/1440
	f.sdkCtx = f.sdkCtx.WithBlockHeight(13)
	_, err = f.stakingKeeper.Slash(f.sdkCtx, consAddr, 9, 10, fraction)
	assert.NilError(t, err)

	ubd, found = f.stakingKeeper.GetUnbondingDelegation(f.sdkCtx, addrDels[0], addrVals[0])
	assert.Assert(t, found)
	assert.Assert(t, len(ubd.Entries) == 1)

	// balance unchanged
	assert.DeepEqual(t, math.NewInt(0), ubd.Entries[0].Balance)

	// just 1 bonded token burned again since that's all the validator now has
	newBondedPoolBalances = f.bankKeeper.GetAllBalances(f.sdkCtx, bondedPool.GetAddress())
	diffTokens = oldBondedPoolBalances.Sub(newBondedPoolBalances...).AmountOf(bondDenom)
	assert.DeepEqual(t, f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 10), diffTokens)

	// apply TM updates
	applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, -1)

	// read updated validator
	// power decreased by 1 again, validator is out of stake
	// validator should be in unbonding period
	validator, _ = f.stakingKeeper.GetValidatorByConsAddr(f.sdkCtx, consAddr)
	assert.Equal(t, validator.GetStatus(), types.Unbonding)
}

// tests Slash at a previous height with a redelegation
func TestSlashWithRedelegation(t *testing.T) {
	f, addrDels, addrVals := bootstrapSlashTest(t, 10)
	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := math.LegacyNewDecWithPrec(5, 1)
	bondDenom, err := f.stakingKeeper.BondDenom(f.sdkCtx)
	assert.NilError(t, err)

	// set a redelegation
	rdTokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 6)
	rd := types.NewRedelegation(addrDels[0], addrVals[0], addrVals[1], 11, time.Unix(0, 0), rdTokens, math.LegacyNewDecFromInt(rdTokens), 0, address.NewBech32Codec("cosmosvaloper"), address.NewBech32Codec("cosmos"))
	assert.NilError(t, f.stakingKeeper.SetRedelegation(f.sdkCtx, rd))

	// set the associated delegation
	del := types.NewDelegation(addrDels[0].String(), addrVals[1].String(), math.LegacyNewDecFromInt(rdTokens))
	assert.NilError(t, f.stakingKeeper.SetDelegation(f.sdkCtx, del))

	// update bonded tokens
	bondedPool := f.stakingKeeper.GetBondedPool(f.sdkCtx)
	notBondedPool := f.stakingKeeper.GetNotBondedPool(f.sdkCtx)
	rdCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, rdTokens.MulRaw(2)))

	assert.NilError(t, banktestutil.FundModuleAccount(f.sdkCtx, f.bankKeeper, bondedPool.GetName(), rdCoins))

	f.accountKeeper.SetModuleAccount(f.sdkCtx, bondedPool)

	oldBonded := f.bankKeeper.GetBalance(f.sdkCtx, bondedPool.GetAddress(), bondDenom).Amount
	oldNotBonded := f.bankKeeper.GetBalance(f.sdkCtx, notBondedPool.GetAddress(), bondDenom).Amount

	// slash validator
	f.sdkCtx = f.sdkCtx.WithBlockHeight(12)
	_, found := f.stakingKeeper.GetValidatorByConsAddr(f.sdkCtx, consAddr)
	assert.Assert(t, found)

	_, err = f.stakingKeeper.Slash(f.sdkCtx, consAddr, 10, 10, fraction)
	assert.NilError(t, err)

	burnAmount := math.LegacyNewDecFromInt(f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 10)).Mul(fraction).TruncateInt()

	bondedPool = f.stakingKeeper.GetBondedPool(f.sdkCtx)
	notBondedPool = f.stakingKeeper.GetNotBondedPool(f.sdkCtx)

	// burn bonded tokens from only from delegations
	bondedPoolBalance := f.bankKeeper.GetBalance(f.sdkCtx, bondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldBonded.Sub(burnAmount), bondedPoolBalance))

	notBondedPoolBalance := f.bankKeeper.GetBalance(f.sdkCtx, notBondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldNotBonded, notBondedPoolBalance))
	oldBonded = f.bankKeeper.GetBalance(f.sdkCtx, bondedPool.GetAddress(), bondDenom).Amount

	// read updating redelegation
	rd, found = f.stakingKeeper.Redelegations.Get(f.sdkCtx, collections.Join3(addrDels[0].Bytes(), addrVals[0].Bytes(), addrVals[1].Bytes()))
	assert.Assert(t, found)
	assert.Assert(t, len(rd.Entries) == 1)
	// read updated validator
	validator, found := f.stakingKeeper.GetValidatorByConsAddr(f.sdkCtx, consAddr)
	assert.Assert(t, found)
	// power decreased by 2 - 4 stake originally bonded at the time of infraction
	// was still bonded at the time of discovery and was slashed by half, 4 stake
	// bonded at the time of discovery hadn't been bonded at the time of infraction
	// and wasn't slashed
	assert.Equal(t, int64(8), validator.GetConsensusPower(f.stakingKeeper.PowerReduction(f.sdkCtx)))

	// slash the validator again
	_, found = f.stakingKeeper.GetValidatorByConsAddr(f.sdkCtx, consAddr)
	assert.Assert(t, found)

	_, err = f.stakingKeeper.Slash(f.sdkCtx, consAddr, 10, 10, math.LegacyOneDec())
	assert.NilError(t, err)
	burnAmount = f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 7)

	// read updated pool
	bondedPool = f.stakingKeeper.GetBondedPool(f.sdkCtx)
	notBondedPool = f.stakingKeeper.GetNotBondedPool(f.sdkCtx)

	// seven bonded tokens burned
	bondedPoolBalance = f.bankKeeper.GetBalance(f.sdkCtx, bondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldBonded.Sub(burnAmount), bondedPoolBalance))
	assert.Assert(math.IntEq(t, oldNotBonded, notBondedPoolBalance))

	bondedPoolBalance = f.bankKeeper.GetBalance(f.sdkCtx, bondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldBonded.Sub(burnAmount), bondedPoolBalance))

	notBondedPoolBalance = f.bankKeeper.GetBalance(f.sdkCtx, notBondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldNotBonded, notBondedPoolBalance))
	oldBonded = f.bankKeeper.GetBalance(f.sdkCtx, bondedPool.GetAddress(), bondDenom).Amount

	// read updating redelegation
	rd, found = f.stakingKeeper.Redelegations.Get(f.sdkCtx, collections.Join3(addrDels[0].Bytes(), addrVals[0].Bytes(), addrVals[1].Bytes()))
	assert.Assert(t, found)
	assert.Assert(t, len(rd.Entries) == 1)
	// read updated validator
	validator, found = f.stakingKeeper.GetValidatorByConsAddr(f.sdkCtx, consAddr)
	assert.Assert(t, found)
	// power decreased by 4
	assert.Equal(t, int64(4), validator.GetConsensusPower(f.stakingKeeper.PowerReduction(f.sdkCtx)))

	// slash the validator again, by 100%
	f.sdkCtx = f.sdkCtx.WithBlockHeight(12)
	_, found = f.stakingKeeper.GetValidatorByConsAddr(f.sdkCtx, consAddr)
	assert.Assert(t, found)

	_, err = f.stakingKeeper.Slash(f.sdkCtx, consAddr, 10, 10, math.LegacyOneDec())
	assert.NilError(t, err)

	burnAmount = math.LegacyNewDecFromInt(f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 10)).Mul(math.LegacyOneDec()).TruncateInt()
	burnAmount = burnAmount.Sub(math.LegacyOneDec().MulInt(rdTokens).TruncateInt())

	// read updated pool
	bondedPool = f.stakingKeeper.GetBondedPool(f.sdkCtx)
	notBondedPool = f.stakingKeeper.GetNotBondedPool(f.sdkCtx)

	bondedPoolBalance = f.bankKeeper.GetBalance(f.sdkCtx, bondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldBonded.Sub(burnAmount), bondedPoolBalance))
	notBondedPoolBalance = f.bankKeeper.GetBalance(f.sdkCtx, notBondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldNotBonded, notBondedPoolBalance))
	oldBonded = f.bankKeeper.GetBalance(f.sdkCtx, bondedPool.GetAddress(), bondDenom).Amount

	// read updating redelegation
	rd, found = f.stakingKeeper.Redelegations.Get(f.sdkCtx, collections.Join3(addrDels[0].Bytes(), addrVals[0].Bytes(), addrVals[1].Bytes()))
	assert.Assert(t, found)
	assert.Assert(t, len(rd.Entries) == 1)
	// apply TM updates
	applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, -1)
	// read updated validator
	// validator decreased to zero power, should be in unbonding period
	validator, _ = f.stakingKeeper.GetValidatorByConsAddr(f.sdkCtx, consAddr)
	assert.Equal(t, validator.GetStatus(), types.Unbonding)

	// slash the validator again, by 100%
	// no stake remains to be slashed
	f.sdkCtx = f.sdkCtx.WithBlockHeight(12)
	// validator still in unbonding period
	validator, _ = f.stakingKeeper.GetValidatorByConsAddr(f.sdkCtx, consAddr)
	assert.Equal(t, validator.GetStatus(), types.Unbonding)

	_, err = f.stakingKeeper.Slash(f.sdkCtx, consAddr, 10, 10, math.LegacyOneDec())
	assert.NilError(t, err)

	// read updated pool
	bondedPool = f.stakingKeeper.GetBondedPool(f.sdkCtx)
	notBondedPool = f.stakingKeeper.GetNotBondedPool(f.sdkCtx)

	bondedPoolBalance = f.bankKeeper.GetBalance(f.sdkCtx, bondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldBonded, bondedPoolBalance))
	notBondedPoolBalance = f.bankKeeper.GetBalance(f.sdkCtx, notBondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldNotBonded, notBondedPoolBalance))

	// read updating redelegation
	rd, found = f.stakingKeeper.Redelegations.Get(f.sdkCtx, collections.Join3(addrDels[0].Bytes(), addrVals[0].Bytes(), addrVals[1].Bytes()))
	assert.Assert(t, found)
	assert.Assert(t, len(rd.Entries) == 1)
	// read updated validator
	// power still zero, still in unbonding period
	validator, _ = f.stakingKeeper.GetValidatorByConsAddr(f.sdkCtx, consAddr)
	assert.Equal(t, validator.GetStatus(), types.Unbonding)
}

// tests Slash at a previous height with both an unbonding delegation and a redelegation
func TestSlashBoth(t *testing.T) {
	f, addrDels, addrVals := bootstrapSlashTest(t, 10)
	fraction := math.LegacyNewDecWithPrec(5, 1)
	bondDenom, err := f.stakingKeeper.BondDenom(f.sdkCtx)
	assert.NilError(t, err)

	// set a redelegation with expiration timestamp beyond which the
	// redelegation shouldn't be slashed
	rdATokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 6)
	rdA := types.NewRedelegation(addrDels[0], addrVals[0], addrVals[1], 11, time.Unix(0, 0), rdATokens, math.LegacyNewDecFromInt(rdATokens), 0, address.NewBech32Codec("cosmosvaloper"), address.NewBech32Codec("cosmos"))
	assert.NilError(t, f.stakingKeeper.SetRedelegation(f.sdkCtx, rdA))

	// set the associated delegation
	delA := types.NewDelegation(addrDels[0].String(), addrVals[1].String(), math.LegacyNewDecFromInt(rdATokens))
	assert.NilError(t, f.stakingKeeper.SetDelegation(f.sdkCtx, delA))

	// set an unbonding delegation with expiration timestamp (beyond which the
	// unbonding delegation shouldn't be slashed)
	ubdATokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 4)
	ubdA := types.NewUnbondingDelegation(addrDels[0], addrVals[0], 11,
		time.Unix(0, 0), ubdATokens, 0, address.NewBech32Codec("cosmosvaloper"), address.NewBech32Codec("cosmos"))
	assert.NilError(t, f.stakingKeeper.SetUnbondingDelegation(f.sdkCtx, ubdA))

	bondedCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, rdATokens.MulRaw(2)))
	notBondedCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, ubdATokens))

	// update bonded tokens
	bondedPool := f.stakingKeeper.GetBondedPool(f.sdkCtx)
	notBondedPool := f.stakingKeeper.GetNotBondedPool(f.sdkCtx)

	assert.NilError(t, banktestutil.FundModuleAccount(f.sdkCtx, f.bankKeeper, bondedPool.GetName(), bondedCoins))
	assert.NilError(t, banktestutil.FundModuleAccount(f.sdkCtx, f.bankKeeper, notBondedPool.GetName(), notBondedCoins))

	f.accountKeeper.SetModuleAccount(f.sdkCtx, bondedPool)
	f.accountKeeper.SetModuleAccount(f.sdkCtx, notBondedPool)

	oldBonded := f.bankKeeper.GetBalance(f.sdkCtx, bondedPool.GetAddress(), bondDenom).Amount
	oldNotBonded := f.bankKeeper.GetBalance(f.sdkCtx, notBondedPool.GetAddress(), bondDenom).Amount
	// slash validator
	f.sdkCtx = f.sdkCtx.WithBlockHeight(12)
	_, found := f.stakingKeeper.GetValidatorByConsAddr(f.sdkCtx, sdk.GetConsAddress(PKs[0]))
	assert.Assert(t, found)
	consAddr0 := sdk.ConsAddress(PKs[0].Address())
	_, err = f.stakingKeeper.Slash(f.sdkCtx, consAddr0, 10, 10, fraction)
	assert.NilError(t, err)

	burnedNotBondedAmount := fraction.MulInt(ubdATokens).TruncateInt()
	burnedBondAmount := math.LegacyNewDecFromInt(f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 10)).Mul(fraction).TruncateInt()
	burnedBondAmount = burnedBondAmount.Sub(burnedNotBondedAmount)

	// read updated pool
	bondedPool = f.stakingKeeper.GetBondedPool(f.sdkCtx)
	notBondedPool = f.stakingKeeper.GetNotBondedPool(f.sdkCtx)

	bondedPoolBalance := f.bankKeeper.GetBalance(f.sdkCtx, bondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldBonded.Sub(burnedBondAmount), bondedPoolBalance))

	notBondedPoolBalance := f.bankKeeper.GetBalance(f.sdkCtx, notBondedPool.GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, oldNotBonded.Sub(burnedNotBondedAmount), notBondedPoolBalance))

	// read updating redelegation
	rdA, found = f.stakingKeeper.Redelegations.Get(f.sdkCtx, collections.Join3(addrDels[0].Bytes(), addrVals[0].Bytes(), addrVals[1].Bytes()))
	assert.Assert(t, found)
	assert.Assert(t, len(rdA.Entries) == 1)
	// read updated validator
	validator, found := f.stakingKeeper.GetValidatorByConsAddr(f.sdkCtx, sdk.GetConsAddress(PKs[0]))
	assert.Assert(t, found)
	// power not decreased, all stake was bonded since
	assert.Equal(t, int64(10), validator.GetConsensusPower(f.stakingKeeper.PowerReduction(f.sdkCtx)))
}

func TestSlashAmount(t *testing.T) {
	f, _, _ := bootstrapSlashTest(t, 10)
	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := math.LegacyNewDecWithPrec(5, 1)
	burnedCoins, err := f.stakingKeeper.Slash(f.sdkCtx, consAddr, f.sdkCtx.BlockHeight(), 10, fraction)
	assert.NilError(t, err)
	assert.Assert(t, burnedCoins.GT(math.ZeroInt()))

	// test the case where the validator was not found, which should return no coins
	_, addrVals := generateAddresses(f, 100)
	noBurned, err := f.stakingKeeper.Slash(f.sdkCtx, sdk.ConsAddress(addrVals[0]), f.sdkCtx.BlockHeight(), 10, fraction)
	assert.NilError(t, err)
	assert.Assert(t, math.NewInt(0).Equal(noBurned))
}
