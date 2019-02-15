package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewDecCoin(t *testing.T) {
	require.NotPanics(t, func() {
		NewDecCoin("atom", 5)
	})
	require.NotPanics(t, func() {
		NewDecCoin("atom", 0)
	})
	require.Panics(t, func() {
		NewDecCoin("Atom", 5)
	})
	require.Panics(t, func() {
		NewDecCoin("atom", -5)
	})
}

func TestNewPositiveDecCoin(t *testing.T) {
	require.NotPanics(t, func() {
		NewPositiveDecCoin("atom", 5)
	})
	require.Panics(t, func() {
		NewPositiveDecCoin("atom", 0)
	})
	require.Panics(t, func() {
		NewPositiveDecCoin("Atom", 5)
	})
	require.Panics(t, func() {
		NewPositiveDecCoin("atom", -5)
	})
}

func TestNewDecCoinFromDec(t *testing.T) {
	require.NotPanics(t, func() {
		NewDecCoinFromDec("atom", NewDec(5))
	})
	require.NotPanics(t, func() {
		NewDecCoinFromDec("atom", ZeroDec())
	})
	require.Panics(t, func() {
		NewDecCoinFromDec("Atom", NewDec(5))
	})
	require.Panics(t, func() {
		NewDecCoinFromDec("atom", NewDec(-5))
	})
}

func TestNewPositiveDecCoinFromDec(t *testing.T) {
	require.NotPanics(t, func() {
		NewPositiveDecCoinFromDec("atom", NewDec(5))
	})
	require.Panics(t, func() {
		NewPositiveDecCoinFromDec("atom", ZeroDec())
	})
	require.Panics(t, func() {
		NewPositiveDecCoinFromDec("Atom", NewDec(5))
	})
	require.Panics(t, func() {
		NewPositiveDecCoinFromDec("atom", NewDec(-5))
	})
}

func TestNewDecCoinFromCoin(t *testing.T) {
	require.NotPanics(t, func() {
		NewDecCoinFromCoin(Coin{"atom", NewInt(5)})
	})
	require.NotPanics(t, func() {
		NewDecCoinFromCoin(Coin{"atom", NewInt(0)})
	})
	require.Panics(t, func() {
		NewDecCoinFromCoin(Coin{"Atom", NewInt(5)})
	})
	require.Panics(t, func() {
		NewDecCoinFromCoin(Coin{"atom", NewInt(-5)})
	})
}

func TestNewPositiveDecCoinFromCoin(t *testing.T) {
	require.NotPanics(t, func() {
		NewPositiveDecCoinFromCoin(Coin{"atom", NewInt(5)})
	})
	require.Panics(t, func() {
		NewPositiveDecCoinFromCoin(Coin{"atom", NewInt(0)})
	})
	require.Panics(t, func() {
		NewPositiveDecCoinFromCoin(Coin{"Atom", NewInt(5)})
	})
	require.Panics(t, func() {
		NewPositiveDecCoinFromCoin(Coin{"atom", NewInt(-5)})
	})
}

func TestDecCoinIsPositive(t *testing.T) {
	dc := NewDecCoin("atom", 5)
	require.True(t, dc.IsPositive())

	dc = NewDecCoin("atom", 0)
	require.False(t, dc.IsPositive())
}

func TestPlusDecCoin(t *testing.T) {
	decCoinA1 := NewDecCoinFromDec("atom", NewDecWithPrec(11, 1))
	decCoinA2 := NewDecCoinFromDec("atom", NewDecWithPrec(22, 1))
	decCoinB1 := NewDecCoinFromDec("btc", NewDecWithPrec(11, 1))

	// regular add
	res := decCoinA1.Plus(decCoinA1)
	require.Equal(t, decCoinA2, res, "sum of coins is incorrect")

	// bad denom add
	require.Panics(t, func() {
		decCoinA1.Plus(decCoinB1)
	}, "expected panic on sum of different denoms")
}

func TestPlusDecCoins(t *testing.T) {
	one := NewDec(1)
	zero := NewDec(0)
	two := NewDec(2)

	cases := []struct {
		inputOne DecCoins
		inputTwo DecCoins
		expected DecCoins
	}{
		{DecCoins{{"atom", one}, {"btc", one}}, DecCoins{{"atom", one}, {"btc", one}}, DecCoins{{"atom", two}, {"btc", two}}},
		{DecCoins{{"atom", zero}, {"btc", one}}, DecCoins{{"atom", zero}, {"btc", zero}}, DecCoins{{"btc", one}}},
		{DecCoins{{"atom", zero}, {"btc", zero}}, DecCoins{{"atom", zero}, {"btc", zero}}, DecCoins(nil)},
	}

	for tcIndex, tc := range cases {
		res := tc.inputOne.Plus(tc.inputTwo)
		require.Equal(t, tc.expected, res, "sum of coins is incorrect, tc #%d", tcIndex)
	}
}

func TestSortDecCoins(t *testing.T) {
	good := DecCoins{
		NewDecCoin("gas", 1),
		NewDecCoin("mineral", 1),
		NewDecCoin("tree", 1),
	}
	empty := DecCoins{
		NewDecCoin("gold", 0),
	}
	badSort1 := DecCoins{
		NewDecCoin("tree", 1),
		NewDecCoin("gas", 1),
		NewDecCoin("mineral", 1),
	}
	badSort2 := DecCoins{ // both are after the first one, but the second and third are in the wrong order
		NewDecCoin("gas", 1),
		NewDecCoin("tree", 1),
		NewDecCoin("mineral", 1),
	}
	badAmt := DecCoins{
		NewDecCoin("gas", 1),
		NewDecCoin("tree", 0),
		NewDecCoin("mineral", 1),
	}
	dup := DecCoins{
		NewDecCoin("gas", 1),
		NewDecCoin("gas", 1),
		NewDecCoin("mineral", 1),
	}

	cases := []struct {
		coins         DecCoins
		before, after bool // valid before/after sort
	}{
		{good, true, true},
		{empty, false, false},
		{badSort1, false, true},
		{badSort2, false, true},
		{badAmt, false, false},
		{dup, false, false},
	}

	for tcIndex, tc := range cases {
		require.Equal(t, tc.before, tc.coins.IsValid(), "coin validity is incorrect before sorting, tc #%d", tcIndex)
		tc.coins.Sort()
		require.Equal(t, tc.after, tc.coins.IsValid(), "coin validity is incorrect after sorting, tc #%d", tcIndex)
	}
}

func TestDecCoinsIsValid(t *testing.T) {
	testCases := []struct {
		input    DecCoins
		expected bool
	}{
		{DecCoins{}, true},
		{DecCoins{DecCoin{"atom", NewDec(5)}}, true},
		{DecCoins{DecCoin{"atom", NewDec(5)}, DecCoin{"btc", NewDec(100000)}}, true},
		{DecCoins{DecCoin{"atom", NewDec(-5)}}, false},
		{DecCoins{DecCoin{"Atom", NewDec(5)}}, false},
		{DecCoins{DecCoin{"atom", NewDec(5)}, DecCoin{"Btc", NewDec(100000)}}, false},
		{DecCoins{DecCoin{"atom", NewDec(5)}, DecCoin{"btc", NewDec(-100000)}}, false},
		{DecCoins{DecCoin{"atom", NewDec(-5)}, DecCoin{"btc", NewDec(100000)}}, false},
		{DecCoins{DecCoin{"Atom", NewDec(5)}, DecCoin{"b", NewDec(100000)}}, false},
	}

	for i, tc := range testCases {
		res := tc.input.IsValid()
		require.Equal(t, tc.expected, res, "unexpected result for test case #%d, input: %v", i, tc.input)
	}
}

func TestParseDecCoins(t *testing.T) {
	testCases := []struct {
		input          string
		expectedResult DecCoins
		expectedErr    bool
	}{
		{"", nil, false},
		{"4stake", nil, true},
		{"5.5atom,4stake", nil, true},
		{"0.0stake", nil, true},
		{"0.004STAKE", nil, true},
		{"-1.0stake", nil, true},
		{"0.004STAKE", nil, true},
		{
			"0.004stake",
			DecCoins{NewDecCoinFromDec("stake", NewDecWithPrec(4000000000000000, Precision))},
			false,
		},
		{
			"5.04atom,0.004stake",
			DecCoins{
				NewDecCoinFromDec("atom", NewDecWithPrec(5040000000000000000, Precision)),
				NewDecCoinFromDec("stake", NewDecWithPrec(4000000000000000, Precision)),
			},
			false,
		},
		{
			"0.004 stake",
			DecCoins{NewDecCoinFromDec("stake", NewDecWithPrec(4000000000000000, Precision))},
			false,
		},
		{
			"\n0.004\tstake",
			DecCoins{NewDecCoinFromDec("stake", NewDecWithPrec(4000000000000000, Precision))},
			false,
		},
	}

	for i, tc := range testCases {
		res, err := ParseDecCoins(tc.input)
		if tc.expectedErr {
			require.Error(t, err, "expected error for test case #%d, input: %v", i, tc.input)
		} else {
			require.NoError(t, err, "unexpected error for test case #%d, input: %v", i, tc.input, err)
			require.Equal(t, tc.expectedResult, res, "unexpected result for test case #%d, input: %v", i, tc.input)
		}
	}
}

func TestDecCoinsString(t *testing.T) {
	testCases := []struct {
		input    DecCoins
		expected string
	}{
		{DecCoins{}, ""},
		{
			DecCoins{
				NewDecCoinFromDec("atom", NewDecWithPrec(5040000000000000000, Precision)),
				NewDecCoinFromDec("stake", NewDecWithPrec(4000000000000000, Precision)),
			},
			"5.040000000000000000atom,0.004000000000000000stake",
		},
	}

	for i, tc := range testCases {
		out := tc.input.String()
		require.Equal(t, tc.expected, out, "unexpected result for test case #%d, input: %v", i, tc.input)
	}
}

func TestDecCoinTruncateDecimal(t *testing.T) {
	type fields struct {
		Denom  string
		Amount Dec
	}
	tests := []struct {
		name   string
		fields fields
		want   Coin
		want1  DecCoin
	}{
		{
			"truncate 1.1",
			fields{"atom", MustNewDecFromStr("1.1")},
			NewCoin("atom", NewInt(1)),
			NewDecCoinFromDec("atom", MustNewDecFromStr("0.1")),
		},
		{
			"truncate 0.1",
			fields{"atom", MustNewDecFromStr("0.1")},
			NewCoin("atom", NewInt(0)),
			NewDecCoinFromDec("atom", MustNewDecFromStr("0.1")),
		},
		{
			"truncate 0.0",
			fields{"atom", MustNewDecFromStr("0.0")},
			NewCoin("atom", NewInt(0)),
			NewDecCoinFromDec("atom", MustNewDecFromStr("0.0")),
		},
		{
			"truncate 0",
			fields{"atom", MustNewDecFromStr("0")},
			NewCoin("atom", NewInt(0)),
			NewDecCoinFromDec("atom", MustNewDecFromStr("0.0")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coin := DecCoin{
				Denom:  tt.fields.Denom,
				Amount: tt.fields.Amount,
			}
			got, got1 := coin.TruncateDecimal()
			require.True(t, tt.want.IsEqual(got))
			require.Equal(t, tt.want1, got1)
		})
	}
}
