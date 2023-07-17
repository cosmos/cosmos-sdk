package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
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

	require.NoError(t, banktestutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(bondDenom, startTokens))))
	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	// create a validator and a delegator to that validator
	validator := testutil.NewValidator(t, addrVals[0], PKs[0])

	validator, issuedShares := validator.AddTokensFromDel(startTokens)
	require.Equal(t, startTokens, issuedShares.RoundInt())

	validator = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validator, true)
	require.True(math.IntEq(t, startTokens, validator.BondedTokens()))
	require.True(t, validator.IsBonded())

	delegation := types.NewDelegation(addrDels[0], addrVals[0], issuedShares, false)
	app.StakingKeeper.SetDelegation(ctx, delegation)

	maxEntries := app.StakingKeeper.MaxEntries(ctx)

	oldBonded := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	oldNotBonded := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// should all pass
	var completionTime time.Time
	for i := int64(0); i < int64(maxEntries); i++ {
		var err error
		ctx = ctx.WithBlockHeight(i)
		completionTime, err = app.StakingKeeper.Undelegate(ctx, addrDels[0], addrVals[0], math.LegacyNewDec(1))
		require.NoError(t, err)
	}

	newBonded := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	require.True(math.IntEq(t, newBonded, oldBonded.SubRaw(int64(maxEntries))))
	require.True(math.IntEq(t, newNotBonded, oldNotBonded.AddRaw(int64(maxEntries))))

	oldBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	oldNotBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// an additional unbond should fail due to max entries
	_, err := app.StakingKeeper.Undelegate(ctx, addrDels[0], addrVals[0], math.LegacyNewDec(1))
	require.Error(t, err)

	newBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	require.True(math.IntEq(t, newBonded, oldBonded))
	require.True(math.IntEq(t, newNotBonded, oldNotBonded))

	// mature unbonding delegations
	ctx = ctx.WithBlockTime(completionTime)
	_, err = app.StakingKeeper.CompleteUnbonding(ctx, addrDels[0], addrVals[0])
	require.NoError(t, err)

	newBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	require.True(math.IntEq(t, newBonded, oldBonded))
	require.True(math.IntEq(t, newNotBonded, oldNotBonded.SubRaw(int64(maxEntries))))

	oldNotBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// unbonding  should work again
	_, err = app.StakingKeeper.Undelegate(ctx, addrDels[0], addrVals[0], math.LegacyNewDec(1))
	require.NoError(t, err)

	newBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	require.True(math.IntEq(t, newBonded, oldBonded.SubRaw(1)))
	require.True(math.IntEq(t, newNotBonded, oldNotBonded.AddRaw(1)))
}

func TestValidatorBondUndelegate(t *testing.T) {
	_, app, ctx := createTestInput(t)

	addrDels := simtestutil.AddTestAddrs(app.BankKeeper, app.StakingKeeper,
		ctx, 2, app.StakingKeeper.TokensFromConsensusPower(ctx, 10000))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	startTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)

	bondDenom := app.StakingKeeper.BondDenom(ctx)
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)

	require.NoError(t, banktestutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(bondDenom, startTokens))))
	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	// create a validator and a delegator to that validator
	validator := testutil.NewValidator(t, addrVals[0], PKs[0])
	validator.Status = types.Bonded
	app.StakingKeeper.SetValidator(ctx, validator)

	// set validator bond factor
	params := app.StakingKeeper.GetParams(ctx)
	params.ValidatorBondFactor = sdk.NewDec(1)
	app.StakingKeeper.SetParams(ctx, params)

	// convert to validator self-bond
	msgServer := keeper.NewMsgServerImpl(app.StakingKeeper)

	validator, _ = app.StakingKeeper.GetValidator(ctx, addrVals[0])
	err := delegateCoinsFromAccount(ctx, app, addrDels[0], startTokens, validator)
	require.NoError(t, err)
	_, err = msgServer.ValidatorBond(sdk.WrapSDKContext(ctx), &types.MsgValidatorBond{
		DelegatorAddress: addrDels[0].String(),
		ValidatorAddress: addrVals[0].String(),
	})
	require.NoError(t, err)

	// tokenize share for 2nd account delegation
	validator, _ = app.StakingKeeper.GetValidator(ctx, addrVals[0])
	err = delegateCoinsFromAccount(ctx, app, addrDels[1], startTokens, validator)
	require.NoError(t, err)
	tokenizeShareResp, err := msgServer.TokenizeShares(sdk.WrapSDKContext(ctx), &types.MsgTokenizeShares{
		DelegatorAddress:    addrDels[1].String(),
		ValidatorAddress:    addrVals[0].String(),
		Amount:              sdk.NewCoin(sdk.DefaultBondDenom, startTokens),
		TokenizedShareOwner: addrDels[0].String(),
	})
	require.NoError(t, err)

	// try undelegating
	_, err = msgServer.Undelegate(sdk.WrapSDKContext(ctx), &types.MsgUndelegate{
		DelegatorAddress: addrDels[0].String(),
		ValidatorAddress: addrVals[0].String(),
		Amount:           sdk.NewCoin(sdk.DefaultBondDenom, startTokens),
	})
	require.Error(t, err)

	// redeem full amount on 2nd account and try undelegation
	validator, _ = app.StakingKeeper.GetValidator(ctx, addrVals[0])
	err = delegateCoinsFromAccount(ctx, app, addrDels[1], startTokens, validator)
	require.NoError(t, err)
	_, err = msgServer.RedeemTokensForShares(sdk.WrapSDKContext(ctx), &types.MsgRedeemTokensForShares{
		DelegatorAddress: addrDels[1].String(),
		Amount:           tokenizeShareResp.Amount,
	})
	require.NoError(t, err)

	// try undelegating
	_, err = msgServer.Undelegate(sdk.WrapSDKContext(ctx), &types.MsgUndelegate{
		DelegatorAddress: addrDels[0].String(),
		ValidatorAddress: addrVals[0].String(),
		Amount:           sdk.NewCoin(sdk.DefaultBondDenom, startTokens),
	})
	require.NoError(t, err)

	validator, _ = app.StakingKeeper.GetValidator(ctx, addrVals[0])
	require.Equal(t, validator.ValidatorBondShares, sdk.ZeroDec())
}

func TestValidatorBondRedelegate(t *testing.T) {
	_, app, ctx := createTestInput(t)

	addrDels := simtestutil.AddTestAddrs(app.BankKeeper, app.StakingKeeper,
		ctx, 2, app.StakingKeeper.TokensFromConsensusPower(ctx, 10000))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	startTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)

	bondDenom := app.StakingKeeper.BondDenom(ctx)
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)

	startPoolToken := sdk.NewCoins(sdk.NewCoin(bondDenom, startTokens.Mul(sdk.NewInt(2))))
	require.NoError(t, banktestutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), startPoolToken))
	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	// create a validator and a delegator to that validator
	validator := testutil.NewValidator(t, addrVals[0], PKs[0])
	validator.Status = types.Bonded
	app.StakingKeeper.SetValidator(ctx, validator)
	validator2 := testutil.NewValidator(t, addrVals[1], PKs[1])
	validator.Status = types.Bonded
	app.StakingKeeper.SetValidator(ctx, validator2)

	// set validator bond factor
	params := app.StakingKeeper.GetParams(ctx)
	params.ValidatorBondFactor = sdk.NewDec(1)
	app.StakingKeeper.SetParams(ctx, params)

	// set total liquid stake
	app.StakingKeeper.SetTotalLiquidStakedTokens(ctx, sdk.NewInt(100))

	// delegate to each validator
	validator, _ = app.StakingKeeper.GetValidator(ctx, addrVals[0])
	err := delegateCoinsFromAccount(ctx, app, addrDels[0], startTokens, validator)
	require.NoError(t, err)

	validator2, _ = app.StakingKeeper.GetValidator(ctx, addrVals[1])
	err = delegateCoinsFromAccount(ctx, app, addrDels[1], startTokens, validator2)
	require.NoError(t, err)

	// convert to validator self-bond
	msgServer := keeper.NewMsgServerImpl(app.StakingKeeper)

	_, err = msgServer.ValidatorBond(sdk.WrapSDKContext(ctx), &types.MsgValidatorBond{
		DelegatorAddress: addrDels[0].String(),
		ValidatorAddress: addrVals[0].String(),
	})
	require.NoError(t, err)

	// tokenize share for 2nd account delegation
	validator, _ = app.StakingKeeper.GetValidator(ctx, addrVals[0])
	err = delegateCoinsFromAccount(ctx, app, addrDels[1], startTokens, validator)
	require.NoError(t, err)
	tokenizeShareResp, err := msgServer.TokenizeShares(sdk.WrapSDKContext(ctx), &types.MsgTokenizeShares{
		DelegatorAddress:    addrDels[1].String(),
		ValidatorAddress:    addrVals[0].String(),
		Amount:              sdk.NewCoin(sdk.DefaultBondDenom, startTokens),
		TokenizedShareOwner: addrDels[0].String(),
	})
	require.NoError(t, err)

	// try undelegating
	_, err = msgServer.BeginRedelegate(sdk.WrapSDKContext(ctx), &types.MsgBeginRedelegate{
		DelegatorAddress:    addrDels[0].String(),
		ValidatorSrcAddress: addrVals[0].String(),
		ValidatorDstAddress: addrVals[1].String(),
		Amount:              sdk.NewCoin(sdk.DefaultBondDenom, startTokens),
	})
	require.Error(t, err)

	// redeem full amount on 2nd account and try undelegation
	validator, _ = app.StakingKeeper.GetValidator(ctx, addrVals[0])
	err = delegateCoinsFromAccount(ctx, app, addrDels[1], startTokens, validator)
	require.NoError(t, err)
	_, err = msgServer.RedeemTokensForShares(sdk.WrapSDKContext(ctx), &types.MsgRedeemTokensForShares{
		DelegatorAddress: addrDels[1].String(),
		Amount:           tokenizeShareResp.Amount,
	})
	require.NoError(t, err)

	// try undelegating
	_, err = msgServer.BeginRedelegate(sdk.WrapSDKContext(ctx), &types.MsgBeginRedelegate{
		DelegatorAddress:    addrDels[0].String(),
		ValidatorSrcAddress: addrVals[0].String(),
		ValidatorDstAddress: addrVals[1].String(),
		Amount:              sdk.NewCoin(sdk.DefaultBondDenom, startTokens),
	})
	require.NoError(t, err)

	validator, _ = app.StakingKeeper.GetValidator(ctx, addrVals[0])
	require.Equal(t, validator.ValidatorBondShares, sdk.ZeroDec())
}
