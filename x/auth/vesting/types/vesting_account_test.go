package types_test

import (
	"testing"
	"time"

	tmtime "github.com/cometbft/cometbft/types/time"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/header"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

var (
	stakeDenom = "stake"
	feeDenom   = "fee"
	emptyCoins = sdk.Coins{}
)

type VestingAccountTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	accountKeeper keeper.AccountKeeper
}

func (s *VestingAccountTestSuite) SetupTest() {
	encCfg := moduletestutil.MakeTestEncodingConfig(vesting.AppModuleBasic{})

	key := storetypes.NewKVStoreKey(authtypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	s.ctx = testCtx.Ctx.WithHeaderInfo(header.Info{})

	maccPerms := map[string][]string{
		"fee_collector":          nil,
		"mint":                   {"minter"},
		"bonded_tokens_pool":     {"burner", "staking"},
		"not_bonded_tokens_pool": {"burner", "staking"},
		"multiPerm":              {"burner", "minter", "staking"},
		"random":                 {"random"},
	}

	s.accountKeeper = keeper.NewAccountKeeper(
		encCfg.Codec,
		storeService,
		authtypes.ProtoBaseAccount,
		maccPerms,
		authcodec.NewBech32Codec("cosmos"),
		"cosmos",
		authtypes.NewModuleAddress("gov").String(),
	)
}

func TestGetVestedCoinsContVestingAcc(t *testing.T) {
	now := tmtime.Now()
	startTime := now.Add(24 * time.Hour)
	endTime := startTime.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()
	cva, err := types.NewContinuousVestingAccount(bacc, origCoins, startTime.Unix(), endTime.Unix())
	require.NoError(t, err)

	// require no coins vested _before_ the start time of the vesting schedule
	vestedCoins := cva.GetVestedCoins(now)
	require.Nil(t, vestedCoins)

	// require no coins vested _before_ the very beginning of the vesting schedule
	vestedCoins = cva.GetVestedCoins(startTime.Add(-1))
	require.Nil(t, vestedCoins)

	// require all coins vested at the end of the vesting schedule
	vestedCoins = cva.GetVestedCoins(endTime)
	require.Equal(t, origCoins, vestedCoins)

	// require 50% of coins vested
	vestedCoins = cva.GetVestedCoins(startTime.Add(12 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestedCoins)

	// require 75% of coins vested
	vestedCoins = cva.GetVestedCoins(startTime.Add(18 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 750), sdk.NewInt64Coin(stakeDenom, 75)}, vestedCoins)

	// require 100% of coins vested
	vestedCoins = cva.GetVestedCoins(endTime)
	require.Equal(t, origCoins, vestedCoins)
}

func TestGetVestingCoinsContVestingAcc(t *testing.T) {
	now := tmtime.Now()
	startTime := now.Add(24 * time.Hour)
	endTime := startTime.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()
	cva, err := types.NewContinuousVestingAccount(bacc, origCoins, startTime.Unix(), endTime.Unix())
	require.NoError(t, err)

	// require all coins vesting before the start time of the vesting schedule
	vestingCoins := cva.GetVestingCoins(now)
	require.Equal(t, origCoins, vestingCoins)

	// require all coins vesting right before the start time of the vesting schedule
	vestingCoins = cva.GetVestingCoins(startTime.Add(-1))
	require.Equal(t, origCoins, vestingCoins)

	// require no coins vesting at the end of the vesting schedule
	vestingCoins = cva.GetVestingCoins(endTime)
	require.Equal(t, emptyCoins, vestingCoins)

	// require 50% of coins vesting in the middle between start and end time
	vestingCoins = cva.GetVestingCoins(startTime.Add(12 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestingCoins)

	// require 25% of coins vesting after 3/4 of the time between start and end time has passed
	vestingCoins = cva.GetVestingCoins(startTime.Add(18 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}, vestingCoins)
}

func TestSpendableCoinsContVestingAcc(t *testing.T) {
	now := tmtime.Now()
	startTime := now.Add(24 * time.Hour)
	endTime := startTime.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()
	cva, err := types.NewContinuousVestingAccount(bacc, origCoins, startTime.Unix(), endTime.Unix())
	require.NoError(t, err)

	// require that all original coins are locked before the beginning of the vesting
	// schedule
	lockedCoins := cva.LockedCoins(now)
	require.Equal(t, origCoins, lockedCoins)

	// require that all original coins are locked at the beginning of the vesting
	// schedule
	lockedCoins = cva.LockedCoins(startTime)
	require.Equal(t, origCoins, lockedCoins)

	// require that there exist no locked coins in the end of the vesting schedule
	lockedCoins = cva.LockedCoins(endTime)
	require.Equal(t, sdk.NewCoins(), lockedCoins)

	// require that all vested coins (50%) are spendable
	lockedCoins = cva.LockedCoins(startTime.Add(12 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, lockedCoins)

	// require 25% of coins vesting after 3/4 of the time between start and end time has passed
	lockedCoins = cva.LockedCoins(startTime.Add(18 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}, lockedCoins)
}

func TestTrackDelegationContVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()

	// require the ability to delegate all vesting coins
	cva, err := types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	require.NoError(t, err)
	cva.TrackDelegation(now, origCoins, origCoins)
	require.Equal(t, origCoins, cva.DelegatedVesting)
	require.Nil(t, cva.DelegatedFree)

	// require the ability to delegate all vested coins
	cva, err = types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	require.NoError(t, err)
	cva.TrackDelegation(endTime, origCoins, origCoins)
	require.Nil(t, cva.DelegatedVesting)
	require.Equal(t, origCoins, cva.DelegatedFree)

	// require the ability to delegate all vesting coins (50%) and all vested coins (50%)
	cva, err = types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	require.NoError(t, err)
	cva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, cva.DelegatedVesting)
	require.Nil(t, cva.DelegatedFree)

	cva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, cva.DelegatedVesting)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, cva.DelegatedFree)

	// require no modifications when delegation amount is zero or not enough funds
	cva, err = types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	require.NoError(t, err)
	require.Panics(t, func() {
		cva.TrackDelegation(endTime, origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 1000000)})
	})
	require.Nil(t, cva.DelegatedVesting)
	require.Nil(t, cva.DelegatedFree)
}

func TestTrackUndelegationContVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()

	// require the ability to undelegate all vesting coins
	cva, err := types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	require.NoError(t, err)
	cva.TrackDelegation(now, origCoins, origCoins)
	cva.TrackUndelegation(origCoins)
	require.Nil(t, cva.DelegatedFree)
	require.Equal(t, emptyCoins, cva.DelegatedVesting)

	// require the ability to undelegate all vested coins
	cva, err = types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	require.NoError(t, err)
	cva.TrackDelegation(endTime, origCoins, origCoins)
	cva.TrackUndelegation(origCoins)
	require.Equal(t, emptyCoins, cva.DelegatedFree)
	require.Nil(t, cva.DelegatedVesting)

	// require no modifications when the undelegation amount is zero
	cva, err = types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	require.NoError(t, err)
	require.Panics(t, func() {
		cva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 0)})
	})
	require.Nil(t, cva.DelegatedFree)
	require.Nil(t, cva.DelegatedVesting)

	// vest 50% and delegate to two validators
	cva, err = types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	require.NoError(t, err)
	cva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	cva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})

	// undelegate from one validator that got slashed 50%
	cva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)}, cva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, cva.DelegatedVesting)

	// undelegate from the other validator that did not get slashed
	cva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Equal(t, emptyCoins, cva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)}, cva.DelegatedVesting)
}

func TestGetVestedCoinsDelVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()

	// require no coins are vested until schedule maturation
	dva, err := types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	require.NoError(t, err)
	vestedCoins := dva.GetVestedCoins(now)
	require.Nil(t, vestedCoins)

	// require all coins be vested at schedule maturation
	vestedCoins = dva.GetVestedCoins(endTime)
	require.Equal(t, origCoins, vestedCoins)
}

func TestGetVestingCoinsDelVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()

	// require all coins vesting at the beginning of the schedule
	dva, err := types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	require.NoError(t, err)
	vestingCoins := dva.GetVestingCoins(now)
	require.Equal(t, origCoins, vestingCoins)

	// require no coins vesting at schedule maturation
	vestingCoins = dva.GetVestingCoins(endTime)
	require.Equal(t, emptyCoins, vestingCoins)
}

func TestSpendableCoinsDelVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()

	// require that all coins are locked in the beginning of the vesting
	// schedule
	dva, err := types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	require.NoError(t, err)
	lockedCoins := dva.LockedCoins(now)
	require.True(t, lockedCoins.Equal(origCoins))

	// require that all coins are spendable after the maturation of the vesting
	// schedule
	lockedCoins = dva.LockedCoins(endTime)
	require.Equal(t, sdk.NewCoins(), lockedCoins)

	// require that all coins are still vesting after some time
	lockedCoins = dva.LockedCoins(now.Add(12 * time.Hour))
	require.True(t, lockedCoins.Equal(origCoins))

	// delegate some locked coins
	// require that locked is reduced
	delegatedAmount := sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 50))
	dva.TrackDelegation(now.Add(12*time.Hour), origCoins, delegatedAmount)
	lockedCoins = dva.LockedCoins(now.Add(12 * time.Hour))
	require.True(t, lockedCoins.Equal(origCoins.Sub(delegatedAmount...)))
}

func TestTrackDelegationDelVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()

	// require the ability to delegate all vesting coins
	dva, err := types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	require.NoError(t, err)
	dva.TrackDelegation(now, origCoins, origCoins)
	require.Equal(t, origCoins, dva.DelegatedVesting)
	require.Nil(t, dva.DelegatedFree)

	// require the ability to delegate all vested coins
	dva, err = types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	require.NoError(t, err)
	dva.TrackDelegation(endTime, origCoins, origCoins)
	require.Nil(t, dva.DelegatedVesting)
	require.Equal(t, origCoins, dva.DelegatedFree)

	// require the ability to delegate all coins half way through the vesting
	// schedule
	dva, err = types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	require.NoError(t, err)
	dva.TrackDelegation(now.Add(12*time.Hour), origCoins, origCoins)
	require.Equal(t, origCoins, dva.DelegatedVesting)
	require.Nil(t, dva.DelegatedFree)

	// require no modifications when delegation amount is zero or not enough funds
	dva, err = types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	require.NoError(t, err)
	require.Panics(t, func() {
		dva.TrackDelegation(endTime, origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 1000000)})
	})
	require.Nil(t, dva.DelegatedVesting)
	require.Nil(t, dva.DelegatedFree)
}

func TestTrackUndelegationDelVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	bacc, origCoins := initBaseAccount()

	// require the ability to undelegate all vesting coins
	dva, err := types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	require.NoError(t, err)
	dva.TrackDelegation(now, origCoins, origCoins)
	dva.TrackUndelegation(origCoins)
	require.Nil(t, dva.DelegatedFree)
	require.Equal(t, emptyCoins, dva.DelegatedVesting)

	// require the ability to undelegate all vested coins
	dva, err = types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	require.NoError(t, err)
	dva.TrackDelegation(endTime, origCoins, origCoins)
	dva.TrackUndelegation(origCoins)
	require.Equal(t, emptyCoins, dva.DelegatedFree)
	require.Nil(t, dva.DelegatedVesting)

	// require no modifications when the undelegation amount is zero
	dva, err = types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	require.NoError(t, err)
	require.Panics(t, func() {
		dva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 0)})
	})
	require.Nil(t, dva.DelegatedFree)
	require.Nil(t, dva.DelegatedVesting)

	// vest 50% and delegate to two validators
	dva, err = types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	require.NoError(t, err)
	dva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	dva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})

	// undelegate from one validator that got slashed 50%
	dva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)})

	require.Nil(t, dva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 75)}, dva.DelegatedVesting)

	// undelegate from the other validator that did not get slashed
	dva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Nil(t, dva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)}, dva.DelegatedVesting)
}

func TestGetVestedCoinsPeriodicVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	periods := types.Periods{
		types.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	bacc, origCoins := initBaseAccount()
	pva, err := types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	require.NoError(t, err)

	// require no coins vested at the beginning of the vesting schedule
	vestedCoins := pva.GetVestedCoins(now)
	require.Nil(t, vestedCoins)

	// require all coins vested at the end of the vesting schedule
	vestedCoins = pva.GetVestedCoins(endTime)
	require.Equal(t, origCoins, vestedCoins)

	// require no coins vested during first vesting period
	vestedCoins = pva.GetVestedCoins(now.Add(6 * time.Hour))
	require.Nil(t, vestedCoins)

	// require 50% of coins vested after period 1
	vestedCoins = pva.GetVestedCoins(now.Add(12 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestedCoins)

	// require period 2 coins don't vest until period is over
	vestedCoins = pva.GetVestedCoins(now.Add(15 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestedCoins)

	// require 75% of coins vested after period 2
	vestedCoins = pva.GetVestedCoins(now.Add(18 * time.Hour))
	require.Equal(t,
		sdk.Coins{
			sdk.NewInt64Coin(feeDenom, 750), sdk.NewInt64Coin(stakeDenom, 75),
		}, vestedCoins)

	// require 100% of coins vested
	vestedCoins = pva.GetVestedCoins(now.Add(48 * time.Hour))
	require.Equal(t, origCoins, vestedCoins)
}

func TestOverflowAndNegativeVestedCoinsPeriods(t *testing.T) {
	now := tmtime.Now()
	tests := []struct {
		name    string
		periods []types.Period
		wantErr string
	}{
		{
			"negative .Length",
			types.Periods{
				types.Period{Length: -1, Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
				types.Period{Length: 6 * 60 * 60, Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
			},
			"period #0 has a negative length: -1",
		},
		{
			"overflow after .Length additions",
			types.Periods{
				types.Period{Length: 9223372036854775108, Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
				types.Period{Length: 6 * 60 * 60, Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
			},
			"vesting start-time cannot be before end-time", // it overflow to a negative number, making start-time > end-time
		},
		{
			"good periods that are not negative nor overflow",
			types.Periods{
				types.Period{Length: now.Unix() - 1000, Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
				types.Period{Length: 60, Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
				types.Period{Length: 30, Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
			},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bacc, origCoins := initBaseAccount()
			pva, err := types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), tt.periods)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			if pbva := pva.BaseVestingAccount; pbva.EndTime < 0 {
				t.Fatalf("Unfortunately we still have negative .EndTime :-(: %d", pbva.EndTime)
			}
		})
	}
}

func TestGetVestingCoinsPeriodicVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	periods := types.Periods{
		types.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	bacc, origCoins := initBaseAccount()
	pva, err := types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	require.NoError(t, err)

	// require all coins vesting at the beginning of the vesting schedule
	vestingCoins := pva.GetVestingCoins(now)
	require.Equal(t, origCoins, vestingCoins)

	// require no coins vesting at the end of the vesting schedule
	vestingCoins = pva.GetVestingCoins(endTime)
	require.Equal(t, emptyCoins, vestingCoins)

	// require 50% of coins vesting
	vestingCoins = pva.GetVestingCoins(now.Add(12 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestingCoins)

	// require 50% of coins vesting after period 1, but before period 2 completes.
	vestingCoins = pva.GetVestingCoins(now.Add(15 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestingCoins)

	// require 25% of coins vesting after period 2
	vestingCoins = pva.GetVestingCoins(now.Add(18 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}, vestingCoins)

	// require 0% of coins vesting after vesting complete
	vestingCoins = pva.GetVestingCoins(now.Add(48 * time.Hour))
	require.Equal(t, emptyCoins, vestingCoins)
}

func TestSpendableCoinsPeriodicVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	periods := types.Periods{
		types.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	bacc, origCoins := initBaseAccount()
	pva, err := types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	require.NoError(t, err)

	// require that there exist no spendable coins at the beginning of the
	// vesting schedule
	lockedCoins := pva.LockedCoins(now)
	require.Equal(t, origCoins, lockedCoins)

	// require that all original coins are spendable at the end of the vesting
	// schedule
	lockedCoins = pva.LockedCoins(endTime)
	require.Equal(t, sdk.NewCoins(), lockedCoins)

	// require that all still vesting coins (50%) are locked
	lockedCoins = pva.LockedCoins(now.Add(12 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, lockedCoins)
}

func TestTrackDelegationPeriodicVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	periods := types.Periods{
		types.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	bacc, origCoins := initBaseAccount()

	// require the ability to delegate all vesting coins
	pva, err := types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	require.NoError(t, err)
	pva.TrackDelegation(now, origCoins, origCoins)
	require.Equal(t, origCoins, pva.DelegatedVesting)
	require.Nil(t, pva.DelegatedFree)

	// require the ability to delegate all vested coins
	pva, err = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	require.NoError(t, err)
	pva.TrackDelegation(endTime, origCoins, origCoins)
	require.Nil(t, pva.DelegatedVesting)
	require.Equal(t, origCoins, pva.DelegatedFree)

	// delegate half of vesting coins
	pva, err = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	require.NoError(t, err)
	pva.TrackDelegation(now, origCoins, periods[0].Amount)
	// require that all delegated coins are delegated vesting
	require.Equal(t, pva.DelegatedVesting, periods[0].Amount)
	require.Nil(t, pva.DelegatedFree)

	// delegate 75% of coins, split between vested and vesting
	pva, err = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	require.NoError(t, err)
	pva.TrackDelegation(now.Add(12*time.Hour), origCoins, periods[0].Amount.Add(periods[1].Amount...))
	// require that the maximum possible amount of vesting coins are chosen for delegation.
	require.Equal(t, pva.DelegatedFree, periods[1].Amount)
	require.Equal(t, pva.DelegatedVesting, periods[0].Amount)

	// require the ability to delegate all vesting coins (50%) and all vested coins (50%)
	pva, err = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	require.NoError(t, err)
	pva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, pva.DelegatedVesting)
	require.Nil(t, pva.DelegatedFree)

	pva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, pva.DelegatedVesting)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, pva.DelegatedFree)

	// require no modifications when delegation amount is zero or not enough funds
	pva, err = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	require.NoError(t, err)
	require.Panics(t, func() {
		pva.TrackDelegation(endTime, origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 1000000)})
	})
	require.Nil(t, pva.DelegatedVesting)
	require.Nil(t, pva.DelegatedFree)
}

func TestTrackUndelegationPeriodicVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	periods := types.Periods{
		types.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	bacc, origCoins := initBaseAccount()

	// require the ability to undelegate all vesting coins at the beginning of vesting
	pva, err := types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	require.NoError(t, err)
	pva.TrackDelegation(now, origCoins, origCoins)
	pva.TrackUndelegation(origCoins)
	require.Nil(t, pva.DelegatedFree)
	require.Equal(t, emptyCoins, pva.DelegatedVesting)

	// require the ability to undelegate all vested coins at the end of vesting
	pva, err = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	require.NoError(t, err)
	pva.TrackDelegation(endTime, origCoins, origCoins)
	pva.TrackUndelegation(origCoins)
	require.Equal(t, emptyCoins, pva.DelegatedFree)
	require.Nil(t, pva.DelegatedVesting)

	// require the ability to undelegate half of coins
	pva, err = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	require.NoError(t, err)
	pva.TrackDelegation(endTime, origCoins, periods[0].Amount)
	pva.TrackUndelegation(periods[0].Amount)
	require.Equal(t, emptyCoins, pva.DelegatedFree)
	require.Nil(t, pva.DelegatedVesting)

	// require no modifications when the undelegation amount is zero
	pva, err = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	require.NoError(t, err)
	require.Panics(t, func() {
		pva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 0)})
	})
	require.Nil(t, pva.DelegatedFree)
	require.Nil(t, pva.DelegatedVesting)

	// vest 50% and delegate to two validators
	pva, err = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	require.NoError(t, err)
	pva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	pva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})

	// undelegate from one validator that got slashed 50%
	pva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)}, pva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, pva.DelegatedVesting)

	// undelegate from the other validator that did not get slashed
	pva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Equal(t, emptyCoins, pva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)}, pva.DelegatedVesting)
}

func TestGetVestedCoinsPermLockedVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(1000 * 24 * time.Hour)

	bacc, origCoins := initBaseAccount()

	// require no coins are vested
	plva, err := types.NewPermanentLockedAccount(bacc, origCoins)
	require.NoError(t, err)
	vestedCoins := plva.GetVestedCoins(now)
	require.Nil(t, vestedCoins)

	// require no coins be vested at end time
	vestedCoins = plva.GetVestedCoins(endTime)
	require.Nil(t, vestedCoins)
}

func TestGetVestingCoinsPermLockedVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(1000 * 24 * time.Hour)

	bacc, origCoins := initBaseAccount()

	// require all coins vesting at the beginning of the schedule
	plva, err := types.NewPermanentLockedAccount(bacc, origCoins)
	require.NoError(t, err)
	vestingCoins := plva.GetVestingCoins(now)
	require.Equal(t, origCoins, vestingCoins)

	// require all coins vesting at the end time
	vestingCoins = plva.GetVestingCoins(endTime)
	require.Equal(t, origCoins, vestingCoins)
}

func TestSpendableCoinsPermLockedVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(1000 * 24 * time.Hour)

	bacc, origCoins := initBaseAccount()

	// require that all coins are locked in the beginning of the vesting
	// schedule
	plva, err := types.NewPermanentLockedAccount(bacc, origCoins)
	require.NoError(t, err)
	lockedCoins := plva.LockedCoins(now)
	require.True(t, lockedCoins.Equal(origCoins))

	// require that all coins are still locked at end time
	lockedCoins = plva.LockedCoins(endTime)
	require.True(t, lockedCoins.Equal(origCoins))

	// delegate some locked coins
	// require that locked is reduced
	delegatedAmount := sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 50))
	plva.TrackDelegation(now.Add(12*time.Hour), origCoins, delegatedAmount)
	lockedCoins = plva.LockedCoins(now.Add(12 * time.Hour))
	require.True(t, lockedCoins.Equal(origCoins.Sub(delegatedAmount...)))
}

func TestTrackDelegationPermLockedVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(1000 * 24 * time.Hour)

	bacc, origCoins := initBaseAccount()

	// require the ability to delegate all vesting coins
	plva, err := types.NewPermanentLockedAccount(bacc, origCoins)
	require.NoError(t, err)
	plva.TrackDelegation(now, origCoins, origCoins)
	require.Equal(t, origCoins, plva.DelegatedVesting)
	require.Nil(t, plva.DelegatedFree)

	// require the ability to delegate all vested coins at endTime
	plva, err = types.NewPermanentLockedAccount(bacc, origCoins)
	require.NoError(t, err)
	plva.TrackDelegation(endTime, origCoins, origCoins)
	require.Equal(t, origCoins, plva.DelegatedVesting)
	require.Nil(t, plva.DelegatedFree)

	// require no modifications when delegation amount is zero or not enough funds
	plva, err = types.NewPermanentLockedAccount(bacc, origCoins)
	require.NoError(t, err)
	require.Panics(t, func() {
		plva.TrackDelegation(endTime, origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 1000000)})
	})
	require.Nil(t, plva.DelegatedVesting)
	require.Nil(t, plva.DelegatedFree)
}

func TestTrackUndelegationPermLockedVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(1000 * 24 * time.Hour)

	bacc, origCoins := initBaseAccount()

	// require the ability to undelegate all vesting coins
	plva, err := types.NewPermanentLockedAccount(bacc, origCoins)
	require.NoError(t, err)
	plva.TrackDelegation(now, origCoins, origCoins)
	plva.TrackUndelegation(origCoins)
	require.Nil(t, plva.DelegatedFree)
	require.Equal(t, emptyCoins, plva.DelegatedVesting)

	// require the ability to undelegate all vesting coins at endTime
	plva, err = types.NewPermanentLockedAccount(bacc, origCoins)
	require.NoError(t, err)
	plva.TrackDelegation(endTime, origCoins, origCoins)
	plva.TrackUndelegation(origCoins)
	require.Nil(t, plva.DelegatedFree)
	require.Equal(t, emptyCoins, plva.DelegatedVesting)

	// require no modifications when the undelegation amount is zero
	plva, err = types.NewPermanentLockedAccount(bacc, origCoins)
	require.NoError(t, err)
	require.Panics(t, func() {
		plva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 0)})
	})
	require.Nil(t, plva.DelegatedFree)
	require.Nil(t, plva.DelegatedVesting)

	// delegate to two validators
	plva, err = types.NewPermanentLockedAccount(bacc, origCoins)
	require.NoError(t, err)
	plva.TrackDelegation(now, origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	plva.TrackDelegation(now, origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})

	// undelegate from one validator that got slashed 50%
	plva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)})

	require.Nil(t, plva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 75)}, plva.DelegatedVesting)

	// undelegate from the other validator that did not get slashed
	plva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Nil(t, plva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)}, plva.DelegatedVesting)
}

func TestGenesisAccountValidate(t *testing.T) {
	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	baseAcc := authtypes.NewBaseAccount(addr, pubkey, 0, 0)
	initialVesting := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 50))
	baseVestingWithCoins, err := types.NewBaseVestingAccount(baseAcc, initialVesting, 100)
	require.NoError(t, err)
	tests := []struct {
		name   string
		acc    authtypes.GenesisAccount
		expErr bool
	}{
		{
			"valid base account",
			baseAcc,
			false,
		},
		{
			"invalid base valid account",
			authtypes.NewBaseAccount(addr, secp256k1.GenPrivKey().PubKey(), 0, 0),
			true,
		},
		{
			"valid base vesting account",
			baseVestingWithCoins,
			false,
		},
		{
			"valid continuous vesting account",
			func() authtypes.GenesisAccount {
				acc, _ := types.NewContinuousVestingAccount(baseAcc, initialVesting, 100, 200)
				return acc
			}(),
			false,
		},
		{
			"invalid vesting times",
			func() authtypes.GenesisAccount {
				acc, _ := types.NewContinuousVestingAccount(baseAcc, initialVesting, 1654668078, 1554668078)
				return acc
			}(),
			true,
		},
		{
			"valid periodic vesting account",
			func() authtypes.GenesisAccount {
				acc, _ := types.NewPeriodicVestingAccount(baseAcc, initialVesting, 0, types.Periods{types.Period{Length: int64(100), Amount: sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 50)}}})
				return acc
			}(),
			false,
		},
		{
			"invalid vesting period lengths",
			types.NewPeriodicVestingAccountRaw(
				baseVestingWithCoins,
				0, types.Periods{types.Period{Length: int64(50), Amount: sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 50)}}}),
			true,
		},
		{
			"invalid vesting period amounts",
			types.NewPeriodicVestingAccountRaw(
				baseVestingWithCoins,
				0, types.Periods{types.Period{Length: int64(100), Amount: sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 25)}}}),
			true,
		},
		{
			"valid permanent locked vesting account",
			func() authtypes.GenesisAccount {
				acc, _ := types.NewPermanentLockedAccount(baseAcc, initialVesting)
				return acc
			}(),
			false,
		},
		{
			"invalid positive end time for permanently locked vest account",
			&types.PermanentLockedAccount{BaseVestingAccount: baseVestingWithCoins},
			true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expErr, tt.acc.Validate() != nil)
		})
	}
}

func initBaseAccount() (*authtypes.BaseAccount, sdk.Coins) {
	_, _, addr := testdata.KeyTestPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := authtypes.NewBaseAccountWithAddress(addr)

	return bacc, origCoins
}

func TestVestingAccountTestSuite(t *testing.T) {
	suite.Run(t, new(VestingAccountTestSuite))
}

func TestUpdateScheduleContinuousVestingAcc(t *testing.T) {
	now := tmtime.Now()
	bacc, _ := initBaseAccount()

	testCases := []struct {
		name            string
		startTime       int64
		endTime         int64
		originalVesting sdk.Coins
		rewardCoins     sdk.Coins
		testTime        int64     // Time at which UpdateSchedule is called
		expectedVesting sdk.Coins // Expected OriginalVesting *after* update
		expectedEndTime int64     // EndTime should not change
		expectError     bool      // Error expected from New... or UpdateSchedule
	}{
		{
			name:            "update halfway through vesting period",
			startTime:       now.Unix(),
			endTime:         now.Add(24 * time.Hour).Unix(),
			originalVesting: sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 1000)),
			rewardCoins:     sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 100)),
			testTime:        now.Add(12 * time.Hour).Unix(),
			// Expected: 1000 (original) + 100 (full reward) = 1100
			expectedVesting: sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 1100)),
			expectedEndTime: now.Add(24 * time.Hour).Unix(),
			expectError:     false,
		},
		{
			name:            "update 75% through vesting period",
			startTime:       now.Unix(),
			endTime:         now.Add(24 * time.Hour).Unix(),
			originalVesting: sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 1000)),
			rewardCoins:     sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 100)),
			testTime:        now.Add(18 * time.Hour).Unix(),
			// Expected: 1000 (original) + 100 (full reward) = 1100
			expectedVesting: sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 1100)),
			expectedEndTime: now.Add(24 * time.Hour).Unix(),
			expectError:     false,
		},
		{
			name:            "update after vesting period completed - should error",
			startTime:       now.Add(-48 * time.Hour).Unix(),
			endTime:         now.Add(-24 * time.Hour).Unix(),
			originalVesting: sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 1000)),
			rewardCoins:     sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 100)),
			testTime:        now.Unix(), // Test time is after end time
			// Expected: UpdateSchedule should error, OriginalVesting remains 1000
			expectedVesting: sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 1000)),
			expectedEndTime: now.Add(-24 * time.Hour).Unix(),
			expectError:     true, // Expect error from UpdateSchedule
		},
		{
			name:            "update at exactly the vesting end time - should error",
			startTime:       now.Add(-24 * time.Hour).Unix(),
			endTime:         now.Unix(),
			originalVesting: sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 1000)),
			rewardCoins:     sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 100)),
			testTime:        now.Unix(), // Test time is exactly end time
			// Expected: UpdateSchedule should error, OriginalVesting remains 1000
			expectedVesting: sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 1000)),
			expectedEndTime: now.Unix(),
			expectError:     true, // Expect error from UpdateSchedule
		},
		{
			name:            "update before start time",
			startTime:       now.Add(1 * time.Hour).Unix(),
			endTime:         now.Add(25 * time.Hour).Unix(),
			originalVesting: sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 1000)),
			rewardCoins:     sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 100)),
			testTime:        now.Unix(), // Test time is before start time
			// Expected: 1000 (original) + 100 (full reward) = 1100
			expectedVesting: sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 1100)),
			expectedEndTime: now.Add(25 * time.Hour).Unix(),
			expectError:     false,
		},
		{
			name:            "update at exactly the start time",
			startTime:       now.Unix(),
			endTime:         now.Add(24 * time.Hour).Unix(),
			originalVesting: sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 1000)),
			rewardCoins:     sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 100)),
			testTime:        now.Unix(), // Test time is exactly start time
			// Expected: 1000 (original) + 100 (full reward) = 1100
			expectedVesting: sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 1100)),
			expectedEndTime: now.Add(24 * time.Hour).Unix(),
			expectError:     false,
		},
		{
			name:            "multiple denominations, update halfway",
			startTime:       now.Unix(),
			endTime:         now.Add(24 * time.Hour).Unix(),
			originalVesting: sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 1000), sdk.NewInt64Coin(feeDenom, 500)),
			rewardCoins:     sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 100), sdk.NewInt64Coin(feeDenom, 50)),
			testTime:        now.Add(12 * time.Hour).Unix(),
			// Expected: stake: 1000 + 100 = 1100, fee: 500 + 50 = 550
			expectedVesting: sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 1100), sdk.NewInt64Coin(feeDenom, 550)),
			expectedEndTime: now.Add(24 * time.Hour).Unix(),
			expectError:     false,
		},
		{
			name:            "rewards contain denom not in original vesting",
			startTime:       now.Unix(),
			endTime:         now.Add(24 * time.Hour).Unix(),
			originalVesting: sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 1000)),
			rewardCoins:     sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 100), sdk.NewInt64Coin(feeDenom, 50)), // feeDenom not in original
			testTime:        now.Add(12 * time.Hour).Unix(),
			// Expected: stake: 1000 + 100 = 1100. Fee denom ignored.
			expectedVesting: sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 1100)),
			expectedEndTime: now.Add(24 * time.Hour).Unix(),
			expectError:     false,
		},
		{
			name:            "zero rewards",
			startTime:       now.Unix(),
			endTime:         now.Add(24 * time.Hour).Unix(),
			originalVesting: sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 1000)),
			rewardCoins:     sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 0)),
			testTime:        now.Add(12 * time.Hour).Unix(),
			expectedVesting: sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 1000)), // No change
			expectedEndTime: now.Add(24 * time.Hour).Unix(),
			expectError:     false,
		},
		{
			name:            "nil rewards",
			startTime:       now.Unix(),
			endTime:         now.Add(24 * time.Hour).Unix(),
			originalVesting: sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 1000)),
			rewardCoins:     nil,
			testTime:        now.Add(12 * time.Hour).Unix(),
			expectedVesting: sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 1000)), // No change
			expectedEndTime: now.Add(24 * time.Hour).Unix(),
			expectError:     false,
		},
		{
			name:            "start time equals end time (zero duration) - should error on creation",
			startTime:       now.Unix(),
			endTime:         now.Unix(), // Zero duration
			originalVesting: sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 1000)),
			rewardCoins:     sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 100)),
			testTime:        now.Add(-1 * time.Hour).Unix(),
			// Expected: Error during NewContinuousVestingAccount due to zero duration
			expectedVesting: nil, // Not relevant as creation fails
			expectedEndTime: now.Unix(),
			expectError:     true, // Expect error from NewContinuousVestingAccount
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cva, err := types.NewContinuousVestingAccount(bacc, tc.originalVesting, tc.startTime, tc.endTime)
			if tc.expectError && err != nil {
				// If we expected an error from New... and got one, test passes
				require.Error(t, err)
				return
			} else if err != nil {
				// If we got an error from New... but didn't expect one
				require.NoError(t, err, "Unexpected error during account creation")
			}

			// Store original vesting before update for comparison in post-update checks
			originalVestingBeforeUpdate := cva.GetOriginalVesting()

			// Update the vesting schedule
			err = cva.UpdateSchedule(time.Unix(tc.testTime, 0), tc.rewardCoins)
			if tc.expectError {
				// Check that OriginalVesting did NOT change
				require.Equal(t, originalVestingBeforeUpdate, cva.OriginalVesting, "OriginalVesting should not change on failed update")
				return
			}

			// If no error was expected from UpdateSchedule
			require.NoError(t, err, "Unexpected error during UpdateSchedule")

			// Verify results after successful update
			require.Equal(t, tc.expectedVesting, cva.OriginalVesting, "OriginalVesting mismatch after update")
			require.Equal(t, tc.expectedEndTime, cva.EndTime, "EndTime mismatch")

			// Verify GetVestedCoins logic still works correctly based on the *new* original vesting amount
			// Check at the test time itself
			currentVested := cva.GetVestedCoins(time.Unix(tc.testTime, 0))

			// Check at a future time (e.g., end time) - only if duration > 0
			if tc.endTime > tc.startTime {
				futureVested := cva.GetVestedCoins(time.Unix(tc.endTime, 0))
				// At the end time, the vested amount should equal the *new* original vesting total
				require.Equal(t, tc.expectedVesting, futureVested, "Vesting at end time should equal the new original vesting")

				// Check that vesting progresses correctly after the update
				if tc.testTime < tc.endTime {
					midPointTime := tc.testTime + (tc.endTime-tc.testTime)/2
					midPointVested := cva.GetVestedCoins(time.Unix(midPointTime, 0))

					// Vested amount should be >= current vested amount (unless currentVested is nil due to testTime < startTime)
					if currentVested != nil {
						require.True(t, midPointVested.IsAllGTE(currentVested), "Vested coins should not decrease over time after update (%v vs %v)", currentVested, midPointVested)
					}

					// If the schedule was actually updated (expectedVesting > originalVestingBeforeUpdate)
					// and the midpoint is after the start time, check that more coins vested than at testTime.
					if !originalVestingBeforeUpdate.Equal(tc.expectedVesting) && midPointTime > tc.startTime {
						// Get vested coins based on OLD schedule at midpoint
						oldScalar := math.LegacyNewDec(midPointTime - tc.startTime).Quo(math.LegacyNewDec(tc.endTime - tc.startTime))
						var oldMidPointVested sdk.Coins
						for _, ovc := range originalVestingBeforeUpdate {
							vestedAmt := math.LegacyNewDecFromInt(ovc.Amount).Mul(oldScalar).RoundInt()
							oldMidPointVested = oldMidPointVested.Add(sdk.NewCoin(ovc.Denom, vestedAmt))
						}

						// New vested amount at midpoint should be greater than old vested amount at midpoint
						require.True(t, midPointVested.IsAnyGT(oldMidPointVested), "Expected more coins to vest at midpoint after update (Old: %v, New: %v)", oldMidPointVested, midPointVested)
					}
				}
			}
		})
	}
}

func TestUpdateScheduleDelayedVestingAcc(t *testing.T) {
	now := tmtime.Now()
	bacc, initialOrigCoins := initBaseAccount()
	rewardCoins := sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 100))
	feeReward := sdk.NewInt64Coin(feeDenom, 50)

	testCases := []struct {
		name            string
		endTime         int64
		originalVesting sdk.Coins
		rewardCoins     sdk.Coins
		testTime        int64 // Time at which UpdateSchedule is called
		expectedVesting sdk.Coins
		expectedEndTime int64 // EndTime should not change
	}{
		{
			name:            "update before end time",
			endTime:         now.Add(24 * time.Hour).Unix(),
			originalVesting: initialOrigCoins,
			rewardCoins:     rewardCoins,
			testTime:        now.Unix(),
			// Expected: Add full reward amount as testTime < endTime
			expectedVesting: initialOrigCoins.Add(rewardCoins...), // stake: 100 + 100 = 200
			expectedEndTime: now.Add(24 * time.Hour).Unix(),
		},
		{
			name:            "update at end time",
			endTime:         now.Add(24 * time.Hour).Unix(),
			originalVesting: initialOrigCoins,
			rewardCoins:     rewardCoins,
			testTime:        now.Add(24 * time.Hour).Unix(), // Exactly end time
			// Expected: No change as testTime >= endTime
			expectedVesting: initialOrigCoins,
			expectedEndTime: now.Add(24 * time.Hour).Unix(),
		},
		{
			name:            "update after end time",
			endTime:         now.Add(-1 * time.Hour).Unix(), // End time already passed
			originalVesting: initialOrigCoins,
			rewardCoins:     rewardCoins,
			testTime:        now.Unix(),
			// Expected: No change as testTime >= endTime
			expectedVesting: initialOrigCoins,
			expectedEndTime: now.Add(-1 * time.Hour).Unix(),
		},
		{
			name:            "zero rewards, update before end time",
			endTime:         now.Add(24 * time.Hour).Unix(),
			originalVesting: initialOrigCoins,
			rewardCoins:     sdk.NewCoins(),
			testTime:        now.Unix(),
			expectedVesting: initialOrigCoins, // No change
			expectedEndTime: now.Add(24 * time.Hour).Unix(),
		},
		{
			name:            "nil rewards, update before end time",
			endTime:         now.Add(24 * time.Hour).Unix(),
			originalVesting: initialOrigCoins,
			rewardCoins:     nil,
			testTime:        now.Unix(),
			expectedVesting: initialOrigCoins, // No change
			expectedEndTime: now.Add(24 * time.Hour).Unix(),
		},
		{
			name:            "rewards contain denom not in original vesting",
			endTime:         now.Add(24 * time.Hour).Unix(),
			originalVesting: sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 100)),            // Only stake
			rewardCoins:     sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 100), feeReward), // stake + fee
			testTime:        now.Unix(),
			// Expected: Add only stake reward as fee is not in original vesting
			expectedVesting: sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 200)),
			expectedEndTime: now.Add(24 * time.Hour).Unix(),
		},
		{
			name:            "multiple denoms, update before end time",
			endTime:         now.Add(24 * time.Hour).Unix(),
			originalVesting: initialOrigCoins, // stake + fee
			rewardCoins:     sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 100), feeReward),
			testTime:        now.Unix(),
			// Expected: Add both rewards
			expectedVesting: initialOrigCoins.Add(sdk.NewCoin(stakeDenom, math.NewInt(100))).Add(feeReward),
			expectedEndTime: now.Add(24 * time.Hour).Unix(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dva, err := types.NewDelayedVestingAccount(bacc, tc.originalVesting, tc.endTime)
			require.NoError(t, err)

			// Update schedule
			err = dva.UpdateSchedule(time.Unix(tc.testTime, 0), tc.rewardCoins)
			require.NoError(t, err)

			// Verify results
			require.Equal(t, tc.expectedVesting, dva.OriginalVesting, "OriginalVesting mismatch")
			require.Equal(t, tc.expectedEndTime, dva.EndTime, "EndTime mismatch")

			// Verify GetVestedCoins logic still works correctly based on the *new* original vesting amount
			vestedAtEndTime := dva.GetVestedCoins(time.Unix(tc.endTime, 0))
			require.Equal(t, tc.expectedVesting, vestedAtEndTime, "Vested coins at end time should equal the new original vesting")
			vestedBeforeEndTime := dva.GetVestedCoins(time.Unix(tc.endTime-1, 0))
			require.Nil(t, vestedBeforeEndTime, "No coins should be vested before end time")
		})
	}
}

// TestGetVestedCoinsAfterMultipleUpdates verifies that GetVestedCoins returns
// the correct amounts after multiple UpdateSchedule calls, including consecutive
// updates at the same time point, according to the *current* implementation
// where the full reward is added to OriginalVesting for existing denominations.
func TestGetVestedCoinsAfterMultipleUpdates(t *testing.T) {
	now := tmtime.Now()
	bacc, _ := initBaseAccount()

	// Set up a continuous vesting account with a 100-second vesting period
	initialVesting := sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 1000))
	startTime := now.Unix()
	duration := int64(100) // 100 seconds vesting duration
	endTime := startTime + duration

	// Create the continuous vesting account
	cva, err := types.NewContinuousVestingAccount(bacc, initialVesting, startTime, endTime)
	require.NoError(t, err)

	// Define checkpoints at 0%, 25%, 50%, 75%, and 100% of vesting period
	checkpoints := []struct {
		offsetPercent int64
		offsetSeconds int64
		label         string
	}{
		{0, 0, "start (0%)"},
		{25, 25, "25%"},
		{50, 50, "50%"},
		{75, 75, "75%"},
		{100, 100, "end (100%)"},
	}

	// STEP 1: Check initial vested coins at each checkpoint
	for _, cp := range checkpoints {
		checkpointTime := time.Unix(startTime+cp.offsetSeconds, 0)
		actual := cva.GetVestedCoins(checkpointTime)

		if cp.offsetPercent == 0 {
			require.Nil(t, actual, "At start, vested coins should be nil")
		} else {
			// Hardcoded expected values based on 1000 total
			expectedValues := map[int64]int64{
				25:  250,  // 25% of 1000
				50:  500,  // 50% of 1000
				75:  750,  // 75% of 1000
				100: 1000, // 100% of 1000
			}
			expected := sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, expectedValues[cp.offsetPercent]))
			require.Equal(t, expected, actual,
				"Initial vested coins at %s mismatch. Expected: %v, Got: %v",
				cp.label, expected, actual)
		}
	}

	// STEP 2: First update at 25% mark (Add full 400 tokens)
	update1Time := time.Unix(startTime+25, 0)
	update1Amount := sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 400))
	err = cva.UpdateSchedule(update1Time, update1Amount)
	require.NoError(t, err)

	// Verify total after first update: 1000 + 400 = 1400 tokens
	expectedTotal1 := int64(1400)
	actualTotal1 := cva.GetOriginalVesting().AmountOf(stakeDenom).Int64()
	require.Equal(t, expectedTotal1, actualTotal1,
		"OriginalVesting mismatch after 25% update. Expected: %v, Got: %v",
		expectedTotal1, actualTotal1)

	// Check vested coins at all checkpoints at or after the 25% mark
	for _, cp := range checkpoints {
		// Skip checkpoints before the update point
		if cp.offsetSeconds < 25 {
			continue
		}

		checkpointTime := time.Unix(startTime+cp.offsetSeconds, 0)
		actual := cva.GetVestedCoins(checkpointTime)

		// Hardcoded expected values based on 1400 total
		expectedValues := map[int64]int64{
			25:  350,  // 25% of 1400
			50:  700,  // 50% of 1400
			75:  1050, // 75% of 1400
			100: 1400, // 100% of 1400
		}
		expected := sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, expectedValues[cp.offsetPercent]))
		require.Equal(t, expected, actual,
			"After 25% update, vested coins at %s mismatch. Expected: %v, Got: %v",
			cp.label, expected, actual)
	}

	// STEP 3: Second update at 50% mark (Add full 400 tokens)
	update2Time := time.Unix(startTime+50, 0)
	update2Amount := sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 400))
	err = cva.UpdateSchedule(update2Time, update2Amount)
	require.NoError(t, err)

	// Verify total after second update: 1400 + 400 = 1800 tokens
	expectedTotal2 := int64(1800)
	actualTotal2 := cva.GetOriginalVesting().AmountOf(stakeDenom).Int64()
	require.Equal(t, expectedTotal2, actualTotal2,
		"OriginalVesting mismatch after first 50% update. Expected: %v, Got: %v",
		expectedTotal2, actualTotal2)

	// Check vested coins at all checkpoints at or after the 50% mark
	for _, cp := range checkpoints {
		// Skip checkpoints before the update point
		if cp.offsetSeconds < 50 {
			continue
		}

		checkpointTime := time.Unix(startTime+cp.offsetSeconds, 0)
		actual := cva.GetVestedCoins(checkpointTime)

		// Hardcoded expected values based on 1800 total
		expectedValues := map[int64]int64{
			50:  900,  // 50% of 1800
			75:  1350, // 75% of 1800
			100: 1800, // 100% of 1800
		}
		expected := sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, expectedValues[cp.offsetPercent]))
		require.Equal(t, expected, actual,
			"After first 50% update, vested coins at %s mismatch. Expected: %v, Got: %v",
			cp.label, expected, actual)
	}

	// STEP 4: Third update at 50% mark (Add full 200 tokens)
	update3Time := time.Unix(startTime+50, 0) // Same timestamp as previous update
	update3Amount := sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 200))
	err = cva.UpdateSchedule(update3Time, update3Amount)
	require.NoError(t, err)

	// Verify total after third update: 1800 + 200 = 2000 tokens
	expectedTotal3 := int64(2000)
	actualTotal3 := cva.GetOriginalVesting().AmountOf(stakeDenom).Int64()
	require.Equal(t, expectedTotal3, actualTotal3,
		"OriginalVesting mismatch after second 50% update. Expected: %v, Got: %v",
		expectedTotal3, actualTotal3)

	// Check vested coins at all checkpoints at or after the 50% mark
	for _, cp := range checkpoints {
		// Skip checkpoints before the update point
		if cp.offsetSeconds < 50 {
			continue
		}

		checkpointTime := time.Unix(startTime+cp.offsetSeconds, 0)
		actual := cva.GetVestedCoins(checkpointTime)

		// Hardcoded expected values based on 2000 total
		expectedValues := map[int64]int64{
			50:  1000, // 50% of 2000
			75:  1500, // 75% of 2000
			100: 2000, // 100% of 2000
		}
		expected := sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, expectedValues[cp.offsetPercent]))
		require.Equal(t, expected, actual,
			"After second 50% update, vested coins at %s mismatch. Expected: %v, Got: %v",
			cp.label, expected, actual)
	}

	// STEP 5: Fourth update at 75% mark (Add full 800 tokens)
	update4Time := time.Unix(startTime+75, 0)
	update4Amount := sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 800))
	err = cva.UpdateSchedule(update4Time, update4Amount)
	require.NoError(t, err)

	// Verify total after fourth update: 2000 + 800 = 2800 tokens
	expectedTotal4 := int64(2800)
	actualTotal4 := cva.GetOriginalVesting().AmountOf(stakeDenom).Int64()
	require.Equal(t, expectedTotal4, actualTotal4,
		"OriginalVesting mismatch after 75% update. Expected: %v, Got: %v",
		expectedTotal4, actualTotal4)

	// Check vested coins at all checkpoints at or after the 75% mark
	for _, cp := range checkpoints {
		// Skip checkpoints before the update point
		if cp.offsetSeconds < 75 {
			continue
		}

		checkpointTime := time.Unix(startTime+cp.offsetSeconds, 0)
		actual := cva.GetVestedCoins(checkpointTime)

		// Hardcoded expected values based on 2800 total
		expectedValues := map[int64]int64{
			75:  2100, // 75% of 2800
			100: 2800, // 100% of 2800
		}
		expected := sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, expectedValues[cp.offsetPercent]))
		require.Equal(t, expected, actual,
			"After 75% update, vested coins at %s mismatch. Expected: %v, Got: %v",
			cp.label, expected, actual)
	}

	// BONUS CHECK: Verify that updates after end time have no effect
	afterEndTime := time.Unix(endTime+10, 0)
	afterEndAmount := sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 1000))
	err = cva.UpdateSchedule(afterEndTime, afterEndAmount)
	// Expect an error because the update time is after the vesting end time
	require.Nil(t, err)

	// Original vesting amount should not change after the failed update attempt
	finalTotal := cva.GetOriginalVesting().AmountOf(stakeDenom).Int64()
	require.Equal(t, expectedTotal4, finalTotal,
		"OriginalVesting should not change after failed end time update. Expected: %v, Got: %v",
		expectedTotal4, finalTotal)
}
