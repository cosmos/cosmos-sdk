package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestDelegationEqual(t *testing.T) {
	d1 := Delegation{
		DelegatorAddr: addr1,
		ValidatorAddr: addr2,
		Shares:        sdk.NewRat(100),
	}
	d2 := Delegation{
		DelegatorAddr: addr1,
		ValidatorAddr: addr2,
		Shares:        sdk.NewRat(100),
	}

	ok := d1.Equal(d2)
	require.True(t, ok)

	d2.ValidatorAddr = addr3
	d2.Shares = sdk.NewRat(200)

	ok = d1.Equal(d2)
	require.False(t, ok)
}

func TestDelegationHumanReadableString(t *testing.T) {
	d := Delegation{
		DelegatorAddr: addr1,
		ValidatorAddr: addr2,
		Shares:        sdk.NewRat(100),
	}

	// NOTE: Being that the validator's keypair is random, we cannot test the
	// actual contents of the string.
	valStr, err := d.HumanReadableString()
	require.Nil(t, err)
	require.NotEmpty(t, valStr)
}

func TestUnbondingDelegationEqual(t *testing.T) {
	ud1 := UnbondingDelegation{
		DelegatorAddr: addr1,
		ValidatorAddr: addr2,
	}
	ud2 := UnbondingDelegation{
		DelegatorAddr: addr1,
		ValidatorAddr: addr2,
	}

	ok := ud1.Equal(ud2)
	require.True(t, ok)

	ud2.ValidatorAddr = addr3
	ud2.MinTime = 20 * 20 * 2

	ok = ud1.Equal(ud2)
	require.False(t, ok)
}

func TestUnbondingDelegationHumanReadableString(t *testing.T) {
	ud := UnbondingDelegation{
		DelegatorAddr: addr1,
		ValidatorAddr: addr2,
	}

	// NOTE: Being that the validator's keypair is random, we cannot test the
	// actual contents of the string.
	valStr, err := ud.HumanReadableString()
	require.Nil(t, err)
	require.NotEmpty(t, valStr)
}

func TestRedelegationEqual(t *testing.T) {
	r1 := Redelegation{
		DelegatorAddr:    addr1,
		ValidatorSrcAddr: addr2,
		ValidatorDstAddr: addr3,
	}
	r2 := Redelegation{
		DelegatorAddr:    addr1,
		ValidatorSrcAddr: addr2,
		ValidatorDstAddr: addr3,
	}

	ok := r1.Equal(r2)
	require.True(t, ok)

	r2.SharesDst = sdk.NewRat(10)
	r2.SharesSrc = sdk.NewRat(20)
	r2.MinTime = 20 * 20 * 2

	ok = r1.Equal(r2)
	require.False(t, ok)
}

func TestRedelegationHumanReadableString(t *testing.T) {
	r := Redelegation{
		DelegatorAddr:    addr1,
		ValidatorSrcAddr: addr2,
		ValidatorDstAddr: addr3,
		SharesDst:        sdk.NewRat(10),
		SharesSrc:        sdk.NewRat(20),
	}

	// NOTE: Being that the validator's keypair is random, we cannot test the
	// actual contents of the string.
	valStr, err := r.HumanReadableString()
	require.Nil(t, err)
	require.NotEmpty(t, valStr)
}
