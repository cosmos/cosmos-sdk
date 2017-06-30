package coin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/types"
)

// these are some constructs for the test cases
var actors = []struct {
	actor basecoin.Actor
	valid bool
}{
	{basecoin.Actor{}, false},
	{basecoin.Actor{App: "fooz"}, false},
	{basecoin.Actor{Address: []byte{1, 2, 3, 4}}, false},
	{basecoin.Actor{App: "fooz", Address: []byte{1, 2, 3, 4}}, true},
	{basecoin.Actor{ChainID: "dings", App: "fooz", Address: []byte{1, 2, 3, 4}}, true},
	{basecoin.Actor{ChainID: "dat", App: "fooz"}, false},
}

var (
	zeroCoin = types.Coin{"zeros", 0}
	plusCoin = types.Coin{"plus", 23}
	negCoin  = types.Coin{"neg", -42}
)

var coins = []struct {
	coins types.Coins
	valid bool
}{
	{types.Coins{}, false},
	{types.Coins{zeroCoin}, false},
	{types.Coins{plusCoin}, true},
	{types.Coins{negCoin}, false},
	{types.Coins{plusCoin, plusCoin}, false},
	{types.Coins{plusCoin, zeroCoin}, false},
	{types.Coins{negCoin, plusCoin}, false},
}

func TestTxValidateInput(t *testing.T) {
	assert := assert.New(t)

	seqs := []struct {
		seq   int
		valid bool
	}{
		{-3, false},
		{0, false},
		{1, true},
		{6571265735, true},
	}

	for i, act := range actors {
		for j, coin := range coins {
			for k, seq := range seqs {
				input := NewTxInput(act.actor, coin.coins, seq.seq)
				err := input.ValidateBasic()
				if act.valid && coin.valid && seq.valid {
					assert.Nil(err, "%d,%d,%d: %+v", i, j, k, err)
				} else {
					assert.NotNil(err, "%d,%d,%d", i, j, k)
				}
			}
		}
	}
}

func TestTxValidateOutput(t *testing.T) {
	assert := assert.New(t)

	for i, act := range actors {
		for j, coin := range coins {
			input := NewTxOutput(act.actor, coin.coins)
			err := input.ValidateBasic()
			if act.valid && coin.valid {
				assert.Nil(err, "%d,%d: %+v", i, j, err)
			} else {
				assert.NotNil(err, "%d,%d", i, j)
			}
		}
	}
}
