package ante_test

import (
	"strings"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
)

func (suite *AnteTestSuite) TestValidateBasic() {
	suite.SetupTest(true) // setup
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	suite.Require().NoError(suite.txBuilder.SetMsgs(msg))
	suite.txBuilder.SetFeeAmount(feeAmount)
	suite.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{}, []uint64{}, []uint64{}
	invalidTx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	suite.Require().NoError(err)

	vbd := ante.NewValidateBasicDecorator()
	antehandler := sdk.ChainAnteDecorators(vbd)
	_, err = antehandler(suite.ctx, invalidTx, false)

	suite.Require().NotNil(err, "Did not error on invalid tx")

	privs, accNums, accSeqs = []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	validTx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	suite.Require().NoError(err)

	_, err = antehandler(suite.ctx, validTx, false)
	suite.Require().Nil(err, "ValidateBasicDecorator returned error on valid tx. err: %v", err)

	// test decorator skips on recheck
	suite.ctx = suite.ctx.WithIsReCheckTx(true)

	// decorator should skip processing invalidTx on recheck and thus return nil-error
	_, err = antehandler(suite.ctx, invalidTx, false)

	suite.Require().Nil(err, "ValidateBasicDecorator ran on ReCheck")
}

func (suite *AnteTestSuite) TestValidateMemo() {
	suite.SetupTest(true) // setup
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	suite.Require().NoError(suite.txBuilder.SetMsgs(msg))
	suite.txBuilder.SetFeeAmount(feeAmount)
	suite.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	suite.txBuilder.SetMemo(strings.Repeat("01234567890", 500))
	invalidTx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	suite.Require().NoError(err)

	// require that long memos get rejected
	vmd := ante.NewValidateMemoDecorator(suite.app.AccountKeeper)
	antehandler := sdk.ChainAnteDecorators(vmd)
	_, err = antehandler(suite.ctx, invalidTx, false)

	suite.Require().NotNil(err, "Did not error on tx with high memo")

	suite.txBuilder.SetMemo(strings.Repeat("01234567890", 10))
	validTx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	suite.Require().NoError(err)

	// require small memos pass ValidateMemo Decorator
	_, err = antehandler(suite.ctx, validTx, false)
	suite.Require().Nil(err, "ValidateBasicDecorator returned error on valid tx. err: %v", err)
}

func (suite *AnteTestSuite) TestConsumeGasForTxSize() {
	suite.SetupTest(true) // setup

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()

	cgtsd := ante.NewConsumeGasForTxSizeDecorator(suite.app.AccountKeeper)
	antehandler := sdk.ChainAnteDecorators(cgtsd)

	testCases := []struct {
		name  string
		sigV2 signing.SignatureV2
	}{
		{"SingleSignatureData", signing.SignatureV2{PubKey: priv1.PubKey()}},
		{"MultiSignatureData", signing.SignatureV2{PubKey: priv1.PubKey(), Data: multisig.NewMultisig(2)}},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
			suite.Require().NoError(suite.txBuilder.SetMsgs(msg))
			suite.txBuilder.SetFeeAmount(feeAmount)
			suite.txBuilder.SetGasLimit(gasLimit)
			suite.txBuilder.SetMemo(strings.Repeat("01234567890", 10))

			privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
			tx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
			suite.Require().NoError(err)

			txBytes, err := suite.clientCtx.TxConfig.TxJSONEncoder()(tx)
			suite.Require().Nil(err, "Cannot marshal tx: %v", err)

			params := suite.app.AccountKeeper.GetParams(suite.ctx)
			expectedGas := sdk.Gas(len(txBytes)) * params.TxSizeCostPerByte

			// Set suite.ctx with TxBytes manually
			suite.ctx = suite.ctx.WithTxBytes(txBytes)

			// track how much gas is necessary to retrieve parameters
			beforeGas := suite.ctx.GasMeter().GasConsumed()
			suite.app.AccountKeeper.GetParams(suite.ctx)
			afterGas := suite.ctx.GasMeter().GasConsumed()
			expectedGas += afterGas - beforeGas

			beforeGas = suite.ctx.GasMeter().GasConsumed()
			suite.ctx, err = antehandler(suite.ctx, tx, false)
			suite.Require().Nil(err, "ConsumeTxSizeGasDecorator returned error: %v", err)

			// require that decorator consumes expected amount of gas
			consumedGas := suite.ctx.GasMeter().GasConsumed() - beforeGas
			suite.Require().Equal(expectedGas, consumedGas, "Decorator did not consume the correct amount of gas")

			// simulation must not underestimate gas of this decorator even with nil signatures
			txBuilder, err := suite.clientCtx.TxConfig.WrapTxBuilder(tx)
			suite.Require().NoError(err)
			suite.Require().NoError(txBuilder.SetSignatures(tc.sigV2))
			tx = txBuilder.GetTx()

			simTxBytes, err := suite.clientCtx.TxConfig.TxJSONEncoder()(tx)
			suite.Require().Nil(err, "Cannot marshal tx: %v", err)
			// require that simulated tx is smaller than tx with signatures
			suite.Require().True(len(simTxBytes) < len(txBytes), "simulated tx still has signatures")

			// Set suite.ctx with smaller simulated TxBytes manually
			suite.ctx = suite.ctx.WithTxBytes(simTxBytes)

			beforeSimGas := suite.ctx.GasMeter().GasConsumed()

			// run antehandler with simulate=true
			suite.ctx, err = antehandler(suite.ctx, tx, true)
			consumedSimGas := suite.ctx.GasMeter().GasConsumed() - beforeSimGas

			// require that antehandler passes and does not underestimate decorator cost
			suite.Require().Nil(err, "ConsumeTxSizeGasDecorator returned error: %v", err)
			suite.Require().True(consumedSimGas >= expectedGas, "Simulate mode underestimates gas on AnteDecorator. Simulated cost: %d, expected cost: %d", consumedSimGas, expectedGas)

		})
	}

}

func (suite *AnteTestSuite) TestTxHeightTimeoutDecorator() {
	suite.SetupTest(true)

	antehandler := sdk.ChainAnteDecorators(ante.TxTimeoutHeightDecorator{})

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()

	testCases := []struct {
		name      string
		timeout   uint64
		height    int64
		expectErr bool
	}{
		{"default value", 0, 10, false},
		{"no timeout (greater height)", 15, 10, false},
		{"no timeout (same height)", 10, 10, false},
		{"timeout (smaller height)", 9, 10, true},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

			suite.Require().NoError(suite.txBuilder.SetMsgs(msg))

			suite.txBuilder.SetFeeAmount(feeAmount)
			suite.txBuilder.SetGasLimit(gasLimit)
			suite.txBuilder.SetMemo(strings.Repeat("01234567890", 10))
			suite.txBuilder.SetTimeoutHeight(tc.timeout)

			privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
			tx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
			suite.Require().NoError(err)

			ctx := suite.ctx.WithBlockHeight(tc.height)
			_, err = antehandler(ctx, tx, true)
			suite.Require().Equal(tc.expectErr, err != nil, err)
		})
	}
}
