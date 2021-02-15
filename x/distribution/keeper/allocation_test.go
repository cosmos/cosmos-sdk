package keeper_test

import (
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

	err := app.BankKeeper.SetBalances(ctx, feeCollector.GetAddress(), fees)
	require.NoError(t, err)
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
	app.DistrKeeper.AllocateTokens(ctx, 200, 200, valConsAddr2, votes)

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

	err := app.BankKeeper.SetBalances(ctx, feeCollector.GetAddress(), fees)
	require.NoError(t, err)

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
