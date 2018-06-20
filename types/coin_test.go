package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPositiveCoin(t *testing.T) {
	assert := assert.New(t)

	cases := []struct {
		inputOne Coin
		expected bool
	}{
		{NewCoin("A", 1), true},
		{NewCoin("A", 0), false},
		{NewCoin("a", -1), false},
	}

	for _, tc := range cases {
		res := tc.inputOne.IsPositive()
		assert.Equal(tc.expected, res)
	}
}

func TestIsNotNegativeCoin(t *testing.T) {
	assert := assert.New(t)

	cases := []struct {
		inputOne Coin
		expected bool
	}{
		{NewCoin("A", 1), true},
		{NewCoin("A", 0), true},
		{NewCoin("a", -1), false},
	}

	for _, tc := range cases {
		res := tc.inputOne.IsNotNegative()
		assert.Equal(tc.expected, res)
	}
}

func TestSameDenomAsCoin(t *testing.T) {
	assert := assert.New(t)

	cases := []struct {
		inputOne Coin
		inputTwo Coin
		expected bool
	}{
		{NewCoin("A", 1), NewCoin("A", 1), true},
		{NewCoin("A", 1), NewCoin("a", 1), false},
		{NewCoin("a", 1), NewCoin("b", 1), false},
		{NewCoin("steak", 1), NewCoin("steak", 10), true},
		{NewCoin("steak", -11), NewCoin("steak", 10), true},
	}

	for _, tc := range cases {
		res := tc.inputOne.SameDenomAs(tc.inputTwo)
		assert.Equal(tc.expected, res)
	}
}

func TestIsGTECoin(t *testing.T) {
	assert := assert.New(t)

	cases := []struct {
		inputOne Coin
		inputTwo Coin
		expected bool
	}{
		{NewCoin("A", 1), NewCoin("A", 1), true},
		{NewCoin("A", 2), NewCoin("A", 1), true},
		{NewCoin("A", -1), NewCoin("A", 5), false},
		{NewCoin("a", 1), NewCoin("b", 1), false},
	}

	for _, tc := range cases {
		res := tc.inputOne.IsGTE(tc.inputTwo)
		assert.Equal(tc.expected, res)
	}
}

func TestIsEqualCoin(t *testing.T) {
	assert := assert.New(t)

	cases := []struct {
		inputOne Coin
		inputTwo Coin
		expected bool
	}{
		{NewCoin("A", 1), NewCoin("A", 1), true},
		{NewCoin("A", 1), NewCoin("a", 1), false},
		{NewCoin("a", 1), NewCoin("b", 1), false},
		{NewCoin("steak", 1), NewCoin("steak", 10), false},
		{NewCoin("steak", -11), NewCoin("steak", 10), false},
	}

	for _, tc := range cases {
		res := tc.inputOne.IsEqual(tc.inputTwo)
		assert.Equal(tc.expected, res)
	}
}

func TestPlusCoin(t *testing.T) {
	assert := assert.New(t)

	cases := []struct {
		inputOne Coin
		inputTwo Coin
		expected Coin
	}{
		{NewCoin("A", 1), NewCoin("A", 1), NewCoin("A", 2)},
		{NewCoin("A", 1), NewCoin("B", 1), NewCoin("A", 1)},
		{NewCoin("asdf", -4), NewCoin("asdf", 5), NewCoin("asdf", 1)},
	}

	for _, tc := range cases {
		res := tc.inputOne.Plus(tc.inputTwo)
		assert.Equal(tc.expected, res)
	}

	tc := struct {
		inputOne Coin
		inputTwo Coin
		expected int64
	}{NewCoin("asdf", -1), NewCoin("asdf", 1), 0}
	res := tc.inputOne.Plus(tc.inputTwo)
	assert.Equal(tc.expected, res.Amount.Int64())
}

func TestMinusCoin(t *testing.T) {
	assert := assert.New(t)

	cases := []struct {
		inputOne Coin
		inputTwo Coin
		expected Coin
	}{

		{NewCoin("A", 1), NewCoin("B", 1), NewCoin("A", 1)},
		{NewCoin("asdf", -4), NewCoin("asdf", 5), NewCoin("asdf", -9)},
		{NewCoin("asdf", 10), NewCoin("asdf", 1), NewCoin("asdf", 9)},
	}

	for _, tc := range cases {
		res := tc.inputOne.Minus(tc.inputTwo)
		assert.Equal(tc.expected, res)
	}

	tc := struct {
		inputOne Coin
		inputTwo Coin
		expected int64
	}{NewCoin("A", 1), NewCoin("A", 1), 0}
	res := tc.inputOne.Minus(tc.inputTwo)
	assert.Equal(tc.expected, res.Amount.Int64())

}

func TestCoins(t *testing.T) {

	//Define the coins to be used in tests
	good := Coins{
		{"GAS", NewInt(1)},
		{"MINERAL", NewInt(1)},
		{"TREE", NewInt(1)},
	}
	neg := good.Negative()
	sum := good.Plus(neg)
	empty := Coins{
		{"GOLD", NewInt(0)},
	}
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
	assert.True(t, good.IsGTE(empty), "Expected %v to be >= %v", good, empty)
	assert.False(t, neg.IsPositive(), "Expected neg coins to not be positive: %v", neg)
	assert.Zero(t, len(sum), "Expected 0 coins")
	assert.False(t, badSort1.IsValid(), "Coins are not sorted")
	assert.False(t, badSort2.IsValid(), "Coins are not sorted")
	assert.False(t, badAmt.IsValid(), "Coins cannot include 0 amounts")
	assert.False(t, dup.IsValid(), "Duplicate coin")

}

func TestPlusCoins(t *testing.T) {
	assert := assert.New(t)

	one := NewInt(1)
	zero := NewInt(0)
	negone := NewInt(-1)
	two := NewInt(2)

	cases := []struct {
		inputOne Coins
		inputTwo Coins
		expected Coins
	}{
		{Coins{{"A", one}, {"B", one}}, Coins{{"A", one}, {"B", one}}, Coins{{"A", two}, {"B", two}}},
		{Coins{{"A", zero}, {"B", one}}, Coins{{"A", zero}, {"B", zero}}, Coins{{"B", one}}},
		{Coins{{"A", zero}, {"B", zero}}, Coins{{"A", zero}, {"B", zero}}, Coins(nil)},
		{Coins{{"A", one}, {"B", zero}}, Coins{{"A", negone}, {"B", zero}}, Coins(nil)},
		{Coins{{"A", negone}, {"B", zero}}, Coins{{"A", zero}, {"B", zero}}, Coins{{"A", negone}}},
	}

	for _, tc := range cases {
		res := tc.inputOne.Plus(tc.inputTwo)
		assert.True(res.IsValid())
		assert.Equal(tc.expected, res)
	}
}

//Test the parsing of Coin and Coins
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

	for _, tc := range cases {
		res, err := ParseCoins(tc.input)
		if !tc.valid {
			assert.NotNil(t, err, "%s: %#v", tc.input, res)
		} else if assert.Nil(t, err, "%s: %+v", tc.input, err) {
			assert.Equal(t, tc.expected, res)
		}
	}

}

func TestSortCoins(t *testing.T) {

	good := Coins{
		NewCoin("GAS", 1),
		NewCoin("MINERAL", 1),
		NewCoin("TREE", 1),
	}
	empty := Coins{
		NewCoin("GOLD", 0),
	}
	badSort1 := Coins{
		NewCoin("TREE", 1),
		NewCoin("GAS", 1),
		NewCoin("MINERAL", 1),
	}
	badSort2 := Coins{ // both are after the first one, but the second and third are in the wrong order
		NewCoin("GAS", 1),
		NewCoin("TREE", 1),
		NewCoin("MINERAL", 1),
	}
	badAmt := Coins{
		NewCoin("GAS", 1),
		NewCoin("TREE", 0),
		NewCoin("MINERAL", 1),
	}
	dup := Coins{
		NewCoin("GAS", 1),
		NewCoin("GAS", 1),
		NewCoin("MINERAL", 1),
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

	for _, tc := range cases {
		assert.Equal(t, tc.before, tc.coins.IsValid())
		tc.coins.Sort()
		assert.Equal(t, tc.after, tc.coins.IsValid())
	}
}

func TestAmountOf(t *testing.T) {

	case0 := Coins{}
	case1 := Coins{
		NewCoin("", 0),
	}
	case2 := Coins{
		NewCoin(" ", 0),
	}
	case3 := Coins{
		NewCoin("GOLD", 0),
	}
	case4 := Coins{
		NewCoin("GAS", 1),
		NewCoin("MINERAL", 1),
		NewCoin("TREE", 1),
	}
	case5 := Coins{
		NewCoin("MINERAL", 1),
		NewCoin("TREE", 1),
	}
	case6 := Coins{
		NewCoin("", 6),
	}
	case7 := Coins{
		NewCoin(" ", 7),
	}
	case8 := Coins{
		NewCoin("GAS", 8),
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
