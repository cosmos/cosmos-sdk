package testutil

import (
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/client"

	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TxGeneratorTestSuite provides a test suite that can be used to test that a TxGenerator implementation is correct
//nolint:golint  // type name will be used as tx.TxGeneratorTestSuite by other packages, and that stutters; consider calling this GeneratorTestSuite
type TxGeneratorTestSuite struct {
	suite.Suite
	TxGenerator client.TxGenerator
}

// NewTxGeneratorTestSuite returns a new TxGeneratorTestSuite with the provided TxGenerator implementation
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
	txBuilder := s.TxGenerator.NewTxBuilder()
	txBuilder.SetFeeAmount(sdk.Coins{sdk.NewInt64Coin("atom", 150)})
	txBuilder.SetGasLimit(50000)
	txBuilder.SetMemo("foomemo")
	tx := txBuilder.GetTx()

	// Encode transaction
	txBytes, err := s.TxGenerator.TxEncoder()(tx)
	s.Require().NoError(err)
	s.Require().NotNil(txBytes)

	tx2, err := s.TxGenerator.TxDecoder()(txBytes)
	s.Require().NoError(err)
	s.Require().Equal(tx, tx2)
}
