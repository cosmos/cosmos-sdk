package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

func TestAllocateTokensToValidatorWithCommission(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	addrs := simapp.AddTestAddrs(app, ctx, 3, sdk.NewInt(1234))
	valAddrs := simapp.ConvertAddrsToValAddrs(addrs)
	valConsPk1 := PKS[0]

	sh := staking.NewHandler(app.StakingKeeper)

	// create validator with 50% commission
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(
		sdk.ValAddress(addrs[0]), valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, sdk.OneInt(),
	)

	res, err := sh(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

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
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	sh := staking.NewHandler(app.StakingKeeper)
	addrs := simapp.AddTestAddrs(app, ctx, 2, sdk.NewInt(1234))
	valAddrs := simapp.ConvertAddrsToValAddrs(addrs)

	valConsPk1, valConsPk2 := PKS[0], PKS[1]
	valOpAddr1, valOpAddr2  := valAddrs[0], valAddrs[1]
	valConsAddr2 := sdk.ConsAddress(valConsPk2.Address())

	// create validator with 50% commission
	commission := staking.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	msg := staking.NewMsgCreateValidator(valOpAddr1, valConsPk1,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, sdk.OneInt())

	res, err := sh(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	// create second validator with 0% commission
	commission = staking.NewCommissionRates(sdk.NewDec(0), sdk.NewDec(0), sdk.NewDec(0))
	msg = staking.NewMsgCreateValidator(valOpAddr2, valConsPk2,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)), staking.Description{}, commission, sdk.OneInt())

	res, err = sh(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	abciValA := abci.Validator{
		Address: valConsPk1.Address(),
		Power:   100,
	}
	abciValB := abci.Validator{
		Address: valConsPk2.Address(),
		Power:   100,
	}

	// assert initial state: zero outstanding rewards, zero community pool, zero commission, zero current rewards
	require.True(t, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valOpAddr1).Rewards.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valOpAddr2).Rewards.IsZero())
	require.True(t, app.DistrKeeper.GetFeePool(ctx).CommunityPool.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valOpAddr1).Commission.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valOpAddr2).Commission.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valOpAddr1).Rewards.IsZero())
	require.True(t, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valOpAddr2).Rewards.IsZero())

	// allocate tokens as if both had voted and second was proposer
	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)))
	feeCollector := app.SupplyKeeper.GetModuleAccount(ctx, types.FeeCollectorName)
	require.NotNil(t, feeCollector)

	err = app.BankKeeper.SetBalances(ctx, feeCollector.GetAddress(), fees)
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
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(465, 1)}}, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valOpAddr1).Rewards)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(515, 1)}}, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, valOpAddr2).Rewards)
	// 2 community pool coins
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(2)}}, app.DistrKeeper.GetFeePool(ctx).CommunityPool)
	// 50% commission for first proposer, (0.5 * 93%) * 100 / 2 = 23.25
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(2325, 2)}}, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valOpAddr1).Commission)
	// zero commission for second proposer
	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valOpAddr2).Commission.IsZero())
	// just staking.proportional for first proposer less commission = (0.5 * 93%) * 100 / 2 = 23.25
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(2325, 2)}}, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valOpAddr1).Rewards)
	// proposer reward + staking.proportional for second proposer = (5 % + 0.5 * (93%)) * 100 = 51.5
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(515, 1)}}, app.DistrKeeper.GetValidatorCurrentRewards(ctx, valOpAddr2).Rewards)
}

