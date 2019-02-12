package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewDecCoin(t *testing.T) {
	require.NotPanics(t, func() {
		NewInt64DecCoin(testDenom1, 5)
	})
	require.NotPanics(t, func() {
		NewInt64DecCoin(testDenom1, 0)
	})
	require.Panics(t, func() {
		NewInt64DecCoin(strings.ToUpper(testDenom1), 5)
	})
	require.Panics(t, func() {
		NewInt64DecCoin(testDenom1, -5)
	})
}

func TestNewDecCoinFromDec(t *testing.T) {
	require.NotPanics(t, func() {
		NewDecCoinFromDec(testDenom1, NewDec(5))
	})
	require.NotPanics(t, func() {
		NewDecCoinFromDec(testDenom1, ZeroDec())
	})
	require.Panics(t, func() {
		NewDecCoinFromDec(strings.ToUpper(testDenom1), NewDec(5))
	})
	require.Panics(t, func() {
		NewDecCoinFromDec(testDenom1, NewDec(-5))
	})
}

func TestNewDecCoinFromCoin(t *testing.T) {
	require.NotPanics(t, func() {
		NewDecCoinFromCoin(Coin{testDenom1, NewInt(5)})
	})
	require.NotPanics(t, func() {
		NewDecCoinFromCoin(Coin{testDenom1, NewInt(0)})
	})
	require.Panics(t, func() {
		NewDecCoinFromCoin(Coin{strings.ToUpper(testDenom1), NewInt(5)})
	})
	require.Panics(t, func() {
		NewDecCoinFromCoin(Coin{testDenom1, NewInt(-5)})
	})
}

func TestDecCoinIsPositive(t *testing.T) {
	dc := NewInt64DecCoin(testDenom1, 5)
	require.True(t, dc.IsPositive())

	dc = NewInt64DecCoin(testDenom1, 0)
	require.False(t, dc.IsPositive())
}

func TestPlusDecCoin(t *testing.T) {
	decCoinA1 := NewDecCoinFromDec(testDenom1, NewDecWithPrec(11, 1))
	decCoinA2 := NewDecCoinFromDec(testDenom1, NewDecWithPrec(22, 1))
	decCoinB1 := NewDecCoinFromDec(testDenom2, NewDecWithPrec(11, 1))

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
		{DecCoins{{testDenom1, one}, {testDenom2, one}}, DecCoins{{testDenom1, one}, {testDenom2, one}}, DecCoins{{testDenom1, two}, {testDenom2, two}}},
		{DecCoins{{testDenom1, zero}, {testDenom2, one}}, DecCoins{{testDenom1, zero}, {testDenom2, zero}}, DecCoins{{testDenom2, one}}},
		{DecCoins{{testDenom1, zero}, {testDenom2, zero}}, DecCoins{{testDenom1, zero}, {testDenom2, zero}}, DecCoins(nil)},
	}

	for tcIndex, tc := range cases {
		res := tc.inputOne.Plus(tc.inputTwo)
		require.Equal(t, tc.expected, res, "sum of coins is incorrect, tc #%d", tcIndex)
	}
}

func TestSortDecCoins(t *testing.T) {
	good := DecCoins{
		NewInt64DecCoin("gas", 1),
		NewInt64DecCoin("mineral", 1),
		NewInt64DecCoin("tree", 1),
	}
	empty := DecCoins{
		NewInt64DecCoin("gold", 0),
	}
	badSort1 := DecCoins{
		NewInt64DecCoin("tree", 1),
		NewInt64DecCoin("gas", 1),
		NewInt64DecCoin("mineral", 1),
	}
	badSort2 := DecCoins{ // both are after the first one, but the second and third are in the wrong order
		NewInt64DecCoin("gas", 1),
		NewInt64DecCoin("tree", 1),
		NewInt64DecCoin("mineral", 1),
	}
	badAmt := DecCoins{
		NewInt64DecCoin("gas", 1),
		NewInt64DecCoin("tree", 0),
		NewInt64DecCoin("mineral", 1),
	}
	dup := DecCoins{
		NewInt64DecCoin("gas", 1),
		NewInt64DecCoin("gas", 1),
		NewInt64DecCoin("mineral", 1),
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
		{DecCoins{DecCoin{testDenom1, NewDec(5)}}, true},
		{DecCoins{DecCoin{testDenom1, NewDec(5)}, DecCoin{testDenom2, NewDec(100000)}}, true},
		{DecCoins{DecCoin{testDenom1, NewDec(-5)}}, false},
		{DecCoins{DecCoin{"AAA", NewDec(5)}}, false},
		{DecCoins{DecCoin{testDenom1, NewDec(5)}, DecCoin{"B", NewDec(100000)}}, false},
		{DecCoins{DecCoin{testDenom1, NewDec(5)}, DecCoin{testDenom2, NewDec(-100000)}}, false},
		{DecCoins{DecCoin{testDenom1, NewDec(-5)}, DecCoin{testDenom2, NewDec(100000)}}, false},
		{DecCoins{DecCoin{"AAA", NewDec(5)}, DecCoin{testDenom2, NewDec(100000)}}, false},
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
	}

	for i, tc := range testCases {
		res, err := ParseDecCoins(tc.input)
		if tc.expectedErr {
			require.Error(t, err, "expected error for test case #%d, input: %v", i, tc.input)
		} else {
			require.NoError(t, err, "unexpected error for test case #%d, input: %v", i, tc.input)
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
