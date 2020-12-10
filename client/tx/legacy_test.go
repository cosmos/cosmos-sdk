package tx_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	tx2 "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/cosmos/cosmos-sdk/types"
	signing2 "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	memo          = "waboom"
	gas           = uint64(10000)
	timeoutHeight = 5
)

var (
	fee            = types.NewCoins(types.NewInt64Coin("bam", 100))
	_, pub1, addr1 = testdata.KeyTestPubAddr()
	_, _, addr2    = testdata.KeyTestPubAddr()
	msg            = banktypes.NewMsgSend(addr1, addr2, types.NewCoins(types.NewInt64Coin("wack", 10000)))
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
	builder.SetTimeoutHeight(timeoutHeight)
}

type TestSuite struct {
	suite.Suite
	encCfg   params.EncodingConfig
	protoCfg client.TxConfig
	aminoCfg client.TxConfig
}

func (s *TestSuite) SetupSuite() {
	encCfg := simapp.MakeTestEncodingConfig()
	s.encCfg = encCfg
	s.protoCfg = tx.NewTxConfig(codec.NewProtoCodec(encCfg.InterfaceRegistry), tx.DefaultSignModes)
	s.aminoCfg = legacytx.StdTxConfig{Cdc: encCfg.Amino}
}

func (s *TestSuite) TestCopyTx() {
	// proto -> amino -> proto
	protoBuilder := s.protoCfg.NewTxBuilder()
	buildTestTx(s.T(), protoBuilder)
	aminoBuilder := s.aminoCfg.NewTxBuilder()
	err := tx2.CopyTx(protoBuilder.GetTx(), aminoBuilder, false)
	s.Require().NoError(err)
	protoBuilder2 := s.protoCfg.NewTxBuilder()
	err = tx2.CopyTx(aminoBuilder.GetTx(), protoBuilder2, false)
	s.Require().NoError(err)
	bz, err := s.protoCfg.TxEncoder()(protoBuilder.GetTx())
	s.Require().NoError(err)
	bz2, err := s.protoCfg.TxEncoder()(protoBuilder2.GetTx())
	s.Require().NoError(err)
	s.Require().Equal(bz, bz2)

	// amino -> proto -> amino
	aminoBuilder = s.aminoCfg.NewTxBuilder()
	buildTestTx(s.T(), aminoBuilder)
	protoBuilder = s.protoCfg.NewTxBuilder()
	err = tx2.CopyTx(aminoBuilder.GetTx(), protoBuilder, false)
	s.Require().NoError(err)
	aminoBuilder2 := s.aminoCfg.NewTxBuilder()
	err = tx2.CopyTx(protoBuilder.GetTx(), aminoBuilder2, false)
	s.Require().NoError(err)
	bz, err = s.aminoCfg.TxEncoder()(aminoBuilder.GetTx())
	s.Require().NoError(err)
	bz2, err = s.aminoCfg.TxEncoder()(aminoBuilder2.GetTx())
	s.Require().NoError(err)
	s.Require().Equal(bz, bz2)
}

func (s *TestSuite) TestConvertTxToStdTx() {
	// proto tx
	protoBuilder := s.protoCfg.NewTxBuilder()
	buildTestTx(s.T(), protoBuilder)
	stdTx, err := tx2.ConvertTxToStdTx(s.encCfg.Amino, protoBuilder.GetTx())
	s.Require().NoError(err)
	s.Require().Equal(memo, stdTx.Memo)
	s.Require().Equal(gas, stdTx.Fee.Gas)
	s.Require().Equal(fee, stdTx.Fee.Amount)
	s.Require().Equal(msg, stdTx.Msgs[0])
	s.Require().Equal(timeoutHeight, stdTx.TimeoutHeight)
	s.Require().Equal(sig.PubKey, stdTx.Signatures[0].PubKey)
	s.Require().Equal(sig.Data.(*signing2.SingleSignatureData).Signature, stdTx.Signatures[0].Signature)

	// SIGN_MODE_DIRECT should fall back to an unsigned tx
	err = protoBuilder.SetSignatures(signing2.SignatureV2{
		PubKey: pub1,
		Data: &signing2.SingleSignatureData{
			SignMode:  signing2.SignMode_SIGN_MODE_DIRECT,
			Signature: []byte("dummy"),
		},
	})
	s.Require().NoError(err)
	stdTx, err = tx2.ConvertTxToStdTx(s.encCfg.Amino, protoBuilder.GetTx())
	s.Require().NoError(err)
	s.Require().Equal(memo, stdTx.Memo)
	s.Require().Equal(gas, stdTx.Fee.Gas)
	s.Require().Equal(fee, stdTx.Fee.Amount)
	s.Require().Equal(msg, stdTx.Msgs[0])
	s.Require().Equal(timeoutHeight, stdTx.TimeoutHeight)
	s.Require().Empty(stdTx.Signatures)

	// std tx
	aminoBuilder := s.aminoCfg.NewTxBuilder()
	buildTestTx(s.T(), aminoBuilder)
	stdTx = aminoBuilder.GetTx().(legacytx.StdTx)
	stdTx2, err := tx2.ConvertTxToStdTx(s.encCfg.Amino, stdTx)
	s.Require().NoError(err)
	s.Require().Equal(stdTx, stdTx2)
}

func (s *TestSuite) TestConvertAndEncodeStdTx() {
	// convert amino -> proto -> amino
	aminoBuilder := s.aminoCfg.NewTxBuilder()
	buildTestTx(s.T(), aminoBuilder)
	stdTx := aminoBuilder.GetTx().(legacytx.StdTx)
	txBz, err := tx2.ConvertAndEncodeStdTx(s.protoCfg, stdTx)
	s.Require().NoError(err)
	decodedTx, err := s.protoCfg.TxDecoder()(txBz)
	s.Require().NoError(err)
	aminoBuilder2 := s.aminoCfg.NewTxBuilder()
	s.Require().NoError(tx2.CopyTx(decodedTx.(signing.Tx), aminoBuilder2, false))
	s.Require().Equal(stdTx, aminoBuilder2.GetTx())

	// just use amino everywhere
	txBz, err = tx2.ConvertAndEncodeStdTx(s.aminoCfg, stdTx)
	s.Require().NoError(err)
	decodedTx, err = s.aminoCfg.TxDecoder()(txBz)
	s.Require().NoError(err)
	s.Require().Equal(stdTx, decodedTx)
}
