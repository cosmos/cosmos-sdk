package types

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	atom  = "atom"
	uatom = "uatom"
	matom = "matom"
	katom = "katom"
)

func TestRegisterDenom(t *testing.T) {
	unit := NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(9), nil))
	require.NoError(t, RegisterDenom("gwei", unit))
	require.Error(t, RegisterDenom("gwei", unit))

	res, ok := GetDenomUnit("gwei")
	require.True(t, ok)
	require.Equal(t, unit, res)

	res, ok = GetDenomUnit("finney")
	require.False(t, ok)
	require.Equal(t, ZeroInt(), res)
}

func TestConvertCoins(t *testing.T) {
	atomUnit := OneInt()
	require.NoError(t, RegisterDenom(atom, atomUnit))

	katomUnit := NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(3), nil))
	require.NoError(t, RegisterDenom(katom, katomUnit))

	uatomUinit := NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil))
	require.NoError(t, RegisterDenom(uatom, uatomUinit))

	matomUnit := NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(5), nil))
	require.NoError(t, RegisterDenom(matom, matomUnit))

	testCases := []struct {
		input  Coin
		denom  string
		result Coin
		expErr bool
	}{
		{NewCoin("foo", ZeroInt()), atom, Coin{}, true},
		{NewCoin(atom, ZeroInt()), "foo", Coin{}, true},
		{NewCoin(atom, ZeroInt()), "FOO", Coin{}, true},

		{NewCoin(atom, NewInt(5)), uatom, NewCoin(uatom, NewInt(5000000)), false}, // atom => uatom
		{NewCoin(atom, NewInt(5)), matom, NewCoin(matom, NewInt(500000)), false},  // atom => matom
		{NewCoin(atom, NewInt(5)), katom, NewCoin(katom, NewInt(5000)), false},    // atom => katom

		{NewCoin(uatom, NewInt(5000000)), matom, NewCoin(matom, NewInt(500000)), false}, // uatom => matom
		{NewCoin(uatom, NewInt(5000000)), katom, NewCoin(katom, NewInt(5000)), false},   // uatom => katom
		{NewCoin(uatom, NewInt(5000000)), atom, NewCoin(atom, NewInt(5)), false},        // uatom => atom

		{NewCoin(matom, NewInt(500000)), katom, NewCoin(katom, NewInt(5000)), false},    // matom => katom
		{NewCoin(matom, NewInt(500000)), uatom, NewCoin(uatom, NewInt(5000000)), false}, // matom => uatom
	}

	for i, tc := range testCases {
		res, err := ConvertCoin(tc.input, tc.denom)
		require.Equal(
			t, tc.expErr, err != nil,
			"unexpected error; tc: #%d, input: %s, denom: %s", i, tc.input, tc.denom,
		)
		require.Equal(
			t, tc.result, res,
			"invalid result; tc: #%d, input: %s, denom: %s", i, tc.input, tc.denom,
		)
	}
}
