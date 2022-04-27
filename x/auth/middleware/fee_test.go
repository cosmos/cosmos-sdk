package middleware_test

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/middleware"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
)

func (s *MWTestSuite) TestEnsureMempoolFees() {
	ctx := s.SetupTest(true) // setup
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()

	txHandler := middleware.ComposeMiddlewares(noopTxHandler, middleware.DeductFeeMiddleware(
		s.app.AccountKeeper,
		s.app.BankKeeper,
		s.app.FeeGrantKeeper,
		nil,
	))

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	atomCoin := sdk.NewCoin("atom", sdk.NewInt(150))
	apeCoin := sdk.NewInt64Coin("ape", 1500000)
	feeAmount := sdk.NewCoins(apeCoin, atomCoin)
	gasLimit := testdata.NewTestGasLimit()
	s.Require().NoError(txBuilder.SetMsgs(msg))
	txBuilder.SetFeeAmount(feeAmount)
	txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	testTx, _, err := s.createTestTx(txBuilder, privs, accNums, accSeqs, ctx.ChainID())
	s.Require().NoError(err)

	// Set high gas price so standard test fee fails
	atomPrice := sdk.NewDecCoinFromDec("atom", sdk.NewDec(200).Quo(sdk.NewDec(100000)))
	highGasPrice := []sdk.DecCoin{atomPrice}
	ctx = ctx.WithMinGasPrices(highGasPrice)

	// txHandler errors with insufficient fees
	_, _, err = txHandler.CheckTx(sdk.WrapSDKContext(ctx), tx.Request{Tx: testTx}, tx.RequestCheckTx{})
	s.Require().NotNil(err, "Middleware should have errored on too low fee for local gasPrice")

	// txHandler should fail since we also check minGasPrice in DeliverTx
	_, err = txHandler.DeliverTx(sdk.WrapSDKContext(ctx), tx.Request{Tx: testTx})
	s.Require().Error(err, "MempoolFeeMiddleware don't error in DeliverTx")

	atomPrice = sdk.NewDecCoinFromDec("atom", sdk.NewDec(0).Quo(sdk.NewDec(100000)))
	lowGasPrice := []sdk.DecCoin{atomPrice}
	ctx = ctx.WithMinGasPrices(lowGasPrice)

	// Set account with sufficient funds
	acc := s.app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	s.app.AccountKeeper.SetAccount(ctx, acc)
	err = testutil.FundAccount(s.app.BankKeeper, ctx, addr1, feeAmount)
	s.Require().NoError(err)

	_, checkTxRes, err := txHandler.CheckTx(sdk.WrapSDKContext(ctx), tx.Request{Tx: testTx}, tx.RequestCheckTx{})
	s.Require().Nil(err, "Middleware should not have errored on fee higher than local gasPrice")
	s.Require().Equal(atomCoin.Amount.Int64(), checkTxRes.Priority, "priority should be atom amount")
}

func (s *MWTestSuite) TestDeductFees() {
	ctx := s.SetupTest(false) // setup
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()
	txHandler := middleware.ComposeMiddlewares(
		noopTxHandler,
		middleware.DeductFeeMiddleware(
			s.app.AccountKeeper,
			s.app.BankKeeper,
			s.app.FeeGrantKeeper,
			nil,
		),
	)

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
	testTx, _, err := s.createTestTx(txBuilder, privs, accNums, accSeqs, ctx.ChainID())
	s.Require().NoError(err)

	// Set account with insufficient funds
	acc := s.app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	s.app.AccountKeeper.SetAccount(ctx, acc)
	coins := sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(10)))
	err = testutil.FundAccount(s.app.BankKeeper, ctx, addr1, coins)
	s.Require().NoError(err)

	// DeliverTx
	_, err = txHandler.DeliverTx(sdk.WrapSDKContext(ctx), tx.Request{Tx: testTx})
	s.Require().NotNil(err, "Tx errored when fee payer had insufficient funds")

	// SimulateTx
	_, err = txHandler.SimulateTx(sdk.WrapSDKContext(ctx), tx.Request{Tx: testTx})
	s.Require().NotNil(err, "Tx errored when fee payer had insufficient funds")

	// Set account with sufficient funds
	s.app.AccountKeeper.SetAccount(ctx, acc)
	err = testutil.FundAccount(s.app.BankKeeper, ctx, addr1, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(200))))
	s.Require().NoError(err)

	// DeliverTx
	_, err = txHandler.DeliverTx(sdk.WrapSDKContext(ctx), tx.Request{Tx: testTx})
	s.Require().Nil(err, "Tx did not error after account has been set with sufficient funds")

	err = testutil.FundAccount(s.app.BankKeeper, ctx, addr1, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(200))))
	s.Require().NoError(err)

	// SimulateTx
	_, err = txHandler.SimulateTx(sdk.WrapSDKContext(ctx), tx.Request{Tx: testTx})
	s.Require().Nil(err, "Tx did not error after account has been set with sufficient funds")
}
