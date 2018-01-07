package coinstore

import (
	"testing"

	"github.com/stretchr/testify/assert"

	crypto "github.com/tendermint/go-crypto"

	"github.com/cosmos/cosmos-sdk/x/coin"
)

func TestTxInputValidation(t *testing.T) {
	addr1 := crypto.Address([]byte{1, 2})
	addr2 := crypto.Address([]byte{7, 8})
	someCoins := coin.Coins{{"atom", 123}}
	multiCoins := coin.Coins{{"atom", 123}, {"eth", 20}}

	var emptyAddr crypto.Address
	emptyCoins := coin.Coins{}
	emptyCoins2 := coin.Coins{{"eth", 0}}
	someEmptyCoins := coin.Coins{{"eth", 10}, {"atom", 0}}
	minusCoins := coin.Coins{{"eth", -34}}
	someMinusCoins := coin.Coins{{"atom", 20}, {"eth", -34}}
	unsortedCoins := coin.Coins{{"eth", 1}, {"atom", 1}}

	cases := []struct {
		valid bool
		txIn  TxInput
	}{
		// auth works with different apps
		{true, NewTxInput(addr1, someCoins)},
		{true, NewTxInputWithSequence(addr1, someCoins, 100)},
		{true, NewTxInputWithSequence(addr2, someCoins, 100)},
		{true, NewTxInputWithSequence(addr2, multiCoins, 100)},

		{false, NewTxInput(emptyAddr, someCoins)},             // empty address
		{false, NewTxInputWithSequence(addr1, someCoins, -1)}, // negative sequence
		{false, NewTxInput(addr1, emptyCoins)},                // invalid coins
		{false, NewTxInput(addr1, emptyCoins2)},               // invalid coins
		{false, NewTxInput(addr1, someEmptyCoins)},            // invalid coins
		{false, NewTxInput(addr1, minusCoins)},                // negative coins
		{false, NewTxInput(addr1, someMinusCoins)},            // negative coins
		{false, NewTxInput(addr1, unsortedCoins)},             // unsorted coins
	}

	for i, tc := range cases {
		err := tc.txIn.ValidateBasic()
		if tc.valid {
			assert.Nil(t, err, "%d: %+v", i, err)
		} else {
			assert.NotNil(t, err, "%d", i)
		}
	}
}

func TestTxOutputValidation(t *testing.T) {
	addr1 := crypto.Address([]byte{1, 2})
	addr2 := crypto.Address([]byte{7, 8})
	someCoins := coin.Coins{{"atom", 123}}
	multiCoins := coin.Coins{{"atom", 123}, {"eth", 20}}

	var emptyAddr crypto.Address
	emptyCoins := coin.Coins{}
	emptyCoins2 := coin.Coins{{"eth", 0}}
	someEmptyCoins := coin.Coins{{"eth", 10}, {"atom", 0}}
	minusCoins := coin.Coins{{"eth", -34}}
	someMinusCoins := coin.Coins{{"atom", 20}, {"eth", -34}}
	unsortedCoins := coin.Coins{{"eth", 1}, {"atom", 1}}

	cases := []struct {
		valid bool
		txOut TxOutput
	}{
		// auth works with different apps
		{true, NewTxOutput(addr1, someCoins)},
		{true, NewTxOutput(addr2, someCoins)},
		{true, NewTxOutput(addr2, multiCoins)},

		{false, NewTxOutput(emptyAddr, someCoins)},  // empty address
		{false, NewTxOutput(addr1, emptyCoins)},     // invalid coins
		{false, NewTxOutput(addr1, emptyCoins2)},    // invalid coins
		{false, NewTxOutput(addr1, someEmptyCoins)}, // invalid coins
		{false, NewTxOutput(addr1, minusCoins)},     // negative coins
		{false, NewTxOutput(addr1, someMinusCoins)}, // negative coins
		{false, NewTxOutput(addr1, unsortedCoins)},  // unsorted coins
	}

	for i, tc := range cases {
		err := tc.txOut.ValidateBasic()
		if tc.valid {
			assert.Nil(t, err, "%d: %+v", i, err)
		} else {
			assert.NotNil(t, err, "%d", i)
		}
	}
}

func TestSendTxValidation(t *testing.T) {

	addr1 := crypto.Address([]byte{1, 2})
	addr2 := crypto.Address([]byte{7, 8})
	atom123 := coin.Coins{{"atom", 123}}
	atom124 := coin.Coins{{"atom", 124}}
	eth123 := coin.Coins{{"eth", 123}}
	atom123eth123 := coin.Coins{{"atom", 123}, {"eth", 123}}

	input1 := NewTxInput(addr1, atom123)
	input2 := NewTxInput(addr1, eth123)
	output1 := NewTxOutput(addr2, atom123)
	output2 := NewTxOutput(addr2, atom124)
	output3 := NewTxOutput(addr2, eth123)
	outputMulti := NewTxOutput(addr2, atom123eth123)

	var emptyAddr crypto.Address

	cases := []struct {
		valid bool
		tx    SendTx
	}{
		{false, SendTx{}},                             // no input or output
		{false, SendTx{Inputs: []TxInput{input1}}},    // just input
		{false, SendTx{Outputs: []TxOutput{output1}}}, // just ouput
		{false, SendTx{
			Inputs:  []TxInput{NewTxInputWithSequence(emptyAddr, atom123, 1)}, // invalid input
			Outputs: []TxOutput{output1}}},
		{false, SendTx{
			Inputs:  []TxInput{input1},
			Outputs: []TxOutput{{emptyAddr, atom123}}}, // invalid ouput
		},
		{false, SendTx{
			Inputs:  []TxInput{input1},
			Outputs: []TxOutput{output2}}, // amounts dont match
		},
		{false, SendTx{
			Inputs:  []TxInput{input1},
			Outputs: []TxOutput{output3}}, // amounts dont match
		},
		{false, SendTx{
			Inputs:  []TxInput{input1},
			Outputs: []TxOutput{outputMulti}}, // amounts dont match
		},
		{false, SendTx{
			Inputs:  []TxInput{input2},
			Outputs: []TxOutput{output1}}, // amounts dont match
		},

		{true, SendTx{
			Inputs:  []TxInput{input1},
			Outputs: []TxOutput{output1}},
		},
		{true, SendTx{
			Inputs:  []TxInput{input1, input2},
			Outputs: []TxOutput{outputMulti}},
		},
	}

	for i, tc := range cases {
		err := tc.tx.ValidateBasic()
		if tc.valid {
			assert.Nil(t, err, "%d: %+v", i, err)
		} else {
			assert.NotNil(t, err, "%d", i)
		}
	}
}

func TestSendTxSigners(t *testing.T) {
	signers := []crypto.Address{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}

	someCoins := coin.Coins{{"atom", 123}}
	inputs := make([]TxInput, len(signers))
	for i, signer := range signers {
		inputs[i] = NewTxInput(signer, someCoins)
	}
	tx := NewSendTx(inputs, nil)

	assert.Equal(t, signers, tx.Signers())
}
