package middleware_test

import (
	"context"
	"errors"

	abci "github.com/tendermint/tendermint/abci/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/middleware"
)

// txTest is a dummy tx that doesn't implement GasTx. It should set the GasMeter
// to 0 in this case.
type txTest struct{}

var _ sdk.Tx = txTest{}

func (t txTest) GetMsgs() []sdk.Msg   { return []sdk.Msg{} }
func (t txTest) ValidateBasic() error { return nil }

func (suite *MWTestSuite) TestSetup() {
	ctx := suite.SetupTest(true)
	txBuilder := suite.clientCtx.TxConfig.NewTxBuilder()

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	suite.Require().NoError(txBuilder.SetMsgs(msg))
	txBuilder.SetFeeAmount(feeAmount)
	txBuilder.SetGasLimit(gasLimit)

	// test tx
	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, _, err := suite.CreateTestTx(txBuilder, privs, accNums, accSeqs, ctx.ChainID())
	suite.Require().NoError(err)

	// Set height to non-zero value for GasMeter to be set
	ctx = ctx.WithBlockHeight(1)

	// Run TxHandler
	txHandler := middleware.ComposeTxMiddleware(noopTxHandler{}, middleware.NewGasTxMiddleware())

	testcases := []struct {
		name        string
		tx          sdk.Tx
		expGasLimit uint64
		expErr      bool
	}{
		{"not a gas tx", txTest{}, 0, true},
		{"tx with its own gas limit", tx, gasLimit, false},
	}
	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			res, err := txHandler.CheckTx(sdk.WrapSDKContext(ctx), tc.tx, abci.RequestCheckTx{})
			if tc.expErr {
				suite.Require().Error(err)
			} else {
				suite.Require().Nil(err, "SetUpContextDecorator returned error")
				suite.Require().Equal(tc.expGasLimit, uint64(res.GasWanted))
			}
		})
	}
}

func (suite *MWTestSuite) TestRecoverPanic() {
	ctx := suite.SetupTest(true)
	txBuilder := suite.clientCtx.TxConfig.NewTxBuilder()

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
	tx, txBytes, err := suite.CreateTestTx(txBuilder, privs, accNums, accSeqs, ctx.ChainID())
	suite.Require().NoError(err)

	txHandler := middleware.ComposeTxMiddleware(outOfGasTxHandler{}, middleware.NewRecoveryTxMiddleware(), middleware.NewGasTxMiddleware())
	res, err := txHandler.CheckTx(sdk.WrapSDKContext(ctx), tx, abci.RequestCheckTx{Tx: txBytes})
	suite.Require().Error(err, "Did not return error on OutOfGas panic")
	suite.Require().True(errors.Is(sdkerrors.ErrOutOfGas, err), "Returned error is not an out of gas error")
	suite.Require().Equal(gasLimit, uint64(res.GasWanted))

	txHandler = middleware.ComposeTxMiddleware(outOfGasTxHandler{}, middleware.NewGasTxMiddleware())
	suite.Require().Panics(func() { txHandler.CheckTx(sdk.WrapSDKContext(ctx), tx, abci.RequestCheckTx{Tx: txBytes}) }, "Recovered from non-Out-of-Gas panic")
}

// outOfGasTxHandler is a test iddleware that will throw OutOfGas panic.
type outOfGasTxHandler struct{}

var _ tx.Handler = outOfGasTxHandler{}

func (txh outOfGasTxHandler) DeliverTx(ctx context.Context, _ sdk.Tx, _ abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	overLimit := sdkCtx.GasMeter().Limit() + 1

	// Should panic with outofgas error
	sdkCtx.GasMeter().ConsumeGas(overLimit, "test panic")

	panic("not reached")
}
func (txh outOfGasTxHandler) CheckTx(ctx context.Context, _ sdk.Tx, _ abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	overLimit := sdkCtx.GasMeter().Limit() + 1

	// Should panic with outofgas error
	sdkCtx.GasMeter().ConsumeGas(overLimit, "test panic")

	panic("not reached")
}
func (txh outOfGasTxHandler) SimulateTx(ctx context.Context, _ sdk.Tx, _ tx.RequestSimulateTx) (tx.ResponseSimulateTx, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	overLimit := sdkCtx.GasMeter().Limit() + 1

	// Should panic with outofgas error
	sdkCtx.GasMeter().ConsumeGas(overLimit, "test panic")

	panic("not reached")
}

// noopTxHandler is a test middleware that will throw OutOfGas panic.
type noopTxHandler struct{}

var _ tx.Handler = noopTxHandler{}

func (txh noopTxHandler) CheckTx(_ context.Context, _ sdk.Tx, _ abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	return abci.ResponseCheckTx{}, nil
}
func (txh noopTxHandler) SimulateTx(_ context.Context, _ sdk.Tx, _ tx.RequestSimulateTx) (tx.ResponseSimulateTx, error) {
	return tx.ResponseSimulateTx{}, nil
}
func (txh noopTxHandler) DeliverTx(ctx context.Context, _ sdk.Tx, _ abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	return abci.ResponseDeliverTx{}, nil
}
