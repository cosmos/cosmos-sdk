package tx_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/x/auth/signing"

	"github.com/stretchr/testify/require"

	tx2 "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	types3 "github.com/cosmos/cosmos-sdk/x/auth/types"

	signing2 "github.com/cosmos/cosmos-sdk/types/tx/signing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/cosmos/cosmos-sdk/types"
	types2 "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	memo = "waboom"
	gas  = uint64(10000)
)

var (
	fee            = types.NewCoins(types.NewInt64Coin("bam", 100))
	_, pub1, addr1 = testdata.KeyTestPubAddr()
	_, _, addr2    = testdata.KeyTestPubAddr()
	msg            = types2.NewMsgSend(addr1, addr2, types.NewCoins(types.NewInt64Coin("wack", 10000)))
	sig            = signing2.SignatureV2{
		PubKey: pub1,
		Data: &signing2.SingleSignatureData{
			SignMode:  signing2.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
			Signature: []byte("dummy"),
		},
	}
)

func buildTestTx(t *testing.T, builder client.TxBuilder) {
	builder.SetMemo(memo)
	builder.SetGasLimit(gas)
	builder.SetFeeAmount(fee)
	err := builder.SetMsgs(msg)
	require.NoError(t, err)
	err = builder.SetSignatures(sig)
	require.NoError(t, err)
}

func TestCopyTx(t *testing.T) {
	encCfg := simapp.MakeEncodingConfig()
	protoCfg := tx.NewTxConfig(codec.NewProtoCodec(encCfg.InterfaceRegistry), std.DefaultPublicKeyCodec{}, tx.DefaultSignModeHandler())
	aminoCfg := types3.StdTxConfig{Cdc: encCfg.Amino}

	// proto -> amino -> proto
	protoBuilder := protoCfg.NewTxBuilder()
	buildTestTx(t, protoBuilder)
	aminoBuilder := aminoCfg.NewTxBuilder()
	err := tx2.CopyTx(protoBuilder.GetTx(), aminoBuilder)
	require.NoError(t, err)
	protoBuilder2 := protoCfg.NewTxBuilder()
	err = tx2.CopyTx(aminoBuilder.GetTx(), protoBuilder2)
	require.NoError(t, err)
	bz, err := protoCfg.TxEncoder()(protoBuilder.GetTx())
	require.NoError(t, err)
	bz2, err := protoCfg.TxEncoder()(protoBuilder2.GetTx())
	require.NoError(t, err)
	require.Equal(t, bz, bz2)

	// amino -> proto -> amino
	aminoBuilder = aminoCfg.NewTxBuilder()
	buildTestTx(t, aminoBuilder)
	protoBuilder = protoCfg.NewTxBuilder()
	err = tx2.CopyTx(aminoBuilder.GetTx(), protoBuilder)
	require.NoError(t, err)
	aminoBuilder2 := aminoCfg.NewTxBuilder()
	err = tx2.CopyTx(protoBuilder.GetTx(), aminoBuilder2)
	require.NoError(t, err)
	bz, err = aminoCfg.TxEncoder()(aminoBuilder.GetTx())
	require.NoError(t, err)
	bz2, err = aminoCfg.TxEncoder()(aminoBuilder2.GetTx())
	require.NoError(t, err)
	require.Equal(t, bz, bz2)
}

func TestConvertTxToStdTx(t *testing.T) {
	encCfg := simapp.MakeEncodingConfig()
	protoCfg := tx.NewTxConfig(codec.NewProtoCodec(encCfg.InterfaceRegistry), std.DefaultPublicKeyCodec{}, tx.DefaultSignModeHandler())

	protoBuilder := protoCfg.NewTxBuilder()
	buildTestTx(t, protoBuilder)
	stdTx, err := tx2.ConvertTxToStdTx(encCfg.Amino, protoBuilder.GetTx())
	require.NoError(t, err)
	require.Equal(t, memo, stdTx.Memo)
	require.Equal(t, gas, stdTx.Fee.Gas)
	require.Equal(t, fee, stdTx.Fee.Amount)
	require.Equal(t, msg, stdTx.Msgs[0])
	require.Equal(t, sig.PubKey.Bytes(), stdTx.Signatures[0].PubKey)
	require.Equal(t, sig.Data.(*signing2.SingleSignatureData).Signature, stdTx.Signatures[0].Signature)
}

func TestConvertAndEncodeStdTx(t *testing.T) {
	encCfg := simapp.MakeEncodingConfig()
	protoCfg := tx.NewTxConfig(codec.NewProtoCodec(encCfg.InterfaceRegistry), std.DefaultPublicKeyCodec{}, tx.DefaultSignModeHandler())
	aminoCfg := types3.StdTxConfig{Cdc: encCfg.Amino}

	// convert amino -> proto -> amino
	aminoBuilder := aminoCfg.NewTxBuilder()
	buildTestTx(t, aminoBuilder)
	stdTx := aminoBuilder.GetTx().(types3.StdTx)
	txBz, err := tx2.ConvertAndEncodeStdTx(protoCfg, stdTx)
	require.NoError(t, err)
	decodedTx, err := protoCfg.TxDecoder()(txBz)
	require.NoError(t, err)
	aminoBuilder2 := aminoCfg.NewTxBuilder()
	require.NoError(t, tx2.CopyTx(decodedTx.(signing.SigFeeMemoTx), aminoBuilder2))
	require.Equal(t, stdTx, aminoBuilder2.GetTx())

	// just use amino everywhere
	txBz, err = tx2.ConvertAndEncodeStdTx(aminoCfg, stdTx)
	require.NoError(t, err)
	decodedTx, err = aminoCfg.TxDecoder()(txBz)
	require.NoError(t, err)
	require.Equal(t, stdTx, decodedTx)
}
