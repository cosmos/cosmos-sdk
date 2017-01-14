package types

import (
	"testing"

	. "github.com/tendermint/go-common"
)

var chainID string = "test_chain"

func TestSendTxSignable(t *testing.T) {
	sendTx := &SendTx{
		Fee: 111,
		Gas: 222,
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
	signBytesHex := Fmt("%X", signBytes)
	expected := "010A746573745F636861696E01000000000000006F00000000000000DE01020106696E7075743101010000000000000030390301093200000106696E70757432010100000000000000006F01DE0000010201076F757470757431010100000000000000014D01076F75747075743201010000000000000001BC"
	if signBytesHex != expected {
		t.Errorf("Got unexpected sign string for SendTx. Expected:\n%v\nGot:\n%v", expected, signBytesHex)
	}
}

func TestAppTxSignable(t *testing.T) {
	callTx := &AppTx{
		Fee:  111,
		Gas:  222,
		Name: "X",
		Input: TxInput{
			Address:  []byte("input1"),
			Coins:    Coins{{"", 12345}},
			Sequence: 67890,
		},
		Data: []byte("data1"),
	}
	signBytes := callTx.SignBytes(chainID)
	signBytesHex := Fmt("%X", signBytes)
	expected := "010A746573745F636861696E01000000000000006F00000000000000DE0101580106696E70757431010100000000000000303903010932000001056461746131"
	if signBytesHex != expected {
		t.Errorf("Got unexpected sign string for AppTx. Expected:\n%v\nGot:\n%v", expected, signBytesHex)
	}
}
