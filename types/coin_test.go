package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.True(good.IsPositive(), fmt.Sprintf("Expected coins to be positive: %v", good))
	assert.True(good.IsGTE(empty), fmt.Sprintf("Expected %v to be >= %v", good, empty))
	assert.False(neg.IsPositive(), fmt.Sprintf("Expected neg coins to not be positive: %v", neg))
	assert.Zero(len(sum), "Expected 0 coins")
	assert.False(badSort1.IsValid(), "Coins are not sorted")
	assert.False(badSort2.IsValid(), "Coins are not sorted")
	assert.False(badAmt.IsValid(), "Coins cannot include 0 amounts")
	assert.False(dup.IsValid(), "Duplicate coin")

}

//Test the parse coin and parse coins functionality
func TestParse(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	makeCoin := func(str string) Coin {
		coin, err := ParseCoin(str)
		require.Nil(err)
		return coin
	}

	makeCoins := func(str string) Coins {
		coin, err := ParseCoins(str)
		require.Nil(err)
		return coin
	}

	//testing ParseCoin Function
	assert.Equal(Coin{}, makeCoin(""), "ParseCoin makes bad empty coin")
	assert.Equal(Coin{"fooCoin", 1}, makeCoin("1fooCoin"), "ParseCoin makes bad coins")
	assert.Equal(Coin{"barCoin", 10}, makeCoin("10 barCoin"), "ParseCoin makes bad coins")

	//testing ParseCoins Function
	assert.True(Coins{{"fooCoin", 1}}.IsEqual(makeCoins("1fooCoin")),
		"ParseCoins doesn't parse a single coin")
	assert.True(Coins{{"barCoin", 99}, {"fooCoin", 1}}.IsEqual(makeCoins("99barCoin,1fooCoin")),
		"ParseCoins doesn't properly parse two coins")
	assert.True(Coins{{"barCoin", 99}, {"fooCoin", 1}}.IsEqual(makeCoins("99 barCoin, 1 fooCoin")),
		"ParseCoins doesn't properly parse two coins which use spaces")
}
