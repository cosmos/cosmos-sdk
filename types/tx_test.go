package types

import (
	"testing"

	. "github.com/tendermint/go-common"
)

var chainID string = "test_chain"

func TestSendTxSignable(t *testing.T) {
	sendTx := &SendTx{
		Inputs: []TxInput{
			TxInput{
				Address:  []byte("input1"),
				Amount:   12345,
				Sequence: 67890,
			},
			TxInput{
				Address:  []byte("input2"),
				Amount:   111,
				Sequence: 222,
			},
		},
		Outputs: []TxOutput{
			TxOutput{
				Address: []byte("output1"),
				Amount:  333,
			},
			TxOutput{
				Address: []byte("output2"),
				Amount:  444,
			},
		},
	}
	signBytes := sendTx.SignBytes(chainID)
	signStr := string(signBytes)
	expected := Fmt(`{"chain_id":"%s","tx":[1,{"inputs":[{"address":"696E70757431","amount":12345,"sequence":67890},{"address":"696E70757432","amount":111,"sequence":222}],"outputs":[{"address":"6F757470757431","amount":333},{"address":"6F757470757432","amount":444}]}]}`,
		chainID)
	if signStr != expected {
		t.Errorf("Got unexpected sign string for SendTx. Expected:\n%v\nGot:\n%v", expected, signStr)
	}
}

func TestCallTxSignable(t *testing.T) {
	callTx := &CallTx{
		Input: TxInput{
			Address:  []byte("input1"),
			Amount:   12345,
			Sequence: 67890,
		},
		Address:  []byte("contract1"),
		GasLimit: 111,
		Fee:      222,
		Data:     []byte("data1"),
	}
	signBytes := callTx.SignBytes(chainID)
	signStr := string(signBytes)
	expected := Fmt(`{"chain_id":"%s","tx":[2,{"address":"636F6E747261637431","data":"6461746131","fee":222,"gas_limit":111,"input":{"address":"696E70757431","amount":12345,"sequence":67890}}]}`,
		chainID)
	if signStr != expected {
		t.Errorf("Got unexpected sign string for CallTx. Expected:\n%v\nGot:\n%v", expected, signStr)
	}
}
