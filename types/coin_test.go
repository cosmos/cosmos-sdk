package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoins(t *testing.T) {

	//Define the coins to be used in tests
	good := Coins{
		{"GAS", 1},
		{"MINERAL", 1},
		{"TREE", 1},
	}
	neg := good.Negative()
	sum := good.Plus(neg)
	empty := Coins{
		{"GOLD", 0},
	}
	badSort1 := Coins{
		{"TREE", 1},
		{"GAS", 1},
		{"MINERAL", 1},
	}
	badSort2 := Coins{ // both are after the first one, but the second and third are in the wrong order
		{"GAS", 1},
		{"TREE", 1},
		{"MINERAL", 1},
	}
	badAmt := Coins{
		{"GAS", 1},
		{"TREE", 0},
		{"MINERAL", 1},
	}
	dup := Coins{
		{"GAS", 1},
		{"GAS", 1},
		{"MINERAL", 1},
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

//Test the parse coin and parse coins functionality
func TestParse(t *testing.T) {

	cases := []struct {
		input    string
		valid    bool  // if false, we expect an error on parse
		expected Coins // if valid is true, make sure this is returned
	}{
		{"", true, nil},
		{"1foo", true, Coins{{"foo", 1}}},
		{"10bar", true, Coins{{"bar", 10}}},
		{"99bar,1foo", true, Coins{{"bar", 99}, {"foo", 1}}},
		{"98 bar , 1 foo  ", true, Coins{{"bar", 98}, {"foo", 1}}},
		{"  55\t \t bling\n", true, Coins{{"bling", 55}}},
		{"2foo, 97 bar", true, Coins{{"bar", 97}, {"foo", 2}}},
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
		{"GAS", 1},
		{"MINERAL", 1},
		{"TREE", 1},
	}
	empty := Coins{
		{"GOLD", 0},
	}
	badSort1 := Coins{
		{"TREE", 1},
		{"GAS", 1},
		{"MINERAL", 1},
	}
	badSort2 := Coins{ // both are after the first one, but the second and third are in the wrong order
		{"GAS", 1},
		{"TREE", 1},
		{"MINERAL", 1},
	}
	badAmt := Coins{
		{"GAS", 1},
		{"TREE", 0},
		{"MINERAL", 1},
	}
	dup := Coins{
		{"GAS", 1},
		{"GAS", 1},
		{"MINERAL", 1},
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
		{"", 0},
	}
	case2 := Coins{
		{" ", 0},
	}
	case3 := Coins{
		{"GOLD", 0},
	}
	case4 := Coins{
		{"GAS", 1},
		{"MINERAL", 1},
		{"TREE", 1},
	}
	case5 := Coins{
		{"MINERAL", 1},
		{"TREE", 1},
	}
	case6 := Coins{
		{"", 6},
	}
	case7 := Coins{
		{" ", 7},
	}
	case8 := Coins{
		{"GAS", 8},
	}

	cases := []struct {
		coins           Coins
		amountOf        int64
		amountOf_       int64
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
		assert.Equal(t, tc.amountOf, tc.coins.AmountOf(""))
		assert.Equal(t, tc.amountOf_, tc.coins.AmountOf(" "))
		assert.Equal(t, tc.amountOfGAS, tc.coins.AmountOf("GAS"))
		assert.Equal(t, tc.amountOfMINERAL, tc.coins.AmountOf("MINERAL"))
		assert.Equal(t, tc.amountOfTREE, tc.coins.AmountOf("TREE"))
	}
}
