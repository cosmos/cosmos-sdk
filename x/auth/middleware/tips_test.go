package middleware_test

import (
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

var initialRegens = sdk.NewCoins(sdk.NewCoin("regen", sdk.NewInt(1000)))
var initialAtoms = sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(1000)))

// setupAcctsForTips sets up 2 accounts:
// - tipper has 1000 regens
// - feePayer has 1000 atoms and 1000 regens
func (s *MWTestSuite) setupAcctsForTips(ctx sdk.Context) (sdk.Context, []testAccount) {
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

func (s *MWTestSuite) TestTips() {
	var msg sdk.Msg

	testcases := []struct {
		name      string
		tip       sdk.Coins
		fee       sdk.Coins
		gasLimit  uint64
		expErr    bool
		expErrStr string
	}{
		{
			"wrong tip denom",
			sdk.NewCoins(sdk.NewCoin("foobar", sdk.NewInt(1000))), initialAtoms, 200000,
			true, "0foobar is smaller than 1000foobar: insufficient funds",
		},
		{
			"insufficient tip from tipper",
			sdk.NewCoins(sdk.NewCoin("regen", sdk.NewInt(5000))), initialAtoms, 200000,
			true, "1000regen is smaller than 5000regen: insufficient funds",
		},
		{
			"insufficient fees from feePayer",
			initialRegens, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(5000))), 200000,
			true, "1000atom is smaller than 5000atom: insufficient funds: insufficient funds",
		},
		{
			"insufficient gas",
			initialRegens, initialAtoms, 100,
			true, "out of gas in location: ReadFlat; gasWanted: 100, gasUsed: 1000: out of gas",
		},
		{
			"happy case",
			initialRegens, initialAtoms, 200000,
			false, "",
		},
	}

	for _, tc := range testcases {
		tc := tc
		s.Run(tc.name, func() {
			ctx := s.SetupTest(false) // reset
			ctx, accts := s.setupAcctsForTips(ctx)
			tipper, feePayer := accts[0], accts[1]

			msg = govtypes.NewMsgVote(tipper.acc.GetAddress(), 1, govtypes.OptionYes)

			auxSignerData := s.mkTipperAuxSignerData(tipper.priv, msg, tc.tip, signing.SignMode_SIGN_MODE_DIRECT_AUX, tipper.accNum, 0, ctx.ChainID())
			feePayerTxBuilder := s.mkFeePayerTxBuilder(s.clientCtx, auxSignerData, feePayer.priv, signing.SignMode_SIGN_MODE_DIRECT, tx.Fee{Amount: tc.fee, GasLimit: tc.gasLimit}, feePayer.accNum, 0, ctx.ChainID())

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

func (s *MWTestSuite) mkTipperAuxSignerData(
	tipperPriv cryptotypes.PrivKey, msg sdk.Msg, tip sdk.Coins,
	signMode signing.SignMode, accNum, accSeq uint64, chainID string,
) tx.AuxSignerData {
	tipperAddr := sdk.AccAddress(tipperPriv.PubKey().Address()).String()
	b := clienttx.NewAuxTxBuilder()
	b.SetAddress(tipperAddr)
	b.SetAccountNumber(accNum)
	b.SetSequence(accSeq)
	err := b.SetMsgs(msg)
	s.Require().NoError(err)
	b.SetTip(&tx.Tip{Amount: tip, Tipper: tipperAddr})
	err = b.SetSignMode(signMode)
	s.Require().NoError(err)
	b.SetSequence(accSeq)
	err = b.SetPubKey(tipperPriv.PubKey())
	s.Require().NoError(err)
	b.SetChainID(chainID)

	signBz, err := b.GetSignBytes()
	s.Require().NoError(err)
	sig, err := tipperPriv.Sign(signBz)
	s.Require().NoError(err)
	b.SetSignature(sig)

	auxSignerData, err := b.GetAuxSignerData()
	s.Require().NoError(err)

	return auxSignerData
}

func (s *MWTestSuite) mkFeePayerTxBuilder(
	clientCtx client.Context,
	auxSignerData tx.AuxSignerData,
	feePayerPriv cryptotypes.PrivKey, signMode signing.SignMode,
	fee tx.Fee, accNum, accSeq uint64, chainID string,
) client.TxBuilder {
	txBuilder := clientCtx.TxConfig.NewTxBuilder()
	err := txBuilder.AddAuxSignerData(auxSignerData)
	s.Require().NoError(err)
	txBuilder.SetFeePayer(sdk.AccAddress(feePayerPriv.PubKey().Address()))
	txBuilder.SetFeeAmount(fee.Amount)
	txBuilder.SetGasLimit(fee.GasLimit)

	// Calling SetSignatures with empty sig to populate AuthInfo.
	tipperSigsV2, err := auxSignerData.GetSignatureV2()
	s.Require().NoError(err)
	feePayerSigV2 := signing.SignatureV2{
		PubKey: feePayerPriv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  signMode,
			Signature: nil,
		}}
	sigsV2 := append([]signing.SignatureV2{tipperSigsV2}, feePayerSigV2)
	txBuilder.SetSignatures(sigsV2...)

	// Actually sign the data.
	signerData := authsigning.SignerData{
		Address:       sdk.AccAddress(feePayerPriv.PubKey().Address()).String(),
		ChainID:       chainID,
		AccountNumber: accNum,
		Sequence:      accSeq,
		PubKey:        feePayerPriv.PubKey(),
	}
	feePayerSigV2, err = clienttx.SignWithPrivKey(
		signMode, signerData,
		txBuilder, feePayerPriv, clientCtx.TxConfig, accSeq)
	s.Require().NoError(err)
	sigsV2 = append([]signing.SignatureV2{tipperSigsV2}, feePayerSigV2)
	err = txBuilder.SetSignatures(sigsV2...)
	s.Require().NoError(err)

	return txBuilder
}
