package ante_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

	privs, accNums, accSeqs := []crypto.PrivKey{}, []uint64{}, []uint64{}
	invalidTx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	suite.Require().NoError(err)

	vbd := ante.NewValidateBasicDecorator()
	antehandler := sdk.ChainAnteDecorators(vbd)
	_, err = antehandler(suite.ctx, invalidTx, false)

	suite.Require().NotNil(err, "Did not error on invalid tx")

	privs, accNums, accSeqs = []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
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

	privs, accNums, accSeqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
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
	suite.txBuilder.SetMemo(strings.Repeat("01234567890", 10))

	privs, accNums, accSeqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	suite.Require().NoError(err)
	txBytes, err := json.Marshal(tx)
	suite.Require().Nil(err, "Cannot marshal tx: %v", err)

	cgtsd := ante.NewConsumeGasForTxSizeDecorator(suite.app.AccountKeeper)
	antehandler := sdk.ChainAnteDecorators(cgtsd)

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
	suite.Require().NoError(txBuilder.SetSignatures(signing.SignatureV2{}))

	simTxBytes, err := json.Marshal(txBuilder.GetTx())
	suite.Require().Nil(err)
	// require that simulated tx is smaller than tx with signatures
	suite.Require().True(len(simTxBytes) < len(txBytes), "simulated tx still has signatures")

	// Set suite.ctx with smaller simulated TxBytes manually
	suite.ctx = suite.ctx.WithTxBytes(txBytes)

	beforeSimGas := suite.ctx.GasMeter().GasConsumed()

	// run antehandler with simulate=true
	suite.ctx, err = antehandler(suite.ctx, txBuilder.GetTx(), true)
	consumedSimGas := suite.ctx.GasMeter().GasConsumed() - beforeSimGas

	// require that antehandler passes and does not underestimate decorator cost
	suite.Require().Nil(err, "ConsumeTxSizeGasDecorator returned error: %v", err)
	suite.Require().True(consumedSimGas >= expectedGas, "Simulate mode underestimates gas on AnteDecorator. Simulated cost: %d, expected cost: %d", consumedSimGas, expectedGas)
}

func TestAnteBasicTestSuite(t *testing.T) {
	suite.Run(t, new(AnteTestSuite))
}
