package app

import (
	"testing"
	"github.com/tendermint/tendermint/crypto"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/stretchr/testify/require"
)

// Test encoding of app2Tx is correct with both msg types
func TestEncoding(t *testing.T) {
	// Create privkeys and addresses
	priv1 := crypto.GenPrivKeyEd25519()
	priv2 := crypto.GenPrivKeyEd25519()
	addr1 := priv1.PubKey().Address().Bytes()
	addr2 := priv2.PubKey().Address().Bytes()

	sendMsg := MsgSend{
		From: addr1,
		To: addr2,
		Amount: sdk.Coins{{"testCoins", sdk.NewInt(100)}},
	}

	// Construct transaction
	fee := auth.StdFee{
		Gas: 1000000000000000,
		Amount: sdk.Coins{{"testCoin", sdk.NewInt(0)}},
	}
	signBytes := auth.StdSignBytes("test-chain", 0, 0, fee, []sdk.Msg{sendMsg}, "")
	sig, err := priv1.Sign(signBytes)
	if err != nil {
		panic(err)
	}
	sigs := []auth.StdSignature{auth.StdSignature{
		PubKey: priv1.PubKey(),
		Signature: sig,
		AccountNumber: 0,
		Sequence: 0,
	}}

	sendTxBefore := app2Tx{
		Msg: sendMsg,
		Signatures: sigs,
	}

	cdc := NewCodec()

	encodedSendTx, err := cdc.MarshalBinary(sendTxBefore)

	require.Nil(t, err, "Error encoding sendTx")

	var sendTxAfter app2Tx
	err = cdc.UnmarshalBinary(encodedSendTx, &sendTxAfter)

	require.Nil(t, err, "Error decoding sendTx")
	require.Equal(t, sendTxBefore, sendTxAfter, "Transaction changed after encoding/decoding")

	issueMsg := MsgIssue{
		Issuer: addr1,
		Receiver: addr2,
		Coin: sdk.Coin{"testCoin", sdk.NewInt(100)},
	}

	signBytes = auth.StdSignBytes("test-chain", 0, 0, fee, []sdk.Msg{issueMsg}, "")
	sig, err = priv1.Sign(signBytes)
	if err != nil {
		panic(err)
	}
	sigs = []auth.StdSignature{auth.StdSignature{
		PubKey: priv1.PubKey(),
		Signature: sig,
		AccountNumber: 0,
		Sequence: 1,
	}}

	issueTxBefore := app2Tx{
		Msg: issueMsg,
		Signatures: sigs,
	}

	encodedIssueTx, err2 := cdc.MarshalBinary(issueTxBefore)

	require.Nil(t, err2, "Error encoding issueTx")

	var issueTxAfter app2Tx
	err2 = cdc.UnmarshalBinary(encodedIssueTx, &issueTxAfter)

	require.Nil(t, err2, "Error decoding issue Tx")
	require.Equal(t, issueTxBefore, issueTxAfter, "Transaction changed after encoding/decoding")

}