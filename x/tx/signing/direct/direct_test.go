package direct_test

import (
	"context"
	"testing"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	"cosmossdk.io/api/cosmos/crypto/secp256k1"
	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"github.com/cosmos/cosmos-proto/any"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/tx/signing/direct"
)

func TestDirectModeHandler(t *testing.T) {
	memo := "sometestmemo"

	msg, err := any.New(&bankv1beta1.MsgSend{})
	require.NoError(t, err)

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

	fee := &txv1beta1.Fee{Amount: []*basev1beta1.Coin{{Denom: "uatom", Amount: "1000"}}, GasLimit: 20000}
	txBody := &txv1beta1.TxBody{
		Messages: []*anypb.Any{msg},
		Memo:     memo,
	}

	authInfo := &txv1beta1.AuthInfo{
		Fee:         fee,
		SignerInfos: signerInfo,
	}

	directHandler := direct.SignModeHandler{}

	signingData := signing.SignerData{
		Address:       "",
		ChainId:       "test-chain",
		AccountNumber: 1,
		PubKey:        pk,
	}

	bodyBz, err := proto.Marshal(txBody)
	require.NoError(t, err)

	authInfoBz, err := proto.Marshal(authInfo)
	require.NoError(t, err)

	txData := signing.TxData{
		Tx: &txv1beta1.Tx{
			Body:     txBody,
			AuthInfo: authInfo,
		},
		TxRaw: &txv1beta1.TxRaw{
			BodyBytes:     bodyBz,
			AuthInfoBytes: authInfoBz,
		},
		BodyHasUnknownNonCriticals: false,
	}

	signBytes, err := directHandler.GetSignBytes(context.Background(), signingData, txData)
	require.NoError(t, err)
	require.NotNil(t, signBytes)

	//authInfo := &txv1beta1.AuthInfo{
	//	Fee:         fee,
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
