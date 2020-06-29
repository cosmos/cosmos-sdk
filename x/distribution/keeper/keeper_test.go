package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func TestSetWithdrawAddr(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addr := simapp.AddTestAddrs(app, ctx, 2, sdk.NewInt(1000000000))

	params := app.DistrKeeper.GetParams(ctx)
	params.WithdrawAddrEnabled = false
	app.DistrKeeper.SetParams(ctx, params)

	err := app.DistrKeeper.SetWithdrawAddr(ctx, addr[0], addr[1])
	require.NotNil(t, err)

	params.WithdrawAddrEnabled = true
	app.DistrKeeper.SetParams(ctx, params)

	err = app.DistrKeeper.SetWithdrawAddr(ctx, addr[0], addr[1])
	require.Nil(t, err)

	require.Error(t, app.DistrKeeper.SetWithdrawAddr(ctx, addr[0], distrAcc.GetAddress()))
}

func TestWithdrawValidatorCommission(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	valCommission := sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", sdk.NewDec(5).Quo(sdk.NewDec(4))),
		sdk.NewDecCoinFromDec("stake", sdk.NewDec(3).Quo(sdk.NewDec(2))),
	}

	addr := simapp.AddTestAddrs(app, ctx, 1, sdk.NewInt(1000000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addr)

	// set module account coins
	distrAcc := app.DistrKeeper.GetDistributionAccount(ctx)
	err := app.BankKeeper.SetBalances(ctx, distrAcc.GetAddress(), sdk.NewCoins(
		sdk.NewCoin("mytoken", sdk.NewInt(2)),
		sdk.NewCoin("stake", sdk.NewInt(2)),
	))
	require.NoError(t, err)
	app.AccountKeeper.SetModuleAccount(ctx, distrAcc)

	// check initial balance
	balance := app.BankKeeper.GetAllBalances(ctx, sdk.AccAddress(valAddrs[0]))
	expTokens := sdk.TokensFromConsensusPower(1000)
	expCoins := sdk.NewCoins(sdk.NewCoin("stake", expTokens))
	require.Equal(t, expCoins, balance)

	// set outstanding rewards
	app.DistrKeeper.SetValidatorOutstandingRewards(ctx, valAddrs[0], types.ValidatorOutstandingRewards{Rewards: valCommission})

	// set commission
	app.DistrKeeper.SetValidatorAccumulatedCommission(ctx, valAddrs[0], types.ValidatorAccumulatedCommission{Commission: valCommission})

	// withdraw commission
	_, err = app.DistrKeeper.WithdrawValidatorCommission(ctx, valAddrs[0])
	require.NoError(t, err)

	// check balance increase
	balance = app.BankKeeper.GetAllBalances(ctx, sdk.AccAddress(valAddrs[0]))
	require.Equal(t, sdk.NewCoins(
		sdk.NewCoin("mytoken", sdk.NewInt(1)),
		sdk.NewCoin("stake", expTokens.AddRaw(1)),
	), balance)

	// check remainder
	remainder := app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission
	require.Equal(t, sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", sdk.NewDec(1).Quo(sdk.NewDec(4))),
		sdk.NewDecCoinFromDec("stake", sdk.NewDec(1).Quo(sdk.NewDec(2))),
	}, remainder)

	require.True(t, true)
}

func TestGetTotalRewards(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	valCommission := sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", sdk.NewDec(5).Quo(sdk.NewDec(4))),
		sdk.NewDecCoinFromDec("stake", sdk.NewDec(3).Quo(sdk.NewDec(2))),
	}

	addr := simapp.AddTestAddrs(app, ctx, 2, sdk.NewInt(1000000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addr)

	app.DistrKeeper.SetValidatorOutstandingRewards(ctx, valAddrs[0], types.ValidatorOutstandingRewards{Rewards: valCommission})
	app.DistrKeeper.SetValidatorOutstandingRewards(ctx, valAddrs[1], types.ValidatorOutstandingRewards{Rewards: valCommission})

	expectedRewards := valCommission.MulDec(sdk.NewDec(2))
	totalRewards := app.DistrKeeper.GetTotalRewards(ctx)

	require.Equal(t, expectedRewards, totalRewards)
}

func TestFundCommunityPool(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addr := simapp.AddTestAddrs(app, ctx, 2, sdk.NewInt(1000000000))

	amount := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	require.NoError(t, app.BankKeeper.SetBalances(ctx, addr[0], amount))

	initPool := app.DistrKeeper.GetFeePool(ctx)
	assert.Empty(t, initPool.CommunityPool)

	err := app.DistrKeeper.FundCommunityPool(ctx, amount, addr[0])
	assert.Nil(t, err)

	assert.Equal(t, initPool.CommunityPool.Add(sdk.NewDecCoinsFromCoins(amount...)...), app.DistrKeeper.GetFeePool(ctx).CommunityPool)
	assert.Empty(t, app.BankKeeper.GetAllBalances(ctx, addr[0]))
}
