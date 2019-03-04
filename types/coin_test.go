package types

import (
	"testing"

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
	t.Parallel()
	require.Panics(t, func() { NewInt64Coin("b", 0) })

	require.NotPanics(t, func() { NewInt64Coin("btc", 0) })
	require.Panics(t, func() { NewPositiveInt64Coin("btc", 0) })
	require.Panics(t, func() { NewInt64Coin("atom", -1) })
	require.Panics(t, func() { NewPositiveInt64Coin("atom", -1) })

	require.NotPanics(t, func() { NewCoin("atom", NewInt(0)) })
	require.Panics(t, func() { NewCoin("atom", NewInt(-1)) })
	require.Panics(t, func() { NewPositiveCoin("atom", NewInt(0)) })
	require.Panics(t, func() { NewPositiveCoin("atom", NewInt(-1)) })

	// test denom case
	require.Panics(t, func() { NewInt64Coin("Atom", 10) })
	require.Panics(t, func() { NewPositiveInt64Coin("Atom", 10) })
	require.Panics(t, func() { NewCoin("Atom", NewInt(10)) })
	require.Panics(t, func() { NewPositiveCoin("Atom", NewInt(10)) })

	// test leading/trailing spaces

	require.Panics(t, func() { NewInt64Coin("atom ", 10) })
	require.Panics(t, func() { NewPositiveInt64Coin("atom ", 10) })
	require.Panics(t, func() { NewCoin("atom ", NewInt(10)) })
	require.Panics(t, func() { NewPositiveCoin("atom ", NewInt(10)) })

	require.Panics(t, func() { NewInt64Coin("atom ", 10) })
	require.Panics(t, func() { NewPositiveInt64Coin("atom ", 10) })
	require.Panics(t, func() { NewCoin("atom ", NewInt(10)) })
	require.Panics(t, func() { NewPositiveCoin("atom ", NewInt(10)) })

	require.Equal(t, NewInt(5), NewInt64Coin("btc", 5).Amount)
	require.Equal(t, NewInt(5), NewCoin("btc", NewInt(5)).Amount)
}

func TestSameDenomAsCoin(t *testing.T) {
	t.Parallel()
	cases := []struct {
		inputOne Coin
		inputTwo Coin
		expected bool
	}{
		{NewInt64Coin("atom", 1), NewInt64Coin("atom", 1), true},
		{NewInt64Coin("atom", 1), NewInt64Coin("btc", 1), false},
		{NewInt64Coin("steak", 1), NewInt64Coin("steak", 10), true},
	}

	for tcIndex, tc := range cases {
		res := tc.inputOne.SameDenomAs(tc.inputTwo)
		require.Equal(t, tc.expected, res, "coin denominations didn't match, tc #%d", tcIndex)
	}
}

func TestIsEqualCoin(t *testing.T) {
	t.Parallel()
	cases := []struct {
		inputOne Coin
		inputTwo Coin
		expected bool
		panics   bool
	}{
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom1, 1), true, false},
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom2, 1), false, true},
		{NewInt64Coin("steak", 1), NewInt64Coin("steak", 10), false, false},
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
			require.Equal(t, tc.expected, res, "Equality is differed from expected. tc #%d, expected %b, actual %b.", tcnum, tc.expected, res)
		}
	}
}

func TestAddCoins(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	assert.True(t, good.IsAllPositive(), "Expected coins to be positive: %v", good)
	assert.False(t, null.IsAllPositive(), "Expected coins to not be positive: %v", null)
	assert.True(t, good.IsAllGTE(empty), "Expected %v to be >= %v", good, empty)
	assert.False(t, good.IsAllLT(empty), "Expected %v to be < %v", good, empty)
	assert.True(t, empty.IsAllLT(good), "Expected %v to be < %v", empty, good)
	assert.Error(t, badSort1.Validate(false, true), "Coins are not sorted")
	assert.Error(t, badSort2.Validate(false, true), "Coins are not sorted")
	assert.Error(t, badAmt.Validate(false, true), "Coins cannot include 0 amounts")
	assert.Error(t, dup.Validate(false, true), "Duplicate coin")
	assert.Error(t, neg.Validate(false, true), "Negative first-denom coin")
}

func TestCoinsGT(t *testing.T) {
	t.Parallel()
	one := NewInt(1)
	two := NewInt(2)

	assert.False(t, Coins{}.IsAllGT(Coins{}))
	assert.True(t, Coins{{testDenom1, one}}.IsAllGT(Coins{}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGT(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGT(Coins{{testDenom2, one}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllGT(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllGT(Coins{{testDenom2, two}}))
}

func TestCoinsGTE(t *testing.T) {
	t.Parallel()
	one := NewInt(1)
	two := NewInt(2)

	assert.True(t, Coins{}.IsAllGTE(Coins{}))
	assert.True(t, Coins{{testDenom1, one}}.IsAllGTE(Coins{}))
	assert.True(t, Coins{{testDenom1, one}}.IsAllGTE(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGTE(Coins{{testDenom2, one}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllGTE(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllGTE(Coins{{testDenom2, two}}))
}

func TestCoinsLT(t *testing.T) {
	t.Parallel()
	one := NewInt(1)
	two := NewInt(2)

	assert.False(t, Coins{}.IsAllLT(Coins{}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllLT(Coins{}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllLT(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllLT(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLT(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLT(Coins{{testDenom2, two}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLT(Coins{{testDenom1, one}, {testDenom2, one}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLT(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.True(t, Coins{}.IsAllLT(Coins{{testDenom1, one}}))
}

func TestCoinsLTE(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	one := NewInt(1)

	cases := []struct {
		input         string
		valid         bool // if false, we expect an error on parse
		validPositive bool
		expected      Coins // if valid is true, make sure this is returned
	}{
		{"", true, false, nil},
		{"0", false, false, nil}, // // empty denom
		{"1foo", true, true, Coins{{"foo", one}}},
		{"10bar", true, true, Coins{{"bar", NewInt(10)}}},
		{"99bar,1foo", true, true, Coins{{"bar", NewInt(99)}, {"foo", one}}},
		{"98 bar , 1 foo  ", true, true, Coins{{"bar", NewInt(98)}, {"foo", one}}},
		{"  55\t \t bling\n", true, true, Coins{{"bling", NewInt(55)}}},
		{"2foo, 97 bar", true, true, Coins{{"bar", NewInt(97)}, {"foo", NewInt(2)}}},
		{"5 mycoin,", false, false, nil},             // no empty coins in a list
		{"2 3foo, 97 bar", false, false, nil},        // 3foo is invalid coin name
		{"11me coin, 12you coin", false, false, nil}, // no spaces in coin names
		{"1.2btc", false, false, nil},                // amount must be integer
		{"5foo-bar", false, false, nil},              // once more, only letters in coin name
		{"5foo,-3bar", false, false, nil},            // all coins must pass validation
		{"5.2foo", false, false, nil},                // decimal coin
		{"-5foo", false, false, nil},                 // negative coin
		{"0foo", false, false, nil},                  // invalid zero
		{"-0foo", false, false, nil},                 // negative zero
	}

	for tcIndex, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			res, err := ParseCoins(tc.input)
			if !tc.valid {
				require.NotNil(t, err, "%s: %#v. tc #%d", tc.input, res, tcIndex)
			} else if assert.Nil(t, err, "%s: %+v", tc.input, err) {
				require.Equal(t, tc.expected, res, "coin parsing was incorrect, tc #%d", tcIndex)
			}

			res, err = ParsePositiveCoins(tc.input)
			if !tc.validPositive {
				require.NotNil(t, err, "%s: %#v. tc #%d", tc.input, res, tcIndex)
			} else if assert.Nil(t, err, "%s: %+v", tc.input, err) {
				require.Equal(t, tc.expected, res, "coin parsing was incorrect, tc #%d", tcIndex)
			}
		})
	}
}

func TestSortCoins(t *testing.T) {
	t.Parallel()
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
		require.Equal(t, tc.before, tc.coins.Validate(false, true) == nil,
			"coin validity is incorrect before sorting, tc #%d", tcIndex)
		tc.coins.Sort()
		require.Equal(t, tc.after, tc.coins.Validate(false, true) == nil,
			"coin validity is incorrect after sorting, tc #%d", tcIndex)
	}
}

func TestAmountOf(t *testing.T) {
	t.Parallel()
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
		amountOfGAS     int64
		amountOfMINERAL int64
		amountOfTREE    int64
	}{
		{case0, 0, 0, 0},
		{case1, 0, 0, 0},
		{case2, 1, 1, 1},
		{case3, 0, 1, 1},
		{case4, 8, 0, 0},
	}

	for _, tc := range cases {
		assert.Equal(t, NewInt(tc.amountOfGAS), tc.coins.AmountOf("gas"))
		assert.Equal(t, NewInt(tc.amountOfMINERAL), tc.coins.AmountOf("mineral"))
		assert.Equal(t, NewInt(tc.amountOfTREE), tc.coins.AmountOf("tree"))
	}

	assert.Panics(t, func() { cases[0].coins.AmountOf("Invalid") })
}

func TestCoinsIsAnyGTE(t *testing.T) {
	t.Parallel()
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

func TestCoinString(t *testing.T) {
	type fields struct {
		Denom  string
		Amount Int
	}
	tests := []struct {
		name string
		coin Coin
		want string
	}{
		{"zero", NewInt64Coin("atom", 0), ""},
		{"value", NewInt64Coin("atom", 10), "10atom"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.coin.String())
		})
	}
}

func TestCoinsString(t *testing.T) {
	zero := NewInt64Coin("atom", 0)
	tests := []struct {
		name  string
		coins Coins
		want  string
	}{
		{"zero", Coins{zero}, ""},
		{"value", Coins{NewInt64Coin("atom", 10)}, "10atom"},
		{"zero,positive", Coins{zero, NewInt64Coin("atom", 10)}, "10atom"},
		{"order does not matter", Coins{zero, NewInt64Coin("atom", 10)}, Coins{NewInt64Coin("atom", 10), zero}.String()},
		{"sort", Coins{NewInt64Coin("btc", 5), NewInt64Coin("atom", 10)}, "10atom,5btc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.coins.String())
		})
	}
}

func TestCoinsValidate(t *testing.T) {
	type args struct {
		failEmpty bool
		failZero  bool
	}
	tests := []struct {
		name    string
		coins   Coins
		args    args
		wantErr bool
	}{
		{"valid", Coins{NewInt64Coin("abc", 10), NewInt64Coin("def", 9)}, args{false, false}, false},
		{"bad sort", Coins{NewInt64Coin("zzz", 10), NewInt64Coin("abc", 9)}, args{false, false}, true},
		{"don't fail on zero", Coins{NewInt64Coin("abc", 0), NewInt64Coin("def", 9)}, args{false, false}, false},
		{"fail on zero", Coins{NewInt64Coin("abc", 0), NewInt64Coin("def", 9)}, args{false, true}, true},
		{"don't fail if empty", Coins{}, args{false, false}, false},
		{"fail if empty", Coins{}, args{true, false}, true},
		{"fail if empty, don't fail on zero", Coins{NewInt64Coin("abc", 0), NewInt64Coin("def", 9)}, args{true, false}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.coins.Validate(tt.args.failEmpty, tt.args.failZero); (err != nil) != tt.wantErr {
				t.Errorf("Coins.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
