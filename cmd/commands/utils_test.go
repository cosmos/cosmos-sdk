package commands

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/basecoin/types"
)

func TestHex(t *testing.T) {

	//test isHex
	hexNoPrefix := hex.EncodeToString([]byte("foobar"))
	hexWPrefix := "0x" + hexNoPrefix
	str := "foobar"
	strWPrefix := "0xfoobar"

	//define the list of coin tests
	var testList = []struct {
		testPass bool
		errMsg   string
	}{
		{isHex(hexWPrefix), "isHex not identifying hex with 0x prefix"},
		{!isHex(hexNoPrefix), "isHex shouldn't identify hex without 0x prefix"},
		{!isHex(str), "isHex shouldn't identify non-hex string"},
		{!isHex(strWPrefix), "isHex shouldn't identify non-hex string with 0x prefix"},
		{StripHex(hexWPrefix) == hexNoPrefix, "StripHex doesn't remove first two characters"},
	}

	//execute the tests
	for _, tl := range testList {
		assert.True(t, tl.testPass, tl.errMsg)
	}

}

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

	//define the list of coin tests
	var testList = []struct {
		testPass bool
		errMsg   string
	}{
		//testing ParseCoin Function
		{types.Coin{} == makeCoin(""), "parseCoin makes bad empty coin"},
		{types.Coin{"fooCoin", 1} == makeCoin("1fooCoin"), "parseCoin makes bad coins"},
		{types.Coin{"barCoin", 10} == makeCoin("10 barCoin"), "parseCoin makes bad coins"},

		//testing ParseCoins Function
		{types.Coins{{"fooCoin", 1}}.IsEqual(makeCoins("1fooCoin")),
			"parseCoins doesn't parse a single coin"},
		{types.Coins{{"barCoin", 99}, {"fooCoin", 1}}.IsEqual(makeCoins("99barCoin,1fooCoin")),
			"parseCoins doesn't properly parse two coins"},
		{types.Coins{{"barCoin", 99}, {"fooCoin", 1}}.IsEqual(makeCoins("99 barCoin, 1 fooCoin")),
			"parseCoins doesn't properly parse two coins which use spaces"},
	}

	//execute the tests
	for _, tl := range testList {
		assert.True(t, tl.testPass, tl.errMsg)
	}
}
