package protocolpool

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/header"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	"cosmossdk.io/x/mint/types"
	protocolpoolkeeper "cosmossdk.io/x/protocolpool/keeper"
	protocolpooltypes "cosmossdk.io/x/protocolpool/types"
	stakingkeeper "cosmossdk.io/x/staking/keeper"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// TestWithdrawAnytime tests if withdrawing funds many times vs withdrawing funds once
// yield the same end balance.
func TestWithdrawAnytime(t *testing.T) {
	var accountKeeper authkeeper.AccountKeeper
	var protocolpoolKeeper protocolpoolkeeper.Keeper
	var bankKeeper bankkeeper.Keeper
	var stakingKeeper *stakingkeeper.Keeper

	app, err := simtestutil.SetupAtGenesis(
		depinject.Configs(
			AppConfig,
			depinject.Supply(log.NewNopLogger()),
		), &accountKeeper, &protocolpoolKeeper, &bankKeeper, &stakingKeeper)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false).WithBlockHeight(1).WithHeaderInfo(header.Info{Height: 1})
	acc := accountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.ModuleName))
	require.NotNil(t, acc)

	testAddrs := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 5, math.NewInt(1))
	testAddr0Str, err := accountKeeper.AddressCodec().BytesToString(testAddrs[0])
	require.NoError(t, err)

	msgServer := protocolpoolkeeper.NewMsgServerImpl(protocolpoolKeeper)
	_, err = msgServer.CreateContinuousFund(
		ctx,
		&protocolpooltypes.MsgCreateContinuousFund{
			Authority:  protocolpoolKeeper.GetAuthority(),
			Recipient:  testAddr0Str,
			Percentage: math.LegacyMustNewDecFromStr("0.5"),
		},
	)
	require.NoError(t, err)

	// increase the community pool by a bunch
	for i := 0; i < 30; i++ {
		ctx, err = simtestutil.NextBlock(app, ctx, time.Minute)
		require.NoError(t, err)

		// withdraw funds randomly, but it must always land on the same end balance
		if rand.Intn(100) > 50 {
			_, err = msgServer.WithdrawContinuousFund(ctx, &protocolpooltypes.MsgWithdrawContinuousFund{
				RecipientAddress: testAddr0Str,
			})
			require.NoError(t, err)
		}
	}

	pool, err := protocolpoolKeeper.GetCommunityPool(ctx)
	require.NoError(t, err)
	require.True(t, pool.IsAllGT(sdk.NewCoins(sdk.NewInt64Coin("stake", 100000))))

	_, err = msgServer.WithdrawContinuousFund(ctx, &protocolpooltypes.MsgWithdrawContinuousFund{
		RecipientAddress: testAddr0Str,
	})
	require.NoError(t, err)

	endBalance := bankKeeper.GetBalance(ctx, testAddrs[0], sdk.DefaultBondDenom)
	require.Equal(t, "11883031stake", endBalance.String())
}

// TestExpireInTheMiddle tests if a continuous fund that expires without anyone
// calling the withdraw function, the funds are still distributed correctly.
func TestExpireInTheMiddle(t *testing.T) {
	var accountKeeper authkeeper.AccountKeeper
	var protocolpoolKeeper protocolpoolkeeper.Keeper
	var bankKeeper bankkeeper.Keeper
	var stakingKeeper *stakingkeeper.Keeper

	app, err := simtestutil.SetupAtGenesis(
		depinject.Configs(
			AppConfig,
			depinject.Supply(log.NewNopLogger()),
		), &accountKeeper, &protocolpoolKeeper, &bankKeeper, &stakingKeeper)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false).WithBlockHeight(1).WithHeaderInfo(header.Info{Height: 1})
	acc := accountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.ModuleName))
	require.NotNil(t, acc)

	testAddrs := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 5, math.NewInt(1))
	testAddr0Str, err := accountKeeper.AddressCodec().BytesToString(testAddrs[0])
	require.NoError(t, err)

	msgServer := protocolpoolkeeper.NewMsgServerImpl(protocolpoolKeeper)

	expirationTime := ctx.BlockTime().Add(time.Minute * 2)
	_, err = msgServer.CreateContinuousFund(
		ctx,
		&protocolpooltypes.MsgCreateContinuousFund{
			Authority:  protocolpoolKeeper.GetAuthority(),
			Recipient:  testAddr0Str,
			Percentage: math.LegacyMustNewDecFromStr("0.1"),
			Expiry:     &expirationTime,
		},
	)
	require.NoError(t, err)

	// increase the community pool by a bunch
	for i := 0; i < 30; i++ {
		ctx, err = simtestutil.NextBlock(app, ctx, time.Minute)
		require.NoError(t, err)
	}

	_, err = msgServer.WithdrawContinuousFund(ctx, &protocolpooltypes.MsgWithdrawContinuousFund{
		RecipientAddress: testAddr0Str,
	})
	require.NoError(t, err)

	endBalance := bankKeeper.GetBalance(ctx, testAddrs[0], sdk.DefaultBondDenom)
	require.Equal(t, "237661stake", endBalance.String())
}
