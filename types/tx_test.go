package types

import (
	"testing"

	cmn "github.com/tendermint/go-common"

	"github.com/stretchr/testify/assert"
)

var chainID string = "test_chain"

func TestSendTxSignable(t *testing.T) {
	sendTx := &SendTx{
		Gas: 222,
		Fee: Coin{"", 111},
		Inputs: []TxInput{
			TxInput{
				Address:  []byte("input1"),
				Coins:    Coins{{"", 12345}},
				Sequence: 67890,
			},
			TxInput{
				Address:  []byte("input2"),
				Coins:    Coins{{"", 111}},
				Sequence: 222,
			},
		},
		Outputs: []TxOutput{
			TxOutput{
				Address: []byte("output1"),
				Coins:   Coins{{"", 333}},
			},
			TxOutput{
				Address: []byte("output2"),
				Coins:   Coins{{"", 444}},
			},
		},
	}
	signBytes := sendTx.SignBytes(chainID)
	signBytesHex := cmn.Fmt("%X", signBytes)
	expected := "010A746573745F636861696E0100000000000000DE00000000000000006F01020106696E7075743101010000000000000030390301093200000106696E70757432010100000000000000006F01DE0000010201076F757470757431010100000000000000014D01076F75747075743201010000000000000001BC"

	assert.True(t, signBytesHex == expected,
		cmn.Fmt("Got unexpected sign string for SendTx. Expected:\n%v\nGot:\n%v", expected, signBytesHex))
}

func TestAppTxSignable(t *testing.T) {
	callTx := &AppTx{
		Gas:  222,
		Fee:  Coin{"", 111},
		Name: "X",
		Input: TxInput{
			Address:  []byte("input1"),
			Coins:    Coins{{"", 12345}},
			Sequence: 67890,
		},
		Data: []byte("data1"),
	}
	signBytes := callTx.SignBytes(chainID)
	signBytesHex := cmn.Fmt("%X", signBytes)
	expected := "010A746573745F636861696E0100000000000000DE00000000000000006F0101580106696E70757431010100000000000000303903010932000001056461746131"

	assert.True(t, signBytesHex == expected,
		cmn.Fmt("Got unexpected sign string for AppTx. Expected:\n%v\nGot:\n%v", expected, signBytesHex))
}
