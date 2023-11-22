package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/header"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/auth"
	authkeeper "cosmossdk.io/x/auth/keeper"
	authtypes "cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/bank"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	"cosmossdk.io/x/protocolpool"
	"cosmossdk.io/x/protocolpool/keeper"
	"cosmossdk.io/x/protocolpool/types"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

func TestContinuousFundEndBlocker(t *testing.T) {
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, types.StoreKey,
	)

	logger := log.NewTestLogger(t)
	cms := integration.CreateMultiStore(keys, logger)
	encCfg := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{}, protocolpool.AppModuleBasic{})
	ctx := sdk.NewContext(cms, true, logger)

	maccPerms := map[string][]string{
		types.ModuleName: {authtypes.Minter},
	}

	authority := authtypes.NewModuleAddress("gov")

	// create account keeper
	accountKeeper := authkeeper.NewAccountKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		addresscodec.NewBech32Codec(sdk.Bech32MainPrefix),
		sdk.Bech32MainPrefix,
		authority.String(),
	)

	// create bank keeper
	bankKeeper := bankkeeper.NewBaseKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		accountKeeper,
		map[string]bool{},
		authority.String(),
		log.NewNopLogger(),
	)

	// create protocolpool keeper
	poolKeeper := keeper.NewKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(keys[types.StoreKey]),
		accountKeeper,
		bankKeeper,
		authority.String(),
	)

	poolAcc := authtypes.NewEmptyModuleAccount(types.ModuleName)

	// mint coins in protocolpool module account
	poolModBal := sdk.NewCoins(sdk.NewInt64Coin("test", 100000000))
	err := bankKeeper.MintCoins(ctx, poolAcc.GetName(), poolModBal)
	require.NoError(t, err)

	addrs := simtestutil.CreateIncrementalAccounts(3)

	// Add a continuous fund proposal to the store with a recipient, percentage, cap, and expiry.
	percentage, err := math.LegacyNewDecFromStr("0.2")
	require.NoError(t, err)
	cap := sdk.NewInt64Coin("test", 100000000)
	oneMonthInSeconds := int64(30 * 24 * 60 * 60) // Approximate number of seconds in 1 month
	expiry := ctx.BlockTime().Add(time.Duration(oneMonthInSeconds) * time.Second)
	cf := types.ContinuousFund{
		Recipient:  addrs[0].String(),
		Percentage: percentage,
		Cap:        &cap,
		Expiry:     &expiry,
	}
	err = poolKeeper.ContinuousFund.Set(ctx, addrs[0], cf)
	require.NoError(t, err)

	// fund addrs[1] with an initial balance
	// Add a continuous fund proposal to the store with a recipient account with more than cap funds, percentage, cap, and expiry.
	err = bankKeeper.SendCoinsFromModuleToAccount(ctx, poolAcc.GetName(), addrs[1], sdk.NewCoins(sdk.NewInt64Coin("test", 10000000)))
	require.NoError(t, err)
	cap = sdk.NewInt64Coin("test", 10000000)
	cf = types.ContinuousFund{
		Recipient:  addrs[1].String(),
		Percentage: percentage,
		Cap:        &cap,
		Expiry:     &expiry,
	}
	err = poolKeeper.ContinuousFund.Set(ctx, addrs[1], cf)
	require.NoError(t, err)

	// Check balances before running EndBlocker
	addr1Bal := bankKeeper.GetAllBalances(ctx, addrs[0])
	require.Equal(t, addr1Bal, sdk.Coins{})
	addr2Bal := bankKeeper.GetAllBalances(ctx, addrs[1])
	require.Equal(t, addr2Bal, sdk.NewCoins(sdk.NewInt64Coin("test", 10000000)))

	err = poolKeeper.EndBlocker(ctx)
	require.NoError(t, err)

	// Check balances after running EndBlocker
	addr1BalAfter := bankKeeper.GetAllBalances(ctx, addrs[0])
	check := addr1BalAfter.IsAllGT(addr1Bal)
	require.True(t, check)
	addr2BalAfter := bankKeeper.GetAllBalances(ctx, addrs[1])
	require.Equal(t, addr2Bal, addr2BalAfter) // since addrs[1] has more account bal than cap, no funds are distributed and balance remains same

	_, err = poolKeeper.ContinuousFund.Get(ctx, addrs[1])
	require.Error(t, err)
	require.ErrorIs(t, err, collections.ErrNotFound)

	// Add a continuous fund proposal to the store with a recipient, percentage, cap, and with exipired time.
	expiry = ctx.BlockTime().AddDate(0, 0, 0)
	cf = types.ContinuousFund{
		Recipient:  addrs[2].String(),
		Percentage: percentage,
		Cap:        &cap,
		Expiry:     &expiry,
	}
	ctx = ctx.WithHeaderInfo(header.Info{Time: time.Unix(10, 0)})
	err = poolKeeper.ContinuousFund.Set(ctx, addrs[2], cf)
	require.NoError(t, err)

	// Check balance before running EndBlocker
	addr3Bal := bankKeeper.GetAllBalances(ctx, addrs[2])
	require.Equal(t, addr3Bal, sdk.Coins{})

	err = poolKeeper.EndBlocker(ctx)
	require.NoError(t, err)

	// Check balance after running EndBloker
	addr3BalAfter := bankKeeper.GetAllBalances(ctx, addrs[2])
	require.Equal(t, addr3Bal, addr3BalAfter) // since addrs[2] is already expired

	_, err = poolKeeper.ContinuousFund.Get(ctx, addrs[2])
	require.Error(t, err)
	require.ErrorIs(t, err, collections.ErrNotFound)
}
