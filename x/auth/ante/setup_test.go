package ante_test

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
)

func (suite *AnteTestSuite) TestSetup() {
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
	tx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	suite.Require().NoError(err)

	sud := ante.NewSetUpContextDecorator()
	antehandler := sdk.ChainAnteDecorators(sud)

	// Set height to non-zero value for GasMeter to be set
	suite.ctx = suite.ctx.WithBlockHeight(1)

	// Context GasMeter Limit not set
	suite.Require().Equal(uint64(0), suite.ctx.GasMeter().Limit(), "GasMeter set with limit before setup")

	newCtx, err := antehandler(suite.ctx, tx, false)
	suite.Require().Nil(err, "SetUpContextDecorator returned error")

	// Context GasMeter Limit should be set after SetUpContextDecorator runs
	suite.Require().Equal(gasLimit, newCtx.GasMeter().Limit(), "GasMeter not set correctly")
}

func (suite *AnteTestSuite) TestRecoverPanic() {
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
	tx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	suite.Require().NoError(err)

	sud := ante.NewSetUpContextDecorator()
	antehandler := sdk.ChainAnteDecorators(sud, OutOfGasDecorator{})

	// Set height to non-zero value for GasMeter to be set
	suite.ctx = suite.ctx.WithBlockHeight(1)

	newCtx, err := antehandler(suite.ctx, tx, false)

	suite.Require().NotNil(err, "Did not return error on OutOfGas panic")

	suite.Require().True(sdkerrors.ErrOutOfGas.Is(err), "Returned error is not an out of gas error")
	suite.Require().Equal(gasLimit, newCtx.GasMeter().Limit())

	antehandler = sdk.ChainAnteDecorators(sud, PanicDecorator{})
	suite.Require().Panics(func() { antehandler(suite.ctx, tx, false) }, "Recovered from non-Out-of-Gas panic") // nolint:errcheck
}

type OutOfGasDecorator struct{}

// AnteDecorator that will throw OutOfGas panic
func (ogd OutOfGasDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	overLimit := ctx.GasMeter().Limit() + 1

	// Should panic with outofgas error
	ctx.GasMeter().ConsumeGas(overLimit, "test panic")

	// not reached
	return next(ctx, tx, simulate)
}

type PanicDecorator struct{}

func (pd PanicDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	panic("random error")
}
