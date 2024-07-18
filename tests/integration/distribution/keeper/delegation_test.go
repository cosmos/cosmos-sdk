package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestWithdrawTokenizeShareRecordReward(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	fundAccounts(t, f.sdkCtx, f.bankKeeper)

	valAddr2 := sdk.ValAddress(sdk.AccAddress(PKS[1].Address()))

	valAddrs := []sdk.ValAddress{f.valAddr, valAddr2}
	tstaking := stakingtestutil.NewHelper(t, f.sdkCtx, f.stakingKeeper)

	// create validator with 50% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))
	valPower := int64(100)
	tstaking.CreateValidatorWithValPower(valAddrs[0], PKS[0], valPower, true)

	// end block to bond validator
	f.stakingKeeper.EndBlocker(f.sdkCtx)

	// next block
	ctx := f.sdkCtx.WithBlockHeight(f.sdkCtx.BlockHeight() + 1)

	// fetch validator and delegation
	val, err := f.stakingKeeper.Validator(ctx, valAddrs[0])
	require.NoError(t, err)
	del, err := f.stakingKeeper.Delegation(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])
	require.NoError(t, err)

	// end period
	endingPeriod, err := f.distrKeeper.IncrementValidatorPeriod(ctx, val)
	require.NoError(t, err)

	// calculate delegation rewards
	rewards, err := f.distrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)
	require.NoError(t, err)

	// rewards should be zero
	require.True(t, rewards.IsZero())

	// start out block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// retrieve validator
	val, err = f.stakingKeeper.Validator(ctx, valAddrs[0])
	require.NoError(t, err)

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some rewards
	initial := f.stakingKeeper.TokensFromConsensusPower(ctx, 10)
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDecFromInt(initial)}}
	f.distrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// end period
	f.distrKeeper.IncrementValidatorPeriod(ctx, val)

	coins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, initial)}
	err = f.mintKeeper.MintCoins(ctx, coins)
	require.NoError(t, err)
	err = f.bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, coins)
	require.NoError(t, err)

	// tokenize share amount
	delTokens := math.NewInt(1000000)
	msgServer := stakingkeeper.NewMsgServerImpl(f.stakingKeeper)
	resp, err := msgServer.TokenizeShares(ctx, &stakingtypes.MsgTokenizeShares{
		DelegatorAddress:    sdk.AccAddress(valAddrs[0]).String(),
		ValidatorAddress:    valAddrs[0].String(),
		TokenizedShareOwner: sdk.AccAddress(valAddrs[1]).String(),
		Amount:              sdk.NewCoin(sdk.DefaultBondDenom, delTokens),
	})
	require.NoError(t, err)

	// try withdrawing rewards before no reward is allocated
	coins, err = f.distrKeeper.WithdrawAllTokenizeShareRecordReward(ctx, sdk.AccAddress(valAddrs[1]))
	require.Nil(t, err)
	require.Equal(t, coins, sdk.Coins{})

	// assert tokenize share response
	require.NoError(t, err)
	require.Equal(t, resp.Amount.Amount, delTokens)

	// end block to bond validator
	f.stakingKeeper.EndBlocker(ctx)
	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	// allocate some rewards
	f.distrKeeper.AllocateTokensToValidator(ctx, val, tokens)
	// end period
	f.distrKeeper.IncrementValidatorPeriod(ctx, val)

	beforeBalance := f.bankKeeper.GetBalance(ctx, sdk.AccAddress(valAddrs[1]), sdk.DefaultBondDenom)

	// withdraw rewards
	coins, err = f.distrKeeper.WithdrawAllTokenizeShareRecordReward(ctx, sdk.AccAddress(valAddrs[1]))
	require.Nil(t, err)

	// check return value
	require.Equal(t, coins.String(), "50000stake")
	// check balance changes
	midBalance := f.bankKeeper.GetBalance(ctx, sdk.AccAddress(valAddrs[1]), sdk.DefaultBondDenom)
	require.Equal(t, beforeBalance.Amount.Add(coins.AmountOf(sdk.DefaultBondDenom)), midBalance.Amount)

	// allocate more rewards manually on module account and try full redeem
	record, err := f.stakingKeeper.GetTokenizeShareRecord(ctx, 1)
	require.NoError(t, err)

	err = f.mintKeeper.MintCoins(ctx, coins)
	require.NoError(t, err)
	err = f.bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, record.GetModuleAddress(), coins)
	require.NoError(t, err)

	shareTokenBalance := f.bankKeeper.GetBalance(ctx, sdk.AccAddress(valAddrs[0]), record.GetShareTokenDenom())

	_, err = msgServer.RedeemTokensForShares(ctx, &stakingtypes.MsgRedeemTokensForShares{
		DelegatorAddress: sdk.AccAddress(valAddrs[0]).String(),
		Amount:           shareTokenBalance,
	})
	require.NoError(t, err)

	finalBalance := f.bankKeeper.GetBalance(ctx, sdk.AccAddress(valAddrs[1]), sdk.DefaultBondDenom)
	require.Equal(t, midBalance.Amount.Add(coins.AmountOf(sdk.DefaultBondDenom)), finalBalance.Amount)
}
