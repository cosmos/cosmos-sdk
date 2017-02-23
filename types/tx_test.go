package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cmn "github.com/tendermint/go-common"
	crypto "github.com/tendermint/go-crypto"
	data "github.com/tendermint/go-data"
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
		cmn.Fmt("Got unexpected sign string for SendTx. Expected:\n%v\nGot:\n%v", expected, signBytesHex))
}

// d'oh, can't use the version in testutils due to circular imports :(
func makePrivAcct() PrivAccount {
	privKey := crypto.PrivKeyS{crypto.GenPrivKeyEd25519()}
	return PrivAccount{
		PrivKeyS: privKey,
		Account: Account{
			PubKey: crypto.PubKeyS{privKey.PubKey()},
		},
	}
}

func TestSendTxJSON(t *testing.T) {
	chainID := "test_chain_id"
	test1PrivAcc := makePrivAcct()
	test2PrivAcc := makePrivAcct()

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
	require.Nil(t, err)
	// fmt.Println(string(js))
	txs := TxS{}
	err = data.FromJSON(js, &txs)
	require.Nil(t, err)
	tx2, ok := txs.Tx.(*SendTx)
	require.True(t, ok)

	// make sure they are the same!
	signBytes := tx.SignBytes(chainID)
	signBytes2 := tx2.SignBytes(chainID)
	assert.Equal(t, signBytes, signBytes2)
	assert.Equal(t, tx, tx2)

	// sign this thing
	sig := test1PrivAcc.Sign(signBytes)
	// we handle both raw sig and wrapped sig the same
	tx.SetSignature(test1PrivAcc.PubKey.Address(), sig)
	tx2.SetSignature(test1PrivAcc.PubKey.Address(), crypto.SignatureS{sig})
	assert.Equal(t, tx, tx2)

	// let's marshal / unmarshal this with signature
	js, err = data.ToJSON(TxS{tx})
	require.Nil(t, err)
	// fmt.Println(string(js))
	err = data.FromJSON(js, &txs)
	require.Nil(t, err)
	tx2, ok = txs.Tx.(*SendTx)
	require.True(t, ok)

	// and make sure the sig is preserved
	assert.Equal(t, tx, tx2)
	assert.False(t, tx2.Inputs[0].Signature.Empty())
}
