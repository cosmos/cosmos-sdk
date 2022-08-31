package keeper_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
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

// STRIDE TESTS

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

	// Store BlacklistedPower
	validatorBlacklistedPowers := []distributiontypes.ValidatorBlacklistedPower{}
	validatorBlacklistedPowers = append(validatorBlacklistedPowers, distributiontypes.ValidatorBlacklistedPower{ValidatorAddress: valAddrs[0].String(), BlacklistedPowerShare: sdk.NewDec(0)})
	validatorBlacklistedPowers = append(validatorBlacklistedPowers, distributiontypes.ValidatorBlacklistedPower{ValidatorAddress: valAddrs[1].String(), BlacklistedPowerShare: sdk.NewDec(1)})
	blacklistedPower := distributiontypes.BlacklistedPower{
		TotalBlacklistedPowerShare: sdk.NewDec(100),
		ValidatorBlacklistedPowers: validatorBlacklistedPowers,
	}
	// set blacklisted power for n-3
	height := strconv.FormatInt(ctx.BlockHeight()-3, 10)
	fmt.Println("height: ", height)
	fmt.Println("blacklistedPower: ", blacklistedPower)
	app.DistrKeeper.SetBlacklistedPower(ctx, height, blacklistedPower)

	power, found := app.DistrKeeper.GetBlacklistedPower(ctx, height)
	fmt.Println("power: ", power)
	fmt.Println("found: ", found)
	require.True(t, found)

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

	sumPreviousPrecommitPower, totalPreviousPower := int64(210), int64(210)
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

	// add the second delegator to the blacklist
	blacklistedDelAddr := addrs[1].String()
	params := app.DistrKeeper.GetParams(ctx)
	params.NoRewardsDelegatorAddresses = append(params.NoRewardsDelegatorAddresses, blacklistedDelAddr)
	app.DistrKeeper.SetParams(ctx, params)
	fmt.Println("added del to blacklist: ", blacklistedDelAddr)

	// delegate some tokens from delegator 1 (blacklisted) to validator 0 (whitelisted)
	// block n-1
	tstaking.Ctx = ctx
	delegatorAddr := sdk.AccAddress(valAddrs[1])
	validatorAddr := valAddrs[0]
	tstaking.Delegate(delegatorAddr, validatorAddr, amtToMixedDelegate)
	app.StakingKeeper.Delegation(ctx, delegatorAddr, validatorAddr)
	// val0 has 10_000_000 blacklisted, 100_000_000 not blacklisted tokens
	val0BlacklistedPowerShare := sdk.NewDec(10).Quo(sdk.NewDec(110))
	fmt.Println("val0BlacklistedPowerShare: ", val0BlacklistedPowerShare)
	// update votes
	votes := []abci.VoteInfo{
		{
			Validator: abci.Validator{
				Address: valConsPk1.Address(),
				Power:   110,
			},
			SignedLastBlock: true,
		},
		{
			Validator: abci.Validator{
				Address: valConsPk2.Address(),
				Power:   100,
			},
			SignedLastBlock: true,
		},
	}

	// Store BlacklistedPower
	// block n-3
	validatorBlacklistedPowers := []distributiontypes.ValidatorBlacklistedPower{}
	validatorBlacklistedPowers = append(validatorBlacklistedPowers, distributiontypes.ValidatorBlacklistedPower{ValidatorAddress: valAddrs[0].String(), BlacklistedPowerShare: val0BlacklistedPowerShare})
	validatorBlacklistedPowers = append(validatorBlacklistedPowers, distributiontypes.ValidatorBlacklistedPower{ValidatorAddress: valAddrs[1].String(), BlacklistedPowerShare: sdk.NewDec(1)})
	blacklistedPower := distributiontypes.BlacklistedPower{
		TotalBlacklistedPowerShare: sdk.NewDec(110),
		ValidatorBlacklistedPowers: validatorBlacklistedPowers,
	}
	height := strconv.FormatInt(ctx.BlockHeight()-3, 10)
	fmt.Println("height: ", height)
	fmt.Println("blacklistedPower: ", blacklistedPower)
	app.DistrKeeper.SetBlacklistedPower(ctx, height, blacklistedPower)

	// above is block n-3
	// block n-2 processes
	// block n-1 processes
	// below is block n

	// block n, for block n-1 (signed by validators from block n-3)
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

	sumPreviousPrecommitPower, totalPreviousPower := int64(210), int64(210)
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

	// add the second delegator to the blacklist
	blacklistedDelAddr := addrs[1].String()
	params := app.DistrKeeper.GetParams(ctx)
	params.NoRewardsDelegatorAddresses = append(params.NoRewardsDelegatorAddresses, blacklistedDelAddr)
	app.DistrKeeper.SetParams(ctx, params)
	fmt.Println("added del to blacklist: ", blacklistedDelAddr)

	// delegate some tokens from delegator 0 (whitelisted) to validator 1 (blacklisted operator)
	tstaking.Ctx = ctx
	delegatorAddr := sdk.AccAddress(valAddrs[0])
	validatorAddr := valAddrs[1]
	tstaking.Delegate(delegatorAddr, validatorAddr, amtToMixedDelegate)
	app.StakingKeeper.Delegation(ctx, delegatorAddr, validatorAddr)
	// val1 has 100_000_000 blacklisted, 10_000_000 whitelisted tokens
	val1BlacklistedPowerShare := sdk.NewDec(100).Quo(sdk.NewDec(110))
	fmt.Println("val1BlacklistedPowerShare: ", val1BlacklistedPowerShare)
	// update votes
	votes := []abci.VoteInfo{
		{
			Validator: abci.Validator{
				Address: valConsPk1.Address(),
				Power:   100,
			},
			SignedLastBlock: true,
		},
		{
			Validator: abci.Validator{
				Address: valConsPk2.Address(),
				Power:   110,
			},
			SignedLastBlock: true,
		},
	}

	// Store BlacklistedPower
	// block n-3
	validatorBlacklistedPowers := []distributiontypes.ValidatorBlacklistedPower{}
	validatorBlacklistedPowers = append(validatorBlacklistedPowers, distributiontypes.ValidatorBlacklistedPower{ValidatorAddress: valAddrs[0].String(), BlacklistedPowerShare: sdk.NewDec(0)})
	validatorBlacklistedPowers = append(validatorBlacklistedPowers, distributiontypes.ValidatorBlacklistedPower{ValidatorAddress: valAddrs[1].String(), BlacklistedPowerShare: val1BlacklistedPowerShare})
	blacklistedPower := distributiontypes.BlacklistedPower{
		TotalBlacklistedPowerShare: sdk.NewDec(100),
		ValidatorBlacklistedPowers: validatorBlacklistedPowers,
	}
	height := strconv.FormatInt(ctx.BlockHeight()-3, 10)
	fmt.Println("height: ", height)
	fmt.Println("blacklistedPower: ", blacklistedPower)
	app.DistrKeeper.SetBlacklistedPower(ctx, height, blacklistedPower)

	// above is block n-3
	// block n-2 processes
	// block n-1 processes
	// below is block n

	// block n, for block n-1 (signed by validators from block n-3)
	app.DistrKeeper.AllocateTokens(ctx, sumPreviousPrecommitPower, totalPreviousPower, valConsAddr2, votes)

	// just staking.proportional for first proposer = (93%) * 100_000_000 = 93_000_000
	//       note 93% is the share that goes to staking rewards
	//       since the only delegator to the second validator is blacklisted, all rewards go to the delegators to the first validator!
	whitelistedValidatorRewards := sdk.NewDec(93_000_000).Mul(sdk.NewDec(100).Quo(sdk.NewDec(110))).RoundInt64()
	require.Equal(t, whitelistedValidatorRewards, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[0]).Rewards[0].Amount.RoundInt64())
	// proposer reward + staking.proportional for second proposer = (5 % commission) * 100 = 5_000_000
	blacklistedValidatorRewards := sdk.NewDec(93_000_000).Mul(sdk.NewDec(10).Quo(sdk.NewDec(110))).Add(sdk.NewDec(5_000_000)).RoundInt64()
	require.Equal(t, blacklistedValidatorRewards, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[1]).Rewards[0].Amount.RoundInt64())

	require.Equal(t, int64(98_000_000), whitelistedValidatorRewards+blacklistedValidatorRewards)

	// zero commission for either validator (including proposer)
	//     0% commission: (0% * 93%) * 100 / 2 = 0
	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[1]).Commission.IsZero())

	// =================== check non-staking rewards =========================
	// 2% to community pool
	require.Equal(t, int64(2_000_000), app.DistrKeeper.GetFeePool(ctx).CommunityPool[0].Amount.RoundInt64())
}
