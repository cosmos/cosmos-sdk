package direct

import (
	"testing"

	"cosmossdk.io/api/cosmos/crypto/secp256k1"
	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"github.com/cosmos/cosmos-proto/any"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/x/tx/signing"
)

func TestDirectModeHandler(t *testing.T) {
	memo := "sometestmemo"

	pk, err := any.New(&secp256k1.PubKey{})
	require.NoError(t, err)
	accSeq := uint64(2) // Arbitrary account sequence

	signerInfo := []*txv1beta1.SignerInfo{
		{
			PublicKey: pk,
			ModeInfo: &txv1beta1.ModeInfo{
				Sum: &txv1beta1.ModeInfo_Single_{
					Single: &txv1beta1.ModeInfo_Single{
						Mode: signingv1beta1.SignMode_SIGN_MODE_DIRECT,
					},
				},
			},
			Sequence: accSeq,
		},
	}
	sigData := &signing.SingleSignatureData{
		SignMode: signingv1beta1.SignMode_SIGN_MODE_DIRECT,
	}
	sig := signing.Signature{
		PubKey:   pk,
		Data:     sigData,
		Sequence: accSeq,
	}

	//fee := &txv1beta1.Fee{Amount: sdk.NewCoins(sdk.NewInt64Coin("atom", 150)), GasLimit: 20000}
	//
	//err = txBuilder.SetMsgs(msgs...)
	//require.NoError(t, err)
	//txBuilder.SetMemo(memo)
	//txBuilder.SetFeeAmount(fee.Amount)
	//txBuilder.SetGasLimit(fee.GasLimit)
	//
	//err = txBuilder.SetSignatures(sig)
	//require.NoError(t, err)
	//
	//t.Log("verify modes and default-mode")
	//modeHandler := txConfig.SignModeHandler()
	//require.Equal(t, modeHandler.DefaultMode(), signingtypes.SignMode_SIGN_MODE_DIRECT)
	//require.Len(t, modeHandler.Modes(), 1)
	//
	//signingData := signing.SignerData{
	//	Address:       addr.String(),
	//	ChainID:       "test-chain",
	//	AccountNumber: 1,
	//	PubKey:        pubkey,
	//}
	//
	//signBytes, err := modeHandler.GetSignBytes(signingtypes.SignMode_SIGN_MODE_DIRECT, signingData, txBuilder.GetTx())
	//
	//require.NoError(t, err)
	//require.NotNil(t, signBytes)
	//
	//authInfo := &txtypes.AuthInfo{
	//	Fee:         &fee,
	//	SignerInfos: signerInfo,
	//}
	//
	//authInfoBytes := marshaler.MustMarshal(authInfo)
	//
	//anys := make([]*codectypes.Any, len(msgs))
	//
	//for i, msg := range msgs {
	//	var err error
	//	anys[i], err = codectypes.NewAnyWithValue(msg)
	//	if err != nil {
	//		panic(err)
	//	}
	//}
	//
	//txBody := &txtypes.TxBody{
	//	Memo:     memo,
	//	Messages: anys,
	//}
	//bodyBytes := marshaler.MustMarshal(txBody)
	//
	//t.Log("verify GetSignBytes with generating sign bytes by marshaling SignDoc")
	//signDoc := txtypes.SignDoc{
	//	AccountNumber: 1,
	//	AuthInfoBytes: authInfoBytes,
	//	BodyBytes:     bodyBytes,
	//	ChainId:       "test-chain",
	//}
	//
	//expectedSignBytes, err := signDoc.Marshal()
	//require.NoError(t, err)
	//require.Equal(t, expectedSignBytes, signBytes)
	//
	//t.Log("verify that setting signature doesn't change sign bytes")
	//sigData.Signature, err = privKey.Sign(signBytes)
	//require.NoError(t, err)
	//err = txBuilder.SetSignatures(sig)
	//require.NoError(t, err)
	//signBytes, err = modeHandler.GetSignBytes(signingtypes.SignMode_SIGN_MODE_DIRECT, signingData, txBuilder.GetTx())
	//require.NoError(t, err)
	//require.Equal(t, expectedSignBytes, signBytes)
	//
	//t.Log("verify GetSignBytes with false txBody data")
	//signDoc.BodyBytes = []byte("dfafdasfds")
	//expectedSignBytes, err = signDoc.Marshal()
	//require.NoError(t, err)
	//require.NotEqual(t, expectedSignBytes, signBytes)
}
