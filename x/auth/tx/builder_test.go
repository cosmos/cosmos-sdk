package tx

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec/testdata"

	tx2 "github.com/cosmos/cosmos-sdk/types/tx"

	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	"github.com/stretchr/testify/require"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestTxBuilder(t *testing.T) {
	_, pubkey, addr := authtypes.KeyTestPubAddr()

	marshaler := codec.NewHybridCodec(codec.New(), codectypes.NewInterfaceRegistry())
	tx := newBuilder(marshaler, std.DefaultPublicKeyCodec{})

	cdc := std.DefaultPublicKeyCodec{}

	memo := "sometestmemo"
	msgs := []sdk.Msg{testdata.NewTestMsg(addr)}

	pk, err := cdc.Encode(pubkey)
	require.NoError(t, err)

	var signerInfo []*tx2.SignerInfo
	signerInfo = append(signerInfo, &tx2.SignerInfo{
		PublicKey: pk,
		ModeInfo: &tx2.ModeInfo{
			Sum: &tx2.ModeInfo_Single_{
				Single: &tx2.ModeInfo_Single{
					Mode: signing.SignMode_SIGN_MODE_DIRECT,
				},
			},
		},
	})

	var sig signing.SignatureV2
	sig = signing.SignatureV2{
		PubKey: pubkey,
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: pubkey.Bytes(),
		},
	}

	fee := tx2.Fee{Amount: sdk.NewCoins(sdk.NewInt64Coin("atom", 150)), GasLimit: 20000}

	t.Log("verify that authInfo bytes encoded with DefaultTxEncoder and decoded with DefaultTxDecoder can be retrieved from GetAuthInfoBytes")
	authInfo := &tx2.AuthInfo{
		Fee:         &fee,
		SignerInfos: signerInfo,
	}

	authInfoBytes := marshaler.MustMarshalBinaryBare(authInfo)

	require.NotEmpty(t, authInfoBytes)

	t.Log("verify that body bytes encoded with DefaultTxEncoder and decoded with DefaultTxDecoder can be retrieved from GetBodyBytes")
	anys := make([]*codectypes.Any, len(msgs))

	for i, msg := range msgs {
		var err error
		anys[i], err = codectypes.NewAnyWithValue(msg)
		if err != nil {
			panic(err)
		}
	}

	txBody := &tx2.TxBody{
		Memo:     memo,
		Messages: anys,
	}
	bodyBytes := marshaler.MustMarshalBinaryBare(txBody)
	require.NotEmpty(t, bodyBytes)
	require.Empty(t, tx.GetBodyBytes())

	t.Log("verify that calling the SetMsgs, SetMemo results in the correct GetBodyBytes")
	require.NotEqual(t, bodyBytes, tx.GetBodyBytes())
	err = tx.SetMsgs(msgs...)
	require.NoError(t, err)
	require.NotEqual(t, bodyBytes, tx.GetBodyBytes())
	tx.SetMemo(memo)
	require.Equal(t, bodyBytes, tx.GetBodyBytes())
	require.Equal(t, len(msgs), len(tx.GetMsgs()))
	require.Equal(t, 0, len(tx.GetPubKeys()))

	t.Log("verify that updated AuthInfo  results in the correct GetAuthInfoBytes and GetPubKeys")
	require.NotEqual(t, authInfoBytes, tx.GetAuthInfoBytes())
	tx.SetFeeAmount(fee.Amount)
	require.NotEqual(t, authInfoBytes, tx.GetAuthInfoBytes())
	tx.SetGasLimit(fee.GasLimit)
	require.NotEqual(t, authInfoBytes, tx.GetAuthInfoBytes())
	err = tx.SetSignatures(sig)
	require.NoError(t, err)

	// once fee, gas and signerInfos are all set, AuthInfo bytes should match
	require.Equal(t, authInfoBytes, tx.GetAuthInfoBytes())

	require.Equal(t, len(msgs), len(tx.GetMsgs()))
	require.Equal(t, 1, len(tx.GetPubKeys()))
	require.Equal(t, pubkey.Bytes(), tx.GetPubKeys()[0].Bytes())
}
