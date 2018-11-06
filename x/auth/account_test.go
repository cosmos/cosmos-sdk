package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmtime "github.com/tendermint/tendermint/types/time"

	codec "github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	testDenom = "steak"
)

func keyPubAddr() (crypto.PrivKey, crypto.PubKey, sdk.AccAddress) {
	key := ed25519.GenPrivKey()
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	return key, pub, addr
}

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

	seq := int64(7)

	err := acc.SetSequence(seq)
	require.Nil(t, err)
	require.Equal(t, seq, acc.GetSequence())
}

func TestBaseAccountMarshal(t *testing.T) {
	_, pub, addr := keyPubAddr()
	acc := NewBaseAccountWithAddress(addr)

	someCoins := sdk.Coins{sdk.NewInt64Coin("atom", 123), sdk.NewInt64Coin("eth", 246)}
	seq := int64(7)

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
	origCoins := sdk.Coins{sdk.NewInt64Coin(testDenom, 100)}
	cva := NewContinuousVestingAccount(addr, origCoins, now, endTime)

	// require no coins vested in the very beginning of the vesting schedule
	vestedCoins := cva.GetVestedCoins(now)
	require.Nil(t, vestedCoins)

	// require all coins vested at the end of the vesting schedule
	vestedCoins = cva.GetVestedCoins(endTime)
	require.Equal(t, origCoins, vestedCoins)

	// require 50% of coins vested
	vestedCoins = cva.GetVestedCoins(now.Add(12 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(testDenom, 50)}, vestedCoins)
}

func TestGetVestingCoinsContVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	_, _, addr := keyPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(testDenom, 100)}
	cva := NewContinuousVestingAccount(addr, origCoins, now, endTime)

	// require all coins vesting in the beginning of the vesting schedule
	vestingCoins := cva.GetVestingCoins(now)
	require.Equal(t, origCoins, vestingCoins)

	// require no coins vesting at the end of the vesting schedule
	vestingCoins = cva.GetVestingCoins(endTime)
	require.Nil(t, vestingCoins)

	// require 50% of coins vesting
	vestingCoins = cva.GetVestingCoins(now.Add(12 * time.Hour))
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(testDenom, 50)}, vestingCoins)
}

func TestSpendableCoinsContVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	_, _, addr := keyPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(testDenom, 100)}
	cva := NewContinuousVestingAccount(addr, origCoins, now, endTime)

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
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(testDenom, 50)}, spendableCoins)

	// receive some coins
	recvAmt := sdk.Coins{sdk.NewInt64Coin(testDenom, 50)}
	cva.SetCoins(cva.GetCoins().Plus(recvAmt))

	// require that all vested coins (50%) are spendable plus any received
	spendableCoins = cva.SpendableCoins(now.Add(12 * time.Hour))
	require.Equal(t, origCoins, spendableCoins)

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
	origCoins := sdk.Coins{sdk.NewInt64Coin(testDenom, 100)}

	// require the ability to delegate all vesting coins
	cva := NewContinuousVestingAccount(addr, origCoins, now, endTime)
	cva.TrackDelegation(now, origCoins)
	require.Equal(t, origCoins, cva.delegatedVesting)
	require.Nil(t, cva.delegatedFree)
	require.Nil(t, cva.GetCoins())

	// require the ability to delegate all vested coins
	cva = NewContinuousVestingAccount(addr, origCoins, now, endTime)
	cva.TrackDelegation(endTime, origCoins)
	require.Nil(t, cva.delegatedVesting)
	require.Equal(t, origCoins, cva.delegatedFree)
	require.Nil(t, cva.GetCoins())

	// require the ability to delegate all vesting coins (50%) and all vested coins (50%)
	cva = NewContinuousVestingAccount(addr, origCoins, now, endTime)
	cva.TrackDelegation(now.Add(12*time.Hour), origCoins)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(testDenom, 50)}, cva.delegatedVesting)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(testDenom, 50)}, cva.delegatedFree)
	require.Nil(t, cva.GetCoins())

	// require no modifications when delegation amount is zero or not enough funds
	cva = NewContinuousVestingAccount(addr, origCoins, now, endTime)
	cva.TrackDelegation(endTime, sdk.Coins{sdk.NewInt64Coin(testDenom, 1000000)})
	require.Nil(t, cva.delegatedVesting)
	require.Nil(t, cva.delegatedFree)
	require.Equal(t, origCoins, cva.GetCoins())
}

func TestTrackUndelegationContVestingAcc(t *testing.T) {
	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	_, _, addr := keyPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(testDenom, 100)}

	// require the ability to undelegate all vesting coins
	cva := NewContinuousVestingAccount(addr, origCoins, now, endTime)
	cva.TrackDelegation(now, origCoins)
	cva.TrackUndelegation(origCoins)
	require.Nil(t, cva.delegatedFree)
	require.Nil(t, cva.delegatedVesting)
	require.Equal(t, origCoins, cva.GetCoins())

	// require the ability to undelegate all vested coins
	cva = NewContinuousVestingAccount(addr, origCoins, now, endTime)
	cva.TrackDelegation(endTime, origCoins)
	cva.TrackUndelegation(origCoins)
	require.Nil(t, cva.delegatedFree)
	require.Nil(t, cva.delegatedVesting)
	require.Equal(t, origCoins, cva.GetCoins())

	// require no modifications when the undelegation amount is zero
	cva = NewContinuousVestingAccount(addr, origCoins, now, endTime)
	cva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(testDenom, 0)})
	require.Nil(t, cva.delegatedFree)
	require.Nil(t, cva.delegatedVesting)
	require.Equal(t, origCoins, cva.GetCoins())

	// vest 50% and delegate to two validators
	cva = NewContinuousVestingAccount(addr, origCoins, now, endTime)
	cva.TrackDelegation(now.Add(12*time.Hour), sdk.Coins{sdk.NewInt64Coin(testDenom, 50)})
	cva.TrackDelegation(now.Add(12*time.Hour), sdk.Coins{sdk.NewInt64Coin(testDenom, 50)})

	// undelegate from one validator that got slashed 50%
	cva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(testDenom, 25)})
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(testDenom, 25)}, cva.delegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(testDenom, 50)}, cva.delegatedVesting)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(testDenom, 25)}, cva.GetCoins())

	// undelegate from the other validator that did not get slashed
	cva.TrackUndelegation(sdk.Coins{sdk.NewInt64Coin(testDenom, 50)})
	require.Nil(t, cva.delegatedFree)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(testDenom, 25)}, cva.delegatedVesting)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin(testDenom, 75)}, cva.GetCoins())
}
