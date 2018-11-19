package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ----------------------------------------------------------------------------
// Coin tests

func TestCoin(t *testing.T) {
	require.Panics(t, func() { NewInt64Coin("A", -1) })
	require.Panics(t, func() { NewCoin("A", NewInt(-1)) })
	require.Equal(t, NewInt(5), NewInt64Coin("A", 5).Amount)
	require.Equal(t, NewInt(5), NewCoin("A", NewInt(5)).Amount)
}

func TestSameDenomAsCoin(t *testing.T) {
	cases := []struct {
		inputOne Coin
		inputTwo Coin
		expected bool
	}{
		{NewInt64Coin("A", 1), NewInt64Coin("A", 1), true},
		{NewInt64Coin("A", 1), NewInt64Coin("a", 1), false},
		{NewInt64Coin("a", 1), NewInt64Coin("b", 1), false},
		{NewInt64Coin("steak", 1), NewInt64Coin("steak", 10), true},
	}

	for tcIndex, tc := range cases {
		res := tc.inputOne.SameDenomAs(tc.inputTwo)
		require.Equal(t, tc.expected, res, "coin denominations didn't match, tc #%d", tcIndex)
	}
}

func TestIsEqualCoin(t *testing.T) {
	cases := []struct {
		inputOne Coin
		inputTwo Coin
		expected bool
	}{
		{NewInt64Coin("A", 1), NewInt64Coin("A", 1), true},
		{NewInt64Coin("A", 1), NewInt64Coin("a", 1), false},
		{NewInt64Coin("a", 1), NewInt64Coin("b", 1), false},
		{NewInt64Coin("steak", 1), NewInt64Coin("steak", 10), false},
	}

	for tcIndex, tc := range cases {
		res := tc.inputOne.IsEqual(tc.inputTwo)
		require.Equal(t, tc.expected, res, "coin equality relation is incorrect, tc #%d", tcIndex)
	}
}

func TestPlusCoin(t *testing.T) {
	cases := []struct {
		inputOne    Coin
		inputTwo    Coin
		expected    Coin
		shouldPanic bool
	}{
		{NewInt64Coin("A", 1), NewInt64Coin("A", 1), NewInt64Coin("A", 2), false},
		{NewInt64Coin("A", 1), NewInt64Coin("A", 0), NewInt64Coin("A", 1), false},
		{NewInt64Coin("A", 1), NewInt64Coin("B", 1), NewInt64Coin("A", 1), true},
	}

	for tcIndex, tc := range cases {
		if tc.shouldPanic {
			require.Panics(t, func() { tc.inputOne.Plus(tc.inputTwo) })
		} else {
			res := tc.inputOne.Plus(tc.inputTwo)
			require.Equal(t, tc.expected, res, "sum of coins is incorrect, tc #%d", tcIndex)
		}
	}
}

func TestMinusCoin(t *testing.T) {
	cases := []struct {
		inputOne    Coin
		inputTwo    Coin
		expected    Coin
		shouldPanic bool
	}{
		{NewInt64Coin("A", 1), NewInt64Coin("B", 1), NewInt64Coin("A", 1), true},
		{NewInt64Coin("A", 10), NewInt64Coin("A", 1), NewInt64Coin("A", 9), false},
		{NewInt64Coin("A", 5), NewInt64Coin("A", 3), NewInt64Coin("A", 2), false},
		{NewInt64Coin("A", 5), NewInt64Coin("A", 0), NewInt64Coin("A", 5), false},
		{NewInt64Coin("A", 1), NewInt64Coin("A", 5), Coin{}, true},
	}

	for tcIndex, tc := range cases {
		if tc.shouldPanic {
			require.Panics(t, func() { tc.inputOne.Minus(tc.inputTwo) })
		} else {
			res := tc.inputOne.Minus(tc.inputTwo)
			require.Equal(t, tc.expected, res, "difference of coins is incorrect, tc #%d", tcIndex)
		}
	}

	tc := struct {
		inputOne Coin
		inputTwo Coin
		expected int64
	}{NewInt64Coin("A", 1), NewInt64Coin("A", 1), 0}
	res := tc.inputOne.Minus(tc.inputTwo)
	require.Equal(t, tc.expected, res.Amount.Int64())
}

func TestIsGTECoin(t *testing.T) {
	cases := []struct {
		inputOne Coin
		inputTwo Coin
		expected bool
	}{
		{NewInt64Coin("A", 1), NewInt64Coin("A", 1), true},
		{NewInt64Coin("A", 2), NewInt64Coin("A", 1), true},
		{NewInt64Coin("A", 1), NewInt64Coin("B", 1), false},
	}

	for tcIndex, tc := range cases {
		res := tc.inputOne.IsGTE(tc.inputTwo)
		require.Equal(t, tc.expected, res, "coin GTE relation is incorrect, tc #%d", tcIndex)
	}
}

func TestIsLTCoin(t *testing.T) {
	cases := []struct {
		inputOne Coin
		inputTwo Coin
		expected bool
	}{
		{NewInt64Coin("A", 1), NewInt64Coin("A", 1), false},
		{NewInt64Coin("A", 2), NewInt64Coin("A", 1), false},
		{NewInt64Coin("a", 0), NewInt64Coin("b", 1), false},
		{NewInt64Coin("a", 1), NewInt64Coin("b", 1), false},
		{NewInt64Coin("a", 1), NewInt64Coin("a", 1), false},
		{NewInt64Coin("a", 1), NewInt64Coin("a", 2), true},
	}

	for tcIndex, tc := range cases {
		res := tc.inputOne.IsLT(tc.inputTwo)
		require.Equal(t, tc.expected, res, "coin LT relation is incorrect, tc #%d", tcIndex)
	}
}

func TestCoinIsZero(t *testing.T) {
	coin := NewInt64Coin("A", 0)
	res := coin.IsZero()
	require.True(t, res)

	coin = NewInt64Coin("A", 1)
	res = coin.IsZero()
	require.False(t, res)
}

// ----------------------------------------------------------------------------
// Coins tests

func TestIsZeroCoins(t *testing.T) {
	cases := []struct {
		inputOne Coins
		expected bool
	}{
		{Coins{}, true},
		{Coins{NewInt64Coin("A", 0)}, true},
		{Coins{NewInt64Coin("A", 0), NewInt64Coin("B", 0)}, true},
		{Coins{NewInt64Coin("A", 1)}, false},
		{Coins{NewInt64Coin("A", 0), NewInt64Coin("B", 1)}, false},
	}

	for _, tc := range cases {
		res := tc.inputOne.IsZero()
		require.Equal(t, tc.expected, res)
	}
}

func TestEqualCoins(t *testing.T) {
	cases := []struct {
		inputOne Coins
		inputTwo Coins
		expected bool
	}{
		{Coins{}, Coins{}, true},
		{Coins{NewInt64Coin("A", 0)}, Coins{NewInt64Coin("A", 0)}, true},
		{Coins{NewInt64Coin("A", 0), NewInt64Coin("B", 1)}, Coins{NewInt64Coin("A", 0), NewInt64Coin("B", 1)}, true},
		{Coins{NewInt64Coin("A", 0)}, Coins{NewInt64Coin("B", 0)}, false},
		{Coins{NewInt64Coin("A", 0)}, Coins{NewInt64Coin("A", 1)}, false},
		{Coins{NewInt64Coin("A", 0)}, Coins{NewInt64Coin("A", 0), NewInt64Coin("B", 1)}, false},
		{Coins{NewInt64Coin("A", 0), NewInt64Coin("B", 1)}, Coins{NewInt64Coin("B", 1), NewInt64Coin("A", 0)}, true},
	}

	for tcnum, tc := range cases {
		res := tc.inputOne.IsEqual(tc.inputTwo)
		require.Equal(t, tc.expected, res, "Equality is differed from expected. tc #%d, expected %b, actual %b.", tcnum, tc.expected, res)
	}
}

func TestPlusCoins(t *testing.T) {
	zero := NewInt(0)
	one := NewInt(1)
	two := NewInt(2)

	cases := []struct {
		inputOne Coins
		inputTwo Coins
		expected Coins
	}{
		{Coins{{"A", one}, {"B", one}}, Coins{{"A", one}, {"B", one}}, Coins{{"A", two}, {"B", two}}},
		{Coins{{"A", zero}, {"B", one}}, Coins{{"A", zero}, {"B", zero}}, Coins{{"B", one}}},
		{Coins{{"A", two}}, Coins{{"B", zero}}, Coins{{"A", two}}},
		{Coins{{"A", one}}, Coins{{"A", one}, {"B", two}}, Coins{{"A", two}, {"B", two}}},
		{Coins{{"A", zero}, {"B", zero}}, Coins{{"A", zero}, {"B", zero}}, Coins(nil)},
	}

	for tcIndex, tc := range cases {
		res := tc.inputOne.Plus(tc.inputTwo)
		assert.True(t, res.IsValid())
		require.Equal(t, tc.expected, res, "sum of coins is incorrect, tc #%d", tcIndex)
	}
}

func TestMinusCoins(t *testing.T) {
	zero := NewInt(0)
	one := NewInt(1)
	two := NewInt(2)

	testCases := []struct {
		inputOne    Coins
		inputTwo    Coins
		expected    Coins
		shouldPanic bool
	}{
		{Coins{{"A", two}}, Coins{{"A", one}, {"B", two}}, Coins{{"A", one}, {"B", two}}, true},
		{Coins{{"A", two}}, Coins{{"B", zero}}, Coins{{"A", two}}, false},
		{Coins{{"A", one}}, Coins{{"B", zero}}, Coins{{"A", one}}, false},
		{Coins{{"A", one}, {"B", one}}, Coins{{"A", one}}, Coins{{"B", one}}, false},
		{Coins{{"A", one}, {"B", one}}, Coins{{"A", two}}, Coins{}, true},
	}

	for i, tc := range testCases {
		if tc.shouldPanic {
			require.Panics(t, func() { tc.inputOne.Minus(tc.inputTwo) })
		} else {
			res := tc.inputOne.Minus(tc.inputTwo)
			assert.True(t, res.IsValid())
			require.Equal(t, tc.expected, res, "sum of coins is incorrect, tc #%d", i)
		}
	}
}

func TestCoins(t *testing.T) {
	good := Coins{
		{"GAS", NewInt(1)},
		{"MINERAL", NewInt(1)},
		{"TREE", NewInt(1)},
	}
	empty := Coins{
		{"GOLD", NewInt(0)},
	}
	null := Coins{}
	badSort1 := Coins{
		{"TREE", NewInt(1)},
		{"GAS", NewInt(1)},
		{"MINERAL", NewInt(1)},
	}

	// both are after the first one, but the second and third are in the wrong order
	badSort2 := Coins{
		{"GAS", NewInt(1)},
		{"TREE", NewInt(1)},
		{"MINERAL", NewInt(1)},
	}
	badAmt := Coins{
		{"GAS", NewInt(1)},
		{"TREE", NewInt(0)},
		{"MINERAL", NewInt(1)},
	}
	dup := Coins{
		{"GAS", NewInt(1)},
		{"GAS", NewInt(1)},
		{"MINERAL", NewInt(1)},
	}

	assert.True(t, good.IsValid(), "Coins are valid")
	assert.True(t, good.IsPositive(), "Expected coins to be positive: %v", good)
	assert.False(t, null.IsPositive(), "Expected coins to not be positive: %v", null)
	assert.True(t, good.IsAllGTE(empty), "Expected %v to be >= %v", good, empty)
	assert.False(t, good.IsAllLT(empty), "Expected %v to be < %v", good, empty)
	assert.True(t, empty.IsAllLT(good), "Expected %v to be < %v", empty, good)
	assert.False(t, badSort1.IsValid(), "Coins are not sorted")
	assert.False(t, badSort2.IsValid(), "Coins are not sorted")
	assert.False(t, badAmt.IsValid(), "Coins cannot include 0 amounts")
	assert.False(t, dup.IsValid(), "Duplicate coin")
}

func TestCoinsGT(t *testing.T) {
	one := NewInt(1)
	two := NewInt(2)

	assert.False(t, Coins{}.IsAllGT(Coins{}))
	assert.True(t, Coins{{"A", one}}.IsAllGT(Coins{}))
	assert.False(t, Coins{{"A", one}}.IsAllGT(Coins{{"A", one}}))
	assert.False(t, Coins{{"A", one}}.IsAllGT(Coins{{"B", one}}))
	assert.True(t, Coins{{"A", one}, {"B", one}}.IsAllGT(Coins{{"B", one}}))
	assert.False(t, Coins{{"A", one}, {"B", one}}.IsAllGT(Coins{{"B", two}}))
}

func TestCoinsGTE(t *testing.T) {
	one := NewInt(1)
	two := NewInt(2)

	assert.True(t, Coins{}.IsAllGTE(Coins{}))
	assert.True(t, Coins{{"A", one}}.IsAllGTE(Coins{}))
	assert.True(t, Coins{{"A", one}}.IsAllGTE(Coins{{"A", one}}))
	assert.False(t, Coins{{"A", one}}.IsAllGTE(Coins{{"B", one}}))
	assert.True(t, Coins{{"A", one}, {"B", one}}.IsAllGTE(Coins{{"B", one}}))
	assert.False(t, Coins{{"A", one}, {"B", one}}.IsAllGTE(Coins{{"B", two}}))
}

func TestCoinsLT(t *testing.T) {
	one := NewInt(1)
	two := NewInt(2)

	assert.False(t, Coins{}.IsAllLT(Coins{}))
	assert.False(t, Coins{{"A", one}}.IsAllLT(Coins{}))
	assert.False(t, Coins{{"A", one}}.IsAllLT(Coins{{"A", one}}))
	assert.False(t, Coins{{"A", one}}.IsAllLT(Coins{{"B", one}}))
	assert.False(t, Coins{{"A", one}, {"B", one}}.IsAllLT(Coins{{"B", one}}))
	assert.False(t, Coins{{"A", one}, {"B", one}}.IsAllLT(Coins{{"B", two}}))
	assert.False(t, Coins{{"A", one}, {"B", one}}.IsAllLT(Coins{{"A", one}, {"B", one}}))
	assert.True(t, Coins{{"A", one}, {"B", one}}.IsAllLT(Coins{{"A", one}, {"B", two}}))
	assert.True(t, Coins{}.IsAllLT(Coins{{"A", one}}))
}

func TestCoinsLTE(t *testing.T) {
	one := NewInt(1)
	two := NewInt(2)

	assert.True(t, Coins{}.IsAllLTE(Coins{}))
	assert.False(t, Coins{{"A", one}}.IsAllLTE(Coins{}))
	assert.True(t, Coins{{"A", one}}.IsAllLTE(Coins{{"A", one}}))
	assert.False(t, Coins{{"A", one}}.IsAllLTE(Coins{{"B", one}}))
	assert.False(t, Coins{{"A", one}, {"B", one}}.IsAllLTE(Coins{{"B", one}}))
	assert.False(t, Coins{{"A", one}, {"B", one}}.IsAllLTE(Coins{{"B", two}}))
	assert.True(t, Coins{{"A", one}, {"B", one}}.IsAllLTE(Coins{{"A", one}, {"B", one}}))
	assert.True(t, Coins{{"A", one}, {"B", one}}.IsAllLTE(Coins{{"A", one}, {"B", two}}))
	assert.True(t, Coins{}.IsAllLTE(Coins{{"A", one}}))
}

func TestParse(t *testing.T) {
	one := NewInt(1)

	cases := []struct {
		input    string
		valid    bool  // if false, we expect an error on parse
		expected Coins // if valid is true, make sure this is returned
	}{
		{"", true, nil},
		{"1foo", true, Coins{{"foo", one}}},
		{"10bar", true, Coins{{"bar", NewInt(10)}}},
		{"99bar,1foo", true, Coins{{"bar", NewInt(99)}, {"foo", one}}},
		{"98 bar , 1 foo  ", true, Coins{{"bar", NewInt(98)}, {"foo", one}}},
		{"  55\t \t bling\n", true, Coins{{"bling", NewInt(55)}}},
		{"2foo, 97 bar", true, Coins{{"bar", NewInt(97)}, {"foo", NewInt(2)}}},
		{"5 mycoin,", false, nil},             // no empty coins in a list
		{"2 3foo, 97 bar", false, nil},        // 3foo is invalid coin name
		{"11me coin, 12you coin", false, nil}, // no spaces in coin names
		{"1.2btc", false, nil},                // amount must be integer
		{"5foo-bar", false, nil},              // once more, only letters in coin name
	}

	for tcIndex, tc := range cases {
		res, err := ParseCoins(tc.input)
		if !tc.valid {
			require.NotNil(t, err, "%s: %#v. tc #%d", tc.input, res, tcIndex)
		} else if assert.Nil(t, err, "%s: %+v", tc.input, err) {
			require.Equal(t, tc.expected, res, "coin parsing was incorrect, tc #%d", tcIndex)
		}
	}
}

func TestSortCoins(t *testing.T) {
	good := Coins{
		NewInt64Coin("GAS", 1),
		NewInt64Coin("MINERAL", 1),
		NewInt64Coin("TREE", 1),
	}
	empty := Coins{
		NewInt64Coin("GOLD", 0),
	}
	badSort1 := Coins{
		NewInt64Coin("TREE", 1),
		NewInt64Coin("GAS", 1),
		NewInt64Coin("MINERAL", 1),
	}
	badSort2 := Coins{ // both are after the first one, but the second and third are in the wrong order
		NewInt64Coin("GAS", 1),
		NewInt64Coin("TREE", 1),
		NewInt64Coin("MINERAL", 1),
	}
	badAmt := Coins{
		NewInt64Coin("GAS", 1),
		NewInt64Coin("TREE", 0),
		NewInt64Coin("MINERAL", 1),
	}
	dup := Coins{
		NewInt64Coin("GAS", 1),
		NewInt64Coin("GAS", 1),
		NewInt64Coin("MINERAL", 1),
	}

	cases := []struct {
		coins         Coins
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

func TestAmountOf(t *testing.T) {
	case0 := Coins{}
	case1 := Coins{
		NewInt64Coin("", 0),
	}
	case2 := Coins{
		NewInt64Coin(" ", 0),
	}
	case3 := Coins{
		NewInt64Coin("GOLD", 0),
	}
	case4 := Coins{
		NewInt64Coin("GAS", 1),
		NewInt64Coin("MINERAL", 1),
		NewInt64Coin("TREE", 1),
	}
	case5 := Coins{
		NewInt64Coin("MINERAL", 1),
		NewInt64Coin("TREE", 1),
	}
	case6 := Coins{
		NewInt64Coin("", 6),
	}
	case7 := Coins{
		NewInt64Coin(" ", 7),
	}
	case8 := Coins{
		NewInt64Coin("GAS", 8),
	}

	cases := []struct {
		coins           Coins
		amountOf        int64
		amountOfSpace   int64
		amountOfGAS     int64
		amountOfMINERAL int64
		amountOfTREE    int64
	}{
		{case0, 0, 0, 0, 0, 0},
		{case1, 0, 0, 0, 0, 0},
		{case2, 0, 0, 0, 0, 0},
		{case3, 0, 0, 0, 0, 0},
		{case4, 0, 0, 1, 1, 1},
		{case5, 0, 0, 0, 1, 1},
		{case6, 6, 0, 0, 0, 0},
		{case7, 0, 7, 0, 0, 0},
		{case8, 0, 0, 8, 0, 0},
	}

	for _, tc := range cases {
		assert.Equal(t, NewInt(tc.amountOf), tc.coins.AmountOf(""))
		assert.Equal(t, NewInt(tc.amountOfSpace), tc.coins.AmountOf(" "))
		assert.Equal(t, NewInt(tc.amountOfGAS), tc.coins.AmountOf("GAS"))
		assert.Equal(t, NewInt(tc.amountOfMINERAL), tc.coins.AmountOf("MINERAL"))
		assert.Equal(t, NewInt(tc.amountOfTREE), tc.coins.AmountOf("TREE"))
	}
}
