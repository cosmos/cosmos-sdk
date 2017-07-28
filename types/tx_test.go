package types

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	data "github.com/tendermint/go-wire/data"
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
	signBytesHex := fmt.Sprintf("%X", signBytes)
	expected := "010A746573745F636861696E0100000000000000DE00000000000000006F01020106696E7075743101010000000000000030390301093200000106696E70757432010100000000000000006F01DE0000010201076F757470757431010100000000000000014D01076F75747075743201010000000000000001BC"

	assert.Equal(t, signBytesHex, expected,
		"Got unexpected sign string for SendTx. Expected:\n%v\nGot:\n%v", expected, signBytesHex)
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
	signBytesHex := fmt.Sprintf("%X", signBytes)
	expected := "010A746573745F636861696E0100000000000000DE00000000000000006F0101580106696E70757431010100000000000000303903010932000001056461746131"

	assert.Equal(t, signBytesHex, expected,
		"Got unexpected sign string for SendTx. Expected:\n%v\nGot:\n%v", expected, signBytesHex)
}

func TestSendTxJSON(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	chainID := "test_chain_id"
	test1PrivAcc := PrivAccountFromSecret("sendtx1")
	test2PrivAcc := PrivAccountFromSecret("sendtx2")

	// Construct a SendTx signature
	tx := &SendTx{
		Gas: 1,
		Fee: Coin{"foo", 2},
		Inputs: []TxInput{
			NewTxInput(test1PrivAcc.PubKey, Coins{{"foo", 10}}, 1),
		},
		Outputs: []TxOutput{
			TxOutput{
				Address: test2PrivAcc.PubKey.Address(),
				Coins:   Coins{{"foo", 8}},
			},
		},
	}

	// serialize this as json and back
	js, err := data.ToJSON(TxS{tx})
	require.Nil(err)
	// fmt.Println(string(js))
	txs := TxS{}
	err = data.FromJSON(js, &txs)
	require.Nil(err)
	tx2, ok := txs.Tx.(*SendTx)
	require.True(ok)

	// make sure they are the same!
	signBytes := tx.SignBytes(chainID)
	signBytes2 := tx2.SignBytes(chainID)
	assert.Equal(signBytes, signBytes2)
	assert.Equal(tx, tx2)

	// sign this thing
	sig := test1PrivAcc.Sign(signBytes)
	// we handle both raw sig and wrapped sig the same
	tx.SetSignature(test1PrivAcc.PubKey.Address(), sig)
	tx2.SetSignature(test1PrivAcc.PubKey.Address(), sig)
	assert.Equal(tx, tx2)

	// let's marshal / unmarshal this with signature
	js, err = data.ToJSON(TxS{tx})
	require.Nil(err)
	// fmt.Println(string(js))
	err = data.FromJSON(js, &txs)
	require.Nil(err)
	tx2, ok = txs.Tx.(*SendTx)
	require.True(ok)

	// and make sure the sig is preserved
	assert.Equal(tx, tx2)
	assert.False(tx2.Inputs[0].Signature.Empty())
}

func TestSendTxIBC(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	good, err := hex.DecodeString("1960CA7E170862837AA8F22A947194F41F61860B")
	require.Nil(err)
	short, err := hex.DecodeString("1960CA7E170862837AA8F22F947194F41F610B")
	require.Nil(err)
	long, err := hex.DecodeString("1960CA7E170862837AA8F22F947194F41F6186120B")
	require.Nil(err)
	slash, err := hex.DecodeString("F40ECECEA86F29D0FDF2980EF72F1708687BD4BF")
	require.Nil(err)

	coins := Coins{{"atom", 5}}

	addrs := []struct {
		addr  []byte
		valid bool
	}{
		{good, true},
		{slash, true},
		{long, false},
		{short, false},
	}

	prefixes := []struct {
		prefix []byte
		valid  bool
	}{
		{nil, true},
		{[]byte("chain-1/"), true},
		{[]byte("chain/with/paths/"), false},
		{[]byte("no-slash-here"), false},
	}

	for i, tc := range addrs {
		for j, pc := range prefixes {
			addr := append(pc.prefix, tc.addr...)
			output := TxOutput{Address: addr, Coins: coins}
			res := output.ValidateBasic()

			if tc.valid && pc.valid {
				assert.True(res.IsOK(), "%d,%d: %s", i, j, res.Log)
			} else {
				assert.False(res.IsOK(), "%d,%d: %s", i, j, res.Log)
			}
		}
	}

}
