package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/basecoin/types"
)

//Test the parse coin and parse coins functionality
func TestParse(t *testing.T) {

	makeCoin := func(str string) types.Coin {
		coin, err := ParseCoin(str)
		if err != nil {
			panic(err.Error())
		}
		return coin
	}

	makeCoins := func(str string) types.Coins {
		coin, err := ParseCoins(str)
		if err != nil {
			panic(err.Error())
		}
		return coin
	}

	//testing ParseCoin Function
	assert.True(t, types.Coin{} == makeCoin(""), "parseCoin makes bad empty coin")
	assert.True(t, types.Coin{"fooCoin", 1} == makeCoin("1fooCoin"), "parseCoin makes bad coins")
	assert.True(t, types.Coin{"barCoin", 10} == makeCoin("10 barCoin"), "parseCoin makes bad coins")

	//testing ParseCoins Function
	assert.True(t, types.Coins{{"fooCoin", 1}}.IsEqual(makeCoins("1fooCoin")), "parseCoins doesn't parse a single coin")
	assert.True(t, types.Coins{{"barCoin", 99}, {"fooCoin", 1}}.IsEqual(makeCoins("99barCoin,1fooCoin")),
		"parseCoins doesn't properly parse two coins")
	assert.True(t, types.Coins{{"barCoin", 99}, {"fooCoin", 1}}.IsEqual(makeCoins("99 barCoin, 1 fooCoin")),
		"parseCoins doesn't properly parse two coins which use spaces")
}
