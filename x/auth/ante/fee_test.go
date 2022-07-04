package ante_test

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
)

func (s *AnteTestSuite) TestDeductFeeDecorator_ZeroGas() {
	s.SetupTest(true) // setup
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

<<<<<<< HEAD
	mfd := ante.NewDeductFeeDecorator(suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.FeeGrantKeeper, nil)
=======
	mfd := ante.NewDeductFeeDecorator(s.accountKeeper, s.bankKeeper, s.feeGrantKeeper, nil)
>>>>>>> 74f265c94 (fix!: prevent 0 gas txs (#12416))
	antehandler := sdk.ChainAnteDecorators(mfd)

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	coins := sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(300)))
<<<<<<< HEAD
	testutil.FundAccount(suite.app.BankKeeper, suite.ctx, addr1, coins)
=======
	testutil.FundAccount(s.bankKeeper, s.ctx, addr1, coins)

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	s.Require().NoError(s.txBuilder.SetMsgs(msg))

	// set zero gas
	s.txBuilder.SetGasLimit(0)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	s.Require().NoError(err)

	// Set IsCheckTx to true
	s.ctx = s.ctx.WithIsCheckTx(true)

	_, err = antehandler(s.ctx, tx, false)
	s.Require().Error(err)
}

func (s *AnteTestSuite) TestEnsureMempoolFees() {
	s.SetupTest(true) // setup
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

	mfd := ante.NewDeductFeeDecorator(s.accountKeeper, s.bankKeeper, s.feeGrantKeeper, nil)
	antehandler := sdk.ChainAnteDecorators(mfd)

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	coins := sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(300)))
	testutil.FundAccount(s.bankKeeper, s.ctx, addr1, coins)
>>>>>>> 74f265c94 (fix!: prevent 0 gas txs (#12416))

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	s.Require().NoError(s.txBuilder.SetMsgs(msg))
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	s.Require().NoError(err)

	// Set high gas price so standard test fee fails
	atomPrice := sdk.NewDecCoinFromDec("atom", sdk.NewDec(200).Quo(sdk.NewDec(100000)))
	highGasPrice := []sdk.DecCoin{atomPrice}
	s.ctx = s.ctx.WithMinGasPrices(highGasPrice)

	// Set IsCheckTx to true
	s.ctx = s.ctx.WithIsCheckTx(true)

	// antehandler errors with insufficient fees
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "Decorator should have errored on too low fee for local gasPrice")

	// Set IsCheckTx to false
	s.ctx = s.ctx.WithIsCheckTx(false)

	// antehandler should not error since we do not check minGasPrice in DeliverTx
	_, err = antehandler(s.ctx, tx, false)
	s.Require().Nil(err, "MempoolFeeDecorator returned error in DeliverTx")

	// Set IsCheckTx back to true for testing sufficient mempool fee
	s.ctx = s.ctx.WithIsCheckTx(true)

	atomPrice = sdk.NewDecCoinFromDec("atom", sdk.NewDec(0).Quo(sdk.NewDec(100000)))
	lowGasPrice := []sdk.DecCoin{atomPrice}
	s.ctx = s.ctx.WithMinGasPrices(lowGasPrice)

	newCtx, err := antehandler(s.ctx, tx, false)
	s.Require().Nil(err, "Decorator should not have errored on fee higher than local gasPrice")
	// Priority is the smallest amount in any denom. Since we have only 1 fee
	// of 150atom, the priority here is 150.
	s.Require().Equal(feeAmount.AmountOf("atom").Int64(), newCtx.Priority())
}

func (s *AnteTestSuite) TestDeductFees() {
	s.SetupTest(false) // setup
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	s.Require().NoError(s.txBuilder.SetMsgs(msg))
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	s.Require().NoError(err)

	// Set account with insufficient funds
<<<<<<< HEAD
	acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr1)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
	coins := sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(10)))
	err = testutil.FundAccount(suite.app.BankKeeper, suite.ctx, addr1, coins)
	suite.Require().NoError(err)

	dfd := ante.NewDeductFeeDecorator(suite.app.AccountKeeper, suite.app.BankKeeper, nil, nil)
=======
	acc := s.accountKeeper.NewAccountWithAddress(s.ctx, addr1)
	s.accountKeeper.SetAccount(s.ctx, acc)
	coins := sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(10)))
	err = testutil.FundAccount(s.bankKeeper, s.ctx, addr1, coins)
	s.Require().NoError(err)

	dfd := ante.NewDeductFeeDecorator(s.accountKeeper, s.bankKeeper, nil, nil)
>>>>>>> 74f265c94 (fix!: prevent 0 gas txs (#12416))
	antehandler := sdk.ChainAnteDecorators(dfd)

	_, err = antehandler(s.ctx, tx, false)

	s.Require().NotNil(err, "Tx did not error when fee payer had insufficient funds")

	// Set account with sufficient funds
<<<<<<< HEAD
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
	err = testutil.FundAccount(suite.app.BankKeeper, suite.ctx, addr1, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(200))))
	suite.Require().NoError(err)
=======
	s.accountKeeper.SetAccount(s.ctx, acc)
	err = testutil.FundAccount(s.bankKeeper, s.ctx, addr1, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(200))))
	s.Require().NoError(err)
>>>>>>> 74f265c94 (fix!: prevent 0 gas txs (#12416))

	_, err = antehandler(s.ctx, tx, false)

	s.Require().Nil(err, "Tx errored after account has been set with sufficient funds")
}
