package tx

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

func TestAminoAuxHandler(t *testing.T) {
	privKey, pubkey, addr := testdata.KeyTestPubAddr()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	interfaceRegistry.RegisterImplementations((*sdk.Msg)(nil), &testdata.TestMsg{})
	marshaler := codec.NewProtoCodec(interfaceRegistry)

	txConfig := NewTxConfig(marshaler, []signingtypes.SignMode{signingtypes.SignMode_SIGN_MODE_AMINO_AUX})
	txBuilder := txConfig.NewTxBuilder()

	accountNumber := uint64(1)
	chainId := "test-chain"
	memo := "sometestmemo"
	msgs := []sdk.Msg{testdata.NewTestMsg(addr)}
	accSeq := uint64(2) // Arbitrary account sequence
	timeout := uint64(10)
	fee := txtypes.Fee{Amount: sdk.NewCoins(sdk.NewInt64Coin("atom", 150)), GasLimit: 20000}
	tip := sdk.NewCoins(sdk.NewCoin("regen", sdk.NewInt(1000)))

	err := txBuilder.SetMsgs(msgs...)
	require.NoError(t, err)
	txBuilder.SetMemo(memo)
	txBuilder.SetFeeAmount(fee.Amount)
	txBuilder.SetGasLimit(fee.GasLimit)
	txBuilder.SetTimeoutHeight(timeout)
	txBuilder.SetTip(&txtypes.Tip{
		Amount: tip,
		Tipper: addr.String(), // Not needed when signing using AMINO_AUX, but putting here for clarity.
	})

	sigData := &signingtypes.SingleSignatureData{
		SignMode: signingtypes.SignMode_SIGN_MODE_AMINO_AUX,
	}

	sig := signingtypes.SignatureV2{
		PubKey:   pubkey,
		Data:     sigData,
		Sequence: accSeq,
	}
	err = txBuilder.SetSignatures(sig)
	require.NoError(t, err)

	signingData := signing.SignerData{
		ChainID:       chainId,
		AccountNumber: accountNumber,
		Sequence:      accSeq,
		SignerIndex:   0,
	}

	handler := signModeAminoAuxHandler{}
	signBytes, err := handler.GetSignBytes(signingtypes.SignMode_SIGN_MODE_AMINO_AUX, signingData, txBuilder.GetTx())
	require.NoError(t, err)
	require.NotNil(t, signBytes)

	expectedSignBytes := legacytx.StdSignAuxBytes(chainId, accountNumber, accSeq, timeout, tip, msgs, memo)
	require.NoError(t, err)
	require.Equal(t, expectedSignBytes, signBytes)

	t.Log("verify that setting signature doesn't change sign bytes")
	sigData.Signature, err = privKey.Sign(signBytes)
	require.NoError(t, err)
	err = txBuilder.SetSignatures(sig)
	require.NoError(t, err)
	signBytes, err = handler.GetSignBytes(signingtypes.SignMode_SIGN_MODE_AMINO_AUX, signingData, txBuilder.GetTx())
	require.NoError(t, err)
	require.Equal(t, expectedSignBytes, signBytes)

	// expect error with wrong sign mode
	_, err = handler.GetSignBytes(signingtypes.SignMode_SIGN_MODE_DIRECT, signingData, txBuilder.GetTx())
	require.Error(t, err)

	// expect error with extension options
	bldr := newBuilder()
	buildTx(t, bldr)
	any, err := codectypes.NewAnyWithValue(testdata.NewTestMsg())
	require.NoError(t, err)
	bldr.tx.Body.ExtensionOptions = []*codectypes.Any{any}
	tx := bldr.GetTx()
	_, err = handler.GetSignBytes(signingtypes.SignMode_SIGN_MODE_AMINO_AUX, signingData, tx)
	require.Error(t, err)

	// expect error with non-critical extension options
	bldr = newBuilder()
	buildTx(t, bldr)
	bldr.tx.Body.NonCriticalExtensionOptions = []*codectypes.Any{any}
	tx = bldr.GetTx()
	_, err = handler.GetSignBytes(signingtypes.SignMode_SIGN_MODE_AMINO_AUX, signingData, tx)
	require.Error(t, err)
}

func TestAminoAuxHandler_DefaultMode(t *testing.T) {
	handler := signModeAminoAuxHandler{}
	require.Equal(t, signingtypes.SignMode_SIGN_MODE_AMINO_AUX, handler.DefaultMode())
}

func TestAminoAuxModeHandler_nonDIRECT_MODE(t *testing.T) {
	invalidModes := []signingtypes.SignMode{
		signingtypes.SignMode_SIGN_MODE_DIRECT,
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
