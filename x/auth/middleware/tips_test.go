package middleware_test

import (
	"fmt"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/middleware"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

var initialRegens = sdk.NewCoins(sdk.NewCoin("regen", sdk.NewInt(1000)))
var initialAtoms = sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(1000)))

// setupMetaTxAccts sets up 2 accounts:
// - tipper has 1000 regens
// - feePayer has 1000 atoms and 1000 regens
func (s *MWTestSuite) setupMetaTxAccts(ctx sdk.Context) (sdk.Context, []testAccount) {
	accts := s.createTestAccounts(ctx, 2, initialRegens)
	feePayer := accts[1]
	err := s.app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initialAtoms)
	s.Require().NoError(err)
	err = s.app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, feePayer.acc.GetAddress(), initialAtoms)
	s.Require().NoError(err)

	// Create dummy proposal for tipper to vote on.
	prop, err := govtypes.NewProposal(govtypes.NewTextProposal("foo", "bar"), 1, time.Now(), time.Now().Add(time.Hour))
	s.Require().NoError(err)
	s.app.GovKeeper.SetProposal(ctx, prop)
	s.app.GovKeeper.ActivateVotingPeriod(ctx, prop)

	// Move to next block to commit previous data to state.
	s.app.EndBlock(abci.RequestEndBlock{Height: ctx.BlockHeight()})
	s.app.Commit()

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: ctx.BlockHeight()}})

	return ctx, accts
}

func (s *MWTestSuite) TestSignModes() {
	ctx := s.SetupTest(false) // reset
	ctx, accts := s.setupMetaTxAccts(ctx)
	tipper, feePayer := accts[0], accts[1]
	msg := govtypes.NewMsgVote(tipper.acc.GetAddress(), 1, govtypes.OptionYes)

	txHandler := middleware.ComposeMiddlewares(noopTxHandler{}, middleware.SignModeTxMiddleware)

	testcases := []struct {
		tipperSignMode   signing.SignMode
		feePayerSignMode signing.SignMode
		expErr           bool
	}{
		{signing.SignMode_SIGN_MODE_DIRECT, signing.SignMode_SIGN_MODE_DIRECT, false},
		{signing.SignMode_SIGN_MODE_DIRECT, signing.SignMode_SIGN_MODE_DIRECT_AUX, true},
		{signing.SignMode_SIGN_MODE_DIRECT, signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, false},
		{signing.SignMode_SIGN_MODE_DIRECT_AUX, signing.SignMode_SIGN_MODE_DIRECT, false},
		{signing.SignMode_SIGN_MODE_DIRECT_AUX, signing.SignMode_SIGN_MODE_DIRECT_AUX, true},
		{signing.SignMode_SIGN_MODE_DIRECT_AUX, signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, false},
		{signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, signing.SignMode_SIGN_MODE_DIRECT, false},
		{signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, signing.SignMode_SIGN_MODE_DIRECT_AUX, true},
		{signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, false},
	}

	for _, tc := range testcases {
		tc := tc
		s.Run(fmt.Sprintf("tipper=%s, feepayer=%s", tc.tipperSignMode, tc.feePayerSignMode), func() {
			tipperTxBuilder := s.mkTipperTxBuilder(tipper.priv, msg, initialRegens, tc.tipperSignMode, tipper.accNum, 0, ctx.ChainID())
			feePayerTxBuilder, err := mkFeePayerTxBuilder(s.clientCtx, feePayer.priv, tc.feePayerSignMode, tx.Fee{Amount: initialAtoms, GasLimit: 200000}, tipperTxBuilder.GetTx(), feePayer.accNum, 0, ctx.ChainID())
			s.Require().NoError(err)

			_, err = txHandler.DeliverTx(sdk.WrapSDKContext(ctx), feePayerTxBuilder.GetTx(), abci.RequestDeliverTx{})
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), "invalid sign mode for")
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *MWTestSuite) TestTips() {
	var msg sdk.Msg
	_, _, randomAddr := testdata.KeyTestPubAddr()

	testcases := []struct {
		name      string
		msgSigner sdk.AccAddress
		tip       sdk.Coins
		fee       sdk.Coins
		gasLimit  uint64
		expErr    bool
		expErrStr string
	}{
		{
			"tipper should be equal to msg signer",
			randomAddr, // arbitrary msg signer, not equal to tipper
			sdk.NewCoins(sdk.NewCoin("regen", sdk.NewInt(5000))), initialAtoms, 200000,
			true, "pubKey does not match signer address",
		},
		{
			"wrong tip denom", nil,
			sdk.NewCoins(sdk.NewCoin("foobar", sdk.NewInt(1000))), initialAtoms, 200000,
			true, "0foobar is smaller than 1000foobar: insufficient funds",
		},
		{
			"insufficient tip from tipper", nil,
			sdk.NewCoins(sdk.NewCoin("regen", sdk.NewInt(5000))), initialAtoms, 200000,
			true, "1000regen is smaller than 5000regen: insufficient funds",
		},
		{
			"insufficient fees from feePayer", nil,
			initialRegens, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(5000))), 200000,
			true, "1000atom is smaller than 5000atom: insufficient funds: insufficient funds",
		},
		{
			"insufficient gas", nil,
			initialRegens, initialAtoms, 100,
			true, "out of gas in location: ReadFlat; gasWanted: 100, gasUsed: 1000: out of gas",
		},
		{
			"happy case", nil,
			initialRegens, initialAtoms, 200000,
			false, "",
		},
	}

	for _, tc := range testcases {
		tc := tc
		s.Run(tc.name, func() {
			ctx := s.SetupTest(false) // reset
			ctx, accts := s.setupMetaTxAccts(ctx)
			tipper, feePayer := accts[0], accts[1]

			voter := tc.msgSigner
			if voter == nil {
				voter = tipper.acc.GetAddress() // Choose tipper as MsgSigner, unless overwritten by testcase.
			}
			msg = govtypes.NewMsgVote(voter, 1, govtypes.OptionYes)

			tipperTxBuilder := s.mkTipperTxBuilder(tipper.priv, msg, tc.tip, signing.SignMode_SIGN_MODE_DIRECT_AUX, tipper.accNum, 0, ctx.ChainID())
			feePayerTxBuilder, err := mkFeePayerTxBuilder(s.clientCtx, feePayer.priv, signing.SignMode_SIGN_MODE_DIRECT, tx.Fee{Amount: tc.fee, GasLimit: tc.gasLimit}, tipperTxBuilder.GetTx(), feePayer.accNum, 0, ctx.ChainID())
			s.Require().NoError(err)

			_, res, err := s.app.SimDeliver(s.clientCtx.TxConfig.TxEncoder(), feePayerTxBuilder.GetTx())

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrStr)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(res)

				// Move to next block to commit previous data to state.
				s.app.EndBlock(abci.RequestEndBlock{Height: ctx.BlockHeight()})
				s.app.Commit()

				ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
				s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: ctx.BlockHeight()}})

				// Make sure tip is correctly transferred to feepayer, and fee is paid.
				expTipperRegens := initialRegens.Sub(tc.tip)
				expFeePayerRegens := initialRegens.Add(tc.tip...)
				expFeePayerAtoms := initialAtoms.Sub(tc.fee)
				s.Require().True(expTipperRegens.AmountOf("regen").Equal(s.app.BankKeeper.GetBalance(ctx, tipper.acc.GetAddress(), "regen").Amount))
				s.Require().True(expFeePayerRegens.AmountOf("regen").Equal(s.app.BankKeeper.GetBalance(ctx, feePayer.acc.GetAddress(), "regen").Amount))
				s.Require().True(expFeePayerAtoms.AmountOf("atom").Equal(s.app.BankKeeper.GetBalance(ctx, feePayer.acc.GetAddress(), "atom").Amount))
				// Make sure MsgVote has been submitted by tipper.
				votes := s.app.GovKeeper.GetAllVotes(ctx)
				s.Require().Len(votes, 1)
				s.Require().Equal(tipper.acc.GetAddress().String(), votes[0].Voter)
			}
		})
	}
}

func (s *MWTestSuite) mkTipperTxBuilder(
	tipperPriv cryptotypes.PrivKey, msg sdk.Msg, tip sdk.Coins,
	signMode signing.SignMode, accNum, accSeq uint64, chainID string,
) client.TxBuilder {
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()
	txBuilder.SetTip(&tx.Tip{
		Amount: tip,
		Tipper: sdk.AccAddress(tipperPriv.PubKey().Address()).String(),
	})
	err := txBuilder.SetMsgs(msg)
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
		Address:       sdk.AccAddress(tipperPriv.PubKey().Address()).String(),
		ChainID:       chainID,
		AccountNumber: accNum,
		Sequence:      accSeq,
		SignerIndex:   0,
	}
	sigV2, err := clienttx.SignWithPrivKey(
		signMode, signerData,
		txBuilder, tipperPriv, s.clientCtx.TxConfig, accSeq)
	s.Require().NoError(err)

	txBuilder.SetSignatures(sigV2)

	return txBuilder
}

func mkFeePayerTxBuilder(
	clientCtx client.Context,
	feePayerPriv cryptotypes.PrivKey, signMode signing.SignMode,
	fee tx.Fee, tipTx tx.TipTx, accNum, accSeq uint64, chainID string,
) (client.TxBuilder, error) {
	txBuilder := clientCtx.TxConfig.NewTxBuilder()
	err := txBuilder.SetMsgs(tipTx.GetMsgs()...)
	if err != nil {
		return nil, err
	}
	txBuilder.SetFeePayer(sdk.AccAddress(feePayerPriv.PubKey().Address()))
	txBuilder.SetFeeAmount(fee.Amount)
	txBuilder.SetGasLimit(fee.GasLimit)
	txBuilder.SetTip(tipTx.GetTip())

	// Calling SetSignatures with empty sig to populate AuthInfo.
	tipperSigsV2, err := tipTx.(authsigning.SigVerifiableTx).GetSignaturesV2()
	if err != nil {
		return nil, err
	}
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
		Address:       sdk.AccAddress(feePayerPriv.PubKey().Address()).String(),
		ChainID:       chainID,
		AccountNumber: accNum,
		Sequence:      accSeq,
		SignerIndex:   1,
	}
	feePayerSigV2, err = clienttx.SignWithPrivKey(
		signMode, signerData,
		txBuilder, feePayerPriv, clientCtx.TxConfig, accSeq)
	if err != nil {
		return nil, err
	}
	sigsV2 = append(tipperSigsV2, feePayerSigV2)
	err = txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	return txBuilder, nil
}
