package middleware_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/suite"
)

func (s *MWTestSuite) TestMetaTransactions() {
	ctx := s.SetupTest(false) // reset

	// Setup 2 accounts:
	// - tipper has 1000 regens
	// - feePayer has 1000 atoms and 1000 regens
	regens := sdk.NewCoins(sdk.NewCoin("regen", sdk.NewInt(1000)))
	atoms := sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(1000)))
	accts := s.createTestAccounts(ctx, 1, regens)
	tipper, feePayer := accts[0], accts[1]
	err := s.app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(1000))))
	s.Require().NoError(err)
	err = s.app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, feePayer.acc.GetAddress(), atoms)
	s.Require().NoError(err)

	testcases := []struct {
		name      string
		tip       sdk.Coins
		fee       sdk.Coins
		gasLimit  uint64
		expErr    bool
		expErrStr string
	}{
		{
			name: "",
			tip:  sdk.NewCoins(sdk.NewCoin("regen", sdk.NewInt(1000))),
			fee:  sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(1000))),
		},
	}
	for _, tc := range testcases {
		s.Run(tc.name, func() {
			tipperTxBuilder := s.mkTipperTxBuilder(tipper.acc.String(), tc.tip)
			feePayerTxBuilder := s.mkFeePayerTxBuilder(feePayer.acc.String(), tx.Fee{Amount: tc.fee, GasLimit: tc.gasLimit}, tipperTxBuilder)
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

func (s *MWTestSuite) mkTipperTxBuilder(tipper string, tip sdk.Coins) client.TxBuilder {
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()
	txBuilder.SetTip(&tx.Tip{
		Amount: tip,
		Tipper: tipper,
	})
	txBuilder.SetMsgs()

	return txBuilder
}

func (s *MWTestSuite) mkFeePayerTxBuilder(feePayer string, fee tx.Fee, tipperTxBuilder client.TxBuilder) client.TxBuilder {
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()
	txBuilder.SetMsgs(tipperTxBuilder.GetTx().GetMsgs()...)
	txBuilder.SetFeeAmount(fee.Amount)
	txBuilder.SetGasLimit(fee.GasLimit)

	return txBuilder
}

func TestMWTestSuite2(t *testing.T) {
	suite.Run(t, new(MWTestSuite))
}
