package tx_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// TestBuilderWithAux creates a tx with 2 aux signers:
// - 1st one is tipper,
// - 2nd one is just an aux signer.
// Then it tests integrating the 2 AuxSignerData into a
// client.TxBuilder created by the fee payer.
func TestBuilderWithAux(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	testdata.RegisterInterfaces(encCfg.InterfaceRegistry)

	// The final TX has 3 signers, in this order.
	tipperPriv, tipperPk, tipperAddr := testdata.KeyTestPubAddr()
	aux2Priv, aux2Pk, aux2Addr := testdata.KeyTestPubAddr()
	feepayerPriv, feepayerPk, feepayerAddr := testdata.KeyTestPubAddr()

	msg := testdata.NewTestMsg(tipperAddr, aux2Addr)
	memo := "test-memo"
	tip := &txtypes.Tip{Tipper: tipperAddr.String(), Amount: sdk.NewCoins(sdk.NewCoin("tip-denom", sdk.NewIntFromUint64(123)))}
	chainID := "test-chain"
	gas := testdata.NewTestGasLimit()
	fee := testdata.NewTestFeeAmount()

	// Create an AuxTxBuilder for tipper (1st signer)
	tipperBuilder := clienttx.NewAuxTxBuilder()
	tipperBuilder.SetAccountNumber(1)
	tipperBuilder.SetSequence(2)
	tipperBuilder.SetTimeoutHeight(3)
	tipperBuilder.SetMemo(memo)
	tipperBuilder.SetChainID(chainID)
	tipperBuilder.SetMsgs(msg)
	tipperBuilder.SetPubKey(tipperPk)
	tipperBuilder.SetTip(tip)
	err := tipperBuilder.SetSignMode(signing.SignMode_SIGN_MODE_DIRECT_AUX)
	require.NoError(t, err)
	signBz, err := tipperBuilder.GetSignBytes()
	require.NoError(t, err)
	tipperSig, err := tipperPriv.Sign(signBz)
	require.NoError(t, err)
	tipperBuilder.SetSignature(tipperSig)
	tipperSignerData, err := tipperBuilder.GetAuxSignerData()
	require.NoError(t, err)

	// Create an AuxTxBuilder for aux2 (2nd signer)
	aux2Builder := clienttx.NewAuxTxBuilder()
	aux2Builder.SetAccountNumber(11)
	aux2Builder.SetSequence(12)
	aux2Builder.SetTimeoutHeight(3)
	aux2Builder.SetMemo(memo)
	aux2Builder.SetChainID(chainID)
	aux2Builder.SetMsgs(msg)
	aux2Builder.SetPubKey(aux2Pk)
	aux2Builder.SetTip(tip)
	err = aux2Builder.SetSignMode(signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)
	require.NoError(t, err)
	signBz, err = aux2Builder.GetSignBytes()
	require.NoError(t, err)
	aux2Sig, err := aux2Priv.Sign(signBz)
	require.NoError(t, err)
	aux2Builder.SetSignature(aux2Sig)
	aux2SignerData, err := aux2Builder.GetAuxSignerData()
	require.NoError(t, err)

	// Fee payer (3rd and last signer) creates a TxBuilder.
	w := encCfg.TxConfig.NewTxBuilder()
	// Note: we're testing calling AddAuxSignerData in the wrong order.
	err = w.AddAuxSignerData(aux2SignerData)
	require.NoError(t, err)
	err = w.AddAuxSignerData(tipperSignerData)
	require.NoError(t, err)
	w.SetFeePayer(feepayerAddr)
	w.SetFeeAmount(fee)
	w.SetGasLimit(gas)
	sigs, err := w.(authsigning.SigVerifiableTx).GetSignaturesV2()
	require.NoError(t, err)
	tipperSigV2 := sigs[0]
	aux2SigV2 := sigs[1]
	// Set all signer infos.
	w.SetSignatures(tipperSigV2, aux2SigV2, signing.SignatureV2{
		PubKey:   feepayerPk,
		Sequence: 15,
	})
	signBz, err = encCfg.TxConfig.SignModeHandler().GetSignBytes(
		signing.SignMode_SIGN_MODE_DIRECT,
		authsigning.SignerData{Address: feepayerAddr.String(), ChainID: chainID, AccountNumber: 11, Sequence: 15, SignerIndex: 1},
		w.GetTx(),
	)
	require.NoError(t, err)
	feepayerSig, err := feepayerPriv.Sign(signBz)
	require.NoError(t, err)
	// Set all signatures.
	w.SetSignatures(tipperSigV2, aux2SigV2, signing.SignatureV2{
		PubKey: feepayerPk,
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: feepayerSig,
		},
		Sequence: 22,
	})

	// Make sure tx is correct.
	txBz, err := encCfg.TxConfig.TxEncoder()(w.GetTx())
	require.NoError(t, err)
	tx, err := encCfg.TxConfig.TxDecoder()(txBz)
	require.NoError(t, err)
	require.Equal(t, tx.(sdk.FeeTx).FeePayer(), feepayerAddr)
	require.Equal(t, tx.(sdk.FeeTx).GetFee(), fee)
	require.Equal(t, tx.(sdk.FeeTx).GetGas(), gas)
	require.Equal(t, tip, tx.(txtypes.TipTx).GetTip())
	require.Equal(t, msg, tx.GetMsgs()[0])
	require.Equal(t, memo, tx.(sdk.TxWithMemo).GetMemo())
	require.Equal(t, uint64(3), tx.(sdk.TxWithTimeoutHeight).GetTimeoutHeight())
	sigs, err = tx.(authsigning.Tx).GetSignaturesV2()
	require.NoError(t, err)
	require.Len(t, sigs, 3)
	require.Equal(t, signing.SignatureV2{
		PubKey:   tipperPk,
		Data:     &signing.SingleSignatureData{SignMode: signing.SignMode_SIGN_MODE_DIRECT_AUX, Signature: tipperSig},
		Sequence: 2,
	}, sigs[0])
	require.Equal(t, signing.SignatureV2{
		PubKey:   aux2Pk,
		Data:     &signing.SingleSignatureData{SignMode: signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, Signature: aux2Sig},
		Sequence: 12,
	}, sigs[1])
	require.Equal(t, signing.SignatureV2{
		PubKey:   feepayerPk,
		Data:     &signing.SingleSignatureData{SignMode: signing.SignMode_SIGN_MODE_DIRECT, Signature: feepayerSig},
		Sequence: 22,
	}, sigs[2])
}
