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

func TestBuilderWithAux(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	testdata.RegisterInterfaces(encCfg.InterfaceRegistry)

	tipperPriv, tipperPk, tipperAddr := testdata.KeyTestPubAddr()
	feepayerPriv, feepayerPk, feepayerAddr := testdata.KeyTestPubAddr()
	msg := testdata.NewTestMsg(tipperAddr)
	memo := "test-memo"
	tip := &txtypes.Tip{Tipper: tipperAddr.String(), Amount: sdk.NewCoins(sdk.NewCoin("tip-denom", sdk.NewIntFromUint64(123)))}
	chainID := "test-chain"
	gas := testdata.NewTestGasLimit()
	fee := testdata.NewTestFeeAmount()

	// Create an AuxTxBuilder
	auxBuilder := clienttx.NewAuxTxBuilder()
	auxBuilder.SetAccountNumber(1)
	auxBuilder.SetSequence(2)
	auxBuilder.SetTimeoutHeight(3)
	auxBuilder.SetMemo(memo)
	auxBuilder.SetChainID(chainID)
	auxBuilder.SetMsgs(msg)
	auxBuilder.SetPubKey(tipperPk)
	auxBuilder.SetTip(tip)
	err := auxBuilder.SetSignMode(signing.SignMode_SIGN_MODE_DIRECT_AUX)
	require.NoError(t, err)
	signBz, err := auxBuilder.GetSignBytes()
	require.NoError(t, err)
	tipperSig, err := tipperPriv.Sign(signBz)
	require.NoError(t, err)
	auxBuilder.SetSignature(tipperSig)
	auxSignerData, err := auxBuilder.GetAuxSignerData()
	require.NoError(t, err)

	// Fee payer creates a TxBuilder.
	w := encCfg.TxConfig.NewTxBuilder()
	err = w.AddAuxSignerData(auxSignerData)
	require.NoError(t, err)
	w.SetFeePayer(feepayerAddr)
	w.SetFeeAmount(fee)
	w.SetGasLimit(gas)
	sigs, err := w.(authsigning.SigVerifiableTx).GetSignaturesV2()
	require.NoError(t, err)
	tipperSigV2 := sigs[0]
	// Set all signer infos.
	w.SetSignatures(tipperSigV2, signing.SignatureV2{
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
	w.SetSignatures(tipperSigV2, signing.SignatureV2{
		PubKey: feepayerPk,
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: feepayerSig,
		},
		Sequence: 15,
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
	require.Equal(t, signing.SignatureV2{
		PubKey:   tipperPk,
		Data:     &signing.SingleSignatureData{SignMode: signing.SignMode_SIGN_MODE_DIRECT_AUX, Signature: tipperSig},
		Sequence: 2,
	}, sigs[0])
	require.Equal(t, signing.SignatureV2{
		PubKey:   feepayerPk,
		Data:     &signing.SingleSignatureData{SignMode: signing.SignMode_SIGN_MODE_DIRECT, Signature: feepayerSig},
		Sequence: 15,
	}, sigs[1])
}
