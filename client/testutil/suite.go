package testutil

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/codec/testdata"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto"

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
	txBuilder := s.TxGenerator.NewTxBuilder()
	tx := txBuilder.GetTx()
	s.Require().NotNil(tx)
	s.Require().Equal(len(tx.GetMsgs()), 0)
}

func (s *TxGeneratorTestSuite) TestTxBuilderSetFeeAmount() {
	txBuilder := s.TxGenerator.NewTxBuilder()
	feeAmount := sdk.Coins{
		sdk.NewInt64Coin("atom", 20000000),
	}
	txBuilder.SetFeeAmount(feeAmount)
	feeTx := txBuilder.GetTx()
	s.Require().Equal(feeAmount, feeTx.GetFee())
}

func (s *TxGeneratorTestSuite) TestTxBuilderSetGasLimit() {
	const newGas uint64 = 300000
	txBuilder := s.TxGenerator.NewTxBuilder()
	txBuilder.SetGasLimit(newGas)
	feeTx := txBuilder.GetTx()
	s.Require().Equal(newGas, feeTx.GetGas())
}

func (s *TxGeneratorTestSuite) TestTxBuilderSetMemo() {
	const newMemo string = "newfoomemo"
	txBuilder := s.TxGenerator.NewTxBuilder()
	txBuilder.SetMemo(newMemo)
	txWithMemo := txBuilder.GetTx()
	s.Require().Equal(txWithMemo.GetMemo(), newMemo)
}

func (s *TxGeneratorTestSuite) TestTxBuilderSetMsgs() {
	_, _, addr1 := authtypes.KeyTestPubAddr()
	_, _, addr2 := authtypes.KeyTestPubAddr()
	msg1 := testdata.NewTestMsg(addr1)
	msg2 := testdata.NewTestMsg(addr2)
	msgs := []sdk.Msg{msg1, msg2}

	txBuilder := s.TxGenerator.NewTxBuilder()

	err := txBuilder.SetMsgs(msgs...)
	s.Require().NoError(err)
	tx := txBuilder.GetTx()
	s.Require().Equal(msgs, tx.GetMsgs())
	s.Require().Equal([]sdk.AccAddress{addr1, addr2}, tx.GetSigners())
	s.Require().Equal(addr1, tx.FeePayer())
	s.Require().Error(tx.ValidateBasic()) // should fail because of no signatures
}

func (s *TxGeneratorTestSuite) TestTxBuilderSetSignatures() {
	privKey, pubkey, addr := authtypes.KeyTestPubAddr()
	privKey2, pubkey2, _ := authtypes.KeyTestPubAddr()
	multisigPk := multisig.NewPubKeyMultisigThreshold(2, []crypto.PubKey{pubkey, pubkey2})

	txBuilder := s.TxGenerator.NewTxBuilder()

	// set test msg
	msg := testdata.NewTestMsg(addr)
	msigAddr := sdk.AccAddress(multisigPk.Address())
	msg2 := testdata.NewTestMsg(msigAddr)
	err := txBuilder.SetMsgs(msg, msg2)
	s.Require().NoError(err)

	// check that validation fails
	s.Require().Error(txBuilder.GetTx().ValidateBasic())

	signModeHandler := s.TxGenerator.SignModeHandler()
	s.Require().Contains(signModeHandler.Modes(), signModeHandler.DefaultMode())

	// set SignatureV2 without actual signature bytes
	sigData1 := &signingtypes.SingleSignatureData{SignMode: signModeHandler.DefaultMode()}
	sig1 := signingtypes.SignatureV2{PubKey: pubkey, Data: sigData1}

	msigData := multisig.NewMultisig(2)
	multisig.AddSignature(msigData, &signingtypes.SingleSignatureData{SignMode: signModeHandler.DefaultMode()}, 0)
	multisig.AddSignature(msigData, &signingtypes.SingleSignatureData{SignMode: signModeHandler.DefaultMode()}, 1)
	msig := signingtypes.SignatureV2{PubKey: multisigPk, Data: msigData}

	// fail validation without required signers
	err = txBuilder.SetSignatures(sig1)
	s.Require().NoError(err)
	sigTx := txBuilder.GetTx()
	s.Require().Error(sigTx.ValidateBasic())

	err = txBuilder.SetSignatures(sig1, msig)
	s.Require().NoError(err)
	sigTx = txBuilder.GetTx()
	s.Require().Len(sigTx.GetSignatures(), 2)
	sigsV2, err := sigTx.GetSignaturesV2()
	s.Require().NoError(err)
	s.Require().Len(sigsV2, 2)
	s.Require().True(sigEquals(sig1, sigsV2[0]))
	s.Require().True(sigEquals(msig, sigsV2[1]))
	s.Require().Equal([]sdk.AccAddress{addr, msigAddr}, sigTx.GetSigners())
	s.Require().NoError(sigTx.ValidateBasic())

	// sign transaction
	signerData := signing.SignerData{
		ChainID:         "test",
		AccountNumber:   1,
		AccountSequence: 2,
	}
	signBytes, err := signModeHandler.GetSignBytes(signModeHandler.DefaultMode(), signerData, sigTx)
	s.Require().NoError(err)
	sigBz, err := privKey.Sign(signBytes)
	s.Require().NoError(err)

	signerData = signing.SignerData{
		ChainID:         "test",
		AccountNumber:   3,
		AccountSequence: 4,
	}
	mSignBytes, err := signModeHandler.GetSignBytes(signModeHandler.DefaultMode(), signerData, sigTx)
	s.Require().NoError(err)
	mSigBz1, err := privKey.Sign(mSignBytes)
	s.Require().NoError(err)
	mSigBz2, err := privKey2.Sign(mSignBytes)
	s.Require().NoError(err)
	msigData = multisig.NewMultisig(2)
	multisig.AddSignature(msigData, &signingtypes.SingleSignatureData{
		SignMode: signModeHandler.DefaultMode(), Signature: mSigBz1}, 0)
	multisig.AddSignature(msigData, &signingtypes.SingleSignatureData{
		SignMode: signModeHandler.DefaultMode(), Signature: mSigBz2}, 0)

	// set signature
	sigData1.Signature = sigBz
	sig1 = signingtypes.SignatureV2{PubKey: pubkey, Data: sigData1}
	msig = signingtypes.SignatureV2{PubKey: multisigPk, Data: msigData}
	err = txBuilder.SetSignatures(sig1, msig)
	s.Require().NoError(err)
	sigTx = txBuilder.GetTx()
	s.Require().Len(sigTx.GetSignatures(), 2)
	sigsV2, err = sigTx.GetSignaturesV2()
	s.Require().NoError(err)
	s.Require().Len(sigsV2, 2)
	s.Require().True(sigEquals(sig1, sigsV2[0]))
	s.Require().True(sigEquals(msig, sigsV2[1]))
	s.Require().Equal([]sdk.AccAddress{addr, msigAddr}, sigTx.GetSigners())
	s.Require().NoError(sigTx.ValidateBasic())
}

func sigEquals(sig1, sig2 signingtypes.SignatureV2) bool {
	if !sig1.PubKey.Equals(sig2.PubKey) {
		return false
	}

	if sig1.Data == nil && sig2.Data == nil {
		return true
	}

	return sigDataEquals(sig1.Data, sig2.Data)
}

func sigDataEquals(data1, data2 signingtypes.SignatureData) bool {
	switch data1 := data1.(type) {
	case *signingtypes.SingleSignatureData:
		data2, ok := data2.(*signingtypes.SingleSignatureData)
		if !ok {
			return false
		}

		if data1.SignMode != data2.SignMode {
			return false
		}

		return bytes.Equal(data1.Signature, data2.Signature)
	case *signingtypes.MultiSignatureData:
		data2, ok := data2.(*signingtypes.MultiSignatureData)
		if !ok {
			return false
		}

		if data1.BitArray.ExtraBitsStored != data2.BitArray.ExtraBitsStored {
			return false
		}

		if !bytes.Equal(data1.BitArray.Elems, data2.BitArray.Elems) {
			return false
		}

		if len(data1.Signatures) != len(data2.Signatures) {
			return false
		}

		for i, s := range data1.Signatures {
			if !sigDataEquals(s, data2.Signatures[i]) {
				return false
			}
		}

		return true
	default:
		return false
	}
}

func (s *TxGeneratorTestSuite) TestTxEncodeDecode() {
	_, pubkey, addr := authtypes.KeyTestPubAddr()
	feeAmount := sdk.Coins{sdk.NewInt64Coin("atom", 150)}
	gasLimit := uint64(50000)
	memo := "foomemo"
	msg := testdata.NewTestMsg(addr)
	dummySig := []byte("dummySig")
	sig := signingtypes.SignatureV2{
		PubKey: pubkey,
		Data: &signingtypes.SingleSignatureData{
			SignMode:  signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
			Signature: dummySig,
		},
	}

	txBuilder := s.TxGenerator.NewTxBuilder()
	txBuilder.SetFeeAmount(feeAmount)
	txBuilder.SetGasLimit(gasLimit)
	txBuilder.SetMemo(memo)
	err := txBuilder.SetMsgs(msg)
	s.Require().NoError(err)
	err = txBuilder.SetSignatures(sig)
	s.Require().NoError(err)
	tx := txBuilder.GetTx()

	s.T().Log("encode transaction")
	txBytes, err := s.TxGenerator.TxEncoder()(tx)
	s.Require().NoError(err)
	s.Require().NotNil(txBytes)

	s.T().Log("decode transaction")
	tx2, err := s.TxGenerator.TxDecoder()(txBytes)
	s.Require().NoError(err)
	tx3, ok := tx2.(signing.SigFeeMemoTx)
	s.Require().True(ok)
	s.Require().Equal([]sdk.Msg{msg}, tx3.GetMsgs())
	s.Require().Equal(feeAmount, tx3.GetFee())
	s.Require().Equal(gasLimit, tx3.GetGas())
	s.Require().Equal(memo, tx3.GetMemo())
	s.Require().Equal([][]byte{dummySig}, tx3.GetSignatures())
	s.Require().Equal([]crypto.PubKey{pubkey}, tx3.GetPubKeys())

	s.T().Log("JSON encode transaction")
	jsonTxBytes, err := s.TxGenerator.TxJSONEncoder()(tx)
	s.Require().NoError(err)
	s.Require().NotNil(jsonTxBytes)

	s.T().Log("JSON decode transaction")
	tx2, err = s.TxGenerator.TxJSONDecoder()(jsonTxBytes)
	s.Require().NoError(err)
	tx3, ok = tx2.(signing.SigFeeMemoTx)
	s.Require().True(ok)
	s.Require().Equal([]sdk.Msg{msg}, tx3.GetMsgs())
	s.Require().Equal(feeAmount, tx3.GetFee())
	s.Require().Equal(gasLimit, tx3.GetGas())
	s.Require().Equal(memo, tx3.GetMemo())
	s.Require().Equal([][]byte{dummySig}, tx3.GetSignatures())
	s.Require().Equal([]crypto.PubKey{pubkey}, tx3.GetPubKeys())
}
