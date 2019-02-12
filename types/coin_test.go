package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ----------------------------------------------------------------------------
// Coin tests

func TestCoin(t *testing.T) {
	require.Panics(t, func() { NewInt64Coin("aaa", -1) })
	require.Panics(t, func() { NewCoin("aaa", NewInt(-1)) })
	require.Panics(t, func() { NewInt64Coin("Atom", 10) })
	require.Panics(t, func() { NewCoin("Atom", NewInt(10)) })
	require.Equal(t, NewInt(5), NewInt64Coin("aaa", 5).Amount)
	require.Equal(t, NewInt(5), NewCoin("aaa", NewInt(5)).Amount)
}

func TestSameDenomAsCoin(t *testing.T) {
	cases := []struct {
		inputOne Coin
		inputTwo Coin
		expected bool
	}{
		{NewInt64Coin("aaa", 1), NewInt64Coin("aaa", 1), true},
		{NewInt64Coin("aaa", 1), NewInt64Coin("bbb", 1), false},
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
		{NewInt64Coin("aaa", 1), NewInt64Coin("aaa", 1), true},
		{NewInt64Coin("aaa", 1), NewInt64Coin("bbb", 1), false},
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
		{NewInt64Coin("aaa", 1), NewInt64Coin("aaa", 1), NewInt64Coin("aaa", 2), false},
		{NewInt64Coin("aaa", 1), NewInt64Coin("aaa", 0), NewInt64Coin("aaa", 1), false},
		{NewInt64Coin("aaa", 1), NewInt64Coin("bbb", 1), NewInt64Coin("aaa", 1), true},
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
		{NewInt64Coin("aaa", 1), NewInt64Coin("bbb", 1), NewInt64Coin("aaa", 1), true},
		{NewInt64Coin("aaa", 10), NewInt64Coin("aaa", 1), NewInt64Coin("aaa", 9), false},
		{NewInt64Coin("aaa", 5), NewInt64Coin("aaa", 3), NewInt64Coin("aaa", 2), false},
		{NewInt64Coin("aaa", 5), NewInt64Coin("aaa", 0), NewInt64Coin("aaa", 5), false},
		{NewInt64Coin("aaa", 1), NewInt64Coin("aaa", 5), Coin{}, true},
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
	}{NewInt64Coin("aaa", 1), NewInt64Coin("aaa", 1), 0}
	res := tc.inputOne.Minus(tc.inputTwo)
	require.Equal(t, tc.expected, res.Amount.Int64())
}

func TestIsGTECoin(t *testing.T) {
	cases := []struct {
		inputOne Coin
		inputTwo Coin
		expected bool
	}{
		{NewInt64Coin("aaa", 1), NewInt64Coin("aaa", 1), true},
		{NewInt64Coin("aaa", 2), NewInt64Coin("aaa", 1), true},
		{NewInt64Coin("aaa", 1), NewInt64Coin("bbb", 1), false},
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
		{NewInt64Coin("aaa", 1), NewInt64Coin("aaa", 1), false},
		{NewInt64Coin("aaa", 2), NewInt64Coin("aaa", 1), false},
		{NewInt64Coin("aaa", 0), NewInt64Coin("bbb", 1), false},
		{NewInt64Coin("aaa", 1), NewInt64Coin("bbb", 1), false},
		{NewInt64Coin("aaa", 1), NewInt64Coin("aaa", 1), false},
		{NewInt64Coin("aaa", 1), NewInt64Coin("aaa", 2), true},
	}

	for tcIndex, tc := range cases {
		res := tc.inputOne.IsLT(tc.inputTwo)
		require.Equal(t, tc.expected, res, "coin LT relation is incorrect, tc #%d", tcIndex)
	}
}

func TestCoinIsZero(t *testing.T) {
	coin := NewInt64Coin("aaa", 0)
	res := coin.IsZero()
	require.True(t, res)

	coin = NewInt64Coin("aaa", 1)
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
		{Coins{NewInt64Coin("aaa", 0)}, true},
		{Coins{NewInt64Coin("aaa", 0), NewInt64Coin("bbb", 0)}, true},
		{Coins{NewInt64Coin("aaa", 1)}, false},
		{Coins{NewInt64Coin("aaa", 0), NewInt64Coin("bbb", 1)}, false},
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
		{Coins{NewInt64Coin("aaa", 0)}, Coins{NewInt64Coin("aaa", 0)}, true},
		{Coins{NewInt64Coin("aaa", 0), NewInt64Coin("bbb", 1)}, Coins{NewInt64Coin("aaa", 0), NewInt64Coin("bbb", 1)}, true},
		{Coins{NewInt64Coin("aaa", 0)}, Coins{NewInt64Coin("bbb", 0)}, false},
		{Coins{NewInt64Coin("aaa", 0)}, Coins{NewInt64Coin("aaa", 1)}, false},
		{Coins{NewInt64Coin("aaa", 0)}, Coins{NewInt64Coin("aaa", 0), NewInt64Coin("bbb", 1)}, false},
		{Coins{NewInt64Coin("aaa", 0), NewInt64Coin("bbb", 1)}, Coins{NewInt64Coin("bbb", 1), NewInt64Coin("aaa", 0)}, true},
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
		{Coins{{"aaa", one}, {"bbb", one}}, Coins{{"aaa", one}, {"bbb", one}}, Coins{{"aaa", two}, {"bbb", two}}},
		{Coins{{"aaa", zero}, {"bbb", one}}, Coins{{"aaa", zero}, {"bbb", zero}}, Coins{{"bbb", one}}},
		{Coins{{"aaa", two}}, Coins{{"bbb", zero}}, Coins{{"aaa", two}}},
		{Coins{{"aaa", one}}, Coins{{"aaa", one}, {"bbb", two}}, Coins{{"aaa", two}, {"bbb", two}}},
		{Coins{{"aaa", zero}, {"bbb", zero}}, Coins{{"aaa", zero}, {"bbb", zero}}, Coins(nil)},
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
		{Coins{{"aaa", two}}, Coins{{"aaa", one}, {"bbb", two}}, Coins{{"aaa", one}, {"bbb", two}}, true},
		{Coins{{"aaa", two}}, Coins{{"bbb", zero}}, Coins{{"aaa", two}}, false},
		{Coins{{"aaa", one}}, Coins{{"bbb", zero}}, Coins{{"aaa", one}}, false},
		{Coins{{"aaa", one}, {"bbb", one}}, Coins{{"aaa", one}}, Coins{{"bbb", one}}, false},
		{Coins{{"aaa", one}, {"bbb", one}}, Coins{{"aaa", two}}, Coins{}, true},
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
		{"gas", NewInt(1)},
		{"mineral", NewInt(1)},
		{"tree", NewInt(1)},
	}
	mixedCase1 := Coins{
		{"gAs", NewInt(1)},
		{"MineraL", NewInt(1)},
		{"TREE", NewInt(1)},
	}
	mixedCase2 := Coins{
		{"gAs", NewInt(1)},
		{"mineral", NewInt(1)},
	}
	mixedCase3 := Coins{
		{"gAs", NewInt(1)},
	}
	empty := Coins{
		{"gold", NewInt(0)},
	}
	null := Coins{}
	badSort1 := Coins{
		{"tree", NewInt(1)},
		{"gas", NewInt(1)},
		{"mineral", NewInt(1)},
	}

	// both are after the first one, but the second and third are in the wrong order
	badSort2 := Coins{
		{"gas", NewInt(1)},
		{"tree", NewInt(1)},
		{"mineral", NewInt(1)},
	}
	badAmt := Coins{
		{"gas", NewInt(1)},
		{"tree", NewInt(0)},
		{"mineral", NewInt(1)},
	}
	dup := Coins{
		{"gas", NewInt(1)},
		{"gas", NewInt(1)},
		{"mineral", NewInt(1)},
	}
	neg := Coins{
		{"gas", NewInt(-1)},
		{"mineral", NewInt(1)},
	}

	assert.True(t, good.IsValid(), "Coins are valid")
	assert.False(t, mixedCase1.IsValid(), "Coins denoms contain upper case characters")
	assert.False(t, mixedCase2.IsValid(), "First Coins denoms contain upper case characters")
	assert.False(t, mixedCase3.IsValid(), "Single denom in Coins contains upper case characters")
	assert.True(t, good.IsPositive(), "Expected coins to be positive: %v", good)
	assert.False(t, null.IsPositive(), "Expected coins to not be positive: %v", null)
	assert.True(t, good.IsAllGTE(empty), "Expected %v to be >= %v", good, empty)
	assert.False(t, good.IsAllLT(empty), "Expected %v to be < %v", good, empty)
	assert.True(t, empty.IsAllLT(good), "Expected %v to be < %v", empty, good)
	assert.False(t, badSort1.IsValid(), "Coins are not sorted")
	assert.False(t, badSort2.IsValid(), "Coins are not sorted")
	assert.False(t, badAmt.IsValid(), "Coins cannot include 0 amounts")
	assert.False(t, dup.IsValid(), "Duplicate coin")
	assert.False(t, neg.IsValid(), "Negative first-denom coin")
}

func TestCoinsGT(t *testing.T) {
	one := NewInt(1)
	two := NewInt(2)

	assert.False(t, Coins{}.IsAllGT(Coins{}))
	assert.True(t, Coins{{"aaa", one}}.IsAllGT(Coins{}))
	assert.False(t, Coins{{"aaa", one}}.IsAllGT(Coins{{"aaa", one}}))
	assert.False(t, Coins{{"aaa", one}}.IsAllGT(Coins{{"bbb", one}}))
	assert.True(t, Coins{{"aaa", one}, {"bbb", one}}.IsAllGT(Coins{{"bbb", one}}))
	assert.False(t, Coins{{"aaa", one}, {"bbb", one}}.IsAllGT(Coins{{"bbb", two}}))
}

func TestCoinsGTE(t *testing.T) {
	one := NewInt(1)
	two := NewInt(2)

	assert.True(t, Coins{}.IsAllGTE(Coins{}))
	assert.True(t, Coins{{"aaa", one}}.IsAllGTE(Coins{}))
	assert.True(t, Coins{{"aaa", one}}.IsAllGTE(Coins{{"aaa", one}}))
	assert.False(t, Coins{{"aaa", one}}.IsAllGTE(Coins{{"bbb", one}}))
	assert.True(t, Coins{{"aaa", one}, {"bbb", one}}.IsAllGTE(Coins{{"bbb", one}}))
	assert.False(t, Coins{{"aaa", one}, {"bbb", one}}.IsAllGTE(Coins{{"bbb", two}}))
}

func TestCoinsLT(t *testing.T) {
	one := NewInt(1)
	two := NewInt(2)

	assert.False(t, Coins{}.IsAllLT(Coins{}))
	assert.False(t, Coins{{"aaa", one}}.IsAllLT(Coins{}))
	assert.False(t, Coins{{"aaa", one}}.IsAllLT(Coins{{"aaa", one}}))
	assert.False(t, Coins{{"aaa", one}}.IsAllLT(Coins{{"bbb", one}}))
	assert.False(t, Coins{{"aaa", one}, {"bbb", one}}.IsAllLT(Coins{{"bbb", one}}))
	assert.False(t, Coins{{"aaa", one}, {"bbb", one}}.IsAllLT(Coins{{"bbb", two}}))
	assert.False(t, Coins{{"aaa", one}, {"bbb", one}}.IsAllLT(Coins{{"aaa", one}, {"bbb", one}}))
	assert.True(t, Coins{{"aaa", one}, {"bbb", one}}.IsAllLT(Coins{{"aaa", one}, {"bbb", two}}))
	assert.True(t, Coins{}.IsAllLT(Coins{{"aaa", one}}))
}

func TestCoinsLTE(t *testing.T) {
	one := NewInt(1)
	two := NewInt(2)

	assert.True(t, Coins{}.IsAllLTE(Coins{}))
	assert.False(t, Coins{{"aaa", one}}.IsAllLTE(Coins{}))
	assert.True(t, Coins{{"aaa", one}}.IsAllLTE(Coins{{"aaa", one}}))
	assert.False(t, Coins{{"aaa", one}}.IsAllLTE(Coins{{"bbb", one}}))
	assert.False(t, Coins{{"aaa", one}, {"bbb", one}}.IsAllLTE(Coins{{"bbb", one}}))
	assert.False(t, Coins{{"aaa", one}, {"bbb", one}}.IsAllLTE(Coins{{"bbb", two}}))
	assert.True(t, Coins{{"aaa", one}, {"bbb", one}}.IsAllLTE(Coins{{"aaa", one}, {"bbb", one}}))
	assert.True(t, Coins{{"aaa", one}, {"bbb", one}}.IsAllLTE(Coins{{"aaa", one}, {"bbb", two}}))
	assert.True(t, Coins{}.IsAllLTE(Coins{{"aaa", one}}))
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
		NewInt64Coin("gas", 1),
		NewInt64Coin("mineral", 1),
		NewInt64Coin("tree", 1),
	}
	empty := Coins{
		NewInt64Coin("gold", 0),
	}
	badSort1 := Coins{
		NewInt64Coin("tree", 1),
		NewInt64Coin("gas", 1),
		NewInt64Coin("mineral", 1),
	}
	badSort2 := Coins{ // both are after the first one, but the second and third are in the wrong order
		NewInt64Coin("gas", 1),
		NewInt64Coin("tree", 1),
		NewInt64Coin("mineral", 1),
	}
	badAmt := Coins{
		NewInt64Coin("gas", 1),
		NewInt64Coin("tree", 0),
		NewInt64Coin("mineral", 1),
	}
	dup := Coins{
		NewInt64Coin("gas", 1),
		NewInt64Coin("gas", 1),
		NewInt64Coin("mineral", 1),
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
		NewInt64Coin("gold", 0),
	}
	case2 := Coins{
		NewInt64Coin("gas", 1),
		NewInt64Coin("mineral", 1),
		NewInt64Coin("tree", 1),
	}
	case3 := Coins{
		NewInt64Coin("mineral", 1),
		NewInt64Coin("tree", 1),
	}
	case4 := Coins{
		NewInt64Coin("gas", 8),
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
		{case2, 0, 0, 1, 1, 1},
		{case3, 0, 0, 0, 1, 1},
		{case4, 0, 0, 8, 0, 0},
	}

	for _, tc := range cases {
		assert.Equal(t, NewInt(tc.amountOfGAS), tc.coins.AmountOf("gas"))
		assert.Equal(t, NewInt(tc.amountOfMINERAL), tc.coins.AmountOf("mineral"))
		assert.Equal(t, NewInt(tc.amountOfTREE), tc.coins.AmountOf("tree"))
	}

	assert.Panics(t, func() { cases[0].coins.AmountOf("Invalid") })
}

func TestCoinsIsAnyGTE(t *testing.T) {
	one := NewInt(1)
	two := NewInt(2)

	assert.False(t, Coins{}.IsAnyGTE(Coins{}))
	assert.False(t, Coins{{"aaa", one}}.IsAnyGTE(Coins{}))
	assert.False(t, Coins{}.IsAnyGTE(Coins{{"aaa", one}}))
	assert.False(t, Coins{{"aaa", one}}.IsAnyGTE(Coins{{"aaa", two}}))
	assert.False(t, Coins{{"aaa", one}}.IsAnyGTE(Coins{{"bbb", one}}))
	assert.True(t, Coins{{"aaa", one}, {"bbb", two}}.IsAnyGTE(Coins{{"aaa", two}, {"bbb", one}}))
	assert.True(t, Coins{{"aaa", one}}.IsAnyGTE(Coins{{"aaa", one}}))
	assert.True(t, Coins{{"aaa", two}}.IsAnyGTE(Coins{{"aaa", one}}))
	assert.True(t, Coins{{"aaa", one}}.IsAnyGTE(Coins{{"aaa", one}, {"bbb", two}}))
	assert.True(t, Coins{{"bbb", two}}.IsAnyGTE(Coins{{"aaa", one}, {"bbb", two}}))
	assert.False(t, Coins{{"bbb", one}}.IsAnyGTE(Coins{{"aaa", one}, {"bbb", two}}))
	assert.True(t, Coins{{"aaa", one}, {"bbb", two}}.IsAnyGTE(Coins{{"aaa", one}, {"bbb", one}}))
	assert.True(t, Coins{{"aaa", one}, {"bbb", one}}.IsAnyGTE(Coins{{"aaa", one}, {"bbb", two}}))
	assert.True(t, Coins{{"xxx", one}, {"yyy", one}}.IsAnyGTE(Coins{{"bbb", one}, {"ccc", one}, {"yyy", one}, {"zzz", one}}))
}
