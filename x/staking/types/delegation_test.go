package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestDelegationEqual(t *testing.T) {
	d1 := NewDelegation(sdk.AccAddress(addr1), addr2, sdk.NewDec(100))
	d2 := d1

	ok := d1.Equal(d2)
	require.True(t, ok)

	d2.ValidatorAddr = addr3
	d2.Shares = sdk.NewDec(200)

	ok = d1.Equal(d2)
	require.False(t, ok)
}

func TestDelegationHumanReadableString(t *testing.T) {
	d := NewDelegation(sdk.AccAddress(addr1), addr2, sdk.NewDec(100))

	// NOTE: Being that the validator's keypair is random, we cannot test the
	// actual contents of the string.
	valStr, err := d.HumanReadableString()
	require.Nil(t, err)
	require.NotEmpty(t, valStr)
}

func TestUnbondingDelegationEqual(t *testing.T) {
	ubd1 := NewUnbondingDelegation(sdk.AccAddress(addr1), addr2, 0,
		time.Unix(0, 0), sdk.NewInt64Coin(DefaultBondDenom, 0))
	ubd2 := ubd1

	ok := ubd1.Equal(ubd2)
	require.True(t, ok)

	ubd2.ValidatorAddr = addr3

	ubd2.Entries[0].CompletionTime = time.Unix(20*20*2, 0)
	ok = ubd1.Equal(ubd2)
	require.False(t, ok)
}

func TestUnbondingDelegationHumanReadableString(t *testing.T) {
	ubd := NewUnbondingDelegation(sdk.AccAddress(addr1), addr2, 0,
		time.Unix(0, 0), sdk.NewInt64Coin(DefaultBondDenom, 0))

	// NOTE: Being that the validator's keypair is random, we cannot test the
	// actual contents of the string.
	valStr, err := ubd.HumanReadableString()
	require.Nil(t, err)
	require.NotEmpty(t, valStr)
}

func TestRedelegationEqual(t *testing.T) {
	r1 := NewRedelegation(sdk.AccAddress(addr1), addr2, addr3, 0,
		time.Unix(0, 0), sdk.NewInt64Coin(DefaultBondDenom, 0),
		sdk.NewDec(0), sdk.NewDec(0))
	r2 := NewRedelegation(sdk.AccAddress(addr1), addr2, addr3, 0,
		time.Unix(0, 0), sdk.NewInt64Coin(DefaultBondDenom, 0),
		sdk.NewDec(0), sdk.NewDec(0))

	ok := r1.Equal(r2)
	require.True(t, ok)

	r2.Entries[0].SharesDst = sdk.NewDec(10)
	r2.Entries[0].SharesSrc = sdk.NewDec(20)
	r2.Entries[0].CompletionTime = time.Unix(20*20*2, 0)

	ok = r1.Equal(r2)
	require.False(t, ok)
}

func TestRedelegationHumanReadableString(t *testing.T) {
	r := NewRedelegation(sdk.AccAddress(addr1), addr2, addr3, 0,
		time.Unix(0, 0), sdk.NewInt64Coin(DefaultBondDenom, 0),
		sdk.NewDec(10), sdk.NewDec(20))

	// NOTE: Being that the validator's keypair is random, we cannot test the
	// actual contents of the string.
	valStr, err := r.HumanReadableString()
	require.Nil(t, err)
	require.NotEmpty(t, valStr)
}
