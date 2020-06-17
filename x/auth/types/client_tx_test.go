package types_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/client"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	priv = secp256k1.GenPrivKey()
	addr = sdk.AccAddress(priv.PubKey().Address())
)

func makeCodec() *codec.Codec {
	cdc := codec.New()
	sdk.RegisterCodec(cdc)
	cdc.RegisterConcrete(sdk.TestMsg{}, "cosmos-sdk/Test", nil)
	return cdc
}

func setupStdTxBuilderTest() client.TxBuilder {
	stdTxGen := types.StdTxGenerator{
		Cdc: makeCodec(),
	}

	return stdTxGen.NewTxBuilder()
}

func TestStdTxBuilder_GetTx(t *testing.T) {
	stdTxBuilder := setupStdTxBuilderTest()
	tx := stdTxBuilder.GetTx()
	require.NotNil(t, tx)
	require.Equal(t, len(tx.GetMsgs()), 0)
}

func TestStdTxBuilder_SetFeeAmount(t *testing.T) {
	stdTxBuilder := setupStdTxBuilderTest()
	feeAmount := sdk.Coins{
		sdk.NewInt64Coin("atom", 20000000),
	}
	tx := stdTxBuilder.GetTx()
	stdTxBuilder.SetFeeAmount(feeAmount)
	feeTx := stdTxBuilder.GetTx().(sdk.FeeTx)
	require.Equal(t, feeTx.GetFee(), feeAmount)
	require.False(t, reflect.DeepEqual(tx, stdTxBuilder.GetTx()))
}

func TestStdTxBuilder_SetGasLimit(t *testing.T) {
	const newGas uint64 = 300000
	stdTxBuilder := setupStdTxBuilderTest()
	tx := stdTxBuilder.GetTx()
	stdTxBuilder.SetGasLimit(newGas)
	feeTx := stdTxBuilder.GetTx().(sdk.FeeTx)
	require.Equal(t, feeTx.GetGas(), newGas)
	require.False(t, reflect.DeepEqual(tx, stdTxBuilder.GetTx()))
}

func TestStdTxBuilder_SetMemo(t *testing.T) {
	const newMemo string = "newfoomemo"
	stdTxBuilder := setupStdTxBuilderTest()
	stdTxBuilder.SetMemo(newMemo)
	txWithMemo := stdTxBuilder.GetTx().(sdk.TxWithMemo)
	require.Equal(t, txWithMemo.GetMemo(), newMemo)
}

func TestStdTxBuilder_SetMsgs(t *testing.T) {
	stdTxBuilder := setupStdTxBuilderTest()
	tx := stdTxBuilder.GetTx()
	stdTxBuilder.SetMsgs(sdk.NewTestMsg(), sdk.NewTestMsg())
	require.NotEqual(t, tx, stdTxBuilder.GetTx())
	require.Equal(t, len(stdTxBuilder.GetTx().GetMsgs()), 2)
}

func TestStdTxBuilder_SetSignatures(t *testing.T) {
	stdTxBuilder := setupStdTxBuilderTest()
	tx := stdTxBuilder.GetTx()
	singleSignatureData := signingtypes.SingleSignatureData{
		Signature: priv.PubKey().Bytes(),
		SignMode:  signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
	}

	err := stdTxBuilder.SetSignatures(signingtypes.SignatureV2{
		PubKey: priv.PubKey(),
		Data:   &singleSignatureData,
	})
	sigTx := stdTxBuilder.GetTx().(ante.SigVerifiableTx)
	require.NoError(t, err)
	require.NotEqual(t, tx, stdTxBuilder.GetTx())
	require.Equal(t, sigTx.GetSignatures()[0], priv.PubKey().Bytes())
}
