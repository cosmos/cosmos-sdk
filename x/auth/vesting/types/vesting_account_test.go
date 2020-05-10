package types_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmtime "github.com/tendermint/tendermint/types/time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

var (
	stakeDenom = "stake"
	feeDenom   = "fee"
)

func TestGetVestedCoinsContVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	_, _, addr := authtypes.KeyTestPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := authtypes.NewBaseAccountWithAddress(addr)
	cva := types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())

	// require no coins vested in the very beginning of the vesting schedule
	vestedCoins := cva.GetVestedCoins(now)
	require.Nil(t, vestedCoins)

	// require all coins vested at the end of the vesting schedule
	vestedCoins = cva.GetVestedCoins(endTime)
	require.Equal(t, origCoins, vestedCoins)

	// require 50% of coins vested
	vestedCoins = cva.GetVestedCoins(now.Add(12 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestedCoins)

	// require 100% of coins vested
	vestedCoins = cva.GetVestedCoins(now.Add(48 * time.Hour))
	require.Equal(t, origCoins, vestedCoins)
}

func TestGetVestingCoinsContVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	_, _, addr := authtypes.KeyTestPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := authtypes.NewBaseAccountWithAddress(addr)
	cva := types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())

	// require all coins vesting in the beginning of the vesting schedule
	vestingCoins := cva.GetVestingCoins(now)
	require.Equal(t, origCoins, vestingCoins)

	// require no coins vesting at the end of the vesting schedule
	vestingCoins = cva.GetVestingCoins(endTime)
	require.Nil(t, vestingCoins)

	// require 50% of coins vesting
	vestingCoins = cva.GetVestingCoins(now.Add(12 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, vestingCoins)
}

func TestSpendableCoinsContVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	_, _, addr := authtypes.KeyTestPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := authtypes.NewBaseAccountWithAddress(addr)

	cva := types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())

	// require that all original coins are locked at the end of the vesting
	// schedule
	lockedCoins := cva.LockedCoins(now)
	require.Equal(t, origCoins, lockedCoins)

	// require that there exist no locked coins in the beginning of the
	lockedCoins = cva.LockedCoins(endTime)
	require.Equal(t, sdk.NewCoins(), lockedCoins)

	// require that all vested coins (50%) are spendable
	lockedCoins = cva.LockedCoins(now.Add(12 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, lockedCoins)

	// require that all vested coins (50%) are spendable plus any received
	lockedCoins = cva.LockedCoins(now.Add(12 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, lockedCoins)
}

func TestTrackDelegationContVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	_, _, addr := authtypes.KeyTestPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := authtypes.NewBaseAccountWithAddress(addr)

	// require the ability to delegate all vesting coins
	cva := types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	cva.TrackDelegation(now, origCoins, origCoins)
	require.Equal(t, origCoins, cva.DelegatedVesting)
	require.Nil(t, cva.DelegatedFree)

	// require the ability to delegate all vested coins
	cva = types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	cva.TrackDelegation(endTime, origCoins, origCoins)
	require.Nil(t, cva.DelegatedVesting)
	require.Equal(t, origCoins, cva.DelegatedFree)

	// require the ability to delegate all vesting coins (50%) and all vested coins (50%)
	cva = types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	cva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, cva.DelegatedVesting)
	require.Nil(t, cva.DelegatedFree)

	cva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, cva.DelegatedVesting)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, cva.DelegatedFree)

	// require no modifications when delegation amount is zero or not enough funds
	cva = types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	require.Panics(t, func() {
		cva.TrackDelegation(endTime, origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 1000000)})
	})
	require.Nil(t, cva.DelegatedVesting)
	require.Nil(t, cva.DelegatedFree)
}

func TestTrackUndelegationContVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	_, _, addr := authtypes.KeyTestPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := authtypes.NewBaseAccountWithAddress(addr)

	// require the ability to undelegate all vesting coins
	cva := types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	cva.TrackDelegation(now, origCoins, origCoins)
	cva.TrackUndelegation(origCoins)
	require.Nil(t, cva.DelegatedFree)
	require.Nil(t, cva.DelegatedVesting)

	// require the ability to undelegate all vested coins
	cva = types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())

	cva.TrackDelegation(endTime, origCoins, origCoins)
	cva.TrackUndelegation(origCoins)
	require.Nil(t, cva.DelegatedFree)
	require.Nil(t, cva.DelegatedVesting)

	// require no modifications when the undelegation amount is zero
	cva = types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())

	require.Panics(t, func() {
		cva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 0)})
	})
	require.Nil(t, cva.DelegatedFree)
	require.Nil(t, cva.DelegatedVesting)

	// vest 50% and delegate to two validators
	cva = types.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())
	cva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	cva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})

	// undelegate from one validator that got slashed 50%
	cva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)}, cva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, cva.DelegatedVesting)

	// undelegate from the other validator that did not get slashed
	cva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Nil(t, cva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)}, cva.DelegatedVesting)
}

func TestGetVestedCoinsDelVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	_, _, addr := authtypes.KeyTestPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := authtypes.NewBaseAccountWithAddress(addr)

	// require no coins are vested until schedule maturation
	dva := types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	vestedCoins := dva.GetVestedCoins(now)
	require.Nil(t, vestedCoins)

	// require all coins be vested at schedule maturation
	vestedCoins = dva.GetVestedCoins(endTime)
	require.Equal(t, origCoins, vestedCoins)
}

func TestGetVestingCoinsDelVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	_, _, addr := authtypes.KeyTestPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := authtypes.NewBaseAccountWithAddress(addr)

	// require all coins vesting at the beginning of the schedule
	dva := types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	vestingCoins := dva.GetVestingCoins(now)
	require.Equal(t, origCoins, vestingCoins)

	// require no coins vesting at schedule maturation
	vestingCoins = dva.GetVestingCoins(endTime)
	require.Nil(t, vestingCoins)
}

func TestSpendableCoinsDelVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	_, _, addr := authtypes.KeyTestPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := authtypes.NewBaseAccountWithAddress(addr)

	// require that all coins are locked in the beginning of the vesting
	// schedule
	dva := types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	lockedCoins := dva.LockedCoins(now)
	require.True(t, lockedCoins.IsEqual(origCoins))

	// require that all coins are spendable after the maturation of the vesting
	// schedule
	lockedCoins = dva.LockedCoins(endTime)
	require.Equal(t, sdk.NewCoins(), lockedCoins)

	// require that all coins are still vesting after some time
	lockedCoins = dva.LockedCoins(now.Add(12 * time.Hour))
	require.True(t, lockedCoins.IsEqual(origCoins))

	// receive some coins
	// require that only received coins are spendable since the account is still
	// vesting
	lockedCoins = dva.LockedCoins(now.Add(12 * time.Hour))
	require.True(t, lockedCoins.IsEqual(origCoins))

	// delegate some locked coins
	// require that locked is reduced
	delegatedAmount := sdk.NewCoins(sdk.NewInt64Coin(stakeDenom, 50))
	dva.TrackDelegation(now.Add(12*time.Hour), origCoins, delegatedAmount)
	lockedCoins = dva.LockedCoins(now.Add(12 * time.Hour))
	require.True(t, lockedCoins.IsEqual(origCoins.Sub(delegatedAmount)))
}

func TestTrackDelegationDelVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	_, _, addr := authtypes.KeyTestPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := authtypes.NewBaseAccountWithAddress(addr)

	// require the ability to delegate all vesting coins
	dva := types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	dva.TrackDelegation(now, origCoins, origCoins)
	require.Equal(t, origCoins, dva.DelegatedVesting)
	require.Nil(t, dva.DelegatedFree)

	// require the ability to delegate all vested coins
	dva = types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	dva.TrackDelegation(endTime, origCoins, origCoins)
	require.Nil(t, dva.DelegatedVesting)
	require.Equal(t, origCoins, dva.DelegatedFree)

	// require the ability to delegate all coins half way through the vesting
	// schedule
	dva = types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	dva.TrackDelegation(now.Add(12*time.Hour), origCoins, origCoins)
	require.Equal(t, origCoins, dva.DelegatedVesting)
	require.Nil(t, dva.DelegatedFree)

	// require no modifications when delegation amount is zero or not enough funds
	dva = types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())

	require.Panics(t, func() {
		dva.TrackDelegation(endTime, origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 1000000)})
	})
	require.Nil(t, dva.DelegatedVesting)
	require.Nil(t, dva.DelegatedFree)
}

func TestTrackUndelegationDelVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	_, _, addr := authtypes.KeyTestPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := authtypes.NewBaseAccountWithAddress(addr)

	// require the ability to undelegate all vesting coins
	dva := types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	dva.TrackDelegation(now, origCoins, origCoins)
	dva.TrackUndelegation(origCoins)
	require.Nil(t, dva.DelegatedFree)
	require.Nil(t, dva.DelegatedVesting)

	// require the ability to undelegate all vested coins
	dva = types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
	dva.TrackDelegation(endTime, origCoins, origCoins)
	dva.TrackUndelegation(origCoins)
	require.Nil(t, dva.DelegatedFree)
	require.Nil(t, dva.DelegatedVesting)

	// require no modifications when the undelegation amount is zero
	dva = types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())

	require.Panics(t, func() {
		dva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 0)})
	})
	require.Nil(t, dva.DelegatedFree)
	require.Nil(t, dva.DelegatedVesting)

	// vest 50% and delegate to two validators
	dva = types.NewDelayedVestingAccount(bacc, origCoins, endTime.Unix())
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

	_, _, addr := authtypes.KeyTestPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := authtypes.NewBaseAccountWithAddress(addr)
	pva := types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)

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
			sdk.NewInt64Coin(feeDenom, 750), sdk.NewInt64Coin(stakeDenom, 75)}, vestedCoins)

	// require 100% of coins vested
	vestedCoins = pva.GetVestedCoins(now.Add(48 * time.Hour))
	require.Equal(t, origCoins, vestedCoins)
}

func TestGetVestingCoinsPeriodicVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	periods := types.Periods{
		types.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	_, _, addr := authtypes.KeyTestPubAddr()
	origCoins := sdk.Coins{
		sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := authtypes.NewBaseAccountWithAddress(addr)
	pva := types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)

	// require all coins vesting at the beginning of the vesting schedule
	vestingCoins := pva.GetVestingCoins(now)
	require.Equal(t, origCoins, vestingCoins)

	// require no coins vesting at the end of the vesting schedule
	vestingCoins = pva.GetVestingCoins(endTime)
	require.Nil(t, vestingCoins)

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
	require.Nil(t, vestingCoins)
}

func TestSpendableCoinsPeriodicVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	periods := types.Periods{
		types.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		types.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	_, _, addr := authtypes.KeyTestPubAddr()
	origCoins := sdk.Coins{
		sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := authtypes.NewBaseAccountWithAddress(addr)
	pva := types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)

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

	// receive some coins
	// require that all still vesting coins (50% of original) are locked plus any received
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

	_, _, addr := authtypes.KeyTestPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := authtypes.NewBaseAccountWithAddress(addr)

	// require the ability to delegate all vesting coins
	pva := types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	pva.TrackDelegation(now, origCoins, origCoins)
	require.Equal(t, origCoins, pva.DelegatedVesting)
	require.Nil(t, pva.DelegatedFree)

	// require the ability to delegate all vested coins
	pva = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	pva.TrackDelegation(endTime, origCoins, origCoins)
	require.Nil(t, pva.DelegatedVesting)
	require.Equal(t, origCoins, pva.DelegatedFree)

	// delegate half of vesting coins
	pva = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	pva.TrackDelegation(now, origCoins, periods[0].Amount)
	// require that all delegated coins are delegated vesting
	require.Equal(t, pva.DelegatedVesting, periods[0].Amount)
	require.Nil(t, pva.DelegatedFree)

	// delegate 75% of coins, split between vested and vesting
	pva = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	pva.TrackDelegation(now.Add(12*time.Hour), origCoins, periods[0].Amount.Add(periods[1].Amount...))
	// require that the maximum possible amount of vesting coins are chosen for delegation.
	require.Equal(t, pva.DelegatedFree, periods[1].Amount)
	require.Equal(t, pva.DelegatedVesting, periods[0].Amount)

	// require the ability to delegate all vesting coins (50%) and all vested coins (50%)
	pva = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	pva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, pva.DelegatedVesting)
	require.Nil(t, pva.DelegatedFree)

	pva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, pva.DelegatedVesting)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, pva.DelegatedFree)

	// require no modifications when delegation amount is zero or not enough funds
	pva = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
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

	_, _, addr := authtypes.KeyTestPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := authtypes.NewBaseAccountWithAddress(addr)

	// require the ability to undelegate all vesting coins at the beginning of vesting
	pva := types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	pva.TrackDelegation(now, origCoins, origCoins)
	pva.TrackUndelegation(origCoins)
	require.Nil(t, pva.DelegatedFree)
	require.Nil(t, pva.DelegatedVesting)

	// require the ability to undelegate all vested coins at the end of vesting
	pva = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)

	pva.TrackDelegation(endTime, origCoins, origCoins)
	pva.TrackUndelegation(origCoins)
	require.Nil(t, pva.DelegatedFree)
	require.Nil(t, pva.DelegatedVesting)

	// require the ability to undelegate half of coins
	pva = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	pva.TrackDelegation(endTime, origCoins, periods[0].Amount)
	pva.TrackUndelegation(periods[0].Amount)
	require.Nil(t, pva.DelegatedFree)
	require.Nil(t, pva.DelegatedVesting)

	// require no modifications when the undelegation amount is zero
	pva = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)

	require.Panics(t, func() {
		pva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 0)})
	})
	require.Nil(t, pva.DelegatedFree)
	require.Nil(t, pva.DelegatedVesting)

	// vest 50% and delegate to two validators
	pva = types.NewPeriodicVestingAccount(bacc, origCoins, now.Unix(), periods)
	pva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	pva.TrackDelegation(now.Add(12*time.Hour), origCoins, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})

	// undelegate from one validator that got slashed 50%
	pva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)}, pva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, pva.DelegatedVesting)

	// undelegate from the other validator that did not get slashed
	pva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Nil(t, pva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)}, pva.DelegatedVesting)
}

func TestGenesisAccountValidate(t *testing.T) {
	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	baseAcc := authtypes.NewBaseAccount(addr, pubkey, 0, 0)
	initialVesting := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 50))
	baseVestingWithCoins := types.NewBaseVestingAccount(baseAcc, initialVesting, 100)
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
			types.NewContinuousVestingAccount(baseAcc, initialVesting, 100, 200),
			false,
		},
		{
			"invalid vesting times",
			types.NewContinuousVestingAccount(baseAcc, initialVesting, 1654668078, 1554668078),
			true,
		},
		{
			"valid periodic vesting account",
			types.NewPeriodicVestingAccount(baseAcc, initialVesting, 0, types.Periods{types.Period{Length: int64(100), Amount: sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 50)}}}),
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
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expErr, tt.acc.Validate() != nil)
		})
	}
}

func TestBaseVestingAccountJSON(t *testing.T) {
	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	coins := sdk.NewCoins(sdk.NewInt64Coin("test", 5))
	baseAcc := authtypes.NewBaseAccount(addr, pubkey, 10, 50)

	acc := types.NewBaseVestingAccount(baseAcc, coins, time.Now().Unix())

	bz, err := json.Marshal(acc)
	require.NoError(t, err)

	bz1, err := acc.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, string(bz1), string(bz))

	var a types.BaseVestingAccount
	require.NoError(t, json.Unmarshal(bz, &a))
	require.Equal(t, acc.String(), a.String())
}

func TestContinuousVestingAccountMarshal(t *testing.T) {
	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	coins := sdk.NewCoins(sdk.NewInt64Coin("test", 5))
	baseAcc := authtypes.NewBaseAccount(addr, pubkey, 10, 50)

	baseVesting := types.NewBaseVestingAccount(baseAcc, coins, time.Now().Unix())
	acc := types.NewContinuousVestingAccountRaw(baseVesting, baseVesting.EndTime)

	bz, err := app.AccountKeeper.MarshalAccount(acc)
	require.Nil(t, err)

	acc2, err := app.AccountKeeper.UnmarshalAccount(bz)
	require.Nil(t, err)
	require.IsType(t, &types.ContinuousVestingAccount{}, acc2)
	require.Equal(t, acc.String(), acc2.String())

	// error on bad bytes
	_, err = app.AccountKeeper.UnmarshalAccount(bz[:len(bz)/2])
	require.NotNil(t, err)
}

func TestContinuousVestingAccountJSON(t *testing.T) {
	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	coins := sdk.NewCoins(sdk.NewInt64Coin("test", 5))
	baseAcc := authtypes.NewBaseAccount(addr, pubkey, 10, 50)

	baseVesting := types.NewBaseVestingAccount(baseAcc, coins, time.Now().Unix())
	acc := types.NewContinuousVestingAccountRaw(baseVesting, baseVesting.EndTime)

	bz, err := json.Marshal(acc)
	require.NoError(t, err)

	bz1, err := acc.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, string(bz1), string(bz))

	var a types.ContinuousVestingAccount
	require.NoError(t, json.Unmarshal(bz, &a))
	require.Equal(t, acc.String(), a.String())
}

func TestPeriodicVestingAccountMarshal(t *testing.T) {
	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	coins := sdk.NewCoins(sdk.NewInt64Coin("test", 5))
	baseAcc := authtypes.NewBaseAccount(addr, pubkey, 10, 50)

	acc := types.NewPeriodicVestingAccount(baseAcc, coins, time.Now().Unix(), types.Periods{types.Period{3600, coins}})

	bz, err := app.AccountKeeper.MarshalAccount(acc)
	require.Nil(t, err)

	acc2, err := app.AccountKeeper.UnmarshalAccount(bz)
	require.Nil(t, err)
	require.IsType(t, &types.PeriodicVestingAccount{}, acc2)
	require.Equal(t, acc.String(), acc2.String())

	// error on bad bytes
	_, err = app.AccountKeeper.UnmarshalAccount(bz[:len(bz)/2])
	require.NotNil(t, err)
}

func TestPeriodicVestingAccountJSON(t *testing.T) {
	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	coins := sdk.NewCoins(sdk.NewInt64Coin("test", 5))
	baseAcc := authtypes.NewBaseAccount(addr, pubkey, 10, 50)

	acc := types.NewPeriodicVestingAccount(baseAcc, coins, time.Now().Unix(), types.Periods{types.Period{3600, coins}})

	bz, err := json.Marshal(acc)
	require.NoError(t, err)

	bz1, err := acc.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, string(bz1), string(bz))

	var a types.PeriodicVestingAccount
	require.NoError(t, json.Unmarshal(bz, &a))
	require.Equal(t, acc.String(), a.String())
}

func TestDelayedVestingAccountMarshal(t *testing.T) {
	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	coins := sdk.NewCoins(sdk.NewInt64Coin("test", 5))
	baseAcc := authtypes.NewBaseAccount(addr, pubkey, 10, 50)

	acc := types.NewDelayedVestingAccount(baseAcc, coins, time.Now().Unix())

	bz, err := app.AccountKeeper.MarshalAccount(acc)
	require.Nil(t, err)

	acc2, err := app.AccountKeeper.UnmarshalAccount(bz)
	require.Nil(t, err)
	require.IsType(t, &types.DelayedVestingAccount{}, acc2)
	require.Equal(t, acc.String(), acc2.String())

	// error on bad bytes
	_, err = app.AccountKeeper.UnmarshalAccount(bz[:len(bz)/2])
	require.NotNil(t, err)
}

func TestDelayedVestingAccountJSON(t *testing.T) {
	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	coins := sdk.NewCoins(sdk.NewInt64Coin("test", 5))
	baseAcc := authtypes.NewBaseAccount(addr, pubkey, 10, 50)

	acc := types.NewDelayedVestingAccount(baseAcc, coins, time.Now().Unix())

	bz, err := json.Marshal(acc)
	require.NoError(t, err)

	bz1, err := acc.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, string(bz1), string(bz))

	var a types.DelayedVestingAccount
	require.NoError(t, json.Unmarshal(bz, &a))
	require.Equal(t, acc.String(), a.String())
}
