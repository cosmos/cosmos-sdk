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

func TestTxValidateTx(t *testing.T) {
	assert := assert.New(t)

	addr1 := basecoin.Actor{App: "coin", Address: []byte{1, 2}}
	addr2 := basecoin.Actor{App: "coin", Address: []byte{3, 4}, ChainID: "over-there"}
	addr3 := basecoin.Actor{App: "role", Address: []byte{7, 8}}
	noAddr := basecoin.Actor{}

	noCoins := types.Coins{}
	someCoins := types.Coins{{"atom", 123}}
	moreCoins := types.Coins{{"atom", 124}}
	otherCoins := types.Coins{{"btc", 15}}
	bothCoins := someCoins.Plus(otherCoins)
	minusCoins := types.Coins{{"eth", -34}}

	// cases: all valid (one), all valid (multi)
	// no input, no outputs, invalid inputs, invalid outputs
	// totals don't match
	cases := []struct {
		valid bool
		tx    SendTx
	}{
		// 0-2. valid cases
		{true, SendTx{
			Inputs:  []TxInput{NewTxInput(addr1, someCoins, 2)},
			Outputs: []TxOutput{NewTxOutput(addr2, someCoins)},
		}},
		{true, SendTx{
			Inputs:  []TxInput{NewTxInput(addr1, someCoins, 2), NewTxInput(addr2, otherCoins, 5)},
			Outputs: []TxOutput{NewTxOutput(addr3, bothCoins)},
		}},
		{true, SendTx{
			Inputs:  []TxInput{NewTxInput(addr1, bothCoins, 42)},
			Outputs: []TxOutput{NewTxOutput(addr2, someCoins), NewTxOutput(addr3, otherCoins)},
		}},

		// 3-4. missing cases
		{false, SendTx{
			Outputs: []TxOutput{NewTxOutput(addr2, someCoins)},
		}},
		{false, SendTx{
			Inputs: []TxInput{NewTxInput(addr1, someCoins, 2)},
		}},

		// 5-8. invalid inputs
		{false, SendTx{
			Inputs:  []TxInput{NewTxInput(noAddr, someCoins, 2)},
			Outputs: []TxOutput{NewTxOutput(addr2, someCoins)},
		}},
		{false, SendTx{
			Inputs:  []TxInput{NewTxInput(addr1, someCoins, -1)},
			Outputs: []TxOutput{NewTxOutput(addr2, someCoins)},
		}},
		{false, SendTx{
			Inputs:  []TxInput{NewTxInput(addr1, noCoins, 2)},
			Outputs: []TxOutput{NewTxOutput(addr2, noCoins)},
		}},
		{false, SendTx{
			Inputs:  []TxInput{NewTxInput(addr1, minusCoins, 2)},
			Outputs: []TxOutput{NewTxOutput(addr2, minusCoins)},
		}},

		// 9-11. totals don't match
		{false, SendTx{
			Inputs:  []TxInput{NewTxInput(addr1, someCoins, 7)},
			Outputs: []TxOutput{NewTxOutput(addr2, moreCoins)},
		}},
		{false, SendTx{
			Inputs:  []TxInput{NewTxInput(addr1, someCoins, 2), NewTxInput(addr2, minusCoins, 5)},
			Outputs: []TxOutput{NewTxOutput(addr3, someCoins)},
		}},
		{false, SendTx{
			Inputs:  []TxInput{NewTxInput(addr1, someCoins, 2), NewTxInput(addr2, moreCoins, 5)},
			Outputs: []TxOutput{NewTxOutput(addr3, bothCoins)},
		}},
	}

	for i, tc := range cases {
		err := tc.tx.ValidateBasic()
		if tc.valid {
			assert.Nil(err, "%d: %+v", i, err)
		} else {
			assert.NotNil(err, "%d", i)
		}
	}

}
