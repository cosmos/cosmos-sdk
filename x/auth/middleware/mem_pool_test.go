package middleware_test

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/middleware"
	abci "github.com/tendermint/tendermint/abci/types"
)

func (suite *MWTestSuite) TestEnsureMempoolFees() {
	ctx := suite.SetupTest(true) // setup
	txBuilder := suite.clientCtx.TxConfig.NewTxBuilder()

	txHandler := middleware.ComposeMiddlewares(noopTxHandler{}, middleware.MempoolFeeMiddleware)

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	suite.Require().NoError(txBuilder.SetMsgs(msg))
	txBuilder.SetFeeAmount(feeAmount)
	txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, _, err := suite.createTestTx(txBuilder, privs, accNums, accSeqs, ctx.ChainID())
	suite.Require().NoError(err)

	// Set high gas price so standard test fee fails
	atomPrice := sdk.NewDecCoinFromDec("atom", sdk.NewDec(200).Quo(sdk.NewDec(100000)))
	highGasPrice := []sdk.DecCoin{atomPrice}
	ctx = ctx.WithMinGasPrices(highGasPrice)

	// antehandler errors with insufficient fees
	_, err = txHandler.CheckTx(sdk.WrapSDKContext(ctx), tx, abci.RequestCheckTx{})
	suite.Require().NotNil(err, "Decorator should have errored on too low fee for local gasPrice")

	// antehandler should not error since we do not check minGasPrice in DeliverTx
	_, err = txHandler.DeliverTx(sdk.WrapSDKContext(ctx), tx, abci.RequestDeliverTx{})
	suite.Require().Nil(err, "MempoolFeeDecorator returned error in DeliverTx")

	atomPrice = sdk.NewDecCoinFromDec("atom", sdk.NewDec(0).Quo(sdk.NewDec(100000)))
	lowGasPrice := []sdk.DecCoin{atomPrice}
	ctx = ctx.WithMinGasPrices(lowGasPrice)

	_, err = txHandler.CheckTx(sdk.WrapSDKContext(ctx), tx, abci.RequestCheckTx{})
	suite.Require().Nil(err, "Decorator should not have errored on fee higher than local gasPrice")
}
