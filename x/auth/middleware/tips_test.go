package middleware_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/suite"
)

var regens = sdk.NewCoins(sdk.NewCoin("regen", sdk.NewInt(1000)))
var atoms = sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(1000)))

// setupMetaTxAccts sets up 2 accounts:
// - tipper has 1000 regens
// - feePayer has 1000 atoms and 1000 regens
func (s *MWTestSuite) setupMetaTxAccts(ctx sdk.Context) []testAccount {
	accts := s.createTestAccounts(ctx, 2, regens)
	feePayer := accts[1]
	err := s.app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, atoms)
	s.Require().NoError(err)
	err = s.app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, feePayer.acc.GetAddress(), atoms)
	s.Require().NoError(err)

	return accts
}

func (s *MWTestSuite) TestMetaTxSignModes() {
	ctx := s.SetupTest(false) // reset
	accts := s.setupMetaTxAccts(ctx)
	tipper, feePayer := accts[0], accts[1]

	testcases := []struct {
		name             string
		tipperSignMode   signing.SignMode
		feePayerSignMode signing.SignMode
		expErr           bool
	}{
		{
			"tipper=DIRECT_AUX, feePayer=DIRECT_AUX",
			signing.SignMode_SIGN_MODE_DIRECT_AUX, signing.SignMode_SIGN_MODE_DIRECT_AUX,
			true,
		},
	}

	for _, tc := range testcases {
		tc := tc
		s.Run(tc.name, func() {
			tipperTxBuilder := s.mkTipperTxBuilder(tipper.priv, regens, tc.tipperSignMode, tipper.accNum, 0, ctx.ChainID())
			feePayerTxBuilder := s.mkFeePayerTxBuilder(feePayer.priv, tc.feePayerSignMode, tx.Fee{Amount: atoms, GasLimit: 200000}, tipperTxBuilder.GetTx(), feePayer.accNum, 0, ctx.ChainID())
			_, res, err := s.app.SimDeliver(s.clientCtx.TxConfig.TxEncoder(), feePayerTxBuilder.GetTx())
			if tc.expErr {
				s.Require().NoError(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(res)
			}
		})
	}
}

func (s *MWTestSuite) TestMetaTransactions() {
	ctx := s.SetupTest(false) // reset
	accts := s.setupMetaTxAccts(ctx)
	tipper, feePayer := accts[0], accts[1]

	testcases := []struct {
		name      string
		tip       sdk.Coins
		fee       sdk.Coins
		gasLimit  uint64
		expErr    bool
		expErrStr string
	}{
		{
			name: "happy case",
			tip:  sdk.NewCoins(sdk.NewCoin("regen", sdk.NewInt(1000))),
			fee:  sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(1000))),
		},
	}

	for _, tc := range testcases {
		tc := tc
		s.Run(tc.name, func() {
			tipperTxBuilder := s.mkTipperTxBuilder(tipper.priv, tc.tip, signing.SignMode_SIGN_MODE_DIRECT_AUX, tipper.accNum, 0, ctx.ChainID())
			feePayerTxBuilder := s.mkFeePayerTxBuilder(feePayer.priv, signing.SignMode_SIGN_MODE_DIRECT, tx.Fee{Amount: tc.fee, GasLimit: tc.gasLimit}, tipperTxBuilder.GetTx(), feePayer.accNum, 0, ctx.ChainID())
			tx, _, err := s.createTestTx(feePayerTxBuilder, []cryptotypes.PrivKey{feePayer.priv}, []uint64{1}, []uint64{0}, ctx.ChainID())
			s.Require().NoError(err)
			_, res, err := s.app.SimDeliver(s.clientCtx.TxConfig.TxEncoder(), tx)
			if tc.expErr {
				s.Require().EqualError(err, tc.expErrStr)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(res)
			}
		})
	}
}

func (s *MWTestSuite) mkTipperTxBuilder(
	tipperPriv cryptotypes.PrivKey, tip sdk.Coins, signMode signing.SignMode,
	accNum, accSeq uint64, chainID string,
) client.TxBuilder {
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()
	txBuilder.SetTip(&tx.Tip{
		Amount: tip,
		Tipper: sdk.AccAddress(tipperPriv.PubKey().Address()).String(),
	})
	err := txBuilder.SetMsgs(govtypes.NewMsgVote(tipperPriv.PubKey().Address().Bytes(), 1, govtypes.OptionYes))
	s.Require().NoError(err)

	// Call SetSignatures with empty sig to populate AuthInfo.
	err = txBuilder.SetSignatures(signing.SignatureV2{
		PubKey: tipperPriv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  signMode,
			Signature: nil,
		}})
	s.Require().NoError(err)

	// Actually sign the data.
	signerData := authsigning.SignerData{
		ChainID:       chainID,
		AccountNumber: accNum,
		Sequence:      accSeq,
	}
	sigV2, err := clienttx.SignWithPrivKey(
		signing.SignMode_SIGN_MODE_DIRECT_AUX, signerData,
		txBuilder, tipperPriv, s.clientCtx.TxConfig, accSeq)
	s.Require().NoError(err)

	txBuilder.SetSignatures(sigV2)

	return txBuilder
}

func (s *MWTestSuite) mkFeePayerTxBuilder(
	feePayerPriv cryptotypes.PrivKey, signMode signing.SignMode,
	fee tx.Fee, tipTx tx.TipTx, accNum, accSeq uint64, chainID string,
) client.TxBuilder {
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()
	err := txBuilder.SetMsgs(tipTx.GetMsgs()...)
	s.Require().NoError(err)
	txBuilder.SetFeePayer(sdk.AccAddress(feePayerPriv.PubKey().Bytes()))
	txBuilder.SetFeeAmount(fee.Amount)
	txBuilder.SetGasLimit(fee.GasLimit)

	// Calling SetSignatures with empty sig to populate AuthInfo.
	tipperSigsV2, err := tipTx.(authsigning.SigVerifiableTx).GetSignaturesV2()
	s.Require().NoError(err)
	feePayerSigV2 := signing.SignatureV2{
		PubKey: feePayerPriv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  signMode,
			Signature: nil,
		}}
	sigsV2 := append(tipperSigsV2, feePayerSigV2)
	txBuilder.SetSignatures(sigsV2...)

	// Actually sign the data.
	signerData := authsigning.SignerData{
		ChainID:       chainID,
		AccountNumber: accNum,
		Sequence:      accSeq,
	}
	feePayerSigV2, err = clienttx.SignWithPrivKey(
		signMode, signerData,
		txBuilder, feePayerPriv, s.clientCtx.TxConfig, accSeq)
	s.Require().NoError(err)
	sigsV2 = append(tipperSigsV2, feePayerSigV2)
	err = txBuilder.SetSignatures(sigsV2...)
	s.Require().NoError(err)

	return txBuilder
}

func TestMWTestSuite2(t *testing.T) {
	suite.Run(t, new(MWTestSuite))
}
