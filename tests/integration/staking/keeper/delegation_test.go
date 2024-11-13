package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	"cosmossdk.io/math"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
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

	validator = keeper.TestingUpdateValidator(f.stakingKeeper, ctx, validator, true)
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
		ctx = ctx.WithBlockHeight(i)
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

func TestValidatorBondUndelegate(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	ctx := f.sdkCtx

	addrDels := simtestutil.AddTestAddrs(f.bankKeeper, f.stakingKeeper, ctx, 2, f.stakingKeeper.TokensFromConsensusPower(ctx, 10000))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	startTokens := f.stakingKeeper.TokensFromConsensusPower(ctx, 10)

	bondDenom, err := f.stakingKeeper.BondDenom(ctx)
	require.NoError(t, err)
	notBondedPool := f.stakingKeeper.GetNotBondedPool(ctx)

	require.NoError(t, banktestutil.FundModuleAccount(ctx, f.bankKeeper, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(bondDenom, startTokens))))
	f.accountKeeper.SetModuleAccount(ctx, notBondedPool)

	// create a validator and a delegator to that validator
	validator := testutil.NewValidator(t, addrVals[0], PKs[0])
	validator.Status = types.Bonded
	f.stakingKeeper.SetValidator(ctx, validator)

	// set validator bond factor
	params, err := f.stakingKeeper.GetParams(ctx)
	require.NoError(t, err)
	params.ValidatorBondFactor = math.LegacyNewDec(1)
	f.stakingKeeper.SetParams(ctx, params)

	// convert to validator self-bond
	msgServer := keeper.NewMsgServerImpl(f.stakingKeeper)

	validator, _ = f.stakingKeeper.GetValidator(ctx, addrVals[0])
	err = delegateCoinsFromAccount(ctx, *f.stakingKeeper, addrDels[0], startTokens, validator)
	require.NoError(t, err)
	_, err = msgServer.ValidatorBond(ctx, &types.MsgValidatorBond{
		DelegatorAddress: addrDels[0].String(),
		ValidatorAddress: addrVals[0].String(),
	})
	require.NoError(t, err)

	// tokenize share for 2nd account delegation
	validator, _ = f.stakingKeeper.GetValidator(ctx, addrVals[0])
	err = delegateCoinsFromAccount(ctx, *f.stakingKeeper, addrDels[1], startTokens, validator)
	require.NoError(t, err)
	tokenizeShareResp, err := msgServer.TokenizeShares(ctx, &types.MsgTokenizeShares{
		DelegatorAddress:    addrDels[1].String(),
		ValidatorAddress:    addrVals[0].String(),
		Amount:              sdk.NewCoin(sdk.DefaultBondDenom, startTokens),
		TokenizedShareOwner: addrDels[0].String(),
	})
	require.NoError(t, err)

	// try undelegating
	_, err = msgServer.Undelegate(ctx, &types.MsgUndelegate{
		DelegatorAddress: addrDels[0].String(),
		ValidatorAddress: addrVals[0].String(),
		Amount:           sdk.NewCoin(sdk.DefaultBondDenom, startTokens),
	})
	require.Error(t, err)

	// redeem full amount on 2nd account and try undelegation
	validator, _ = f.stakingKeeper.GetValidator(ctx, addrVals[0])
	err = delegateCoinsFromAccount(ctx, *f.stakingKeeper, addrDels[1], startTokens, validator)
	require.NoError(t, err)
	_, err = msgServer.RedeemTokensForShares(ctx, &types.MsgRedeemTokensForShares{
		DelegatorAddress: addrDels[1].String(),
		Amount:           tokenizeShareResp.Amount,
	})
	require.NoError(t, err)

	// try undelegating
	_, err = msgServer.Undelegate(ctx, &types.MsgUndelegate{
		DelegatorAddress: addrDels[0].String(),
		ValidatorAddress: addrVals[0].String(),
		Amount:           sdk.NewCoin(sdk.DefaultBondDenom, startTokens),
	})
	require.NoError(t, err)

	validator, _ = f.stakingKeeper.GetValidator(ctx, addrVals[0])
	require.Equal(t, validator.ValidatorBondShares, math.LegacyZeroDec())
}

func TestValidatorBondRedelegate(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	ctx := f.sdkCtx

	addrDels := simtestutil.AddTestAddrs(f.bankKeeper, f.stakingKeeper, ctx, 2, f.stakingKeeper.TokensFromConsensusPower(ctx, 10000))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	startTokens := f.stakingKeeper.TokensFromConsensusPower(ctx, 10)

	bondDenom, err := f.stakingKeeper.BondDenom(ctx)
	require.NoError(t, err)
	notBondedPool := f.stakingKeeper.GetNotBondedPool(ctx)

	startPoolToken := sdk.NewCoins(sdk.NewCoin(bondDenom, startTokens.Mul(math.NewInt(2))))
	require.NoError(t, banktestutil.FundModuleAccount(ctx, f.bankKeeper, notBondedPool.GetName(), startPoolToken))
	f.accountKeeper.SetModuleAccount(ctx, notBondedPool)

	// create a validator and a delegator to that validator
	validator := testutil.NewValidator(t, addrVals[0], PKs[0])
	validator.Status = types.Bonded
	f.stakingKeeper.SetValidator(ctx, validator)
	validator2 := testutil.NewValidator(t, addrVals[1], PKs[1])
	validator.Status = types.Bonded
	f.stakingKeeper.SetValidator(ctx, validator2)

	// set validator bond factor
	params, err := f.stakingKeeper.GetParams(ctx)
	require.NoError(t, err)
	params.ValidatorBondFactor = math.LegacyNewDec(1)
	f.stakingKeeper.SetParams(ctx, params)

	// set total liquid stake
	f.stakingKeeper.SetTotalLiquidStakedTokens(ctx, math.NewInt(100))

	// delegate to each validator
	validator, _ = f.stakingKeeper.GetValidator(ctx, addrVals[0])
	err = delegateCoinsFromAccount(ctx, *f.stakingKeeper, addrDels[0], startTokens, validator)
	require.NoError(t, err)

	validator2, _ = f.stakingKeeper.GetValidator(ctx, addrVals[1])
	err = delegateCoinsFromAccount(ctx, *f.stakingKeeper, addrDels[1], startTokens, validator2)
	require.NoError(t, err)

	// convert to validator self-bond
	msgServer := keeper.NewMsgServerImpl(f.stakingKeeper)
	_, err = msgServer.ValidatorBond(ctx, &types.MsgValidatorBond{
		DelegatorAddress: addrDels[0].String(),
		ValidatorAddress: addrVals[0].String(),
	})
	require.NoError(t, err)

	// tokenize share for 2nd account delegation
	validator, _ = f.stakingKeeper.GetValidator(ctx, addrVals[0])
	err = delegateCoinsFromAccount(ctx, *f.stakingKeeper, addrDels[1], startTokens, validator)
	require.NoError(t, err)
	tokenizeShareResp, err := msgServer.TokenizeShares(ctx, &types.MsgTokenizeShares{
		DelegatorAddress:    addrDels[1].String(),
		ValidatorAddress:    addrVals[0].String(),
		Amount:              sdk.NewCoin(sdk.DefaultBondDenom, startTokens),
		TokenizedShareOwner: addrDels[0].String(),
	})
	require.NoError(t, err)

	// try undelegating
	_, err = msgServer.BeginRedelegate(ctx, &types.MsgBeginRedelegate{
		DelegatorAddress:    addrDels[0].String(),
		ValidatorSrcAddress: addrVals[0].String(),
		ValidatorDstAddress: addrVals[1].String(),
		Amount:              sdk.NewCoin(sdk.DefaultBondDenom, startTokens),
	})
	require.Error(t, err)

	// redeem full amount on 2nd account and try undelegation
	validator, _ = f.stakingKeeper.GetValidator(ctx, addrVals[0])
	err = delegateCoinsFromAccount(ctx, *f.stakingKeeper, addrDels[1], startTokens, validator)
	require.NoError(t, err)
	_, err = msgServer.RedeemTokensForShares(ctx, &types.MsgRedeemTokensForShares{
		DelegatorAddress: addrDels[1].String(),
		Amount:           tokenizeShareResp.Amount,
	})
	require.NoError(t, err)

	// try undelegating
	_, err = msgServer.BeginRedelegate(ctx, &types.MsgBeginRedelegate{
		DelegatorAddress:    addrDels[0].String(),
		ValidatorSrcAddress: addrVals[0].String(),
		ValidatorDstAddress: addrVals[1].String(),
		Amount:              sdk.NewCoin(sdk.DefaultBondDenom, startTokens),
	})
	require.NoError(t, err)

	validator, _ = f.stakingKeeper.GetValidator(ctx, addrVals[0])
	require.Equal(t, validator.ValidatorBondShares, math.LegacyZeroDec())
}

func TestSendTokenizedSharesToValidatorBondedAndRedeem(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	ctx := f.sdkCtx

	addrDels := simtestutil.AddTestAddrs(f.bankKeeper, f.stakingKeeper, ctx, 2, f.stakingKeeper.TokensFromConsensusPower(ctx, 10000))

	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	startTokens := f.stakingKeeper.TokensFromConsensusPower(ctx, 10)

	bondDenom, err := f.stakingKeeper.BondDenom(ctx)
	require.NoError(t, err)
	notBondedPool := f.stakingKeeper.GetNotBondedPool(ctx)

	require.NoError(t, banktestutil.FundModuleAccount(ctx, f.bankKeeper, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(bondDenom, startTokens))))
	f.accountKeeper.SetModuleAccount(ctx, notBondedPool)

	// create a validator and a delegator to that validator
	validator := testutil.NewValidator(t, addrVals[0], PKs[0])
	validator.Status = types.Bonded
	f.stakingKeeper.SetValidator(ctx, validator)

	// set validator bond factor
	params, err := f.stakingKeeper.GetParams(ctx)
	require.NoError(t, err)
	params.ValidatorBondFactor = math.LegacyNewDec(1)
	f.stakingKeeper.SetParams(ctx, params)

	// convert to validator self-bond
	msgServer := keeper.NewMsgServerImpl(f.stakingKeeper)

	validator, _ = f.stakingKeeper.GetValidator(ctx, addrVals[0])
	err = delegateCoinsFromAccount(ctx, *f.stakingKeeper, addrDels[0], startTokens, validator)
	require.NoError(t, err)
	_, err = msgServer.ValidatorBond(ctx, &types.MsgValidatorBond{
		DelegatorAddress: addrDels[0].String(),
		ValidatorAddress: addrVals[0].String(),
	})
	require.NoError(t, err)

	// confirm that the delegation is marked as validator bond
	delegation1, err := f.stakingKeeper.GetDelegation(ctx, addrDels[0], addrVals[0])
	require.NoError(t, err)
	require.True(t, delegation1.ValidatorBond)

	// tokenize share for 2nd account delegation
	validator, _ = f.stakingKeeper.GetValidator(ctx, addrVals[0])
	err = delegateCoinsFromAccount(ctx, *f.stakingKeeper, addrDels[1], startTokens, validator)
	require.NoError(t, err)
	// confirm that the delegation was NOT marked as validator bond
	delegation2, err := f.stakingKeeper.GetDelegation(ctx, addrDels[1], addrVals[0])
	require.NoError(t, err)
	require.False(t, delegation2.ValidatorBond)

	// confirm that the ValidatorBond delegation cannot be tokenized
	_, err = msgServer.TokenizeShares(ctx, &types.MsgTokenizeShares{
		DelegatorAddress:    addrDels[0].String(),
		ValidatorAddress:    addrVals[0].String(),
		TokenizedShareOwner: addrDels[0].String(),
		Amount:              sdk.NewCoin(sdk.DefaultBondDenom, startTokens),
	})
	require.Error(t, err)
	require.EqualError(t, err, "validator bond delegation is not allowed to tokenize share")

	// tokenize share for 2nd account delegation
	tokenizeShareResp, err := msgServer.TokenizeShares(ctx, &types.MsgTokenizeShares{
		DelegatorAddress:    addrDels[1].String(),
		TokenizedShareOwner: addrDels[1].String(),
		ValidatorAddress:    addrVals[0].String(),
		Amount:              sdk.NewCoin(sdk.DefaultBondDenom, startTokens),
	})
	require.NoError(t, err)

	// transfer tokenized shares (as coins) to the delegator with validator bond
	err = f.bankKeeper.SendCoins(ctx, addrDels[1], addrDels[0], sdk.NewCoins(tokenizeShareResp.Amount))
	require.NoError(t, err)

	// confirm that the tokenized shares are now owned by the delegator with validator bond
	balanceSender := f.bankKeeper.GetBalance(ctx, addrDels[1], tokenizeShareResp.Amount.Denom)
	require.Equal(t, balanceSender.Amount, math.ZeroInt())
	balanceReceiver := f.bankKeeper.GetBalance(ctx, addrDels[0], tokenizeShareResp.Amount.Denom)
	require.Equal(t, balanceReceiver.Amount, tokenizeShareResp.Amount.Amount)

	// redeem shares and assert that the validator bond shares are increased
	validator, _ = f.stakingKeeper.GetValidator(ctx, addrVals[0])
	beforeRedeemShares := validator.ValidatorBondShares

	redeemResp, err := msgServer.RedeemTokensForShares(ctx, &types.MsgRedeemTokensForShares{
		DelegatorAddress: addrDels[0].String(),
		Amount:           tokenizeShareResp.Amount,
	})
	require.NoError(t, err)

	validator, _ = f.stakingKeeper.GetValidator(ctx, addrVals[0])
	afterRedeemShares := validator.ValidatorBondShares
	require.True(t, afterRedeemShares.GT(beforeRedeemShares))
	require.Equal(t, afterRedeemShares, beforeRedeemShares.Add(math.LegacyNewDecFromInt(tokenizeShareResp.Amount.Amount)))

	// undelegate the delegator with ValidatorBond and assert that the validator bond shares are decreased
	_, err = msgServer.Undelegate(ctx, &types.MsgUndelegate{
		DelegatorAddress: addrDels[0].String(),
		ValidatorAddress: addrVals[0].String(),
		Amount:           sdk.NewCoin(sdk.DefaultBondDenom, redeemResp.Amount.Amount),
	})
	require.NoError(t, err)

	validator, _ = f.stakingKeeper.GetValidator(ctx, addrVals[0])
	afterUndelegateShares := validator.ValidatorBondShares
	require.True(t, afterUndelegateShares.LT(afterRedeemShares))
	require.Equal(t, afterUndelegateShares, afterRedeemShares.Sub(math.LegacyNewDecFromInt(redeemResp.Amount.Amount)))
	require.Equal(t, afterUndelegateShares, beforeRedeemShares)
}
