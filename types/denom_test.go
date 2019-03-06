package types

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterDenom(t *testing.T) {
	require.Error(t, RegisterDenom(Uatom, ZeroInt()))

	unit := NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(9), nil))
	require.NoError(t, RegisterDenom("gwei", unit))

	res, ok := GetDenomUnit("gwei")
	require.True(t, ok)
	require.Equal(t, unit, res)

	res, ok = GetDenomUnit("finney")
	require.False(t, ok)
	require.Equal(t, ZeroInt(), res)
}

func TestConvertCoins(t *testing.T) {
	katomUnit := NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(3), nil))
	require.NoError(t, RegisterDenom("katom", katomUnit))

	testCases := []struct {
		input  Coin
		denom  string
		result Coin
		expErr bool
	}{
		{NewCoin("foo", ZeroInt()), Atom, Coin{}, true},
		{NewCoin(Atom, ZeroInt()), "foo", Coin{}, true},
		{NewCoin(Atom, ZeroInt()), "FOO", Coin{}, true},

		{NewCoin(Atom, NewInt(5)), Uatom, NewCoin(Uatom, NewInt(5000000)), false},  // atom => uatom
		{NewCoin(Atom, NewInt(5)), Matom, NewCoin(Matom, NewInt(500000)), false},   // atom => matom
		{NewCoin(Atom, NewInt(5)), "katom", NewCoin("katom", NewInt(5000)), false}, // atom => katom
	}

	for i, tc := range testCases {
		res, err := ConvertCoins(tc.input, tc.denom)
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
