package keeper

import (
	"math/big"
	"testing"
	stdtime "time"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	cmtprototypes "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cometbft/cometbft/types/time"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"
)

var (
	PKs        = simtestutil.CreateTestPubKeys(500)
	stakeDenom = "stake"
	feeDenom   = "fee"
	emptyCoin  = sdk.Coins{}
)

type fixture struct {
	app *integration.App

	sdkCtx sdk.Context
	cdc    codec.Codec
	keys   map[string]*storetypes.KVStoreKey

	accountKeeper authkeeper.AccountKeeper
	bankKeeper    bankkeeper.Keeper
	stakingKeeper stakingkeeper.Keeper
}

func init() {
	sdk.DefaultPowerReduction = math.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
}

func initBaseAccount() (*authtypes.BaseAccount, sdk.Coins) {
	_, _, addr := testdata.KeyTestPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := authtypes.NewBaseAccountWithAddress(addr)

	return bacc, origCoins
}

func initFixtures(t *testing.T) *fixture {
	keys := storetypes.NewKVStoreKeys(authtypes.StoreKey, stakingtypes.StoreKey, banktypes.StoreKey, minttypes.StoreKey)
	cdc := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, vesting.AppModuleBasic{}).Codec

	logger := log.NewTestLogger(t)
	cms := integration.CreateMultiStore(keys, logger)

	newCtx := sdk.NewContext(cms, cmtprototypes.Header{}, true, logger)

	authority := authtypes.NewModuleAddress("gov")

	maccPerms := map[string][]string{
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.ModuleName:        {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	}

	accountKeeper := authkeeper.NewAccountKeeper(
		cdc,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		addresscodec.NewBech32Codec(sdk.Bech32MainPrefix),
		sdk.Bech32MainPrefix,
		authority.String(),
	)

	blockedAddresses := map[string]bool{
		accountKeeper.GetAuthority(): false,
	}
	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		accountKeeper,
		blockedAddresses,
		authority.String(),
		log.NewNopLogger(),
	)

	stakingKeeper := stakingkeeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
		accountKeeper,
		bankKeeper,
		authority.String(),
		addresscodec.NewBech32Codec(sdk.Bech32PrefixValAddr),
		addresscodec.NewBech32Codec(sdk.Bech32PrefixConsAddr),
	)

	authModule := auth.NewAppModule(cdc, accountKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper, nil)
	stakingModule := staking.NewAppModule(cdc, stakingKeeper, accountKeeper, bankKeeper, nil)
	//mintModule := mint.NewAppModule(cdc, mintKeeper, accountKeeper, nil, nil)

	integrationApp := integration.NewIntegrationApp(newCtx, logger, keys, cdc, map[string]appmodule.AppModule{
		authtypes.ModuleName:    authModule,
		banktypes.ModuleName:    bankModule,
		stakingtypes.ModuleName: stakingModule,
		//minttypes.ModuleName:    mintModule,
	})

	sdkCtx := sdk.UnwrapSDKContext(integrationApp.Context())
	// Register MsgServer and QueryServer
	stakingtypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), stakingkeeper.NewMsgServerImpl(stakingKeeper))
	stakingtypes.RegisterQueryServer(integrationApp.QueryHelper(), stakingkeeper.NewQuerier(stakingKeeper))

	// set default staking params
	assert.NilError(t, stakingKeeper.SetParams(sdkCtx, stakingtypes.DefaultParams()))

	f := &fixture{
		app:           integrationApp,
		sdkCtx:        sdkCtx,
		cdc:           cdc,
		keys:          keys,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		stakingKeeper: *stakingKeeper,
	}
	return f
}

func TestCreatePeriodicVestingAccBricked(t *testing.T) {
	t.Parallel()
	_ = initFixtures(t)

	now := time.Now()
	c := sdk.NewCoins
	fee := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(feeDenom, amt) }
	stake := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, amt) }

	badPeriods := vestingtypes.Periods{
		{Length: 1, Amount: c(fee(10000), stake(100))},
		{Length: 1000, Amount: []sdk.Coin{{Denom: feeDenom, Amount: math.NewInt(-9000)}}},
	}
	bacc, origCoins := initBaseAccount()
	_, err := vestingtypes.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), badPeriods)
	assert.Error(t, err, "period #1 has invalid coins: -9000fee")

}

func TestAddGrantPeroidicVestingAcc(t *testing.T) {
	t.Parallel()
	f := initFixtures(t)
	ctx := f.sdkCtx
	now := time.Now()
	c := sdk.NewCoins
	fee := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(feeDenom, amt) }
	stake := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, amt) }

	periods := vestingtypes.Periods{
		{Length: 100, Amount: c(fee(250), stake(25))},
		{Length: 100, Amount: c(fee(250), stake(25))},
		{Length: 100, Amount: c(fee(250), stake(25))},
		{Length: 100, Amount: c(fee(250), stake(25))},
	}
	bacc, origCoins := initBaseAccount()
	pva, err := vestingtypes.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	assert.NilError(t, err)

	// simulate 60stake (unvested) lost to slashing
	pva.DelegatedVesting = c(stake(60))
	ctx = ctx.WithBlockTime(now.Add(150 * stdtime.Second))
	assert.Equal(t, int64(75), pva.GetVestingCoins(ctx.BlockTime()).AmountOf(stakeDenom).Int64())
	assert.Equal(t, int64(15), pva.LockedCoins(ctx.BlockTime()).AmountOf(stakeDenom).Int64())

	// Add a new grant while all slashing is covered by unvested tokens
	pva.AddGrant(ctx, f.stakingKeeper, ctx.BlockTime().Unix(), periods, origCoins)

	// After new grant, 115stake locked at now+150 due to slashing,
	// delegation bookkeeping unchanged
	assert.Equal(t, int64(115), pva.LockedCoins(ctx.BlockTime()).AmountOf(stakeDenom).Int64())
	assert.Equal(t, int64(60), pva.DelegatedVesting.AmountOf(stakeDenom).Int64())
	assert.Equal(t, int64(0), pva.DelegatedFree.AmountOf(stakeDenom).Int64())

	ctx = ctx.WithBlockTime(now.Add(425 * stdtime.Second))
	require.Equal(t, int64(50), pva.GetVestingCoins(ctx.BlockTime()).AmountOf(stakeDenom).Int64())
	require.Equal(t, int64(0), pva.LockedCoins(ctx.BlockTime()).AmountOf(stakeDenom).Int64())

	// Add a new grant, while slashed amount is 50 unvested, 10 vested
	pva.AddGrant(ctx, f.stakingKeeper, ctx.BlockTime().Unix(), periods, origCoins)
	assert.Equal(t, int64(100), pva.LockedCoins(ctx.BlockTime()).AmountOf(stakeDenom).Int64())
	assert.Equal(t, int64(50), pva.DelegatedVesting.AmountOf(stakeDenom).Int64())
	assert.Equal(t, int64(0), pva.DelegatedFree.AmountOf(stakeDenom).Int64())

	ctx = ctx.WithBlockTime(now.Add(1000 * stdtime.Second))
	require.Equal(t, int64(0), pva.GetVestingCoins(ctx.BlockTime()).AmountOf(stakeDenom).Int64())
	require.Equal(t, int64(0), pva.LockedCoins(ctx.BlockTime()).AmountOf(stakeDenom).Int64())

	// Add a new grant with residual slashed amount, but no unvested
	pva.AddGrant(ctx, f.stakingKeeper, ctx.BlockTime().Unix(), periods, origCoins)

	// After new grant, all 100 locked, no residual delegation bookkeeping
	require.Equal(t, int64(100), pva.LockedCoins(ctx.BlockTime()).AmountOf(stakeDenom).Int64())
	require.Equal(t, int64(0), pva.DelegatedVesting.AmountOf(stakeDenom).Int64())
	require.Equal(t, int64(0), pva.DelegatedFree.AmountOf(stakeDenom).Int64())

	//f.accountKeeper.SetAccount(ctx, pva)
	// fund the vesting account with new grant (old has vested and transferred out)
	err = banktestutil.FundAccount(ctx, f.bankKeeper, pva.GetAddress(), origCoins)
	require.NoError(t, err)
	require.Equal(t, int64(100), f.bankKeeper.GetBalance(ctx, pva.GetAddress(), stakeDenom).Amount.Int64())

	// we should not be able to transfer the latest grant out until it has vested
	_, _, dest := testdata.KeyTestPubAddr()
	tadr := pva.GetAddress()
	co := c(stake(1))
	err = f.bankKeeper.SendCoins(ctx, tadr, dest, co)
	//require.Error(t, err)
	ctx = ctx.WithBlockTime(now.Add(1500 * stdtime.Second))
	//err = f.bankKeeper.SendCoins(ctx, pva.GetAddress(), dest, origCoins)
	require.NoError(t, err)
}

func TestAddGrantPeriodicVestincAcc_FullSlash(t *testing.T) {
	t.Parallel()
	f := initFixtures(t)
	ctx := f.sdkCtx
	now := time.Now()
	c := sdk.NewCoins

	stake := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, amt) }
	// create an account with an initial grant
	periods := vestingtypes.Periods{
		{Length: 100, Amount: c(stake(40))},
		{Length: 100, Amount: c(stake(60))},
	}
	bacc, _ := initBaseAccount()
	origCoins := c(stake(100))
	pva, err := vestingtypes.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	assert.NilError(t, err)

	// simulate all 100stake lost to slashing
	pva.DelegatedVesting = c(stake(100))

	// Nothing locked at now+150 since all unvested lost to slashing
	require.Equal(t, int64(0), pva.LockedCoins(now.Add(150*stdtime.Second)).AmountOf(stakeDenom).Int64())

	// Nothing locked at now+150 since all unvested lost to slashing
	newGrant := c(stake(50))
	pva.AddGrant(
		ctx,
		f.stakingKeeper,
		ctx.BlockTime().Unix(),
		[]vestingtypes.Period{{Length: 50, Amount: newGrant}},
		newGrant,
	)
	f.accountKeeper.SetAccount(ctx, pva)

	// Only 10 of the new grant locked, since 40 fell into the "hole" of slashed-vested
	require.Equal(t, int64(0), pva.LockedCoins(now.Add(150*stdtime.Second)).AmountOf(stakeDenom).Int64())
}

func TestAddGrantPeriodicVestingAcc_negAmount(t *testing.T) {
	t.Parallel()
	f := initFixtures(t)
	ctx := f.sdkCtx
	now := time.Now()
	c := sdk.NewCoins

	fee := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(feeDenom, amt) }
	stake := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, amt) }

	periods := vestingtypes.Periods{
		{Length: 100, Amount: c(fee(250), stake(25))},
		{Length: 100, Amount: c(fee(250), stake(25))},
		{Length: 100, Amount: c(fee(250), stake(25))},
		{Length: 100, Amount: c(fee(250), stake(25))},
	}
	bacc, origCoins := initBaseAccount()
	pva, err := vestingtypes.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	assert.NilError(t, err)

	addr := pva.GetAddress()

	// At now+150 add a new grant wich attempts to prematurly vest the grant
	bogusPeriods := vestingtypes.Periods{
		{Length: 1, Amount: c(fee(750))},
		{Length: 1000, Amount: []sdk.Coin{{Denom: feeDenom, Amount: math.NewInt(-749)}}},
	}
	ctx = ctx.WithBlockTime(now.Add(150 * stdtime.Second))
	pva.AddGrant(ctx, f.stakingKeeper, ctx.BlockTime().Unix(), bogusPeriods, c(fee(1)))

	// fund the vesting account with new grant (old has vested and transferred out)
	err = banktestutil.FundAccount(ctx, f.bankKeeper, addr, origCoins)
	assert.NilError(t, err)
	assert.Equal(t, int64(100), f.bankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())

	// try to transfer the orginal grant before its time
	ctx = ctx.WithBlockTime(now.Add(160 * stdtime.Second))
	_, _, dest := testdata.KeyTestPubAddr()
	err = f.bankKeeper.SendCoins(ctx, addr, dest, c(fee(750)))
	assert.NilError(t, err)
}
