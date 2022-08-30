package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestAllocateTokensToValidatorWithCommission(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addrs := simapp.AddTestAddrs(app, ctx, 3, sdk.NewInt(1234))
	valAddrs := simapp.ConvertAddrsToValAddrs(addrs)
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// create validator with 50% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	tstaking.CreateValidator(sdk.ValAddress(addrs[0]), valConsPk1, sdk.NewInt(100), true)
	val := app.StakingKeeper.Validator(ctx, valAddrs[0])

	// allocate tokens
	tokens := sdk.DecCoins{
		{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(10)},
	}
	app.DistrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// check commission
	expected := sdk.DecCoins{
		{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(5)},
	}
	require.Equal(t, expected, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, val.GetOperator()).Commission)

	// check current rewards
	require.Equal(t, expected, app.DistrKeeper.GetValidatorCurrentRewards(ctx, val.GetOperator()).Rewards)
}

func TestAllocateTokensToManyValidators(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addrs := simapp.AddTestAddrs(app, ctx, 2, sdk.NewInt(1234))
	valAddrs := simapp.ConvertAddrsToValAddrs(addrs)
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// create validator with 50% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	tstaking.CreateValidator(valAddrs[0], valConsPk1, sdk.NewInt(100), true)

	// create second validator with 0% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDec(0), sdk.NewDec(0), sdk.NewDec(0))
	tstaking.CreateValidator(valAddrs[1], valConsPk2, sdk.NewInt(100), true)

	abciValA := abci.Validator{
		Address: valConsPk1.Address(),
		Power:   100,
	}
	abciValB := abci.Validator{
		Address: valConsPk2.Address(),
		Power:   100,
	}

	// assert initial state: zero outstanding rewards, zero community pool, zero commission, zero current rewards
	require.True(t, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[0]).Rewards.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[1]).Rewards.IsZero())
	require.True(t, app.DistrKeeper.GetFeePool(ctx).CommunityPool.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[1]).Commission.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[0]).Rewards.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[1]).Rewards.IsZero())

	// allocate tokens as if both had voted and second was proposer
	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)))
	feeCollector := app.AccountKeeper.GetModuleAccount(ctx, types.FeeCollectorName)
	require.NotNil(t, feeCollector)

	// fund fee collector
	require.NoError(t, simapp.FundModuleAccount(app.BankKeeper, ctx, feeCollector.GetName(), fees))

	app.AccountKeeper.SetAccount(ctx, feeCollector)

	votes := []abci.VoteInfo{
		{
			Validator:       abciValA,
			SignedLastBlock: true,
		},
		{
			Validator:       abciValB,
			SignedLastBlock: true,
		},
	}
	app.DistrKeeper.AllocateTokens(ctx, 200_000, 200_000, valConsAddr2, votes)

	// 98 outstanding rewards (100 less 2 to community pool)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(465, 1)}}, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[0]).Rewards)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(515, 1)}}, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[1]).Rewards)
	// 2 community pool coins
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(2)}}, app.DistrKeeper.GetFeePool(ctx).CommunityPool)
	// 50% commission for first proposer, (0.5 * 93%) * 100 / 2 = 23.25
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(2325, 2)}}, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission)
	// zero commission for second proposer
	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[1]).Commission.IsZero())
	// just staking.proportional for first proposer less commission = (0.5 * 93%) * 100 / 2 = 23.25
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(2325, 2)}}, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[0]).Rewards)
	// proposer reward + staking.proportional for second proposer = (5 % + 0.5 * (93%)) * 100 = 51.5
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(515, 1)}}, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[1]).Rewards)
}

func TestAllocateTokensTruncation(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addrs := simapp.AddTestAddrs(app, ctx, 3, sdk.NewInt(1234))
	valAddrs := simapp.ConvertAddrsToValAddrs(addrs)
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// create validator with 10% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), sdk.NewDec(0))
	tstaking.CreateValidator(valAddrs[0], valConsPk1, sdk.NewInt(110), true)

	// create second validator with 10% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), sdk.NewDec(0))
	tstaking.CreateValidator(valAddrs[1], valConsPk2, sdk.NewInt(100), true)

	// create third validator with 10% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), sdk.NewDec(0))
	tstaking.CreateValidator(valAddrs[2], valConsPk3, sdk.NewInt(100), true)

	abciValA := abci.Validator{
		Address: valConsPk1.Address(),
		Power:   11,
	}
	abciValB := abci.Validator{
		Address: valConsPk2.Address(),
		Power:   10,
	}
	abciValС := abci.Validator{
		Address: valConsPk3.Address(),
		Power:   10,
	}

	// assert initial state: zero outstanding rewards, zero community pool, zero commission, zero current rewards
	require.True(t, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[0]).Rewards.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[1]).Rewards.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[1]).Rewards.IsZero())
	require.True(t, app.DistrKeeper.GetFeePool(ctx).CommunityPool.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[1]).Commission.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[0]).Rewards.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[1]).Rewards.IsZero())

	// allocate tokens as if both had voted and second was proposer
	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(634195840)))

	feeCollector := app.AccountKeeper.GetModuleAccount(ctx, types.FeeCollectorName)
	require.NotNil(t, feeCollector)

	require.NoError(t, simapp.FundModuleAccount(app.BankKeeper, ctx, feeCollector.GetName(), fees))

	app.AccountKeeper.SetAccount(ctx, feeCollector)

	votes := []abci.VoteInfo{
		{
			Validator:       abciValA,
			SignedLastBlock: true,
		},
		{
			Validator:       abciValB,
			SignedLastBlock: true,
		},
		{
			Validator:       abciValС,
			SignedLastBlock: true,
		},
	}
	app.DistrKeeper.AllocateTokens(ctx, 31, 31, sdk.ConsAddress(valConsPk2.Address()), votes)

	require.True(t, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[0]).Rewards.IsValid())
	require.True(t, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[1]).Rewards.IsValid())
	require.True(t, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[2]).Rewards.IsValid())
}

// func TestAllocateTokensToManyValidatorsWithBlacklist(t *testing.T) {

// 	startingAcctBalance := sdk.NewInt(5_000_000_000)
// 	amtToSelfBond := sdk.NewInt(100_000_000)
// 	sumPreviousPrecommitPower, totalPreviousPower := int64(200), int64(200)
// 	feesAmt := sdk.NewInt(100_000_000)

// 	app := simapp.Setup(false)
// 	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

// 	addrs := simapp.AddTestAddrs(app, ctx, 2, startingAcctBalance)

// 	// add the second delegator to the blacklist
// 	blacklistedDelAddr := addrs[1].String()
// 	params := app.DistrKeeper.GetParams(ctx)
// 	params.NoRewardsDelegatorAddresses = append(params.NoRewardsDelegatorAddresses, blacklistedDelAddr)
// 	app.DistrKeeper.SetParams(ctx, params)
// 	fmt.Println("added del to blacklist: ", blacklistedDelAddr)

// 	valAddrs := simapp.ConvertAddrsToValAddrs(addrs)
// 	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

// 	// create validator with 0%
// 	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDec(0), sdk.NewDec(0), sdk.NewDec(0))
// 	tstaking.CreateValidator(valAddrs[0], valConsPk1, amtToSelfBond, true)
// 	// valObj := app.StakingKeeper.Validator(ctx, valAddrs[0])

// 	// fmt.Println("tokens on val:", valObj.GetTokens(), "power red", sdk.DefaultPowerReduction)
// 	// fmt.Println("tokens on val:", valObj.GetTokens(), "power red", sdk.DefaultPowerReduction)
// 	// fmt.Println("shares on val:", valObj.GetDelegatorShares())

// 	// valTotPower := sdk.TokensToConsensusPower(valObj.GetTokens(), sdk.DefaultPowerReduction)
// 	// fmt.Println("power on val:", valTotPower)

// 	// MOOSE test getting power share
// 	// taintedVals := app.DistrKeeper.GetTaintedValidators(ctx)
// 	// fmt.Println("tainted val: ", taintedVals) //, taintedValsBlacklistedPowerShare, taintedVals)

// 	// valsBlacklistedPower, taintedValsBlacklistedPowerShare, taintedVals := app.DistrKeeper.GetValsBlacklistedPowerShare(ctx)
// 	// fmt.Println(valsBlacklistedPower, taintedValsBlacklistedPowerShare, taintedVals)

// 	// create second validator with 0% commission
// 	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDec(0), sdk.NewDec(0), sdk.NewDec(0))
// 	tstaking.CreateValidator(valAddrs[1], valConsPk2, amtToSelfBond, true)

// 	abciValA := abci.Validator{
// 		Address: valConsPk1.Address(),
// 		Power:   100,
// 	}
// 	abciValB := abci.Validator{
// 		Address: valConsPk2.Address(),
// 		Power:   100,
// 	}

// 	// assert initial state: zero outstanding rewards, zero community pool, zero commission, zero current rewards
// 	require.True(t, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[0]).Rewards.IsZero())
// 	require.True(t, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[1]).Rewards.IsZero())
// 	require.True(t, app.DistrKeeper.GetFeePool(ctx).CommunityPool.IsZero())
// 	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission.IsZero())
// 	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[1]).Commission.IsZero())
// 	require.True(t, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[0]).Rewards.IsZero())
// 	require.True(t, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[1]).Rewards.IsZero())

// 	// allocate tokens as if both had voted and second was proposer
// 	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, feesAmt))
// 	feeCollector := app.AccountKeeper.GetModuleAccount(ctx, types.FeeCollectorName)
// 	require.NotNil(t, feeCollector)

// 	// fund fee collector
// 	require.NoError(t, simapp.FundModuleAccount(app.BankKeeper, ctx, feeCollector.GetName(), fees))

// 	app.AccountKeeper.SetAccount(ctx, feeCollector)

// 	votes := []abci.VoteInfo{
// 		{
// 			Validator:       abciValA,
// 			SignedLastBlock: true,
// 		},
// 		{
// 			Validator:       abciValB,
// 			SignedLastBlock: true,
// 		},
// 	}

// 	app.DistrKeeper.AllocateTokens(ctx, sumPreviousPrecommitPower, totalPreviousPower, valConsAddr2, votes)
// 	// valsBlacklistedPower, taintedValsBlacklistedPowerShare, _ = app.DistrKeeper.GetValsBlacklistedPowerShare(ctx)
// 	// totalWhitelistedPowerShare := sdk.NewDec(1).Sub(valsBlacklistedPower.Quo(sdk.NewDec(200)))
// 	// require.Equal(t, totalWhitelistedPowerShare, 0)
// 	// fmt.Println("totalWhitelistedPowerShare: ", totalWhitelistedPowerShare)

// 	// =================== check staking rewards =========================

// 	// just staking.proportional for first proposer less commission = (0.5 * 93%) * 100 / 2 = 23.25
// 	// require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(2325, 2)}}, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[0]).Rewards)
// 	// proposer reward + staking.proportional for second proposer = (5 % + 0.5 * (93%)) * 100 = 51.5
// 	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(515_000_000, 1)}}, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[1]).Rewards)

// 	// 50% commission for first proposer, (0.5 * 93%) * 100 / 2 = 23.25
// 	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(2_325_000_000, 2)}}, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission)
// 	// zero commission for second proposer
// 	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[1]).Commission.IsZero())

// 	// =================== check non-staking rewards =========================
// 	// 98 outstanding rewards (100 less 2 to community pool)
// 	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(465_000_000, 1)}}, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[0]).Rewards)
// 	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(515_000_000, 1)}}, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[1]).Rewards)
// 	// 2 community pool coins
// 	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(2_000_000)}}, app.DistrKeeper.GetFeePool(ctx).CommunityPool)
// }

// test the case in which there are two validators and two delegators, one of which is blacklisted
//     all rewards should flow to the validator and delegator who is NOT blacklisted
func TestAllocateTokensToManyValidatorsWithBlacklist(t *testing.T) {
	// params
	startingAcctBalance := sdk.NewInt(5_000_000_000)
	amtToSelfBond := sdk.NewInt(100_000_000)
	sumPreviousPrecommitPower, totalPreviousPower := int64(200), int64(200)
	feesAmt := sdk.NewInt(100_000_000)

	// chain setup
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	addrs := simapp.AddTestAddrs(app, ctx, 2, startingAcctBalance)

	valAddrs := simapp.ConvertAddrsToValAddrs(addrs)
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// validator setup
	// create validator with 0%
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDec(0), sdk.NewDec(0), sdk.NewDec(0))
	tstaking.CreateValidator(valAddrs[0], valConsPk1, amtToSelfBond, true)

	// create second validator with 0% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDec(0), sdk.NewDec(0), sdk.NewDec(0))
	tstaking.CreateValidator(valAddrs[1], valConsPk2, amtToSelfBond, true)

	abciValA := abci.Validator{
		Address: valConsPk1.Address(),
		Power:   100,
	}
	abciValB := abci.Validator{
		Address: valConsPk2.Address(),
		Power:   100,
	}

	// assert initial state: zero outstanding rewards, zero community pool, zero commission, zero current rewards
	require.True(t, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[0]).Rewards.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[1]).Rewards.IsZero())
	require.True(t, app.DistrKeeper.GetFeePool(ctx).CommunityPool.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[1]).Commission.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[0]).Rewards.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[1]).Rewards.IsZero())

	// allocate tokens as if both had voted and second was proposer
	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, feesAmt))
	feeCollector := app.AccountKeeper.GetModuleAccount(ctx, types.FeeCollectorName)
	require.NotNil(t, feeCollector)

	// fund fee collector
	require.NoError(t, simapp.FundModuleAccount(app.BankKeeper, ctx, feeCollector.GetName(), fees))

	app.AccountKeeper.SetAccount(ctx, feeCollector)

	votes := []abci.VoteInfo{
		{
			Validator:       abciValA,
			SignedLastBlock: true,
		},
		{
			Validator:       abciValB,
			SignedLastBlock: true,
		},
	}

	// add the second delegator to the blacklist
	blacklistedDelAddr := addrs[1].String()
	params := app.DistrKeeper.GetParams(ctx)
	params.NoRewardsDelegatorAddresses = append(params.NoRewardsDelegatorAddresses, blacklistedDelAddr)
	app.DistrKeeper.SetParams(ctx, params)
	fmt.Println("added del to blacklist: ", blacklistedDelAddr)

	app.DistrKeeper.AllocateTokens(ctx, sumPreviousPrecommitPower, totalPreviousPower, valConsAddr2, votes)

	// just staking.proportional for first proposer = (93%) * 100_000_000 = 93_000_000
	//       note 93% is the share that goes to staking rewards
	//       since the only delegator to the second validator is blacklisted, all rewards go to the delegators to the first validator!
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(93_000_000)}}, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[0]).Rewards)
	// proposer reward + staking.proportional for second proposer = (5 % commission) * 100 = 5_000_000
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(5_000_000)}}, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[1]).Rewards)

	// zero commission for either validator (including proposer)
	//     0% commission: (0% * 93%) * 100 / 2 = 0
	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[1]).Commission.IsZero())

	// =================== check non-staking rewards =========================
	// 98 outstanding rewards (100 less 2 to community pool)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(93_000_000)}}, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[0]).Rewards)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(5_000_000)}}, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[1]).Rewards)
	// 2% to community pool
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(2_000_000)}}, app.DistrKeeper.GetFeePool(ctx).CommunityPool)
}

// test the case in which a blacklisted and a whitelisted delegator delegate to val1
//     - val1 should NOT accrue rewards for the blacklisted delegator's share
//     - all other rewards should upscale by the blacklisted share
// 			- val1 should only accrue rewards for the whitelisted delegator's share, upscaled
// 			- val2 should accrue rewards for its delegator, upscaled
func TestAllocateTokensToManyValidatorsWithBlacklist_AddlBlacklistedDelegation(t *testing.T) {
	// params
	startingAcctBalance := sdk.NewInt(5_000_000_000)
	amtToSelfBond := sdk.NewInt(100_000_000)
	amtToMixedDelegate := sdk.NewInt(10_000_000)

	sumPreviousPrecommitPower, totalPreviousPower := int64(200), int64(200)
	feesAmt := sdk.NewInt(100_000_000)

	// chain setup
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	addrs := simapp.AddTestAddrs(app, ctx, 2, startingAcctBalance)

	valAddrs := simapp.ConvertAddrsToValAddrs(addrs)
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// validator setup
	// create validator with 0%
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDec(0), sdk.NewDec(0), sdk.NewDec(0))
	tstaking.CreateValidator(valAddrs[0], valConsPk1, amtToSelfBond, true)

	// create second validator with 0% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDec(0), sdk.NewDec(0), sdk.NewDec(0))
	tstaking.CreateValidator(valAddrs[1], valConsPk2, amtToSelfBond, true)

	abciValA := abci.Validator{
		Address: valConsPk1.Address(),
		Power:   100,
	}
	abciValB := abci.Validator{
		Address: valConsPk2.Address(),
		Power:   100,
	}

	// assert initial state: zero outstanding rewards, zero community pool, zero commission, zero current rewards
	require.True(t, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[0]).Rewards.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[1]).Rewards.IsZero())
	require.True(t, app.DistrKeeper.GetFeePool(ctx).CommunityPool.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[1]).Commission.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[0]).Rewards.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[1]).Rewards.IsZero())

	// allocate tokens as if both had voted and second was proposer
	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, feesAmt))
	feeCollector := app.AccountKeeper.GetModuleAccount(ctx, types.FeeCollectorName)
	require.NotNil(t, feeCollector)

	// fund fee collector
	require.NoError(t, simapp.FundModuleAccount(app.BankKeeper, ctx, feeCollector.GetName(), fees))

	app.AccountKeeper.SetAccount(ctx, feeCollector)

	votes := []abci.VoteInfo{
		{
			Validator:       abciValA,
			SignedLastBlock: true,
		},
		{
			Validator:       abciValB,
			SignedLastBlock: true,
		},
	}

	// add the second delegator to the blacklist
	blacklistedDelAddr := addrs[1].String()
	params := app.DistrKeeper.GetParams(ctx)
	params.NoRewardsDelegatorAddresses = append(params.NoRewardsDelegatorAddresses, blacklistedDelAddr)
	app.DistrKeeper.SetParams(ctx, params)
	fmt.Println("added del to blacklist: ", blacklistedDelAddr)

	// delegate some tokens from delegator 1 (blacklisted) to validator 1
	tstaking.Ctx = ctx
	delegatorAddr := sdk.AccAddress(valAddrs[1])
	validatorAddr := valAddrs[0]
	tstaking.Delegate(delegatorAddr, validatorAddr, amtToMixedDelegate)
	app.StakingKeeper.Delegation(ctx, delegatorAddr, validatorAddr)

	app.DistrKeeper.AllocateTokens(ctx, sumPreviousPrecommitPower, totalPreviousPower, valConsAddr2, votes)

	// just staking.proportional for first proposer = (93%) * 100_000_000 = 93_000_000
	//       note 93% is the share that goes to staking rewards
	//       since the only delegator to the second validator is blacklisted, all rewards go to the delegators to the first validator!
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(93_000_000)}}, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[0]).Rewards)
	// proposer reward + staking.proportional for second proposer = (5 % commission) * 100 = 5_000_000
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(5_000_000)}}, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[1]).Rewards)

	// zero commission for either validator (including proposer)
	//     0% commission: (0% * 93%) * 100 / 2 = 0
	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[1]).Commission.IsZero())

	// =================== check non-staking rewards =========================
	// 98 outstanding rewards (100 less 2 to community pool)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(93_000_000)}}, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[0]).Rewards)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(5_000_000)}}, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[1]).Rewards)
	// 2% to community pool
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(2_000_000)}}, app.DistrKeeper.GetFeePool(ctx).CommunityPool)
}
func TestAllocateTokensToManyValidatorsWithBlacklist_MixedDelegations(t *testing.T) {

	// params
	startingAcctBalance := sdk.NewInt(5_000_000_000)
	amtToSelfBond := sdk.NewInt(100_000_000)
	amtToMixedDelegate := sdk.NewInt(10_000_000)
	sumPreviousPrecommitPower, totalPreviousPower := int64(200), int64(200)
	feesAmt := sdk.NewInt(100_000_000)

	// chain setup
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	addrs := simapp.AddTestAddrs(app, ctx, 2, startingAcctBalance)

	// add the second delegator to the blacklist
	blacklistedDelAddr := addrs[1].String()
	params := app.DistrKeeper.GetParams(ctx)
	params.NoRewardsDelegatorAddresses = append(params.NoRewardsDelegatorAddresses, blacklistedDelAddr)
	app.DistrKeeper.SetParams(ctx, params)
	fmt.Println("added del to blacklist: ", blacklistedDelAddr)

	valAddrs := simapp.ConvertAddrsToValAddrs(addrs)
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// validator setup
	// create validator with 0%
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDec(0), sdk.NewDec(0), sdk.NewDec(0))
	tstaking.CreateValidator(valAddrs[0], valConsPk1, amtToSelfBond, true)

	// create second validator with 0% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDec(0), sdk.NewDec(0), sdk.NewDec(0))
	tstaking.CreateValidator(valAddrs[1], valConsPk2, amtToSelfBond, true)

	abciValA := abci.Validator{
		Address: valConsPk1.Address(),
		Power:   100,
	}
	abciValB := abci.Validator{
		Address: valConsPk2.Address(),
		Power:   100,
	}

	// delegate some tokens from delegator 1 (blacklisted) to validator 1
	tstaking.Ctx = ctx
	delegatorAddr := sdk.AccAddress(valAddrs[1])
	validatorAddr := valAddrs[0]
	tstaking.Delegate(delegatorAddr, validatorAddr, amtToMixedDelegate)
	app.StakingKeeper.Delegation(ctx, delegatorAddr, validatorAddr)

	// assert initial state: zero outstanding rewards, zero community pool, zero commission, zero current rewards
	require.True(t, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[0]).Rewards.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[1]).Rewards.IsZero())
	require.True(t, app.DistrKeeper.GetFeePool(ctx).CommunityPool.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[1]).Commission.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[0]).Rewards.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[1]).Rewards.IsZero())

	// allocate tokens as if both had voted and second was proposer
	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, feesAmt))
	feeCollector := app.AccountKeeper.GetModuleAccount(ctx, types.FeeCollectorName)
	require.NotNil(t, feeCollector)

	// fund fee collector
	require.NoError(t, simapp.FundModuleAccount(app.BankKeeper, ctx, feeCollector.GetName(), fees))

	app.AccountKeeper.SetAccount(ctx, feeCollector)

	votes := []abci.VoteInfo{
		{
			Validator:       abciValA,
			SignedLastBlock: true,
		},
		{
			Validator:       abciValB,
			SignedLastBlock: true,
		},
	}

	app.DistrKeeper.AllocateTokens(ctx, sumPreviousPrecommitPower, totalPreviousPower, valConsAddr2, votes)

	// just staking.proportional for first proposer = (93%) * 100_000_000 = 93_000_000
	// nbote 93% is the share that goes to staking rewards
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(93_000_000)}}, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[0]).Rewards)
	// proposer reward only, staking is blacklisted = (0.05 * 100_000_000)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(5_000_000)}}, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[1]).Rewards)

	// zero commission for either validator (including proposer)
	//     0% commission: (0% * 93%) * 100 / 2 = 0
	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[1]).Commission.IsZero())

	// =================== check non-staking rewards =========================
	// 98 outstanding rewards (100 less 2 to community pool)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(93_000_000)}}, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[0]).Rewards)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(5_000_000)}}, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[1]).Rewards)
	// 2% to community pool
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(2_000_000)}}, app.DistrKeeper.GetFeePool(ctx).CommunityPool)
}

// func TestAllocateTokensToManyValidatorsWithBlacklist_MixedDelegations(t *testing.T) {

// 	// params
// 	startingAcctBalance := sdk.NewInt(5_000_000_000)
// 	amtToSelfBond := sdk.NewInt(100_000_000)
// 	// amtToMixedDelegate := sdk.NewInt(10_000_000)
// 	sumPreviousPrecommitPower, totalPreviousPower := int64(200), int64(200)
// 	feesAmt := sdk.NewInt(100_000_000)

// 	// chain setup
// 	app := simapp.Setup(false)
// 	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
// 	addrs := simapp.AddTestAddrs(app, ctx, 2, startingAcctBalance)

// 	// add the second delegator to the blacklist
// 	blacklistedDelAddr := addrs[1].String()
// 	params := app.DistrKeeper.GetParams(ctx)
// 	params.NoRewardsDelegatorAddresses = append(params.NoRewardsDelegatorAddresses, blacklistedDelAddr)
// 	app.DistrKeeper.SetParams(ctx, params)
// 	fmt.Println("added del to blacklist: ", blacklistedDelAddr)

// 	valAddrs := simapp.ConvertAddrsToValAddrs(addrs)
// 	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

// 	// validator setup
// 	// create validator with 0%
// 	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDec(0), sdk.NewDec(0), sdk.NewDec(0))
// 	tstaking.CreateValidator(valAddrs[0], valConsPk1, amtToSelfBond, true)

// 	// create second validator with 0% commission
// 	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDec(0), sdk.NewDec(0), sdk.NewDec(0))
// 	tstaking.CreateValidator(valAddrs[1], valConsPk2, amtToSelfBond, true)

// 	abciValA := abci.Validator{
// 		Address: valConsPk1.Address(),
// 		Power:   100,
// 	}
// 	abciValB := abci.Validator{
// 		Address: valConsPk2.Address(),
// 		Power:   100,
// 	}

// 	// delegate some tokens from delegator 0 (whitelisted) to validator 1
// 	tstaking.Ctx = ctx
// 	delegatorAddr := sdk.AccAddress(valAddrs[1])
// 	validatorAddr := valAddrs[1]
// 	tstaking.Delegate(delegatorAddr, validatorAddr, amtToMixedDelegate)
// 	app.StakingKeeper.Delegation(ctx, delegatorAddr, validatorAddr)

// 	// assert initial state: zero outstanding rewards, zero community pool, zero commission, zero current rewards
// 	require.True(t, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[0]).Rewards.IsZero())
// 	require.True(t, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[1]).Rewards.IsZero())
// 	require.True(t, app.DistrKeeper.GetFeePool(ctx).CommunityPool.IsZero())
// 	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission.IsZero())
// 	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[1]).Commission.IsZero())
// 	require.True(t, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[0]).Rewards.IsZero())
// 	require.True(t, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[1]).Rewards.IsZero())

// 	// allocate tokens as if both had voted and second was proposer
// 	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, feesAmt))
// 	feeCollector := app.AccountKeeper.GetModuleAccount(ctx, types.FeeCollectorName)
// 	require.NotNil(t, feeCollector)

// 	// fund fee collector
// 	require.NoError(t, simapp.FundModuleAccount(app.BankKeeper, ctx, feeCollector.GetName(), fees))

// 	app.AccountKeeper.SetAccount(ctx, feeCollector)

// 	votes := []abci.VoteInfo{
// 		{
// 			Validator:       abciValA,
// 			SignedLastBlock: true,
// 		},
// 		{
// 			Validator:       abciValB,
// 			SignedLastBlock: true,
// 		},
// 	}

// 	app.DistrKeeper.AllocateTokens(ctx, sumPreviousPrecommitPower, totalPreviousPower, valConsAddr2, votes)

// 	// just staking.proportional for first proposer = (93%) * 100_000_000 = 93_000_000
// 	// nbote 93% is the share that goes to staking rewards
// 	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(93_000_000)}}, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[0]).Rewards)
// 	// proposer reward only, staking is blacklisted = (0.05 * 100_000_000)
// 	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(5_000_000)}}, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[1]).Rewards)

// 	// zero commission for either validator (including proposer)
// 	//     0% commission: (0% * 93%) * 100 / 2 = 0
// 	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission.IsZero())
// 	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[1]).Commission.IsZero())

// 	// =================== check non-staking rewards =========================
// 	// 98 outstanding rewards (100 less 2 to community pool)
// 	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(93_000_000)}}, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[0]).Rewards)
// 	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(5_000_000)}}, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[1]).Rewards)
// 	// 2% to community pool
// 	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(2_000_000)}}, app.DistrKeeper.GetFeePool(ctx).CommunityPool)
// }
