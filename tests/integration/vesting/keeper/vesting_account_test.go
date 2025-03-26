package keeper

import (
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmtime "github.com/cometbft/cometbft/types/time"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtestutil "github.com/cosmos/cosmos-sdk/x/auth/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/stretchr/testify/suite"
)

var (
	stakeDenom = "stake"
	feeDenom   = "fee"
	emptyCoins = sdk.Coins{}
)

type KeeperTestSuite struct {
	suite.Suite

	ctx               sdk.Context
	accountKeeper     keeper.AccountKeeper
	stakingKeeper     stakingkeeper.Keeper
	interfaceRegistry codectypes.InterfaceRegistry
	bankKeeper        bankkeeper.Keeper
}

func (s *KeeperTestSuite) SetupTest() {
	app, err := simtestutil.Setup(authtestutil.AppConfig, s.accountKeeper, s.stakingKeeper, s.bankKeeper, s.interfaceRegistry)
	s.Require().NoError(err)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	s.ctx = ctx
}

func (s *KeeperTestSuite) TestGetVestedCoinsContVestingAcc() {
	now := cmtime.Now()
	endTime := now.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()
	continuousVestingAccount := vestingtypes.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())

	// require no coins vested in the very beginning of the vesting schedule
	vestedCoins := continuousVestingAccount.GetVestedCoins(now)
	s.Require().Nil(vestedCoins)

	// require all coins vested at the end of the vesting schedule
	vestedCoins = continuousVestingAccount.GetVestedCoins(endTime)
	s.Require().Equal(origCoins, vestedCoins)

	// require 50% of coins vested
	vestedCoins = continuousVestingAccount.GetVestedCoins(now.Add(12 * time.Hour))
	s.Require().Equal(sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestedCoins)

	// require 100% of coins vested
	vestedCoins = continuousVestingAccount.GetVestedCoins(now.Add(48 * time.Hour))
	s.Require().Equal(origCoins, vestedCoins)
}

func (s *KeeperTestSuite) TestGetVestingCoinsContVestingAcc() {
	now := cmtime.Now()
	endTime := now.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()
	cva := vestingtypes.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())

	// require all coins vesting in the beginning of the vesting schedule
	vestingCoins := cva.GetVestingCoins(now)
	s.Require().Equal(origCoins, vestingCoins)

	// require no coins vesting at the end of the vesting schedule
	vestingCoins = cva.GetVestingCoins(endTime)
	s.Require().Equal(emptyCoins, vestingCoins)

	// require 50% of coins vesting
	vestingCoins = cva.GetVestingCoins(now.Add(12 * time.Hour))
	s.Require().Equal(sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestingCoins)
}

func (s *KeeperTestSuite) TestSpendableCoinsContVestingAcc() {
	now := cmtime.Now()
	endTime := now.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()
	cva := vestingtypes.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())

	// require that all original coins are locked at the end of the vesting schedule
	lockedCoins := cva.LockedCoins(now)
	s.Require().Equal(origCoins, lockedCoins)

	// require that there exist no locked coins in the beginning of the
	lockedCoins = cva.LockedCoins(endTime)
	s.Require().Equal(emptyCoins, lockedCoins)

	// require that all vested coins (50%) are spendable
	lockedCoins = cva.LockedCoins(now.Add(12 * time.Hour))
	s.Require().Equal(sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, lockedCoins)
}

func (s *KeeperTestSuite) TestTrackDelegationContVestingAcc() {
	now := cmtime.Now()
	endTime := now.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()
	cva := vestingtypes.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	cva.TrackDelegation(now, origCoins, origCoins)
	s.Require().Equal(origCoins, cva.DelegatedVesting)
	s.Require().Nil(cva.DelegatedFree)

	// require the ability to delegate all vested coins
	cva = vestingtypes.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	cva.TrackDelegation(endTime, origCoins, origCoins)
	s.Require().Nil(cva.DelegatedVesting)
	s.Require().Equal(origCoins, cva.DelegatedFree)

	// require the ability to delegate all vesting coins (50%) and all vested coins (50%)
	cva = vestingtypes.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	cva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	s.Require().Equal(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, cva.DelegatedVesting)
	s.Require().Nil(cva.DelegatedFree)

	cva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	s.Require().Equal(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, cva.DelegatedVesting)
	s.Require().Equal(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, cva.DelegatedFree)

	cva = vestingtypes.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	s.Require().Panics(func() {
		cva.TrackDelegation(endTime, origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 1000000)})
	})
	s.Require().Nil(cva.DelegatedVesting)
	s.Require().Nil(cva.DelegatedFree)

}

func (s *KeeperTestSuite) TestTrackUndelegationContVestingAcc() {
	now := cmtime.Now()
	endTime := now.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()
	// require the ability to undelegate all vesting coins
	cva := vestingtypes.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	cva.TrackDelegation(now, origCoins, origCoins)
	cva.TrackUndelegation(origCoins)
	s.Require().Nil(cva.DelegatedFree)
	s.Require().Equal(emptyCoins, cva.DelegatedVesting)

	// require the ability to undelegate all vested coins
	cva = vestingtypes.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	cva.TrackDelegation(endTime, origCoins, origCoins)
	cva.TrackUndelegation(origCoins)
	s.Require().Equal(emptyCoins, cva.DelegatedFree)
	s.Require().Nil(cva.DelegatedVesting)

	// require no modifications when the undelegation amount is zero
	cva = vestingtypes.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())

	s.Require().Panics(func() {
		cva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 0)})
	})
	s.Require().Nil(cva.DelegatedFree)
	s.Require().Nil(cva.DelegatedVesting)

	// undelegate from one validator that got slashed 50%
	cva = vestingtypes.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	cva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	cva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})

	// undelegate from the other validator that did not get slashed
	cva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)})
	s.Require().Equal(s.T(), sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)}, cva.DelegatedFree)
	s.Require().Equal(s.T(), sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, cva.DelegatedVesting)
}

func (s *KeeperTestSuite) TestGetVestedCoinsDelVestingAcc() {
	now := cmtime.Now()
	endTime := now.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()

	// require no coins are vested until schedule maturation
	dva := vestingtypes.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	vestedCoins := dva.GetVestedCoins(now)
	s.Require().Nil(vestedCoins)

	// require all coins be vested at schedule maturation
	vestedCoins = dva.GetVestedCoins(endTime)
	s.Require().Equal(origCoins, vestedCoins)
}

func (s *KeeperTestSuite) TestGetVestingCoinsDelVestingAcc() {
	now := cmtime.Now()
	endTime := now.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()

	// require all coins vesting at the beginning of the schedule
	dva := vestingtypes.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	vestingCoins := dva.GetVestingCoins(now)
	s.Require().Equal(origCoins, vestingCoins)

	// require no coins vesting at schedule maturation
	vestingCoins = dva.GetVestingCoins(endTime)
	s.Require().Equal(s.T(), emptyCoins, vestingCoins)
}

func (s *KeeperTestSuite) TestCreatePeriodicVestingAccBricked() {
	s.SetupTest()
	c := sdk.NewCoins
	fee := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(feeDenom, amt) }
	stake := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, amt) }
	now := cmtime.Now()

	ctx := s.ctx.WithBlockTime(now)
	s.Require().Equal(s.T(), "stake", s.stakingKeeper.BondDenom(ctx))

	bogusPeriods := vestingtypes.Periods{
		{Length: 1, Amount: c(fee(10000), stake(100))},
		{Length: 1000, Amount: []sdk.Coin{{Denom: feeDenom, Amount: sdk.NewInt(-9000)}}},
	}
	bacc, origCoins := initBaseAccount()
	pva := vestingtypes.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), bogusPeriods)
	addr := pva.GetAddress()
	s.accountKeeper.SetAccount(ctx, pva)

	err := banktestutil.FundAccount(s.bankKeeper, ctx, addr, c(fee(6000), stake(100)))
	s.Require().NoError(err)
	s.Require().Equal(int64(100), s.bankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())

	ctx = ctx.WithBlockTime(now.Add(160 * time.Second))
	_, _, dest := testdata.KeyTestPubAddr()
	s.Require().Panics(func() { s.bankKeeper.SendCoins(ctx, addr, dest, c(fee(750))) })
}

func (s *KeeperTestSuite) TestAddGrantPeriodicVestingAcc() {
	s.SetupTest()
	c := sdk.NewCoins
	fee := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(feeDenom, amt) }
	stake := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, amt) }
	now := cmtime.Now()

	ctx := s.ctx.WithBlockTime(now)
	s.Require().Equal(s.T(), "stake", s.stakingKeeper.BondDenom(ctx))

	// create an account with an initial grant
	periods := vestingtypes.Periods{
		{Length: 100, Amount: c(fee(250), stake(25))},
		{Length: 100, Amount: c(fee(250), stake(25))},
		{Length: 100, Amount: c(fee(250), stake(25))},
		{Length: 100, Amount: c(fee(250), stake(25))},
	}
	bacc, origCoins := initBaseAccount()
	pva := vestingtypes.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	addr := pva.GetAddress()

	// simulate 60stake (unvested) lost to slashing
	pva.DelegatedVesting = c(stake(60))

	// At now+150, 75stake unvested but only 15stake locked, due to slashing
	ctx = ctx.WithBlockTime(now.Add(150 * time.Second))
	s.Require().Equal(int64(75), pva.GetVestingCoins(ctx.BlockTime()).AmountOf(stakeDenom).Int64())
	s.Require().Equal(int64(15), pva.LockedCoins(ctx.BlockTime()).AmountOf(stakeDenom).Int64())

	// Add a new grant while all slashing is covered by unvested tokens
	pva.AddGrant(ctx, s.stakingKeeper, ctx.BlockTime().Unix(), periods, origCoins)

	// After new grant, 115stake locked at now+150 due to slashing,
	s.Require().Equal(int64(115), pva.LockedCoins(ctx.BlockTime()).AmountOf(stakeDenom).Int64())
	s.Require().Equal(int64(60), pva.DelegatedVesting.AmountOf(stakeDenom).Int64())
	s.Require().Equal(int64(0), pva.DelegatedFree.AmountOf(stakeDenom).Int64())

	// At now+425, 50stake unvested, nothing locked due to slashing
	ctx = ctx.WithBlockTime(now.Add(425 * time.Second))
	s.Require().Equal(int64(50), pva.GetVestingCoins(ctx.BlockTime()).AmountOf(stakeDenom).Int64())
	s.Require().Equal(int64(0), pva.LockedCoins(ctx.BlockTime()).AmountOf(stakeDenom).Int64())

	// Add a new grant, while slashed amount is 50 unvested, 10 vested
	pva.AddGrant(ctx, s.stakingKeeper, ctx.BlockTime().Unix(), periods, origCoins)

	// After new grant, slashed amount reduced to 50 vested, locked is 100
	s.Require().Equal(int64(100), pva.LockedCoins(ctx.BlockTime()).AmountOf(stakeDenom).Int64())
	s.Require().Equal(int64(50), pva.DelegatedVesting.AmountOf(stakeDenom).Int64())
	s.Require().Equal(int64(0), pva.DelegatedFree.AmountOf(stakeDenom).Int64())

	// At now+1000, nothing unvested, nothing locked
	ctx = ctx.WithBlockTime(now.Add(1000 * time.Second))
	s.Require().Equal(int64(0), pva.GetVestingCoins(ctx.BlockTime()).AmountOf(stakeDenom).Int64())
	s.Require().Equal(int64(0), pva.LockedCoins(ctx.BlockTime()).AmountOf(stakeDenom).Int64())

	// Add a new grant with residual slashed amount, but no unvested
	pva.AddGrant(ctx, s.stakingKeeper, ctx.BlockTime().Unix(), periods, origCoins)

	// After new grant, all 100 locked, no residual delegation bookkeeping
	s.Require().Equal(int64(100), pva.LockedCoins(ctx.BlockTime()).AmountOf(stakeDenom).Int64())
	s.Require().Equal(int64(0), pva.DelegatedVesting.AmountOf(stakeDenom).Int64())
	s.Require().Equal(int64(0), pva.DelegatedFree.AmountOf(stakeDenom).Int64())

	s.accountKeeper.SetAccount(ctx, pva)

	// fund the vesting account with new grant (old has vested and transferred out)
	err := banktestutil.FundAccount(s.bankKeeper, ctx, addr, origCoins)
	s.Require().NoError(err)
	s.Require().Equal(int64(100), s.bankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())

	// we should not be able to transfer the latest grant out until it has vested
	_, _, dest := testdata.KeyTestPubAddr()
	err = s.bankKeeper.SendCoins(ctx, addr, dest, c(stake(1)))
	s.Require().Error(err)
	ctx = ctx.WithBlockTime(now.Add(1500 * time.Second))
	err = s.bankKeeper.SendCoins(ctx, addr, dest, origCoins)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestAddGrantPeriodicVestingAcc_FullSlash() {
	s.SetupTest()
	c := sdk.NewCoins
	stake := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, amt) }
	now := cmtime.Now()

	ctx := s.ctx.WithBlockTime(now)
	s.Require().Equal(s.T(), "stake", s.stakingKeeper.BondDenom(ctx))

	// create an account with an initial grant
	periods := vestingtypes.Periods{
		{Length: 100, Amount: c(stake(40))},
		{Length: 100, Amount: c(stake(60))},
	}
	bacc, _ := initBaseAccount()
	origCoins := c(stake(100))
	pva := vestingtypes.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)

	// simulate all 100stake lost to slashing
	pva.DelegatedVesting = c(stake(100))

	// Nothing locked at now+150 since all unvested lost to slashing
	s.Require().Equal(int64(0), pva.LockedCoins(now.Add(150*time.Second)).AmountOf(stakeDenom).Int64())

	// Add a new grant of 50stake
	newGrant := c(stake(50))
	pva.AddGrant(ctx, s.stakingKeeper, now.Add(500*time.Second).Unix(), []types.Period{{Length: 50, Amount: newGrant}}, newGrant)
	s.accountKeeper.SetAccount(ctx, pva)

	// Only 10 of the new grant locked, since 40 fell into the "hole" of slashed-vested
	s.Require().Equal(int64(10), pva.LockedCoins(now.Add(150*time.Second)).AmountOf(stakeDenom).Int64())
}

func (s *KeeperTestSuite) TestAddGrantPeriodicVestingAcc_negAmount() {
	s.SetupTest()
	c := sdk.NewCoins
	fee := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(feeDenom, amt) }
	stake := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, amt) }
	now := cmtime.Now()

	ctx := s.ctx.WithBlockTime(now)
	s.Require().Equal(s.T(), "stake", s.stakingKeeper.BondDenom(ctx))

	// create an account with an initial grant
	periods := vestingtypes.Periods{
		{Length: 100, Amount: c(fee(250), stake(25))},
		{Length: 100, Amount: c(fee(250), stake(25))},
		{Length: 100, Amount: c(fee(250), stake(25))},
		{Length: 100, Amount: c(fee(250), stake(25))},
	}
	bacc, origCoins := initBaseAccount()
	pva := vestingtypes.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	addr := pva.GetAddress()

	// at now+150, 50stake unvested, 50stake locked
	bogusPeriods := vestingtypes.Periods{
		{Length: 1, Amount: c(fee(750))},
		{Length: 1000, Amount: []sdk.Coin{{Denom: feeDenom, Amount: sdk.NewInt(-749)}}},
	}
	ctx = ctx.WithBlockTime(now.Add(150 * time.Second))
	pva.AddGrant(ctx, s.stakingKeeper, ctx.BlockTime().Unix(), bogusPeriods, c(fee(1)))
	s.accountKeeper.SetAccount(ctx, pva)

	// fund the vesting account with new grant (old has vested and transferred out)
	err := banktestutil.FundAccount(s.bankKeeper, ctx, addr, c(fee(1001), stake(100)))
	s.Require().NoError(err)
	s.Require().Equal(int64(100), s.bankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())

	// try to transfer the original grant before its time
	ctx = ctx.WithBlockTime(now.Add(160 * time.Second))
	_, _, dest := testdata.KeyTestPubAddr()
	err = s.bankKeeper.SendCoins(ctx, addr, dest, c(fee(750)))
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestGetVestedCoinsPeriodicVestingAcc() {
	s.SetupTest()
	now := cmtime.Now()
	endTime := now.Add(24 * time.Hour)
	periods := types.Periods{
		types.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	bacc, origCoins := initBaseAccount()
	pva := vestingtypes.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)

	// require no coins vested at the beginning of the vesting schedule
	vestedCoins := pva.GetVestedCoins(now)
	s.Require().Nil(vestedCoins)

	// require all coins vested at the end of the vesting schedule
	vestedCoins = pva.GetVestedCoins(endTime)
	s.Require().Equal(origCoins, vestedCoins)

	// require no coins vested during first vesting period
	vestedCoins = pva.GetVestedCoins(now.Add(6 * time.Hour))
	s.Require().Nil(vestedCoins)

	//require 50% of coins vested after period 1
	vestedCoins = pva.GetVestedCoins(now.Add(12 * time.Hour))
	s.Require().Equal(sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestedCoins)

	//require period 2 coins don't vest until period is over
	vestedCoins = pva.GetVestedCoins(now.Add(15 * time.Hour))
	s.Require().Equal(sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestedCoins)

	//require 75% of coins vested after period 2
	vestedCoins = pva.GetVestedCoins(now.Add(18 * time.Hour))
	s.Require().Equal(sdk.Coins{sdk.NewInt64Coin(feeDenom, 750), sdk.NewInt64Coin(stakeDenom, 75)}, vestedCoins)

	//require 100% of coins vested
	vestedCoins = pva.GetVestedCoins(now.Add(48 * time.Hour))
	s.Require().Equal(origCoins, vestedCoins)
}

func (s *KeeperTestSuite) TestGetVestingCoinsPeriodicVestingAcc() {
	s.SetupTest()
	now := cmtime.Now()
	endTime := now.Add(24 * time.Hour)
	periods := types.Periods{
		types.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	bacc, origCoins := initBaseAccount()
	pva := vestingtypes.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)

	// require all coins vesting at the beginning of the vesting schedule
	vestingCoins := pva.GetVestingCoins(now)
	s.Require().Equal(origCoins, vestingCoins)

	// require no coins vesting at the end of the vesting schedule
	vestingCoins = pva.GetVestingCoins(endTime)
	s.Require().Equal(emptyCoins, vestingCoins)

	// require 50% of coins vesting
	vestingCoins = pva.GetVestingCoins(now.Add(12 * time.Hour))
	s.Require().Equal(sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestingCoins)

	// require 50% of coins vesting after period 1, but before period 2 completes.
	vestingCoins = pva.GetVestingCoins(now.Add(15 * time.Hour))
	s.Require().Equal(sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestingCoins)

	// require 25% of coins vesting after period 2
	vestingCoins = pva.GetVestingCoins(now.Add(18 * time.Hour))
	s.Require().Equal(sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}, vestingCoins)

	// require 0% of coins vesting after vesting complete
	vestingCoins = pva.GetVestingCoins(now.Add(48 * time.Hour))
	s.Require().Equal(emptyCoins, vestingCoins)
}

func (s *KeeperTestSuite) TestSpendableCoinsPeriodicVestingAcc() {
	s.SetupTest()
	now := cmtime.Now()
	endTime := now.Add(24 * time.Hour)
	periods := types.Periods{
		types.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	bacc, origCoins := initBaseAccount()
	pva := vestingtypes.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)

	// require that there exist no spendable coins at the beginning of the
	// vesting schedule
	lockedCoins := pva.LockedCoins(now)
	s.Require().Equal(origCoins, lockedCoins)

	// require that all original coins are spendable at the end of the vesting
	lockedCoins = pva.LockedCoins(endTime)
	s.Require().Equal(emptyCoins, lockedCoins)

	// require that all still vesting coins (50%) are locked
	lockedCoins = pva.LockedCoins(now.Add(12 * time.Hour))
	s.Require().Equal(sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, lockedCoins)
}

func initBaseAccount() (*authtypes.BaseAccount, sdk.Coins) {
	_, _, addr := testdata.KeyTestPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := authtypes.NewBaseAccountWithAddress(addr)

	return bacc, origCoins
}
