package tx

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestTxWrapper(t *testing.T) {
	_, pubkey, addr := authtypes.KeyTestPubAddr()

	marshaler := codec.NewHybridCodec(codec.New(), codectypes.NewInterfaceRegistry())
	tx := NewBuilder(marshaler, std.DefaultPublicKeyCodec{})

	cdc := std.DefaultPublicKeyCodec{}

	memo := "sometestmemo"
	msgs := []sdk.Msg{types.NewTestMsg(addr)}

	pk, err := cdc.Encode(pubkey)
	require.NoError(t, err)

	var signerInfo []*SignerInfo
	signerInfo = append(signerInfo, &SignerInfo{
		PublicKey: pk,
		ModeInfo: &ModeInfo{
			Sum: &ModeInfo_Single_{
				Single: &ModeInfo_Single{
					Mode: signing.SignMode_SIGN_MODE_DIRECT,
				},
			},
		},
	})

	var sig signing.SignatureV2
	sig = signing.SignatureV2{
		PubKey: pubkey,
		Data: &signing.SingleSignatureData{
			SignMode: signing.SignMode_SIGN_MODE_DIRECT,
			Signature: pubkey.Bytes(),
		},
	}

	fee := Fee{Amount: sdk.NewCoins(sdk.NewInt64Coin("atom", 150)), GasLimit: 20000}

	t.Log("verify that authInfo bytes encoded with DefaultTxEncoder and decoded with DefaultTxDecoder can be retrieved from GetAuthInfoBytes")
	authInfo := &AuthInfo{
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

	txBody := &TxBody{
		Memo:     memo,
		Messages: anys,
	}
	bodyBytes := marshaler.MustMarshalBinaryBare(txBody)
	require.NotEmpty(t, bodyBytes)
	require.Empty(t, tx.GetBodyBytes())

	t.Log("verify that calling the SetMsgs, SetMemo results in the correct GetBodyBytes")
	require.NotEqual(t, bodyBytes, tx.GetBodyBytes())
	tx.SetMsgs(msgs)
	require.NotEqual(t, bodyBytes, tx.GetBodyBytes())
	tx.SetMemo(memo)
	require.Equal(t, bodyBytes, tx.GetBodyBytes())
	require.Equal(t, len(msgs), len(tx.GetMsgs()))
	require.Equal(t, 0, len(tx.GetPubKeys()))

	t.Log("verify that updated AuthInfo  results in the correct GetAuthInfoBytes and GetPubKeys")
	require.NotEqual(t, authInfoBytes, tx.GetAuthInfoBytes())
	tx.SetFee(fee.Amount)
	require.NotEqual(t, authInfoBytes, tx.GetAuthInfoBytes())
	tx.SetGas(fee.GasLimit)
	require.NotEqual(t, authInfoBytes, tx.GetAuthInfoBytes())
	tx.SetSignaturesV2(sig)

	// once fee, gas and signerInfos are all set, AuthInfo bytes should match
	require.Equal(t, authInfoBytes, tx.GetAuthInfoBytes())

	require.Equal(t, len(msgs), len(tx.GetMsgs()))
	require.Equal(t, 1, len(tx.GetPubKeys()))
	require.Equal(t, pubkey.Bytes(), tx.GetPubKeys()[0].Bytes())
}
