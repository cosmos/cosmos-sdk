package tx

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/stretchr/testify/require"
)

func TestAminoAuxHandler(t *testing.T) {
	privKey, pubkey, addr := testdata.KeyTestPubAddr()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	interfaceRegistry.RegisterImplementations((*sdk.Msg)(nil), &testdata.TestMsg{})
	marshaler := codec.NewProtoCodec(interfaceRegistry)

	txConfig := NewTxConfig(marshaler, []signingtypes.SignMode{signingtypes.SignMode_SIGN_MODE_AMINO_AUX})
	txBuilder := txConfig.NewTxBuilder()

	memo := "sometestmemo"
	msgs := []sdk.Msg{testdata.NewTestMsg(addr)}
	accSeq := uint64(2) // Arbitrary account sequence

	// any, err := codectypes.NewAnyWithValue(pubkey)
	// require.NoError(t, err)

	sigData := &signingtypes.SingleSignatureData{
		SignMode: signingtypes.SignMode_SIGN_MODE_AMINO_AUX,
	}

	sig := signingtypes.SignatureV2{
		PubKey:   pubkey,
		Data:     sigData,
		Sequence: accSeq,
	}

	fee := txtypes.Fee{Amount: sdk.NewCoins(sdk.NewInt64Coin("atom", 150)), GasLimit: 20000}

	err := txBuilder.SetMsgs(msgs...)
	require.NoError(t, err)
	txBuilder.SetMemo(memo)
	txBuilder.SetFeeAmount(fee.Amount)
	txBuilder.SetGasLimit(fee.GasLimit)

	err = txBuilder.SetSignatures(sig)
	require.NoError(t, err)

	signingData := signing.SignerData{
		ChainID:       "test-chain",
		AccountNumber: 1,
	}

	modeHandler := signModeAminoAuxHandler{}
	signBytes, err := modeHandler.GetSignBytes(signingtypes.SignMode_SIGN_MODE_AMINO_AUX, signingData, txBuilder.GetTx())

	require.NoError(t, err)
	require.NotNil(t, signBytes)

	anys := make([]*codectypes.Any, len(msgs))
	for i, msg := range msgs {
		var err error
		anys[i], err = codectypes.NewAnyWithValue(msg)
		if err != nil {
			panic(err)
		}
	}

	// txBody := &txtypes.TxBody{
	// 	Memo:     memo,
	// 	Messages: anys,
	// }
	// bodyBytes := marshaler.MustMarshal(txBody)
	txBytes, err := txConfig.TxJSONEncoder()(txBuilder.GetTx())
	raw := make([]json.RawMessage, 1)
	raw = append(raw, txBytes)
	require.NoError(t, err)

	t.Log("verify GetSignBytes with generating sign bytes by marshaling signDocDirectAux")
	signDocDirectAux := txtypes.StdSignDocAux{
		AccountNumber: 1,
		ChainId:       "test-chain",
		TimeoutHeight: uint64(10),
		Memo:          memo,
		Msgs:          raw,
	}

	expectedSignBytes, err := signDocDirectAux.Marshal()
	require.NoError(t, err)
	require.Equal(t, expectedSignBytes, signBytes)

	t.Log("verify that setting signature doesn't change sign bytes")
	sigData.Signature, err = privKey.Sign(signBytes)
	require.NoError(t, err)
	err = txBuilder.SetSignatures(sig)
	require.NoError(t, err)
	signBytes, err = modeHandler.GetSignBytes(signingtypes.SignMode_SIGN_MODE_DIRECT_AUX, signingData, txBuilder.GetTx())
	require.NoError(t, err)
	require.Equal(t, expectedSignBytes, signBytes)

	// t.Log("verify GetSignBytes with false txBody data")
	// signDocDirectAux.BodyBytes = []byte("dfafdasfds")
	// expectedSignBytes, err = signDocDirectAux.Marshal()
	// require.NoError(t, err)
	// require.NotEqual(t, expectedSignBytes, signBytes)
}

func TestAminoAuxHandler_DefaultMode(t *testing.T) {
	handler := signModeAminoAuxHandler{}
	require.Equal(t, signingtypes.SignMode_SIGN_MODE_AMINO_AUX, handler.DefaultMode())
}

func TestAminoAuxModeHandler_nonDIRECT_MODE(t *testing.T) {
	invalidModes := []signingtypes.SignMode{
		signingtypes.SignMode_SIGN_MODE_DIRECT,
		signingtypes.SignMode_SIGN_MODE_DIRECT_JSON,
		signingtypes.SignMode_SIGN_MODE_TEXTUAL,
		signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
		signingtypes.SignMode_SIGN_MODE_UNSPECIFIED,
		signingtypes.SignMode_SIGN_MODE_DIRECT_AUX,
	}
	for _, invalidMode := range invalidModes {
		t.Run(invalidMode.String(), func(t *testing.T) {
			var dh signModeAminoAuxHandler
			var signingData signing.SignerData
			_, err := dh.GetSignBytes(invalidMode, signingData, nil)
			require.Error(t, err)
			wantErr := fmt.Errorf("expected %s, got %s", signingtypes.SignMode_SIGN_MODE_AMINO_AUX, invalidMode)
			require.Equal(t, err, wantErr)
		})
	}
}

func TestAminoAuxModeHandler_nonProtoTx(t *testing.T) {
	var ah signModeAminoAuxHandler
	var signingData signing.SignerData
	tx := new(nonProtoTx)
	_, err := ah.GetSignBytes(signingtypes.SignMode_SIGN_MODE_AMINO_AUX, signingData, tx)
	require.Error(t, err)
	wantErr := fmt.Errorf("can only handle a protobuf Tx, got %T", tx)
	require.Equal(t, err, wantErr)
}
