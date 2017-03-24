package commands

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/basecoin/types"
)

func TestHex(t *testing.T) {
	assert := assert.New(t)

	//test isHex
	hexNoPrefix := hex.EncodeToString([]byte("foobar"))
	hexWPrefix := "0x" + hexNoPrefix
	str := "foobar"
	strWPrefix := "0xfoobar"

	assert.True(isHex(hexWPrefix), "isHex not identifying hex with 0x prefix")
	assert.False(isHex(hexNoPrefix), "isHex shouldn't identify hex without 0x prefix")
	assert.False(isHex(str), "isHex shouldn't identify non-hex string")
	assert.False(isHex(strWPrefix), "isHex shouldn't identify non-hex string with 0x prefix")
	assert.True(StripHex(hexWPrefix) == hexNoPrefix, "StripHex doesn't remove first two characters")

}

//Test the parse coin and parse coins functionality
func TestParse(t *testing.T) {
	assert := assert.New(t)

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
	assert.Equal(types.Coin{}, makeCoin(""), "parseCoin makes bad empty coin")
	assert.Equal(types.Coin{"fooCoin", 1}, makeCoin("1fooCoin"), "parseCoin makes bad coins")
	assert.Equal(types.Coin{"barCoin", 10}, makeCoin("10 barCoin"), "parseCoin makes bad coins")

	//testing ParseCoins Function
	assert.True(types.Coins{{"fooCoin", 1}}.IsEqual(makeCoins("1fooCoin")),
		"parseCoins doesn't parse a single coin")
	assert.True(types.Coins{{"barCoin", 99}, {"fooCoin", 1}}.IsEqual(makeCoins("99barCoin,1fooCoin")),
		"parseCoins doesn't properly parse two coins")
	assert.True(types.Coins{{"barCoin", 99}, {"fooCoin", 1}}.IsEqual(makeCoins("99 barCoin, 1 fooCoin")),
		"parseCoins doesn't properly parse two coins which use spaces")
}
