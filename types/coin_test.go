package types

import (
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testDenom1 = "atom"
	testDenom2 = "muon"
)

// ----------------------------------------------------------------------------
// Coin tests

func TestCoin(t *testing.T) {
	require.Panics(t, func() { NewInt64Coin(testDenom1, -1) })
	require.Panics(t, func() { NewCoin(testDenom1, NewInt(-1)) })
	require.Panics(t, func() { NewInt64Coin(strings.ToUpper(testDenom1), 10) })
	require.Panics(t, func() { NewCoin(strings.ToUpper(testDenom1), NewInt(10)) })
	require.Equal(t, NewInt(5), NewInt64Coin(testDenom1, 5).Amount)
	require.Equal(t, NewInt(5), NewCoin(testDenom1, NewInt(5)).Amount)
}

func TestIsEqualCoin(t *testing.T) {
	cases := []struct {
		inputOne Coin
		inputTwo Coin
		expected bool
		panics   bool
	}{
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom1, 1), true, false},
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom2, 1), false, true},
		{NewInt64Coin("stake", 1), NewInt64Coin("stake", 10), false, false},
	}

	for tcIndex, tc := range cases {
		if tc.panics {
			require.Panics(t, func() { tc.inputOne.IsEqual(tc.inputTwo) })
		} else {
			res := tc.inputOne.IsEqual(tc.inputTwo)
			require.Equal(t, tc.expected, res, "coin equality relation is incorrect, tc #%d", tcIndex)
		}
	}
}

func TestCoinIsValid(t *testing.T) {
	cases := []struct {
		coin       Coin
		expectPass bool
	}{
		{Coin{testDenom1, NewInt(-1)}, false},
		{Coin{testDenom1, NewInt(0)}, true},
		{Coin{testDenom1, NewInt(1)}, true},
		{Coin{"Atom", NewInt(1)}, false},
		{Coin{"a", NewInt(1)}, false},
		{Coin{"a very long coin denom", NewInt(1)}, false},
		{Coin{"atOm", NewInt(1)}, false},
		{Coin{"     ", NewInt(1)}, false},
	}

	for i, tc := range cases {
		require.Equal(t, tc.expectPass, tc.coin.IsValid(), "unexpected result for IsValid, tc #%d", i)
	}
}

func TestAddCoin(t *testing.T) {
	cases := []struct {
		inputOne    Coin
		inputTwo    Coin
		expected    Coin
		shouldPanic bool
	}{
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom1, 2), false},
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom1, 0), NewInt64Coin(testDenom1, 1), false},
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom2, 1), NewInt64Coin(testDenom1, 1), true},
	}

	for tcIndex, tc := range cases {
		if tc.shouldPanic {
			require.Panics(t, func() { tc.inputOne.Add(tc.inputTwo) })
		} else {
			res := tc.inputOne.Add(tc.inputTwo)
			require.Equal(t, tc.expected, res, "sum of coins is incorrect, tc #%d", tcIndex)
		}
	}
}

func TestSubCoin(t *testing.T) {
	cases := []struct {
		inputOne    Coin
		inputTwo    Coin
		expected    Coin
		shouldPanic bool
	}{
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom2, 1), NewInt64Coin(testDenom1, 1), true},
		{NewInt64Coin(testDenom1, 10), NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom1, 9), false},
		{NewInt64Coin(testDenom1, 5), NewInt64Coin(testDenom1, 3), NewInt64Coin(testDenom1, 2), false},
		{NewInt64Coin(testDenom1, 5), NewInt64Coin(testDenom1, 0), NewInt64Coin(testDenom1, 5), false},
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom1, 5), Coin{}, true},
	}

	for tcIndex, tc := range cases {
		if tc.shouldPanic {
			require.Panics(t, func() { tc.inputOne.Sub(tc.inputTwo) })
		} else {
			res := tc.inputOne.Sub(tc.inputTwo)
			require.Equal(t, tc.expected, res, "difference of coins is incorrect, tc #%d", tcIndex)
		}
	}

	tc := struct {
		inputOne Coin
		inputTwo Coin
		expected int64
	}{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom1, 1), 0}
	res := tc.inputOne.Sub(tc.inputTwo)
	require.Equal(t, tc.expected, res.Amount.Int64())
}

func TestIsGTECoin(t *testing.T) {
	cases := []struct {
		inputOne Coin
		inputTwo Coin
		expected bool
		panics   bool
	}{
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom1, 1), true, false},
		{NewInt64Coin(testDenom1, 2), NewInt64Coin(testDenom1, 1), true, false},
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom2, 1), false, true},
	}

	for tcIndex, tc := range cases {
		if tc.panics {
			require.Panics(t, func() { tc.inputOne.IsGTE(tc.inputTwo) })
		} else {
			res := tc.inputOne.IsGTE(tc.inputTwo)
			require.Equal(t, tc.expected, res, "coin GTE relation is incorrect, tc #%d", tcIndex)
		}
	}
}

func TestIsLTCoin(t *testing.T) {
	cases := []struct {
		inputOne Coin
		inputTwo Coin
		expected bool
		panics   bool
	}{
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom1, 1), false, false},
		{NewInt64Coin(testDenom1, 2), NewInt64Coin(testDenom1, 1), false, false},
		{NewInt64Coin(testDenom1, 0), NewInt64Coin(testDenom2, 1), false, true},
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom2, 1), false, true},
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom1, 1), false, false},
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom1, 2), true, false},
	}

	for tcIndex, tc := range cases {
		if tc.panics {
			require.Panics(t, func() { tc.inputOne.IsLT(tc.inputTwo) })
		} else {
			res := tc.inputOne.IsLT(tc.inputTwo)
			require.Equal(t, tc.expected, res, "coin LT relation is incorrect, tc #%d", tcIndex)
		}
	}
}

func TestCoinIsZero(t *testing.T) {
	coin := NewInt64Coin(testDenom1, 0)
	res := coin.IsZero()
	require.True(t, res)

	coin = NewInt64Coin(testDenom1, 1)
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
		{Coins{NewInt64Coin(testDenom1, 0)}, true},
		{Coins{NewInt64Coin(testDenom1, 0), NewInt64Coin(testDenom2, 0)}, true},
		{Coins{NewInt64Coin(testDenom1, 1)}, false},
		{Coins{NewInt64Coin(testDenom1, 0), NewInt64Coin(testDenom2, 1)}, false},
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
		panics   bool
	}{
		{Coins{}, Coins{}, true, false},
		{Coins{NewInt64Coin(testDenom1, 0)}, Coins{NewInt64Coin(testDenom1, 0)}, true, false},
		{Coins{NewInt64Coin(testDenom1, 0), NewInt64Coin(testDenom2, 1)}, Coins{NewInt64Coin(testDenom1, 0), NewInt64Coin(testDenom2, 1)}, true, false},
		{Coins{NewInt64Coin(testDenom1, 0)}, Coins{NewInt64Coin(testDenom2, 0)}, false, true},
		{Coins{NewInt64Coin(testDenom1, 0)}, Coins{NewInt64Coin(testDenom1, 1)}, false, false},
		{Coins{NewInt64Coin(testDenom1, 0)}, Coins{NewInt64Coin(testDenom1, 0), NewInt64Coin(testDenom2, 1)}, false, false},
		{Coins{NewInt64Coin(testDenom1, 0), NewInt64Coin(testDenom2, 1)}, Coins{NewInt64Coin(testDenom1, 0), NewInt64Coin(testDenom2, 1)}, true, false},
	}

	for tcnum, tc := range cases {
		if tc.panics {
			require.Panics(t, func() { tc.inputOne.IsEqual(tc.inputTwo) })
		} else {
			res := tc.inputOne.IsEqual(tc.inputTwo)
			require.Equal(t, tc.expected, res, "Equality is differed from exported. tc #%d, expected %b, actual %b.", tcnum, tc.expected, res)
		}
	}
}

func TestAddCoins(t *testing.T) {
	zero := NewInt(0)
	one := NewInt(1)
	two := NewInt(2)

	cases := []struct {
		inputOne Coins
		inputTwo Coins
		expected Coins
	}{
		{Coins{{testDenom1, one}, {testDenom2, one}}, Coins{{testDenom1, one}, {testDenom2, one}}, Coins{{testDenom1, two}, {testDenom2, two}}},
		{Coins{{testDenom1, zero}, {testDenom2, one}}, Coins{{testDenom1, zero}, {testDenom2, zero}}, Coins{{testDenom2, one}}},
		{Coins{{testDenom1, two}}, Coins{{testDenom2, zero}}, Coins{{testDenom1, two}}},
		{Coins{{testDenom1, one}}, Coins{{testDenom1, one}, {testDenom2, two}}, Coins{{testDenom1, two}, {testDenom2, two}}},
		{Coins{{testDenom1, zero}, {testDenom2, zero}}, Coins{{testDenom1, zero}, {testDenom2, zero}}, Coins(nil)},
	}

	for tcIndex, tc := range cases {
		res := tc.inputOne.Add(tc.inputTwo)
		assert.True(t, res.IsValid())
		require.Equal(t, tc.expected, res, "sum of coins is incorrect, tc #%d", tcIndex)
	}
}

func TestSubCoins(t *testing.T) {
	zero := NewInt(0)
	one := NewInt(1)
	two := NewInt(2)

	testCases := []struct {
		inputOne    Coins
		inputTwo    Coins
		expected    Coins
		shouldPanic bool
	}{
		{Coins{{testDenom1, two}}, Coins{{testDenom1, one}, {testDenom2, two}}, Coins{{testDenom1, one}, {testDenom2, two}}, true},
		{Coins{{testDenom1, two}}, Coins{{testDenom2, zero}}, Coins{{testDenom1, two}}, false},
		{Coins{{testDenom1, one}}, Coins{{testDenom2, zero}}, Coins{{testDenom1, one}}, false},
		{Coins{{testDenom1, one}, {testDenom2, one}}, Coins{{testDenom1, one}}, Coins{{testDenom2, one}}, false},
		{Coins{{testDenom1, one}, {testDenom2, one}}, Coins{{testDenom1, two}}, Coins{}, true},
	}

	for i, tc := range testCases {
		if tc.shouldPanic {
			require.Panics(t, func() { tc.inputOne.Sub(tc.inputTwo) })
		} else {
			res := tc.inputOne.Sub(tc.inputTwo)
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
	empty := NewCoins()
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
	assert.True(t, good.IsAllPositive(), "Expected coins to be positive: %v", good)
	assert.False(t, empty.IsAllPositive(), "Expected coins to not be positive: %v", empty)
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
	assert.True(t, Coins{{testDenom1, one}}.IsAllGT(Coins{}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGT(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGT(Coins{{testDenom2, one}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, two}}.IsAllGT(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllGT(Coins{{testDenom2, two}}))
}

func TestCoinsLT(t *testing.T) {
	one := NewInt(1)
	two := NewInt(2)

	assert.False(t, Coins{}.IsAllLT(Coins{}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllLT(Coins{}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllLT(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllLT(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLT(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLT(Coins{{testDenom2, two}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLT(Coins{{testDenom1, one}, {testDenom2, one}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLT(Coins{{testDenom1, two}, {testDenom2, two}}))
	assert.True(t, Coins{}.IsAllLT(Coins{{testDenom1, one}}))
}

func TestCoinsLTE(t *testing.T) {
	one := NewInt(1)
	two := NewInt(2)

	assert.True(t, Coins{}.IsAllLTE(Coins{}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllLTE(Coins{}))
	assert.True(t, Coins{{testDenom1, one}}.IsAllLTE(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllLTE(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLTE(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLTE(Coins{{testDenom2, two}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLTE(Coins{{testDenom1, one}, {testDenom2, one}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.True(t, Coins{}.IsAllLTE(Coins{{testDenom1, one}}))
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
	assert.False(t, Coins{{testDenom1, one}}.IsAnyGTE(Coins{}))
	assert.False(t, Coins{}.IsAnyGTE(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAnyGTE(Coins{{testDenom1, two}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAnyGTE(Coins{{testDenom2, one}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, two}}.IsAnyGTE(Coins{{testDenom1, two}, {testDenom2, one}}))
	assert.True(t, Coins{{testDenom1, one}}.IsAnyGTE(Coins{{testDenom1, one}}))
	assert.True(t, Coins{{testDenom1, two}}.IsAnyGTE(Coins{{testDenom1, one}}))
	assert.True(t, Coins{{testDenom1, one}}.IsAnyGTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.True(t, Coins{{testDenom2, two}}.IsAnyGTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.False(t, Coins{{testDenom2, one}}.IsAnyGTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, two}}.IsAnyGTE(Coins{{testDenom1, one}, {testDenom2, one}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAnyGTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.True(t, Coins{{"xxx", one}, {"yyy", one}}.IsAnyGTE(Coins{{testDenom2, one}, {"ccc", one}, {"yyy", one}, {"zzz", one}}))
}

func TestCoinsIsAllGT(t *testing.T) {
	one := NewInt(1)
	two := NewInt(2)

	assert.False(t, Coins{}.IsAllGT(Coins{}))
	assert.True(t, Coins{{testDenom1, one}}.IsAllGT(Coins{}))
	assert.False(t, Coins{}.IsAllGT(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGT(Coins{{testDenom1, two}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGT(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, two}}.IsAllGT(Coins{{testDenom1, two}, {testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGT(Coins{{testDenom1, one}}))
	assert.True(t, Coins{{testDenom1, two}}.IsAllGT(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGT(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.False(t, Coins{{testDenom2, two}}.IsAllGT(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.False(t, Coins{{testDenom2, one}}.IsAllGT(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, two}}.IsAllGT(Coins{{testDenom1, one}, {testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllGT(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.False(t, Coins{{"xxx", one}, {"yyy", one}}.IsAllGT(Coins{{testDenom2, one}, {"ccc", one}, {"yyy", one}, {"zzz", one}}))
}

func TestCoinsIsAllGTE(t *testing.T) {
	one := NewInt(1)
	two := NewInt(2)

	assert.True(t, Coins{}.IsAllGTE(Coins{}))
	assert.True(t, Coins{{testDenom1, one}}.IsAllGTE(Coins{}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllGTE(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllGTE(Coins{{testDenom2, two}}))
	assert.False(t, Coins{}.IsAllGTE(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGTE(Coins{{testDenom1, two}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGTE(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, two}}.IsAllGTE(Coins{{testDenom1, two}, {testDenom2, one}}))
	assert.True(t, Coins{{testDenom1, one}}.IsAllGTE(Coins{{testDenom1, one}}))
	assert.True(t, Coins{{testDenom1, two}}.IsAllGTE(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.False(t, Coins{{testDenom2, two}}.IsAllGTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.False(t, Coins{{testDenom2, one}}.IsAllGTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, two}}.IsAllGTE(Coins{{testDenom1, one}, {testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllGTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.False(t, Coins{{"xxx", one}, {"yyy", one}}.IsAllGTE(Coins{{testDenom2, one}, {"ccc", one}, {"yyy", one}, {"zzz", one}}))
}

func TestNewCoins(t *testing.T) {
	tenatom := NewInt64Coin("atom", 10)
	tenbtc := NewInt64Coin("btc", 10)
	zeroeth := NewInt64Coin("eth", 0)
	tests := []struct {
		name      string
		coins     Coins
		want      Coins
		wantPanic bool
	}{
		{"empty args", []Coin{}, Coins{}, false},
		{"one coin", []Coin{tenatom}, Coins{tenatom}, false},
		{"sort after create", []Coin{tenbtc, tenatom}, Coins{tenatom, tenbtc}, false},
		{"sort and remove zeroes", []Coin{zeroeth, tenbtc, tenatom}, Coins{tenatom, tenbtc}, false},
		{"panic on dups", []Coin{tenatom, tenatom}, Coins{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				require.Panics(t, func() { NewCoins(tt.coins...) })
				return
			}
			got := NewCoins(tt.coins...)
			require.True(t, got.IsEqual(tt.want))
		})
	}
}

func TestCoinsIsAnyGT(t *testing.T) {
	twoAtom := NewInt64Coin("atom", 2)
	fiveAtom := NewInt64Coin("atom", 5)
	threeEth := NewInt64Coin("eth", 3)
	sixEth := NewInt64Coin("eth", 6)
	twoBtc := NewInt64Coin("btc", 2)

	require.False(t, Coins{}.IsAnyGT(Coins{}))

	require.False(t, Coins{fiveAtom}.IsAnyGT(Coins{}))
	require.False(t, Coins{}.IsAnyGT(Coins{fiveAtom}))
	require.True(t, Coins{fiveAtom}.IsAnyGT(Coins{twoAtom}))
	require.False(t, Coins{twoAtom}.IsAnyGT(Coins{fiveAtom}))

	require.True(t, Coins{twoAtom, sixEth}.IsAnyGT(Coins{twoBtc, fiveAtom, threeEth}))
	require.False(t, Coins{twoBtc, twoAtom, threeEth}.IsAnyGT(Coins{fiveAtom, sixEth}))
	require.False(t, Coins{twoAtom, sixEth}.IsAnyGT(Coins{twoBtc, fiveAtom}))
}

func TestFindDup(t *testing.T) {
	abc := NewInt64Coin("abc", 10)
	def := NewInt64Coin("def", 10)
	ghi := NewInt64Coin("ghi", 10)

	type args struct {
		coins Coins
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"empty", args{NewCoins()}, -1},
		{"one coin", args{NewCoins(NewInt64Coin("xyz", 10))}, -1},
		{"no dups", args{Coins{abc, def, ghi}}, -1},
		{"dup at first position", args{Coins{abc, abc, def}}, 1},
		{"dup after first position", args{Coins{abc, def, def}}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findDup(tt.args.coins); got != tt.want {
				t.Errorf("findDup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarshalJSONCoins(t *testing.T) {
	cdc := codec.New()
	RegisterCodec(cdc)

	testCases := []struct {
		name      string
		input     Coins
		strOutput string
	}{
		{"nil coins", nil, `[]`},
		{"empty coins", Coins{}, `[]`},
		{"non-empty coins", NewCoins(NewInt64Coin("foo", 50)), `[{"denom":"foo","amount":"50"}]`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bz, err := cdc.MarshalJSON(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.strOutput, string(bz))

			var newCoins Coins
			require.NoError(t, cdc.UnmarshalJSON(bz, &newCoins))

			if tc.input.Empty() {
				require.Nil(t, newCoins)
			} else {
				require.Equal(t, tc.input, newCoins)
			}
		})
	}
}
