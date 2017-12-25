package coin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/go-wire/data"

	sdk "github.com/cosmos/cosmos-sdk"
)

// these are some constructs for the test cases
var actors = []struct {
	actor sdk.Actor
	valid bool
}{
	{sdk.Actor{}, false},
	{sdk.Actor{App: "fooz"}, false},
	{sdk.Actor{Address: []byte{1, 2, 3, 4}}, false},
	{sdk.Actor{App: "fooz", Address: []byte{1, 2, 3, 4}}, true},
	{sdk.Actor{ChainID: "dings", App: "fooz", Address: []byte{1, 2, 3, 4}}, true},
	{sdk.Actor{ChainID: "dat", App: "fooz"}, false},
}

var (
	zeroCoin = Coin{"zeros", 0}
	plusCoin = Coin{"plus", 23}
	negCoin  = Coin{"neg", -42}
)

var coins = []struct {
	coins Coins
	valid bool
}{
	{Coins{}, false},
	{Coins{zeroCoin}, false},
	{Coins{plusCoin}, true},
	{Coins{negCoin}, false},
	{Coins{plusCoin, plusCoin}, false},
	{Coins{plusCoin, zeroCoin}, false},
	{Coins{negCoin, plusCoin}, false},
}

func TestTxValidateInput(t *testing.T) {
	assert := assert.New(t)

	for i, act := range actors {
		for j, coin := range coins {
			input := NewTxInput(act.actor, coin.coins)
			err := input.ValidateBasic()
			if act.valid && coin.valid {
				assert.Nil(err, "%d,%d: %+v", i, j, err)
			} else {
				assert.NotNil(err, "%d,%d", i, j)
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

	addr1 := sdk.Actor{App: "coin", Address: []byte{1, 2}}
	addr2 := sdk.Actor{App: "coin", Address: []byte{3, 4}, ChainID: "over-there"}
	addr3 := sdk.Actor{App: "role", Address: []byte{7, 8}}
	noAddr := sdk.Actor{}

	noCoins := Coins{}
	someCoins := Coins{{"atom", 123}}
	moreCoins := Coins{{"atom", 124}}
	otherCoins := Coins{{"btc", 15}}
	bothCoins := someCoins.Plus(otherCoins)
	minusCoins := Coins{{"eth", -34}}

	// cases: all valid (one), all valid (multi)
	// no input, no outputs, invalid inputs, invalid outputs
	// totals don't match
	cases := []struct {
		valid bool
		tx    sdk.Tx
	}{
		// 0-2. valid cases
		{true, NewSendTx(
			[]TxInput{NewTxInput(addr1, someCoins)},
			[]TxOutput{NewTxOutput(addr2, someCoins)},
		)},
		{true, NewSendTx(
			[]TxInput{NewTxInput(addr1, someCoins), NewTxInput(addr2, otherCoins)},
			[]TxOutput{NewTxOutput(addr3, bothCoins)},
		)},
		{true, NewSendTx(
			[]TxInput{NewTxInput(addr1, bothCoins)},
			[]TxOutput{NewTxOutput(addr2, someCoins), NewTxOutput(addr3, otherCoins)},
		)},

		// 3-4. missing cases
		{false, NewSendTx(
			nil,
			[]TxOutput{NewTxOutput(addr2, someCoins)},
		)},
		{false, NewSendTx(
			[]TxInput{NewTxInput(addr1, someCoins)},
			nil,
		)},

		// 5-7. invalid inputs
		{false, NewSendTx(
			[]TxInput{NewTxInput(noAddr, someCoins)},
			[]TxOutput{NewTxOutput(addr2, someCoins)},
		)},
		{false, NewSendTx(
			[]TxInput{NewTxInput(addr1, noCoins)},
			[]TxOutput{NewTxOutput(addr2, noCoins)},
		)},
		{false, NewSendTx(
			[]TxInput{NewTxInput(addr1, minusCoins)},
			[]TxOutput{NewTxOutput(addr2, minusCoins)},
		)},

		// 8-10. totals don't match
		{false, NewSendTx(
			[]TxInput{NewTxInput(addr1, someCoins)},
			[]TxOutput{NewTxOutput(addr2, moreCoins)},
		)},
		{false, NewSendTx(
			[]TxInput{NewTxInput(addr1, someCoins), NewTxInput(addr2, minusCoins)},
			[]TxOutput{NewTxOutput(addr3, someCoins)},
		)},
		{false, NewSendTx(
			[]TxInput{NewTxInput(addr1, someCoins), NewTxInput(addr2, moreCoins)},
			[]TxOutput{NewTxOutput(addr3, bothCoins)},
		)},
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

func TestTxSerializeTx(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	addr1 := sdk.Actor{App: "coin", Address: []byte{1, 2}}
	addr2 := sdk.Actor{App: "coin", Address: []byte{3, 4}}
	someCoins := Coins{{"atom", 123}}

	send := NewSendTx(
		[]TxInput{NewTxInput(addr1, someCoins)},
		[]TxOutput{NewTxOutput(addr2, someCoins)},
	)

	js, err := data.ToJSON(send)
	require.Nil(err)
	var tx sdk.Tx
	err = data.FromJSON(js, &tx)
	require.Nil(err)
	assert.Equal(send, tx)

	bin, err := data.ToWire(send)
	require.Nil(err)
	var tx2 sdk.Tx
	err = data.FromWire(bin, &tx2)
	require.Nil(err)
	assert.Equal(send, tx2)

}
