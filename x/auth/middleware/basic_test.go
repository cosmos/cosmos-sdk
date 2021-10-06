package middleware_test

import (
	"strings"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/middleware"
	"github.com/tendermint/tendermint/abci/types"
)

func (s *MWTestSuite) TestValidateBasic() {
	ctx := s.SetupTest(true) // setup
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()

	txHandler := middleware.ComposeMiddlewares(noopTxHandler{}, middleware.ValidateBasicMiddleware)

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	s.Require().NoError(txBuilder.SetMsgs(msg))
	txBuilder.SetFeeAmount(feeAmount)
	txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{}, []uint64{}, []uint64{}
	invalidTx, _, err := s.createTestTx(txBuilder, privs, accNums, accSeqs, ctx.ChainID())
	s.Require().NoError(err)

	_, err = txHandler.DeliverTx(sdk.WrapSDKContext(ctx), invalidTx, types.RequestDeliverTx{})
	s.Require().NotNil(err, "Did not error on invalid tx")

	privs, accNums, accSeqs = []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	validTx, _, err := s.createTestTx(txBuilder, privs, accNums, accSeqs, ctx.ChainID())
	s.Require().NoError(err)

	_, err = txHandler.DeliverTx(sdk.WrapSDKContext(ctx), validTx, types.RequestDeliverTx{})
	s.Require().Nil(err, "ValidateBasicMiddleware returned error on valid tx. err: %v", err)

	// test middleware skips on recheck
	ctx = ctx.WithIsReCheckTx(true)

	// middleware should skip processing invalidTx on recheck and thus return nil-error
	_, err = txHandler.DeliverTx(sdk.WrapSDKContext(ctx), invalidTx, types.RequestDeliverTx{})
	s.Require().Nil(err, "ValidateBasicMiddleware ran on ReCheck")
}

func (s *MWTestSuite) TestValidateMemo() {
	ctx := s.SetupTest(true) // setup
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()
	txHandler := middleware.ComposeMiddlewares(noopTxHandler{}, middleware.ValidateMemoMiddleware(s.app.AccountKeeper))

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	s.Require().NoError(txBuilder.SetMsgs(msg))
	txBuilder.SetFeeAmount(feeAmount)
	txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	txBuilder.SetMemo(strings.Repeat("01234567890", 500))
	invalidTx, _, err := s.createTestTx(txBuilder, privs, accNums, accSeqs, ctx.ChainID())
	s.Require().NoError(err)

	// require that long memos get rejected
	_, err = txHandler.DeliverTx(sdk.WrapSDKContext(ctx), invalidTx, types.RequestDeliverTx{})

	s.Require().NotNil(err, "Did not error on tx with high memo")

	txBuilder.SetMemo(strings.Repeat("01234567890", 10))
	validTx, _, err := s.createTestTx(txBuilder, privs, accNums, accSeqs, ctx.ChainID())
	s.Require().NoError(err)

	// require small memos pass ValidateMemo middleware
	_, err = txHandler.DeliverTx(sdk.WrapSDKContext(ctx), validTx, types.RequestDeliverTx{})
	s.Require().Nil(err, "ValidateBasicMiddleware returned error on valid tx. err: %v", err)
}

func (s *MWTestSuite) TestConsumeGasForTxSize() {
	ctx := s.SetupTest(true) // setup
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()

	txHandler := middleware.ComposeMiddlewares(noopTxHandler{}, middleware.ConsumeTxSizeGasMiddleware(s.app.AccountKeeper))

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()

	testCases := []struct {
		name  string
		sigV2 signing.SignatureV2
	}{
		{"SingleSignatureData", signing.SignatureV2{PubKey: priv1.PubKey()}},
		{"MultiSignatureData", signing.SignatureV2{PubKey: priv1.PubKey(), Data: multisig.NewMultisig(2)}},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			txBuilder = s.clientCtx.TxConfig.NewTxBuilder()
			s.Require().NoError(txBuilder.SetMsgs(msg))
			txBuilder.SetFeeAmount(feeAmount)
			txBuilder.SetGasLimit(gasLimit)
			txBuilder.SetMemo(strings.Repeat("01234567890", 10))

			privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
			testTx, _, err := s.createTestTx(txBuilder, privs, accNums, accSeqs, ctx.ChainID())
			s.Require().NoError(err)

			txBytes, err := s.clientCtx.TxConfig.TxJSONEncoder()(testTx)
			s.Require().Nil(err, "Cannot marshal tx: %v", err)

			params := s.app.AccountKeeper.GetParams(ctx)
			expectedGas := sdk.Gas(len(txBytes)) * params.TxSizeCostPerByte

			// Set ctx with TxBytes manually
			ctx = ctx.WithTxBytes(txBytes)

			// track how much gas is necessary to retrieve parameters
			beforeGas := ctx.GasMeter().GasConsumed()
			s.app.AccountKeeper.GetParams(ctx)
			afterGas := ctx.GasMeter().GasConsumed()
			expectedGas += afterGas - beforeGas

			beforeGas = ctx.GasMeter().GasConsumed()
			_, err = txHandler.DeliverTx(sdk.WrapSDKContext(ctx), testTx, types.RequestDeliverTx{Tx: txBytes})

			s.Require().Nil(err, "ConsumeTxSizeGasMiddleware returned error: %v", err)

			// require that middleware consumes expected amount of gas
			consumedGas := ctx.GasMeter().GasConsumed() - beforeGas
			s.Require().Equal(expectedGas, consumedGas, "Middleware did not consume the correct amount of gas")

			// simulation must not underestimate gas of this middleware even with nil signatures
			txBuilder, err := s.clientCtx.TxConfig.WrapTxBuilder(testTx)
			s.Require().NoError(err)
			s.Require().NoError(txBuilder.SetSignatures(tc.sigV2))
			testTx = txBuilder.GetTx()

			simTxBytes, err := s.clientCtx.TxConfig.TxJSONEncoder()(testTx)
			s.Require().Nil(err, "Cannot marshal tx: %v", err)
			// require that simulated tx is smaller than tx with signatures
			s.Require().True(len(simTxBytes) < len(txBytes), "simulated tx still has signatures")

			// Set s.ctx with smaller simulated TxBytes manually
			ctx = ctx.WithTxBytes(simTxBytes)

			beforeSimGas := ctx.GasMeter().GasConsumed()

			// run txhandler in simulate mode
			_, err = txHandler.SimulateTx(sdk.WrapSDKContext(ctx), testTx, tx.RequestSimulateTx{TxBytes: simTxBytes})
			consumedSimGas := ctx.GasMeter().GasConsumed() - beforeSimGas

			// require that txhandler passes and does not underestimate middleware cost
			s.Require().Nil(err, "ConsumeTxSizeGasMiddleware returned error: %v", err)
			s.Require().True(consumedSimGas >= expectedGas, "Simulate mode underestimates gas on Middleware. Simulated cost: %d, expected cost: %d", consumedSimGas, expectedGas)
		})
	}
}

func (s *MWTestSuite) TestTxHeightTimeoutMiddleware() {
	ctx := s.SetupTest(true)

	txHandler := middleware.ComposeMiddlewares(noopTxHandler{}, middleware.TxTimeoutHeightMiddleware)

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

		s.Run(tc.name, func() {
			txBuilder := s.clientCtx.TxConfig.NewTxBuilder()

			s.Require().NoError(txBuilder.SetMsgs(msg))

			txBuilder.SetFeeAmount(feeAmount)
			txBuilder.SetGasLimit(gasLimit)
			txBuilder.SetMemo(strings.Repeat("01234567890", 10))
			txBuilder.SetTimeoutHeight(tc.timeout)

			privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
			testTx, _, err := s.createTestTx(txBuilder, privs, accNums, accSeqs, ctx.ChainID())
			s.Require().NoError(err)

			ctx := ctx.WithBlockHeight(tc.height)
			_, err = txHandler.SimulateTx(sdk.WrapSDKContext(ctx), testTx, tx.RequestSimulateTx{})
			s.Require().Equal(tc.expectErr, err != nil, err)
		})
	}
}
