package tx

import (
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/cosmos-sdk/client"

	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type TxGeneratorTestSuite struct {
	suite.Suite
	TxGenerator client.TxGenerator
}

func NewTxGeneratorTestSuite(txGenerator client.TxGenerator) *TxGeneratorTestSuite {
	return &TxGeneratorTestSuite{TxGenerator: txGenerator}
}

func (s *TxGeneratorTestSuite) TestTxBuilderGetTx() {
	stdTxBuilder := s.TxGenerator.NewTxBuilder()
	tx := stdTxBuilder.GetTx()
	s.Require().NotNil(tx)
	s.Require().Equal(len(tx.GetMsgs()), 0)
}

func (s *TxGeneratorTestSuite) TestTxBuilderSetFeeAmount() {
	stdTxBuilder := s.TxGenerator.NewTxBuilder()
	feeAmount := sdk.Coins{
		sdk.NewInt64Coin("atom", 20000000),
	}
	stdTxBuilder.SetFeeAmount(feeAmount)
	feeTx := stdTxBuilder.GetTx().(sdk.FeeTx)
	s.Require().Equal(feeAmount, feeTx.GetFee())
}

func (s *TxGeneratorTestSuite) TestTxBuilderSetGasLimit() {
	const newGas uint64 = 300000
	stdTxBuilder := s.TxGenerator.NewTxBuilder()
	stdTxBuilder.SetGasLimit(newGas)
	feeTx := stdTxBuilder.GetTx().(sdk.FeeTx)
	s.Require().Equal(newGas, feeTx.GetGas())
}

func (s *TxGeneratorTestSuite) TestTxBuilderSetMemo() {
	const newMemo string = "newfoomemo"
	stdTxBuilder := s.TxGenerator.NewTxBuilder()
	stdTxBuilder.SetMemo(newMemo)
	txWithMemo := stdTxBuilder.GetTx().(sdk.TxWithMemo)
	s.Require().Equal(txWithMemo.GetMemo(), newMemo)
}

func (s *TxGeneratorTestSuite) TestTxBuilderSetMsgs() {
	stdTxBuilder := s.TxGenerator.NewTxBuilder()
	tx := stdTxBuilder.GetTx()
	err := stdTxBuilder.SetMsgs(sdk.NewTestMsg(), sdk.NewTestMsg())
	s.Require().NoError(err)
	s.Require().NotEqual(tx, stdTxBuilder.GetTx())
	s.Require().Equal(len(stdTxBuilder.GetTx().GetMsgs()), 2)
}

type HasSignaturesTx interface {
	GetSignatures() [][]byte
	GetSigners() []sdk.AccAddress
	GetPubKeys() []crypto.PubKey // If signer already has pubkey in context, this list will have nil in its place
}

func (s *TxGeneratorTestSuite) TestTxBuilderSetSignatures() {
	priv := secp256k1.GenPrivKey()

	stdTxBuilder := s.TxGenerator.NewTxBuilder()
	tx := stdTxBuilder.GetTx()
	singleSignatureData := signingtypes.SingleSignatureData{
		Signature: priv.PubKey().Bytes(),
		SignMode:  signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
	}

	err := stdTxBuilder.SetSignatures(signingtypes.SignatureV2{
		PubKey: priv.PubKey(),
		Data:   &singleSignatureData,
	})
	sigTx := stdTxBuilder.GetTx().(HasSignaturesTx)
	s.Require().NoError(err)
	s.Require().NotEqual(tx, stdTxBuilder.GetTx())
	s.Require().Equal(sigTx.GetSignatures()[0], priv.PubKey().Bytes())
}

func (s *TxGeneratorTestSuite) TestTxEncodeDecode() {
	// Build a test transaction
	fee := types.NewStdFee(50000, sdk.Coins{sdk.NewInt64Coin("atom", 150)})
	stdTx := types.NewStdTx([]sdk.Msg{nil}, fee, []types.StdSignature{}, "foomemo")

	// Encode transaction
	txBytes, err := s.TxGenerator.TxEncoder()(stdTx)
	s.Require().NoError(err)
	s.Require().NotNil(txBytes)

	tx, err := s.TxGenerator.TxDecoder()(txBytes)
	s.Require().NoError(err)
	s.Require().Equal([]sdk.Msg{nil}, tx.GetMsgs())
}
