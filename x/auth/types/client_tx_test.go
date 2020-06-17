package types

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/tests"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func makeCodec() *codec.Codec {
	cdc := codec.New()
	sdk.RegisterCodec(cdc)
	//types.RegisterCodec(cdc)
	cdc.RegisterConcrete(sdk.TestMsg{}, "cosmos-sdk/Test", nil)
	return cdc
}

func setupStdTxBuilderTest(t *testing.T) (client.TxBuilder, keyring.Info) {
	const fromkey = "fromkey"

	// Now add a temporary keybase
	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	kr, err := keyring.New(t.Name(), "test", dir, nil)
	require.NoError(t, err)
	path := hd.CreateHDPath(118, 0, 0).String()

	_, seed, err := kr.NewMnemonic(fromkey, keyring.English, path, hd.Secp256k1)
	require.NoError(t, err)
	require.NoError(t, kr.Delete(fromkey))

	info, err := kr.NewAccount(fromkey, seed, "", path, hd.Secp256k1)
	require.NoError(t, err)

	msgs := []sdk.Msg{sdk.NewTestMsg(info.GetAddress())}

	stdTx := StdTx{
		Memo:       "foomemo",
		Msgs:       msgs,
		Signatures: nil,
		Fee: StdFee{
			Amount: NewTestCoins(),
			Gas:    200000,
		},
	}

	return &StdTxBuilder{
		stdTx,
		makeCodec(),
	}, info
}

func TestStdTxBuilder_GetTx(t *testing.T) {
	stdTxBuilder, info := setupStdTxBuilderTest(t)
	tx := stdTxBuilder.GetTx()
	require.NotNil(t, tx)
	require.NotNil(t, tx.GetMsgs())
	require.Equal(t, tx.GetMsgs()[0].GetSigners()[0], info.GetAddress())
	require.Equal(t, len(tx.GetMsgs()), 1)
	require.Equal(t, len(tx.GetMsgs()[0].GetSigners()), 1)
}

func TestStdTxBuilder_SetFeeAmount(t *testing.T) {
	stdTxBuilder, _ := setupStdTxBuilderTest(t)
	feeAmount := sdk.Coins{
		sdk.NewInt64Coin("atom", 20000000),
	}
	tx := stdTxBuilder.GetTx()
	stdTxBuilder.SetFeeAmount(feeAmount)
	feeTx := stdTxBuilder.GetTx().(sdk.FeeTx)
	require.NotEqual(t, tx, stdTxBuilder.GetTx())
	require.Equal(t, feeTx.GetFee(), feeAmount)
	require.False(t, reflect.DeepEqual(tx, stdTxBuilder.GetTx()))
}

func TestStdTxBuilder_SetGasLimit(t *testing.T) {
	stdTxBuilder, _ := setupStdTxBuilderTest(t)
	tx := stdTxBuilder.GetTx()
	stdTxBuilder.SetGasLimit(300000)
	feeTx := stdTxBuilder.GetTx().(sdk.FeeTx)
	require.NotEqual(t, tx, stdTxBuilder.GetTx())
	require.Equal(t, feeTx.GetGas(), uint64(300000))
	require.False(t, reflect.DeepEqual(tx, stdTxBuilder.GetTx()))
}

func TestStdTxBuilder_SetMemo(t *testing.T) {
	stdTxBuilder, _ := setupStdTxBuilderTest(t)
	tx := stdTxBuilder.GetTx()
	stdTxBuilder.SetMemo("newfoomemo")
	txWithMemo := stdTxBuilder.GetTx().(sdk.TxWithMemo)
	require.NotEqual(t, tx, stdTxBuilder.GetTx())
	require.Equal(t, txWithMemo.GetMemo(), "newfoomemo")
	require.False(t, reflect.DeepEqual(tx, stdTxBuilder.GetTx()))
}

func TestStdTxBuilder_SetMsgs(t *testing.T) {
	stdTxBuilder, _ := setupStdTxBuilderTest(t)
	tx := stdTxBuilder.GetTx()
	stdTxBuilder.SetMsgs(NewTestMsg(), NewTestMsg())
	require.NotEqual(t, tx, stdTxBuilder.GetTx())
	require.False(t, reflect.DeepEqual(tx, stdTxBuilder.GetTx()))
	require.Equal(t, len(stdTxBuilder.GetTx().GetMsgs()), 2)
}

func TestStdTxBuilder_SetSignatures(t *testing.T) {
	stdTxBuilder, info := setupStdTxBuilderTest(t)
	tx := stdTxBuilder.GetTx()
	require.Equal(t, tx.GetMsgs()[0].GetSigners()[0], info.GetAddress())

	singleSignatureData := signingtypes.SingleSignatureData{
		Signature: priv.PubKey().Bytes(),
		SignMode:  signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
	}

	err := stdTxBuilder.SetSignatures(signingtypes.SignatureV2{
		PubKey: priv.PubKey(),
		Data:   &singleSignatureData,
	})
	//sigTx := stdTxBuilder.GetTx().(ante.SigVerifiableTx)

	require.Equal(t, 1, len(stdTxBuilder.GetTx().GetMsgs()[0].GetSigners()))
	require.NoError(t, err)
	require.NotEqual(t, tx, stdTxBuilder.GetTx())
	//require.NotEqual(t, sigTx.GetSignatures()[0], priv.PubKey().Bytes())
	require.False(t, reflect.DeepEqual(tx, stdTxBuilder.GetTx()))
}
