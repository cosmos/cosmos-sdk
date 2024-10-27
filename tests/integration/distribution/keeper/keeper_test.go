package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/testutil"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

func TestSetWithdrawAddr(t *testing.T) {
	var (
		bankKeeper    bankkeeper.Keeper
		distrKeeper   keeper.Keeper
		stakingKeeper *stakingkeeper.Keeper
	)

	app, err := simtestutil.Setup(testutil.AppConfig,
		&bankKeeper,
		&distrKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addr := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 2, sdk.NewInt(1000000000))

	params := distrKeeper.GetParams(ctx)
	params.WithdrawAddrEnabled = false
	require.NoError(t, distrKeeper.SetParams(ctx, params))

	err = distrKeeper.SetWithdrawAddr(ctx, addr[0], addr[1])
	require.NotNil(t, err)

	params.WithdrawAddrEnabled = true
	require.NoError(t, distrKeeper.SetParams(ctx, params))

	err = distrKeeper.SetWithdrawAddr(ctx, addr[0], addr[1])
	require.Nil(t, err)

	require.Error(t, distrKeeper.SetWithdrawAddr(ctx, addr[0], distrAcc.GetAddress()))
}

func TestWithdrawValidatorCommission(t *testing.T) {
	var (
		accountKeeper authkeeper.AccountKeeper
		bankKeeper    bankkeeper.Keeper
		distrKeeper   keeper.Keeper
		stakingKeeper *stakingkeeper.Keeper
	)

	app, err := simtestutil.Setup(testutil.AppConfig,
		&accountKeeper,
		&bankKeeper,
		&distrKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	valCommission := sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", math.LegacyNewDec(5).Quo(math.LegacyNewDec(4))),
		sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(3).Quo(math.LegacyNewDec(2))),
	}

	addr := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 1, sdk.NewInt(1000000000))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addr)

	// set module account coins
	distrAcc := distrKeeper.GetDistributionAccount(ctx)
	coins := sdk.NewCoins(sdk.NewCoin("mytoken", sdk.NewInt(2)), sdk.NewCoin("stake", sdk.NewInt(2)))
	require.NoError(t, banktestutil.FundModuleAccount(bankKeeper, ctx, distrAcc.GetName(), coins))

	accountKeeper.SetModuleAccount(ctx, distrAcc)

	// check initial balance
	balance := bankKeeper.GetAllBalances(ctx, sdk.AccAddress(valAddrs[0]))
	expTokens := stakingKeeper.TokensFromConsensusPower(ctx, 1000)
	expCoins := sdk.NewCoins(sdk.NewCoin("stake", expTokens))
	require.Equal(t, expCoins, balance)

	// set outstanding rewards
	distrKeeper.SetValidatorOutstandingRewards(ctx, valAddrs[0], types.ValidatorOutstandingRewards{Rewards: valCommission})

	// set commission
	distrKeeper.SetValidatorAccumulatedCommission(ctx, valAddrs[0], types.ValidatorAccumulatedCommission{Commission: valCommission})

	// withdraw commission
	_, err = distrKeeper.WithdrawValidatorCommission(ctx, valAddrs[0])
	require.NoError(t, err)

	// check balance increase
	balance = bankKeeper.GetAllBalances(ctx, sdk.AccAddress(valAddrs[0]))
	require.Equal(t, sdk.NewCoins(
		sdk.NewCoin("mytoken", sdk.NewInt(1)),
		sdk.NewCoin("stake", expTokens.AddRaw(1)),
	), balance)

	// check remainder
	remainder := distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission
	require.Equal(t, sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", math.LegacyNewDec(1).Quo(math.LegacyNewDec(4))),
		sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(1).Quo(math.LegacyNewDec(2))),
	}, remainder)

	require.True(t, true)
}

func TestGetTotalRewards(t *testing.T) {
	var (
		bankKeeper    bankkeeper.Keeper
		distrKeeper   keeper.Keeper
		stakingKeeper *stakingkeeper.Keeper
	)

	app, err := simtestutil.Setup(testutil.AppConfig,
		&bankKeeper,
		&distrKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	valCommission := sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", math.LegacyNewDec(5).Quo(math.LegacyNewDec(4))),
		sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(3).Quo(math.LegacyNewDec(2))),
	}

	addr := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 2, sdk.NewInt(1000000000))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addr)

	distrKeeper.SetValidatorOutstandingRewards(ctx, valAddrs[0], types.ValidatorOutstandingRewards{Rewards: valCommission})
	distrKeeper.SetValidatorOutstandingRewards(ctx, valAddrs[1], types.ValidatorOutstandingRewards{Rewards: valCommission})

	expectedRewards := valCommission.MulDec(math.LegacyNewDec(2))
	totalRewards := distrKeeper.GetTotalRewards(ctx)

	require.Equal(t, expectedRewards, totalRewards)
}

func TestFundCommunityPool(t *testing.T) {
	var (
		bankKeeper    bankkeeper.Keeper
		distrKeeper   keeper.Keeper
		stakingKeeper *stakingkeeper.Keeper
	)

	app, err := simtestutil.Setup(testutil.AppConfig,
		&bankKeeper,
		&distrKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// reset fee pool
	distrKeeper.SetFeePool(ctx, types.InitialFeePool())

	addr := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 2, math.ZeroInt())

	amount := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	require.NoError(t, banktestutil.FundAccount(bankKeeper, ctx, addr[0], amount))

	initPool := distrKeeper.GetFeePool(ctx)
	require.Empty(t, initPool.CommunityPool)

	err = distrKeeper.FundCommunityPool(ctx, amount, addr[0])
	require.Nil(t, err)

	require.Equal(t, initPool.CommunityPool.Add(sdk.NewDecCoinsFromCoins(amount...)...), distrKeeper.GetFeePool(ctx).CommunityPool)
	require.Empty(t, bankKeeper.GetAllBalances(ctx, addr[0]))
}
