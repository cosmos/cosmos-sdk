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
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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

func (s *KeeperTestSuite) TestClawback() {
	s.SetupTest()
	c := sdk.NewCoins
	fee := func(x int64) sdk.Coin { return sdk.NewInt64Coin(feeDenom, x) }
	stake := func(x int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, x) }
	now := cmtime.Now()

	// set up simapp and validators
	ctx := s.ctx.WithBlockTime(now)
	valAddr, val, err := createValidator(ctx, s.bankKeeper, s.stakingKeeper, 100)
	s.Require().NoError(err)
	s.Require().Equal(s.T(), "stake", s.stakingKeeper.BondDenom(ctx))

	lockupPeriods := types.Periods{
		{Length: int64(12 * 3600), Amount: c(fee(1000), stake(100))}, // noon
	}
	vestingPeriods := types.Periods{
		{Length: int64(8 * 3600), Amount: c(fee(200))},            // 8am
		{Length: int64(1 * 3600), Amount: c(fee(200), stake(50))}, // 9am
		{Length: int64(6 * 3600), Amount: c(fee(200), stake(50))}, // 3pm
		{Length: int64(2 * 3600), Amount: c(fee(200))},            // 5pm
		{Length: int64(1 * 3600), Amount: c(fee(200))},            // 6pm
	}

	bacc, origCoins := initBaseAccount()
	_, _, funder := testdata.KeyTestPubAddr()
	va := types.NewClawbackVestingAccount(bacc, funder, origCoins, now.Unix(), lockupPeriods, vestingPeriods)
	// simulate 17stake lost to slashing
	va.DelegatedVesting = c(stake(17))
	addr := va.GetAddress()
	s.accountKeeper.SetAccount(ctx, va)

	// fund the vesting account with 17 take lost to slashing
	err = testutil.FundAccount(s.bankKeeper, ctx, addr, c(fee(1000), stake(83)))

	s.Require().NoError(err)
	s.Require().Equal(int64(1000), s.bankKeeper.GetBalance(ctx, addr, feeDenom).Amount.Int64())
	s.Require().Equal(int64(83), s.bankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())
	ctx = ctx.WithBlockTime(now.Add(11 * time.Hour))

	// delegate 65
	shares, err := s.stakingKeeper.Delegate(ctx, addr, sdk.NewInt(65), stakingtypes.Unbonded, val, true)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt(65), shares.TruncateInt())
	s.Require().Equal(int64(18), s.bankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())

	// undelegate 5
	_, err = s.stakingKeeper.Undelegate(ctx, addr, valAddr, sdk.NewDec(5))
	s.Require().NoError(err)

	// clawback the unvested funds (600fee, 50stake)
	_, _, dest := testdata.KeyTestPubAddr()
	va2 := s.accountKeeper.GetAccount(ctx, addr).(*types.ClawbackVestingAccount)
	err = va2.Clawback(ctx, funder, dest, s.accountKeeper, s.bankKeeper, s.stakingKeeper)
	s.Require().NoError(err)

	// check vesting account
	// want 400fee, 33stake (delegated), all vested
	feeAmt := s.bankKeeper.GetBalance(ctx, addr, feeDenom).Amount
	s.Require().Equal(int64(400), feeAmt.Int64())
	stakeAmt := s.bankKeeper.GetBalance(ctx, addr, stakeDenom).Amount
	s.Require().Equal(int64(0), stakeAmt.Int64())
	del, found := s.stakingKeeper.GetDelegation(ctx, addr, valAddr)
	s.Require().True(found)
	shares = del.GetShares()
	val, found = s.stakingKeeper.GetValidator(ctx, valAddr)
	s.Require().True(found)
	stakeAmt = val.TokensFromSharesTruncated(shares).RoundInt()
	s.Require().Equal(sdk.NewInt(33), stakeAmt)

	// check destination account
	// want 600fee, 50stake (18 unbonded, 5 unboinding, 27 bonded)
	feeAmt = s.bankKeeper.GetBalance(ctx, dest, feeDenom).Amount
	s.Require().Equal(int64(600), feeAmt.Int64())
	stakeAmt = s.bankKeeper.GetBalance(ctx, dest, stakeDenom).Amount
	s.Require().Equal(int64(18), stakeAmt.Int64())
	del, found = s.stakingKeeper.GetDelegation(ctx, dest, valAddr)
	s.Require().True(found)
	shares = del.GetShares()
	stakeAmt = val.TokensFromSharesTruncated(shares).RoundInt()
	s.Require().Equal(sdk.NewInt(27), stakeAmt)
	ubd, found := s.stakingKeeper.GetUnbondingDelegation(ctx, dest, valAddr)
	s.Require().True(found)
	s.Require().Equal(sdk.NewInt(5), ubd.Entries[0].Balance)

}

func (s *KeeperTestSuite) TestClawback_finalUnlock() {
	s.SetupTest()
	c := sdk.NewCoins
	fee := func(x int64) sdk.Coin { return sdk.NewInt64Coin(feeDenom, x) }
	stake := func(x int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, x) }
	now := cmtime.Now()

	ctx := s.ctx.WithBlockTime(now)
	s.Require().Equal(s.T(), "stake", s.stakingKeeper.BondDenom(ctx))

	lockupPeriods := types.Periods{
		{Length: int64(20 * 3600), Amount: c(fee(1000), stake(100))}, // 8pm
	}
	vestingPeriods := types.Periods{
		{Length: int64(8 * 3600), Amount: c(fee(200))},            // 8am
		{Length: int64(1 * 3600), Amount: c(fee(200), stake(50))}, // 9am
		{Length: int64(6 * 3600), Amount: c(fee(200), stake(50))}, // 3pm
		{Length: int64(2 * 3600), Amount: c(fee(200))},            // 5pm
		{Length: int64(1 * 3600), Amount: c(fee(200))},            // 6pm
	}

	bacc, origCoins := initBaseAccount()
	_, _, funder := testdata.KeyTestPubAddr()
	va := types.NewClawbackVestingAccount(bacc, funder, origCoins, now.Unix(), lockupPeriods, vestingPeriods)
	addr := va.GetAddress()
	s.accountKeeper.SetAccount(ctx, va)

	// fund the vesting account with new grant (old has vested and transferred out)
	err := banktestutil.FundAccount(s.bankKeeper, ctx, addr, c(fee(1000), stake(100)))
	s.Require().NoError(err)
	s.Require().Equal(int64(1000), s.bankKeeper.GetBalance(ctx, addr, feeDenom).Amount.Int64())
	s.Require().Equal(int64(100), s.bankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())
	ctx = ctx.WithBlockTime(now.Add(11 * time.Hour))

	// clawback the unvested funds (600fee, 50stake)
	_, _, dest := testdata.KeyTestPubAddr()
	va2 := s.accountKeeper.GetAccount(ctx, addr).(*types.ClawbackVestingAccount)
	err = va2.Clawback(ctx, funder, dest, s.accountKeeper, s.bankKeeper, s.stakingKeeper)
	s.Require().NoError(err)

	// check vesting account
	// want 400fee, 33stake (delegated), all vested
	feeAmt := s.bankKeeper.GetBalance(ctx, addr, feeDenom).Amount
	s.Require().Equal(int64(400), feeAmt.Int64())
	stakeAmt := s.bankKeeper.GetBalance(ctx, addr, stakeDenom).Amount
	s.Require().Equal(int64(0), stakeAmt.Int64())

	// check destination account
	// want 600fee, 50stake (18 unbonded, 5 unboinding, 27 bonded)
	feeAmt = s.bankKeeper.GetBalance(ctx, dest, feeDenom).Amount
	s.Require().Equal(int64(600), feeAmt.Int64())
	stakeAmt = s.bankKeeper.GetBalance(ctx, dest, stakeDenom).Amount
	s.Require().Equal(int64(18), stakeAmt.Int64())

	// Remaining funds in vesting account should still be locked at 7pm
	ctx = ctx.WithBlockTime(now.Add(19 * time.Hour))
	spendable := s.bankKeeper.SpendableCoins(ctx, addr)
	s.Require().True(spendable.IsZero())
	err = s.bankKeeper.SendCoins(ctx, addr, dest, c(fee(10)))
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestRewards() {
	s.SetupTest()
	c := sdk.NewCoins
	stake := func(x int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, x) }
	now := cmtime.Now()

	// set up simapp and validators
	ctx := s.ctx.WithBlockTime(now)
	_, val, err := createValidator(ctx, s.bankKeeper, s.stakingKeeper, 100)
	s.Require().NoError(err)
	s.Require().Equal(s.T(), "stake", s.stakingKeeper.BondDenom(ctx))

	lockupPeriods := types.Periods{
		{Length: 1, Amount: c(stake(4000))},
	}

	vestingPeriods := types.Periods{
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
	}

	bacc, origCoins := initBaseAccount()
	_, _, funder := testdata.KeyTestPubAddr()
	va := types.NewClawbackVestingAccount(bacc, funder, origCoins, now.Unix(), lockupPeriods, vestingPeriods)
	addr := va.GetAddress()
	s.accountKeeper.SetAccount(ctx, va)

	// fund the vesting account with 300stake lost to transfer
	err = testutil.FundAccount(s.bankKeeper, ctx, addr, c(stake(3700)))
	s.Require().NoError(err)
	s.Require().Equal(int64(3700), s.bankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())
	ctx = ctx.WithBlockTime(now.Add(650 * time.Second))

	// delegate 1600
	shares, err := s.stakingKeeper.Delegate(ctx, addr, sdk.NewInt(1600), stakingtypes.Unbonded, val, true)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt(1600), shares.TruncateInt())
	s.Require().Equal(int64(2100), s.bankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())
	va = s.accountKeeper.GetAccount(ctx, addr).(*types.ClawbackVestingAccount)
	s.Require().Equal(int64(1000), va.GetVestingCoins(ctx.BlockTime()).AmountOf(stakeDenom).Int64())

	// distribute a reward of 120stake
	err = testutil.FundAccount(s.bankKeeper, ctx, addr, c(stake(120)))
	s.Require().NoError(err)
	va.PostReward(ctx, c(stake(120)), s.accountKeeper, s.bankKeeper, s.stakingKeeper)

	// With 1600 delegated, 1000 unvested, reward should be 75 unvested
	va = s.accountKeeper.GetAccount(ctx, addr).(*types.ClawbackVestingAccount)
	s.Require().Equal(int64(4075), va.OriginalVesting.AmountOf(stakeDenom).Int64())
	s.Require().Equal(8, len(va.VestingPeriods))
	for i := 0; i < 6; i++ {
		s.Require().Equal(int64(500), va.VestingPeriods[i].Amount.AmountOf(stakeDenom).Int64())
	}
	s.Require().Equal(int64(537), va.VestingPeriods[6].Amount.AmountOf(stakeDenom).Int64())
	s.Require().Equal(int64(538), va.VestingPeriods[7].Amount.AmountOf(stakeDenom).Int64())
}

func (s *KeeperTestSuite) TestRewards_PostSlash() {
	c := sdk.NewCoins
	stake := func(x int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, x) }
	now := cmtime.Now()

	ctx := s.ctx.WithBlockTime(now)
	_, val, err := createValidator(ctx, s.bankKeeper, s.stakingKeeper, 100)
	s.Require().NoError(err)
	s.Require().Equal(s.T(), "stake", s.stakingKeeper.BondDenom(ctx))

	lockupPeriods := types.Periods{
		{Length: 1, Amount: c(stake(4000))},
	}
	vestingPeriods := types.Periods{
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
		{Length: int64(100), Amount: c(stake(500))},
	}

	bacc, origCoins := initBaseAccount()
	_, _, funder := testdata.KeyTestPubAddr()
	va := types.NewClawbackVestingAccount(bacc, funder, origCoins, now.Unix(), lockupPeriods, vestingPeriods)
	addr := va.GetAddress()
	s.accountKeeper.SetAccount(ctx, va)

	// fund the vesting account with 350stake lost to slashing
	err = testutil.FundAccount(s.bankKeeper, ctx, addr, c(stake(3650)))
	s.Require().NoError(err)
	s.Require().Equal(int64(3650), s.bankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())

	// delegate all 3650stake
	shares, err := s.stakingKeeper.Delegate(ctx, addr, sdk.NewInt(3650), stakingtypes.Unbonded, val, true)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt(3650), shares.TruncateInt())
	s.Require().Equal(int64(0), s.bankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())
	va = s.accountKeeper.GetAccount(ctx, addr).(*types.ClawbackVestingAccount)

	// distribute a reward of 160stake - should all be unvested
	err = testutil.FundAccount(s.bankKeeper, ctx, addr, c(stake(160)))
	s.Require().NoError(err)
	va.PostReward(ctx, c(stake(160)), s.accountKeeper, s.bankKeeper, s.stakingKeeper)
	va = s.accountKeeper.GetAccount(ctx, addr).(*types.ClawbackVestingAccount)
	s.Require().Equal(int64(4160), va.OriginalVesting.AmountOf(stakeDenom).Int64())
	s.Require().Equal(8, len(va.VestingPeriods))
	for i := 0; i < 8; i++ {
		s.Require().Equal(int64(520), va.VestingPeriods[i].Amount.AmountOf(stakeDenom).Int64())
	}

	// must not be able to transfer reward until it vests
	_, _, dest := testdata.KeyTestPubAddr()
	err = s.bankKeeper.SendCoins(ctx, addr, dest, c(stake(1)))
	s.Require().Error(err)
	ctx = ctx.WithBlockTime(now.Add(600 * time.Second))
	err = s.bankKeeper.SendCoins(ctx, addr, dest, c(stake(160)))
	s.Require().NoError(err)

	// distribute another reward once everything has vested
	ctx = ctx.WithBlockTime(now.Add(1200 * time.Second))
	err = testutil.FundAccount(s.bankKeeper, ctx, addr, c(stake(160)))
	s.Require().NoError(err)
	va.PostReward(ctx, c(stake(160)), s.accountKeeper, s.bankKeeper, s.stakingKeeper)
	va = s.accountKeeper.GetAccount(ctx, addr).(*types.ClawbackVestingAccount)
	// shouldn't be added to vesting schedule
	s.Require().Equal(int64(4160), va.OriginalVesting.AmountOf(stakeDenom).Int64())
}

func (s *KeeperTestSuite) TestAddGrantClawbackVestingAcc() {
	c := sdk.NewCoins
	fee := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(feeDenom, amt) }
	stake := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, amt) }
	now := cmtime.Now()

	ctx := s.ctx.WithBlockTime(now)
	s.Require().Equal(s.T(), "stake", s.stakingKeeper.BondDenom(ctx))

	// create an account with an initial grant
	_, _, funder := testdata.KeyTestPubAddr()
	lockupPeriods := types.Periods{{Length: 1, Amount: c(fee(1000), stake(100))}}
	vestingPeriods := types.Periods{
		{Length: 100, Amount: c(fee(250), stake(25))},
		{Length: 100, Amount: c(fee(250), stake(25))},
		{Length: 100, Amount: c(fee(250), stake(25))},
		{Length: 100, Amount: c(fee(250), stake(25))},
	}
	bacc, origCoins := initBaseAccount()
	va := types.NewClawbackVestingAccount(bacc, funder, origCoins, now.Unix(), lockupPeriods, vestingPeriods)
	addr := va.GetAddress()

	// simulate 60stake (unvested) lost to slashing
	va.DelegatedVesting = c(stake(60))

	// At now+150, 75stake unvested but only 15stake are locked, due to slashing
	ctx = ctx.WithBlockTime(now.Add(150 * time.Second))
	s.Require().Equal(int64(75), va.GetVestingCoins(ctx.BlockTime()).AmountOf(stakeDenom).Int64())
	s.Require().Equal(int64(15), va.LockedCoins(ctx.BlockTime()).AmountOf(stakeDenom).Int64())

	// Add a new grant while all slashing is covered by unvested tokens
	err := va.AddGrant(ctx, funder.String(), s.stakingKeeper, ctx.BlockTime().Unix(),
		lockupPeriods, vestingPeriods, origCoins)
	s.Require().NoError(err)

	// After new grant, 115stake locked at now+150, due to slashing,
	// delegation bookkeeping unchanged
	s.Require().Equal(int64(115), va.LockedCoins(ctx.BlockTime()).AmountOf(stakeDenom).Int64())
	s.Require().Equal(int64(60), va.DelegatedVesting.AmountOf(stakeDenom).Int64())
	s.Require().Equal(int64(0), va.DelegatedFree.AmountOf(stakeDenom).Int64())

	// At now+1000, nothing unvested, nothing locked
	ctx = ctx.WithBlockTime(now.Add(1000 * time.Second))
	s.Require().Equal(int64(0), va.GetVestingCoins(ctx.BlockTime()).AmountOf(stakeDenom).Int64())
	s.Require().Equal(int64(0), va.LockedCoins(ctx.BlockTime()).AmountOf(stakeDenom).Int64())

	// Add a new grant with residual slashed amount, but no unvested
	err = va.AddGrant(ctx, funder.String(), s.stakingKeeper, ctx.BlockTime().Unix(),
		lockupPeriods, vestingPeriods, origCoins)
	s.Require().NoError(err)

	// After new grant, all 100 locked, no residual delegation bookkeeping
	s.Require().Equal(int64(100), va.LockedCoins(ctx.BlockTime()).AmountOf(stakeDenom).Int64())
	s.Require().Equal(int64(0), va.DelegatedVesting.AmountOf(stakeDenom).Int64())
	s.Require().Equal(int64(0), va.DelegatedFree.AmountOf(stakeDenom).Int64())

	s.accountKeeper.SetAccount(ctx, va)

	// fund the vesting account with new grant (old has vested and transferred out)
	err = testutil.FundAccount(s.bankKeeper, ctx, addr, origCoins)
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

func (s *KeeperTestSuite) TestReturnGrants() {
	c := sdk.NewCoins
	fee := func(x int64) sdk.Coin { return sdk.NewInt64Coin(feeDenom, x) }
	stake := func(x int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, x) }
	now := cmtime.Now()

	// set up simapp and validators
	ctx := s.ctx.WithBlockTime(now)
	valAddr, val, err := createValidator(ctx, s.bankKeeper, s.stakingKeeper, 100)
	s.Require().NoError(err)
	s.Require().Equal(s.T(), "stake", s.stakingKeeper.BondDenom(ctx))

	lockupPeriods := types.Periods{
		{Length: int64(12 * 3600), Amount: c(fee(1000), stake(100))}, // noon
	}
	vestingPeriods := types.Periods{
		{Length: int64(8 * 3600), Amount: c(fee(200))},            // 8am
		{Length: int64(1 * 3600), Amount: c(fee(200), stake(50))}, // 9am
		{Length: int64(6 * 3600), Amount: c(fee(200), stake(50))}, // 3pm
		{Length: int64(2 * 3600), Amount: c(fee(200))},            // 5pm
		{Length: int64(1 * 3600), Amount: c(fee(200))},            // 6pm
	}

	bacc, origCoins := initBaseAccount()
	_, _, funder := testdata.KeyTestPubAddr()
	va := types.NewClawbackVestingAccount(bacc, funder, origCoins, now.Unix(), lockupPeriods, vestingPeriods)
	// simulate 17stake lost to slashing
	va.DelegatedVesting = c(stake(17))
	addr := va.GetAddress()
	s.accountKeeper.SetAccount(ctx, va)

	// fund the vesting account with an extra 200fee but 17stake lost to slashing
	err = testutil.FundAccount(s.bankKeeper, ctx, addr, c(fee(1200), stake(83)))
	s.Require().NoError(err)
	s.Require().Equal(int64(1200), s.bankKeeper.GetBalance(ctx, addr, feeDenom).Amount.Int64())
	s.Require().Equal(int64(83), s.bankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())
	ctx = ctx.WithBlockTime(now.Add(11 * time.Hour))

	// delegate 65
	shares, err := s.stakingKeeper.Delegate(ctx, addr, sdk.NewInt(65), stakingtypes.Unbonded, val, true)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt(65), shares.TruncateInt())
	s.Require().Equal(int64(18), s.bankKeeper.GetBalance(ctx, addr, stakeDenom).Amount.Int64())

	// undelegate 5
	_, err = s.stakingKeeper.Undelegate(ctx, addr, valAddr, sdk.NewDec(5))
	s.Require().NoError(err)

	// Return the grant (1000fee, 100stake) with (1200fee, 83stake) available
	va2 := s.accountKeeper.GetAccount(ctx, addr).(*types.ClawbackVestingAccount)
	va2.ReturnGrants(ctx, s.accountKeeper, s.bankKeeper, s.stakingKeeper)

	// check vesting account
	// want 200fee all vested
	dest := funder
	feeAmt := s.bankKeeper.GetBalance(ctx, addr, feeDenom).Amount
	s.Require().Equal(int64(200), feeAmt.Int64())
	stakeAmt := s.bankKeeper.GetBalance(ctx, addr, stakeDenom).Amount
	s.Require().Equal(int64(0), stakeAmt.Int64())
	spendable := s.bankKeeper.SpendableCoins(ctx, addr)
	s.Require().Equal(int64(200), spendable.AmountOf(feeDenom).Int64())
	_, found := s.stakingKeeper.GetDelegation(ctx, addr, valAddr)
	s.Require().False(found)
	_, found = s.stakingKeeper.GetUnbondingDelegation(ctx, addr, valAddr)
	s.Require().False(found)

	// check destination account
	// want 1000fee, 83stake (18 unbonded, 5 unbonding, 60 bonded)
	feeAmt = s.bankKeeper.GetBalance(ctx, dest, feeDenom).Amount
	s.Require().Equal(int64(1000), feeAmt.Int64())
	stakeAmt = s.bankKeeper.GetBalance(ctx, dest, stakeDenom).Amount
	s.Require().Equal(int64(18), stakeAmt.Int64())
	del, found := s.stakingKeeper.GetDelegation(ctx, dest, valAddr)
	s.Require().True(found)
	s.Require().Equal(sdk.NewInt(60), del.Shares.TruncateInt())
	val, found = s.stakingKeeper.GetValidator(ctx, valAddr)
	s.Require().True(found)
	stakeAmt = val.TokensFromSharesTruncated(shares).RoundInt()
	s.Require().Equal(sdk.NewInt(60), stakeAmt)
	ubd, found := s.stakingKeeper.GetUnbondingDelegation(ctx, dest, valAddr)
	s.Require().True(found)
	s.Require().Equal(1, len(ubd.Entries))
	s.Require().Equal(sdk.NewInt(5), ubd.Entries[0].Balance)
}

func (s *KeeperTestSuite) TestClawbackVestingAccountStore() {
	baseAcc, coins := initBaseAccount()
	addr := sdk.AccAddress([]byte("the funder"))
	acc := types.NewClawbackVestingAccount(baseAcc, addr, coins, time.Now().Unix(),
		types.Periods{types.Period{3600, coins}}, types.Periods{types.Period{3600, coins}})

	ctx := s.ctx
	_, _, err := createValidator(ctx, s.bankKeeper, s.stakingKeeper, 100)
	s.Require().NoError(err)

	s.accountKeeper.SetAccount(ctx, acc)
	acc2 := s.accountKeeper.GetAccount(ctx, acc.GetAddress())
	s.Require().IsType(&types.ClawbackVestingAccount{}, acc2)
	s.Require().Equal(acc.String(), acc2.String())
}

func (s *KeeperTestSuite) TestClawbackVestingAccountMarshal() {
	baseAcc, coins := initBaseAccount()
	addr := sdk.AccAddress([]byte("the funder"))
	acc := types.NewClawbackVestingAccount(baseAcc, addr, coins, time.Now().Unix(),
		types.Periods{types.Period{3600, coins}}, types.Periods{types.Period{3600, coins}})

	bz, err := s.accountKeeper.MarshalAccount(acc)
	s.Require().Nil(err)

	acc2, err := s.accountKeeper.UnmarshalAccount(bz)
	s.Require().Nil(err)
	s.Require().IsType(&types.ClawbackVestingAccount{}, acc2)
	s.Require().Equal(acc.String(), acc2.String())

	// error on bad bytes
	_, err = s.accountKeeper.UnmarshalAccount(bz[:len(bz)/2])
	s.Require().NotNil(err)
}

func initBaseAccount() (*authtypes.BaseAccount, sdk.Coins) {
	_, _, addr := testdata.KeyTestPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := authtypes.NewBaseAccountWithAddress(addr)

	return bacc, origCoins
}

func createValidator(ctx sdk.Context, bankKeeper bankkeeper.Keeper, stakingKeeper stakingkeeper.Keeper, powers int64) (sdk.ValAddress, stakingtypes.Validator, error) {
	valTokens := sdk.TokensFromConsensusPower(powers, sdk.DefaultPowerReduction)
	addrs := simtestutil.AddTestAddrsIncremental(bankKeeper, stakingKeeper, ctx, 1, valTokens)
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrs)
	pks := simtestutil.CreateTestPubKeys(1)

	val, err := stakingtypes.NewValidator(valAddrs[0], pks[0], stakingtypes.Description{})
	if err != nil {
		return nil, stakingtypes.Validator{}, err
	}

	stakingKeeper.SetValidator(ctx, val)
	if err := stakingKeeper.SetValidatorByConsAddr(ctx, val); err != nil {
		return nil, stakingtypes.Validator{}, err
	}
	stakingKeeper.SetNewValidatorByPowerIndex(ctx, val)
	_, err = stakingKeeper.Delegate(ctx, addrs[0], valTokens, stakingtypes.Unbonded, val, true)
	if err != nil {
		return nil, stakingtypes.Validator{}, err
	}
	_ = staking.EndBlocker(ctx, &stakingKeeper)

	return valAddrs[0], val, nil
}
