package types

import (
	"strings"
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
	require.Panics(t, func() { NewUint64Coin(strings.ToUpper(testDenom1), 10) })
	require.Panics(t, func() { NewCoin(strings.ToUpper(testDenom1), NewUint(10)) })
	require.Equal(t, NewUint(5), NewUint64Coin(testDenom1, 5).Amount)
	require.Equal(t, NewUint(5), NewCoin(testDenom1, NewUint(5)).Amount)
}

func TestIsEqualCoin(t *testing.T) {
	cases := []struct {
		inputOne Coin
		inputTwo Coin
		expected bool
		panics   bool
	}{
		{NewUint64Coin(testDenom1, 1), NewUint64Coin(testDenom1, 1), true, false},
		{NewUint64Coin(testDenom1, 1), NewUint64Coin(testDenom2, 1), false, true},
		{NewUint64Coin("steak", 1), NewUint64Coin("steak", 10), false, false},
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

func TestPlusCoin(t *testing.T) {
	cases := []struct {
		inputOne    Coin
		inputTwo    Coin
		expected    Coin
		shouldPanic bool
	}{
		{NewUint64Coin(testDenom1, 1), NewUint64Coin(testDenom1, 1), NewUint64Coin(testDenom1, 2), false},
		{NewUint64Coin(testDenom1, 1), NewUint64Coin(testDenom1, 0), NewUint64Coin(testDenom1, 1), false},
		{NewUint64Coin(testDenom1, 1), NewUint64Coin(testDenom2, 1), NewUint64Coin(testDenom1, 1), true},
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
		{NewUint64Coin(testDenom1, 1), NewUint64Coin(testDenom2, 1), NewUint64Coin(testDenom1, 1), true},
		{NewUint64Coin(testDenom1, 10), NewUint64Coin(testDenom1, 1), NewUint64Coin(testDenom1, 9), false},
		{NewUint64Coin(testDenom1, 5), NewUint64Coin(testDenom1, 3), NewUint64Coin(testDenom1, 2), false},
		{NewUint64Coin(testDenom1, 5), NewUint64Coin(testDenom1, 0), NewUint64Coin(testDenom1, 5), false},
		{NewUint64Coin(testDenom1, 1), NewUint64Coin(testDenom1, 5), Coin{}, true},
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
		expected uint64
	}{NewUint64Coin(testDenom1, 1), NewUint64Coin(testDenom1, 1), 0}
	res := tc.inputOne.Minus(tc.inputTwo)
	require.Equal(t, tc.expected, res.Amount.Uint64())
}

func TestIsGTECoin(t *testing.T) {
	cases := []struct {
		inputOne Coin
		inputTwo Coin
		expected bool
		panics   bool
	}{
		{NewUint64Coin(testDenom1, 1), NewUint64Coin(testDenom1, 1), true, false},
		{NewUint64Coin(testDenom1, 2), NewUint64Coin(testDenom1, 1), true, false},
		{NewUint64Coin(testDenom1, 1), NewUint64Coin(testDenom2, 1), false, true},
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
		{NewUint64Coin(testDenom1, 1), NewUint64Coin(testDenom1, 1), false, false},
		{NewUint64Coin(testDenom1, 2), NewUint64Coin(testDenom1, 1), false, false},
		{NewUint64Coin(testDenom1, 0), NewUint64Coin(testDenom2, 1), false, true},
		{NewUint64Coin(testDenom1, 1), NewUint64Coin(testDenom2, 1), false, true},
		{NewUint64Coin(testDenom1, 1), NewUint64Coin(testDenom1, 1), false, false},
		{NewUint64Coin(testDenom1, 1), NewUint64Coin(testDenom1, 2), true, false},
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
	coin := NewUint64Coin(testDenom1, 0)
	res := coin.IsZero()
	require.True(t, res)

	coin = NewUint64Coin(testDenom1, 1)
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
		{Coins{NewUint64Coin(testDenom1, 0)}, true},
		{Coins{NewUint64Coin(testDenom1, 0), NewUint64Coin(testDenom2, 0)}, true},
		{Coins{NewUint64Coin(testDenom1, 1)}, false},
		{Coins{NewUint64Coin(testDenom1, 0), NewUint64Coin(testDenom2, 1)}, false},
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
		{Coins{NewUint64Coin(testDenom1, 0)}, Coins{NewUint64Coin(testDenom1, 0)}, true, false},
		{Coins{NewUint64Coin(testDenom1, 0), NewUint64Coin(testDenom2, 1)}, Coins{NewUint64Coin(testDenom1, 0), NewUint64Coin(testDenom2, 1)}, true, false},
		{Coins{NewUint64Coin(testDenom1, 0)}, Coins{NewUint64Coin(testDenom2, 0)}, false, true},
		{Coins{NewUint64Coin(testDenom1, 0)}, Coins{NewUint64Coin(testDenom1, 1)}, false, false},
		{Coins{NewUint64Coin(testDenom1, 0)}, Coins{NewUint64Coin(testDenom1, 0), NewUint64Coin(testDenom2, 1)}, false, false},
		{Coins{NewUint64Coin(testDenom1, 0), NewUint64Coin(testDenom2, 1)}, Coins{NewUint64Coin(testDenom1, 0), NewUint64Coin(testDenom2, 1)}, true, false},
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

func TestPlusCoins(t *testing.T) {
	zero := ZeroUint()
	one := OneUint()
	two := NewUint(2)

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
		res := tc.inputOne.Plus(tc.inputTwo)
		assert.True(t, res.IsValid())
		require.Equal(t, tc.expected, res, "sum of coins is incorrect, tc #%d", tcIndex)
	}
}

func TestMinusCoins(t *testing.T) {
	zero := ZeroUint()
	one := OneUint()
	two := NewUint(2)

	testCases := []struct {
		name        string
		inputOne    Coins
		inputTwo    Coins
		expected    Coins
		shouldPanic bool
	}{
		{"ok", NewCoins(NewCoin(testDenom1, two)), NewCoinsFromDenomAmountPairs([]string{testDenom1, testDenom2}, []Uint{one, two}), nil, true},
		{"inputTwoDenoms sanitised", NewCoins(NewCoin(testDenom1, two)), NewCoinsFromDenomAmountPairs([]string{testDenom2, testDenom1}, []Uint{zero, two}), ZeroCoins(), false},
		{"inputTwoDenoms not sanitised", NewCoins(NewCoin(testDenom1, two)), Coins{{testDenom2, zero}, {testDenom1, two}}, nil, true},
		{"different coins", Coins{{testDenom1, one}}, Coins{{testDenom2, zero}}, Coins{{testDenom1, one}}, false},
		// {Coins{{testDenom1, one}, {testDenom2, one}}, Coins{{testDenom1, one}}, Coins{{testDenom2, one}}, false},
		// {Coins{{testDenom1, one}, {testDenom2, one}}, Coins{{testDenom1, two}}, Coins{}, true},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				require.Panics(t, func() { tc.inputOne.Minus(tc.inputTwo) })
			} else {
				res := tc.inputOne.Minus(tc.inputTwo)
				assert.True(t, res.IsValid())
				require.Equal(t, tc.expected, res, "sum of coins is incorrect, tc #%d", i)
			}
		})
	}
}

func TestCoins(t *testing.T) {
	good := Coins{
		{"gas", NewUint(1)},
		{"mineral", NewUint(1)},
		{"tree", NewUint(1)},
	}
	mixedCase1 := Coins{
		{"gAs", NewUint(1)},
		{"MineraL", NewUint(1)},
		{"TREE", NewUint(1)},
	}
	mixedCase2 := Coins{
		{"gAs", NewUint(1)},
		{"mineral", NewUint(1)},
	}
	mixedCase3 := Coins{
		{"gAs", NewUint(1)},
	}
	empty := Coins{
		{"gold", NewUint(0)},
	}
	null := Coins{}
	badSort1 := Coins{
		{"tree", NewUint(1)},
		{"gas", NewUint(1)},
		{"mineral", NewUint(1)},
	}

	// both are after the first one, but the second and third are in the wrong order
	badSort2 := Coins{
		{"gas", NewUint(1)},
		{"tree", NewUint(1)},
		{"mineral", NewUint(1)},
	}
	badAmt := Coins{
		{"gas", NewUint(1)},
		{"tree", NewUint(0)},
		{"mineral", NewUint(1)},
	}
	dup := Coins{
		{"gas", NewUint(1)},
		{"gas", NewUint(1)},
		{"mineral", NewUint(1)},
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
	assert.False(t, badSort1.IsValid(), "Coins are not sorted")
	assert.False(t, badSort2.IsValid(), "Coins are not sorted")
	assert.False(t, badAmt.IsValid(), "Coins cannot include 0 amounts")
	assert.False(t, dup.IsValid(), "Duplicate coin")
	//	assert.False(t, neg.IsValid(), "Negative first-denom coin")
}

func TestCoinsIsAllGT(t *testing.T) {
	abc := MustParseCoins("1aaa,2bbb,3ccc")
	ab := MustParseCoins("1aaa,2bbb")
	oneaoneb := MustParseCoins("1aaa,1bbb")
	onebonec := MustParseCoins("1bbb,1ccc")
	oneconed := MustParseCoins("1ccc,1ddd")

	assert.False(t, ZeroCoins().IsAllGT(ZeroCoins()))
	assert.False(t, ZeroCoins().IsAllGT(abc))
	assert.True(t, abc.IsAllGT(ZeroCoins()))
	assert.False(t, abc.IsAllGT(ab))
	assert.False(t, ab.IsAllGT(abc))
	assert.False(t, abc.IsAllGT(oneaoneb))
	assert.True(t, abc.IsAllGT(onebonec))
	assert.False(t, abc.IsAllGT(oneconed))
}

func TestCoinsGTE(t *testing.T) {
	one := NewUint(1)
	two := NewUint(2)

	assert.True(t, Coins{}.IsAllGTE(Coins{}))
	assert.True(t, Coins{{testDenom1, one}}.IsAllGTE(Coins{}))
	assert.True(t, Coins{{testDenom1, one}}.IsAllGTE(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGTE(Coins{{testDenom2, one}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllGTE(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllGTE(Coins{{testDenom2, two}}))
}

func TestCoinsIsAllLTE(t *testing.T) {
	abc := MustParseCoins("1aaa,2bbb,3ccc")
	ab := MustParseCoins("1aaa,2bbb")
	onea := MustParseCoins("1aaa")
	oneb := MustParseCoins("1bbb")

	assert.True(t, ZeroCoins().IsAllLTE(ZeroCoins()))
	assert.True(t, ZeroCoins().IsAllLTE(onea))
	assert.True(t, onea.IsAllLTE(onea))
	assert.False(t, onea.IsAllLTE(oneb))
	assert.True(t, ab.IsAllLTE(abc))
	assert.False(t, abc.IsAllLTE(ab))
	assert.True(t, onea.IsAllLTE(abc))
}

func TestParse(t *testing.T) {
	one := NewUint(1)

	cases := []struct {
		input    string
		valid    bool  // if false, we expect an error on parse
		expected Coins // if valid is true, make sure this is returned
	}{
		{"", true, nil},
		{"1foo", true, Coins{{"foo", one}}},
		{"10bar", true, Coins{{"bar", NewUint(10)}}},
		{"99bar,1foo", true, Coins{{"bar", NewUint(99)}, {"foo", one}}},
		{"98 bar , 1 foo  ", true, Coins{{"bar", NewUint(98)}, {"foo", one}}},
		{"  55\t \t bling\n", true, Coins{{"bling", NewUint(55)}}},
		{"2foo, 97 bar", true, Coins{{"bar", NewUint(97)}, {"foo", NewUint(2)}}},
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
		NewUint64Coin("gas", 1),
		NewUint64Coin("mineral", 1),
		NewUint64Coin("tree", 1),
	}
	empty := Coins{
		NewUint64Coin("gold", 0),
	}
	badSort1 := Coins{
		NewUint64Coin("tree", 1),
		NewUint64Coin("gas", 1),
		NewUint64Coin("mineral", 1),
	}
	badSort2 := Coins{ // both are after the first one, but the second and third are in the wrong order
		NewUint64Coin("gas", 1),
		NewUint64Coin("tree", 1),
		NewUint64Coin("mineral", 1),
	}
	badAmt := Coins{
		NewUint64Coin("gas", 1),
		NewUint64Coin("tree", 0),
		NewUint64Coin("mineral", 1),
	}
	dup := Coins{
		NewUint64Coin("gas", 1),
		NewUint64Coin("gas", 1),
		NewUint64Coin("mineral", 1),
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
		NewUint64Coin("gold", 0),
	}
	case2 := Coins{
		NewUint64Coin("gas", 1),
		NewUint64Coin("mineral", 1),
		NewUint64Coin("tree", 1),
	}
	case3 := Coins{
		NewUint64Coin("mineral", 1),
		NewUint64Coin("tree", 1),
	}
	case4 := Coins{
		NewUint64Coin("gas", 8),
	}

	cases := []struct {
		coins           Coins
		amountOf        uint64
		amountOfSpace   uint64
		amountOfGAS     uint64
		amountOfMINERAL uint64
		amountOfTREE    uint64
	}{
		{case0, 0, 0, 0, 0, 0},
		{case1, 0, 0, 0, 0, 0},
		{case2, 0, 0, 1, 1, 1},
		{case3, 0, 0, 0, 1, 1},
		{case4, 0, 0, 8, 0, 0},
	}

	for _, tc := range cases {
		assert.Equal(t, NewUint(tc.amountOfGAS), tc.coins.AmountOf("gas"))
		assert.Equal(t, NewUint(tc.amountOfMINERAL), tc.coins.AmountOf("mineral"))
		assert.Equal(t, NewUint(tc.amountOfTREE), tc.coins.AmountOf("tree"))
	}

	assert.Panics(t, func() { cases[0].coins.AmountOf("Invalid") })
}

func TestCoinsIsAnyGTE(t *testing.T) {
	abc := MustParseCoins("1aaa,2bbb,3ccc")
	bac := MustParseCoins("2aaa,1bbb,3ccc")
	ab := MustParseCoins("1aaa,2bbb")
	onea := MustParseCoins("1aaa")
	twoa := MustParseCoins("2aaa")
	oneb := MustParseCoins("1bbb")

	assert.False(t, ZeroCoins().IsAnyGTE(ZeroCoins()))
	assert.True(t, onea.IsAnyGTE(ZeroCoins()))
	assert.False(t, ZeroCoins().IsAnyGTE(onea))
	assert.False(t, onea.IsAnyGTE(twoa))
	assert.True(t, onea.IsAnyGTE(onea))
	assert.True(t, onea.IsAnyGTE(oneb))
	assert.True(t, oneb.IsAnyGTE(onea))
	assert.True(t, abc.IsAnyGTE(abc))
	assert.True(t, abc.IsAnyGTE(bac))
	assert.True(t, ab.IsAnyGTE(abc))
	assert.False(t, oneb.IsAnyGTE(abc))
}

func TestCoinsContainsDenomsOf(t *testing.T) {
	abc := MustParseCoins("1aaa,2bbb,3ccc")
	ab := MustParseCoins("1aaa,2bbb")

	tests := []struct {
		name   string
		coins  Coins
		coinsB Coins
		want   bool
	}{
		{"empty sets", ZeroCoins(), ZeroCoins(), true},
		{"empty coinsB", abc, ZeroCoins(), true},
		{"empty coins", ZeroCoins(), abc, false},
		{"contains", abc, ab, true},
		{"does not contain", ab, abc, false},
		{"set contains itself", abc, abc, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.coins.ContainsDenomsOf(tt.coinsB))
		})
	}
}

func TestCoinsDifference(t *testing.T) {
	abc, err := ParseCoins("1aaa,2bbb,3ccc")
	require.NoError(t, err)
	ab, err := ParseCoins("1aaa,2bbb")
	require.NoError(t, err)

	tests := []struct {
		name   string
		coins  Coins
		coinsB Coins
		want   Coins
	}{
		{"empty sets", ZeroCoins(), ZeroCoins(), ZeroCoins()},
		{"empty A", ZeroCoins(), abc, ZeroCoins()},
		{"empty B", abc, ZeroCoins(), abc},
		{"A greater than B", abc, ab, NewCoins(NewUint64Coin("ccc", 3))},
		{"A smaller than B", ab, abc, ZeroCoins()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.True(t, tt.want.IsEqual(tt.coins.Difference(tt.coinsB)))
		})
	}
}
