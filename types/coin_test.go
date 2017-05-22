package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoins(t *testing.T) {
	assert := assert.New(t)

	//Define the coins to be used in tests
	good := Coins{
		Coin{"GAS", 1},
		Coin{"MINERAL", 1},
		Coin{"TREE", 1},
	}
	neg := good.Negative()
	sum := good.Plus(neg)
	empty := Coins{
		Coin{"GOLD", 0},
	}
	badSort1 := Coins{
		Coin{"TREE", 1},
		Coin{"GAS", 1},
		Coin{"MINERAL", 1},
	}
	badSort2 := Coins{ // both are after the first one, but the second and third are in the wrong order
		Coin{"GAS", 1},
		Coin{"TREE", 1},
		Coin{"MINERAL", 1},
	}
	badAmt := Coins{
		Coin{"GAS", 1},
		Coin{"TREE", 0},
		Coin{"MINERAL", 1},
	}
	dup := Coins{
		Coin{"GAS", 1},
		Coin{"GAS", 1},
		Coin{"MINERAL", 1},
	}

	assert.True(good.IsValid(), "Coins are valid")
	assert.True(good.IsPositive(), "Expected coins to be positive: %v", good)
	assert.True(good.IsGTE(empty), "Expected %v to be >= %v", good, empty)
	assert.False(neg.IsPositive(), "Expected neg coins to not be positive: %v", neg)
	assert.Zero(len(sum), "Expected 0 coins")
	assert.False(badSort1.IsValid(), "Coins are not sorted")
	assert.False(badSort2.IsValid(), "Coins are not sorted")
	assert.False(badAmt.IsValid(), "Coins cannot include 0 amounts")
	assert.False(dup.IsValid(), "Duplicate coin")

}

//Test the parse coin and parse coins functionality
func TestParse(t *testing.T) {
	assert := assert.New(t)

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
		{"2foo, 97 bar", true, Coins{{"bar", 97}, {"foo", 2}}},
	}

	for _, tc := range cases {
		res, err := ParseCoins(tc.input)
		if !tc.valid {
			assert.NotNil(err, tc.input)
		} else if assert.Nil(err, "%s: %+v", tc.input, err) {
			assert.Equal(tc.expected, res)
		}
	}

}

func TestSortCoins(t *testing.T) {
	assert := assert.New(t)

	good := Coins{
		Coin{"GAS", 1},
		Coin{"MINERAL", 1},
		Coin{"TREE", 1},
	}
	empty := Coins{
		Coin{"GOLD", 0},
	}
	badSort1 := Coins{
		Coin{"TREE", 1},
		Coin{"GAS", 1},
		Coin{"MINERAL", 1},
	}
	badSort2 := Coins{ // both are after the first one, but the second and third are in the wrong order
		Coin{"GAS", 1},
		Coin{"TREE", 1},
		Coin{"MINERAL", 1},
	}
	badAmt := Coins{
		Coin{"GAS", 1},
		Coin{"TREE", 0},
		Coin{"MINERAL", 1},
	}
	dup := Coins{
		Coin{"GAS", 1},
		Coin{"GAS", 1},
		Coin{"MINERAL", 1},
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
		assert.Equal(tc.before, tc.coins.IsValid())
		tc.coins.Sort()
		assert.Equal(tc.after, tc.coins.IsValid())
	}
}
