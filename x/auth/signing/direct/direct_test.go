package direct_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/x/auth/tx"

	"github.com/cosmos/cosmos-sdk/x/auth/signing/direct"

	"github.com/cosmos/cosmos-sdk/codec/testdata"

	"github.com/cosmos/cosmos-sdk/codec"

	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"

	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestDirectModeHandler(t *testing.T) {
	privKey, pubkey, addr := authtypes.KeyTestPubAddr()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	interfaceRegistry.RegisterImplementations((*sdk.Msg)(nil), &testdata.TestMsg{})
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	pubKeyCdc := std.DefaultPublicKeyCodec{}

	txGen := tx.NewTxGenerator(marshaler, pubKeyCdc, tx.DefaultSignModeHandler())
	txBuilder := txGen.NewTxBuilder()

	memo := "sometestmemo"
	msgs := []sdk.Msg{testdata.NewTestMsg(addr)}

	pk, err := pubKeyCdc.Encode(pubkey)
	require.NoError(t, err)

	var signerInfo []*txtypes.SignerInfo
	signerInfo = append(signerInfo, &txtypes.SignerInfo{
		PublicKey: pk,
		ModeInfo: &txtypes.ModeInfo{
			Sum: &txtypes.ModeInfo_Single_{
				Single: &txtypes.ModeInfo_Single{
					Mode: signingtypes.SignMode_SIGN_MODE_DIRECT,
				},
			},
		},
	})

	sigData := &signingtypes.SingleSignatureData{
		SignMode: signingtypes.SignMode_SIGN_MODE_DIRECT,
	}
	sig := signingtypes.SignatureV2{
		PubKey: pubkey,
		Data:   sigData,
	}

	fee := txtypes.Fee{Amount: sdk.NewCoins(sdk.NewInt64Coin("atom", 150)), GasLimit: 20000}

	err = txBuilder.SetMsgs(msgs...)
	require.NoError(t, err)
	txBuilder.SetMemo(memo)
	txBuilder.SetFeeAmount(fee.Amount)
	txBuilder.SetGasLimit(fee.GasLimit)

	err = txBuilder.SetSignatures(sig)
	require.NoError(t, err)

	t.Log("verify modes and default-mode")
	directModeHandler := direct.ModeHandler{}
	require.Equal(t, directModeHandler.DefaultMode(), signingtypes.SignMode_SIGN_MODE_DIRECT)
	require.Len(t, directModeHandler.Modes(), 1)

	signingData := signing.SignerData{
		ChainID:         "test-chain",
		AccountNumber:   1,
		AccountSequence: 1,
	}

	signBytes, err := directModeHandler.GetSignBytes(signingtypes.SignMode_SIGN_MODE_DIRECT, signingData, txBuilder.GetTx())

	require.NoError(t, err)
	require.NotNil(t, signBytes)

	authInfo := &txtypes.AuthInfo{
		Fee:         &fee,
		SignerInfos: signerInfo,
	}

	authInfoBytes := marshaler.MustMarshalBinaryBare(authInfo)

	anys := make([]*codectypes.Any, len(msgs))

	for i, msg := range msgs {
		var err error
		anys[i], err = codectypes.NewAnyWithValue(msg)
		if err != nil {
			panic(err)
		}
	}

	txBody := &txtypes.TxBody{
		Memo:     memo,
		Messages: anys,
	}
	bodyBytes := marshaler.MustMarshalBinaryBare(txBody)

	t.Log("verify GetSignBytes with generating sign bytes by marshaling SignDoc")
	signDoc := txtypes.SignDoc{
		AccountNumber:   1,
		AccountSequence: 1,
		AuthInfoBytes:   authInfoBytes,
		BodyBytes:       bodyBytes,
		ChainId:         "test-chain",
	}

	expectedSignBytes, err := signDoc.Marshal()
	require.NoError(t, err)
	require.Equal(t, expectedSignBytes, signBytes)

	t.Log("verify that setting signature doesn't change sign bytes")
	sigData.Signature, err = privKey.Sign(signBytes)
	require.NoError(t, err)
	err = txBuilder.SetSignatures(sig)
	require.NoError(t, err)
	signBytes, err = directModeHandler.GetSignBytes(signingtypes.SignMode_SIGN_MODE_DIRECT, signingData, txBuilder.GetTx())
	require.NoError(t, err)
	require.Equal(t, expectedSignBytes, signBytes)

	t.Log("verify GetSignBytes with false txBody data")
	signDoc.BodyBytes = []byte("dfafdasfds")
	expectedSignBytes, err = signDoc.Marshal()
	require.NoError(t, err)
	require.NotEqual(t, expectedSignBytes, signBytes)
}
