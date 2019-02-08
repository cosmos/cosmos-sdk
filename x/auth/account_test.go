package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	stakeDenom = "stake"
	feeDenom   = "fee"
)

func TestBaseAddressPubKey(t *testing.T) {
	_, pub1, addr1 := keyPubAddr()
	_, pub2, addr2 := keyPubAddr()
	acc := NewBaseAccountWithAddress(addr1)

	// check the address (set) and pubkey (not set)
	require.EqualValues(t, addr1, acc.GetAddress())
	require.EqualValues(t, nil, acc.GetPubKey())

	// can't override address
	err := acc.SetAddress(addr2)
	require.NotNil(t, err)
	require.EqualValues(t, addr1, acc.GetAddress())

	// set the pubkey
	err = acc.SetPubKey(pub1)
	require.Nil(t, err)
	require.Equal(t, pub1, acc.GetPubKey())

	// can override pubkey
	err = acc.SetPubKey(pub2)
	require.Nil(t, err)
	require.Equal(t, pub2, acc.GetPubKey())

	//------------------------------------

	// can set address on empty account
	acc2 := BaseAccount{}
	err = acc2.SetAddress(addr2)
	require.Nil(t, err)
	require.EqualValues(t, addr2, acc2.GetAddress())
}

func TestBaseAccountCoins(t *testing.T) {
	_, _, addr := keyPubAddr()
	acc := NewBaseAccountWithAddress(addr)

	someCoins := sdk.Coins{sdk.NewInt64Coin("atom", 123), sdk.NewInt64Coin("eth", 246)}

	err := acc.SetCoins(someCoins)
	require.Nil(t, err)
	require.Equal(t, someCoins, acc.GetCoins())
}

func TestBaseAccountSequence(t *testing.T) {
	_, _, addr := keyPubAddr()
	acc := NewBaseAccountWithAddress(addr)

	seq := uint64(7)

	err := acc.SetSequence(seq)
	require.Nil(t, err)
	require.Equal(t, seq, acc.GetSequence())
}

func TestBaseAccountMarshal(t *testing.T) {
	_, pub, addr := keyPubAddr()
	acc := NewBaseAccountWithAddress(addr)

	someCoins := sdk.Coins{sdk.NewInt64Coin("atom", 123), sdk.NewInt64Coin("eth", 246)}
	seq := uint64(7)

	// set everything on the account
	err := acc.SetPubKey(pub)
	require.Nil(t, err)
	err = acc.SetSequence(seq)
	require.Nil(t, err)
	err = acc.SetCoins(someCoins)
	require.Nil(t, err)

	// need a codec for marshaling
	cdc := codec.New()
	codec.RegisterCrypto(cdc)

	b, err := cdc.MarshalBinaryLengthPrefixed(acc)
	require.Nil(t, err)

	acc2 := BaseAccount{}
	err = cdc.UnmarshalBinaryLengthPrefixed(b, &acc2)
	require.Nil(t, err)
	require.Equal(t, acc, acc2)

	// error on bad bytes
	acc2 = BaseAccount{}
	err = cdc.UnmarshalBinaryLengthPrefixed(b[:len(b)/2], &acc2)
	require.NotNil(t, err)
}

func TestGetVestedCoinsContVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	_, _, addr := keyPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := NewBaseAccountWithAddress(addr)
	bacc.SetCoins(origCoins)
	cva := NewContinuousVestingAccount(&bacc, now.Unix(), endTime.Unix())

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

	_, _, addr := keyPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := NewBaseAccountWithAddress(addr)
	bacc.SetCoins(origCoins)
	cva := NewContinuousVestingAccount(&bacc, now.Unix(), endTime.Unix())

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

	_, _, addr := keyPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := NewBaseAccountWithAddress(addr)
	bacc.SetCoins(origCoins)
	cva := NewContinuousVestingAccount(&bacc, now.Unix(), endTime.Unix())

	// require that there exist no spendable coins in the beginning of the
	// vesting schedule
	spendableCoins := cva.SpendableCoins(now)
	require.Nil(t, spendableCoins)

	// require that all original coins are spendable at the end of the vesting
	// schedule
	spendableCoins = cva.SpendableCoins(endTime)
	require.Equal(t, origCoins, spendableCoins)

	// require that all vested coins (50%) are spendable
	spendableCoins = cva.SpendableCoins(now.Add(12 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}, spendableCoins)

	// receive some coins
	recvAmt := sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}
	cva.SetCoins(cva.GetCoins().Plus(recvAmt))

	// require that all vested coins (50%) are spendable plus any received
	spendableCoins = cva.SpendableCoins(now.Add(12 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 100)}, spendableCoins)

	// spend all spendable coins
	cva.SetCoins(cva.GetCoins().Minus(spendableCoins))

	// require that no more coins are spendable
	spendableCoins = cva.SpendableCoins(now.Add(12 * time.Hour))
	require.Nil(t, spendableCoins)
}

func TestTrackDelegationContVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	_, _, addr := keyPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := NewBaseAccountWithAddress(addr)
	bacc.SetCoins(origCoins)

	// require the ability to delegate all vesting coins
	cva := NewContinuousVestingAccount(&bacc, now.Unix(), endTime.Unix())
	cva.TrackDelegation(now, origCoins)
	require.Equal(t, origCoins, cva.DelegatedVesting)
	require.Nil(t, cva.DelegatedFree)
	require.Nil(t, cva.GetCoins())

	// require the ability to delegate all vested coins
	bacc.SetCoins(origCoins)
	cva = NewContinuousVestingAccount(&bacc, now.Unix(), endTime.Unix())
	cva.TrackDelegation(endTime, origCoins)
	require.Nil(t, cva.DelegatedVesting)
	require.Equal(t, origCoins, cva.DelegatedFree)
	require.Nil(t, cva.GetCoins())

	// require the ability to delegate all vesting coins (50%) and all vested coins (50%)
	bacc.SetCoins(origCoins)
	cva = NewContinuousVestingAccount(&bacc, now.Unix(), endTime.Unix())
	cva.TrackDelegation(now.Add(12*time.Hour), sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, cva.DelegatedVesting)
	require.Nil(t, cva.DelegatedFree)

	cva.TrackDelegation(now.Add(12*time.Hour), sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, cva.DelegatedVesting)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, cva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000)}, cva.GetCoins())

	// require no modifications when delegation amount is zero or not enough funds
	bacc.SetCoins(origCoins)
	cva = NewContinuousVestingAccount(&bacc, now.Unix(), endTime.Unix())
	require.Panics(t, func() {
		cva.TrackDelegation(endTime, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 1000000)})
	})
	require.Nil(t, cva.DelegatedVesting)
	require.Nil(t, cva.DelegatedFree)
	require.Equal(t, origCoins, cva.GetCoins())
}

func TestTrackUndelegationContVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	_, _, addr := keyPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := NewBaseAccountWithAddress(addr)
	bacc.SetCoins(origCoins)

	// require the ability to undelegate all vesting coins
	cva := NewContinuousVestingAccount(&bacc, now.Unix(), endTime.Unix())
	cva.TrackDelegation(now, origCoins)
	cva.TrackUndelegation(origCoins)
	require.Nil(t, cva.DelegatedFree)
	require.Nil(t, cva.DelegatedVesting)
	require.Equal(t, origCoins, cva.GetCoins())

	// require the ability to undelegate all vested coins
	bacc.SetCoins(origCoins)
	cva = NewContinuousVestingAccount(&bacc, now.Unix(), endTime.Unix())

	cva.TrackDelegation(endTime, origCoins)
	cva.TrackUndelegation(origCoins)
	require.Nil(t, cva.DelegatedFree)
	require.Nil(t, cva.DelegatedVesting)
	require.Equal(t, origCoins, cva.GetCoins())

	// require no modifications when the undelegation amount is zero
	bacc.SetCoins(origCoins)
	cva = NewContinuousVestingAccount(&bacc, now.Unix(), endTime.Unix())

	require.Panics(t, func() {
		cva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 0)})
	})
	require.Nil(t, cva.DelegatedFree)
	require.Nil(t, cva.DelegatedVesting)
	require.Equal(t, origCoins, cva.GetCoins())

	// vest 50% and delegate to two validators
	cva = NewContinuousVestingAccount(&bacc, now.Unix(), endTime.Unix())
	cva.TrackDelegation(now.Add(12*time.Hour), sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	cva.TrackDelegation(now.Add(12*time.Hour), sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})

	// undelegate from one validator that got slashed 50%
	cva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)}, cva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}, cva.DelegatedVesting)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 25)}, cva.GetCoins())

	// undelegate from the other validator that did not get slashed
	cva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Nil(t, cva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)}, cva.DelegatedVesting)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 75)}, cva.GetCoins())
}

func TestGetVestedCoinsDelVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	_, _, addr := keyPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := NewBaseAccountWithAddress(addr)
	bacc.SetCoins(origCoins)

	// require no coins are vested until schedule maturation
	dva := NewDelayedVestingAccount(&bacc, endTime.Unix())
	vestedCoins := dva.GetVestedCoins(now)
	require.Nil(t, vestedCoins)

	// require all coins be vested at schedule maturation
	vestedCoins = dva.GetVestedCoins(endTime)
	require.Equal(t, origCoins, vestedCoins)
}

func TestGetVestingCoinsDelVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	_, _, addr := keyPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := NewBaseAccountWithAddress(addr)
	bacc.SetCoins(origCoins)

	// require all coins vesting at the beginning of the schedule
	dva := NewDelayedVestingAccount(&bacc, endTime.Unix())
	vestingCoins := dva.GetVestingCoins(now)
	require.Equal(t, origCoins, vestingCoins)

	// require no coins vesting at schedule maturation
	vestingCoins = dva.GetVestingCoins(endTime)
	require.Nil(t, vestingCoins)
}

func TestSpendableCoinsDelVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	_, _, addr := keyPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := NewBaseAccountWithAddress(addr)
	bacc.SetCoins(origCoins)

	// require that no coins are spendable in the beginning of the vesting
	// schedule
	dva := NewDelayedVestingAccount(&bacc, endTime.Unix())
	spendableCoins := dva.SpendableCoins(now)
	require.Nil(t, spendableCoins)

	// require that all coins are spendable after the maturation of the vesting
	// schedule
	spendableCoins = dva.SpendableCoins(endTime)
	require.Equal(t, origCoins, spendableCoins)

	// require that all coins are still vesting after some time
	spendableCoins = dva.SpendableCoins(now.Add(12 * time.Hour))
	require.Nil(t, spendableCoins)

	// receive some coins
	recvAmt := sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)}
	dva.SetCoins(dva.GetCoins().Plus(recvAmt))

	// require that only received coins are spendable since the account is still
	// vesting
	spendableCoins = dva.SpendableCoins(now.Add(12 * time.Hour))
	require.Equal(t, recvAmt, spendableCoins)

	// spend all spendable coins
	dva.SetCoins(dva.GetCoins().Minus(spendableCoins))

	// require that no more coins are spendable
	spendableCoins = dva.SpendableCoins(now.Add(12 * time.Hour))
	require.Nil(t, spendableCoins)
}

func TestTrackDelegationDelVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	_, _, addr := keyPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := NewBaseAccountWithAddress(addr)
	bacc.SetCoins(origCoins)

	// require the ability to delegate all vesting coins
	bacc.SetCoins(origCoins)
	dva := NewDelayedVestingAccount(&bacc, endTime.Unix())
	dva.TrackDelegation(now, origCoins)
	require.Equal(t, origCoins, dva.DelegatedVesting)
	require.Nil(t, dva.DelegatedFree)
	require.Nil(t, dva.GetCoins())

	// require the ability to delegate all vested coins
	bacc.SetCoins(origCoins)
	dva = NewDelayedVestingAccount(&bacc, endTime.Unix())
	dva.TrackDelegation(endTime, origCoins)
	require.Nil(t, dva.DelegatedVesting)
	require.Equal(t, origCoins, dva.DelegatedFree)
	require.Nil(t, dva.GetCoins())

	// require the ability to delegate all coins half way through the vesting
	// schedule
	bacc.SetCoins(origCoins)
	dva = NewDelayedVestingAccount(&bacc, endTime.Unix())
	dva.TrackDelegation(now.Add(12*time.Hour), origCoins)
	require.Equal(t, origCoins, dva.DelegatedVesting)
	require.Nil(t, dva.DelegatedFree)
	require.Nil(t, dva.GetCoins())

	// require no modifications when delegation amount is zero or not enough funds
	bacc.SetCoins(origCoins)
	dva = NewDelayedVestingAccount(&bacc, endTime.Unix())

	require.Panics(t, func() {
		dva.TrackDelegation(endTime, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 1000000)})
	})
	require.Nil(t, dva.DelegatedVesting)
	require.Nil(t, dva.DelegatedFree)
	require.Equal(t, origCoins, dva.GetCoins())
}

func TestTrackUndelegationDelVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	_, _, addr := keyPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := NewBaseAccountWithAddress(addr)
	bacc.SetCoins(origCoins)

	// require the ability to undelegate all vesting coins
	bacc.SetCoins(origCoins)
	dva := NewDelayedVestingAccount(&bacc, endTime.Unix())
	dva.TrackDelegation(now, origCoins)
	dva.TrackUndelegation(origCoins)
	require.Nil(t, dva.DelegatedFree)
	require.Nil(t, dva.DelegatedVesting)
	require.Equal(t, origCoins, dva.GetCoins())

	// require the ability to undelegate all vested coins
	bacc.SetCoins(origCoins)
	dva = NewDelayedVestingAccount(&bacc, endTime.Unix())
	dva.TrackDelegation(endTime, origCoins)
	dva.TrackUndelegation(origCoins)
	require.Nil(t, dva.DelegatedFree)
	require.Nil(t, dva.DelegatedVesting)
	require.Equal(t, origCoins, dva.GetCoins())

	// require no modifications when the undelegation amount is zero
	bacc.SetCoins(origCoins)
	dva = NewDelayedVestingAccount(&bacc, endTime.Unix())

	require.Panics(t, func() {
		dva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 0)})
	})
	require.Nil(t, dva.DelegatedFree)
	require.Nil(t, dva.DelegatedVesting)
	require.Equal(t, origCoins, dva.GetCoins())

	// vest 50% and delegate to two validators
	bacc.SetCoins(origCoins)
	dva = NewDelayedVestingAccount(&bacc, endTime.Unix())
	dva.TrackDelegation(now.Add(12*time.Hour), sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	dva.TrackDelegation(now.Add(12*time.Hour), sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})

	// undelegate from one validator that got slashed 50%
	dva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)})

	require.Nil(t, dva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 75)}, dva.DelegatedVesting)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 25)}, dva.GetCoins())

	// undelegate from the other validator that did not get slashed
	dva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(stakeDenom, 50)})
	require.Nil(t, dva.DelegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(stakeDenom, 25)}, dva.DelegatedVesting)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 75)}, dva.GetCoins())
}
