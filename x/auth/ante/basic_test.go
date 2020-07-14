package ante_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec/testdata"
	"github.com/stretchr/testify/suite"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func (suite *AnteTestSuite) TestValidateBasic() {
	suite.SetupTest(true) // setup
	suite.txBuilder = suite.clientCtx.TxGenerator.NewTxBuilder()

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	fee := types.NewTestStdFee()
	suite.txBuilder.SetMsgs(msg)
	suite.txBuilder.SetFeeAmount(fee.GetAmount())
	suite.txBuilder.SetGasLimit(fee.GetGas())

	privs, accNums, seqs := []crypto.PrivKey{}, []uint64{}, []uint64{}
	invalidTx := suite.CreateTestTx(privs, accNums, seqs, suite.ctx.ChainID())

	vbd := ante.NewValidateBasicDecorator()
	antehandler := sdk.ChainAnteDecorators(vbd)
	_, err := antehandler(suite.ctx, invalidTx, false)

	suite.Require().NotNil(err, "Did not error on invalid tx")

	privs, accNums, seqs = []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	validTx := suite.CreateTestTx(privs, accNums, seqs, suite.ctx.ChainID())

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
	suite.txBuilder = suite.clientCtx.TxGenerator.NewTxBuilder()

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	fee := types.NewTestStdFee()
	suite.txBuilder.SetMsgs(msg)
	suite.txBuilder.SetFeeAmount(fee.GetAmount())
	suite.txBuilder.SetGasLimit(fee.GetGas())

	privs, accNums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	suite.txBuilder.SetMemo(strings.Repeat("01234567890", 500))
	invalidTx := suite.CreateTestTx(privs, accNums, seqs, suite.ctx.ChainID())

	// require that long memos get rejected
	vmd := ante.NewValidateMemoDecorator(suite.app.AccountKeeper)
	antehandler := sdk.ChainAnteDecorators(vmd)
	_, err := antehandler(suite.ctx, invalidTx, false)

	suite.Require().NotNil(err, "Did not error on tx with high memo")

	suite.txBuilder.SetMemo(strings.Repeat("01234567890", 10))
	validTx := suite.CreateTestTx(privs, accNums, seqs, suite.ctx.ChainID())

	// require small memos pass ValidateMemo Decorator
	_, err = antehandler(suite.ctx, validTx, false)
	suite.Require().Nil(err, "ValidateBasicDecorator returned error on valid tx. err: %v", err)
}

func (suite *AnteTestSuite) TestConsumeGasForTxSize() {
	suite.SetupTest(true) // setup
	suite.txBuilder = suite.clientCtx.TxGenerator.NewTxBuilder()

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	fee := types.NewTestStdFee()
	suite.txBuilder.SetMsgs(msg)
	suite.txBuilder.SetFeeAmount(fee.GetAmount())
	suite.txBuilder.SetGasLimit(fee.GetGas())
	suite.txBuilder.SetMemo(strings.Repeat("01234567890", 10))

	privs, accNums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx := suite.CreateTestTx(privs, accNums, seqs, suite.ctx.ChainID())
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
	sigTx := tx.(types.StdTx)
	sigTx.Signatures = []types.StdSignature{{}}

	simTxBytes, err := json.Marshal(sigTx)
	suite.Require().Nil(err)
	// require that simulated tx is smaller than tx with signatures
	suite.Require().True(len(simTxBytes) < len(txBytes), "simulated tx still has signatures")

	// Set suite.ctx with smaller simulated TxBytes manually
	suite.ctx = suite.ctx.WithTxBytes(txBytes)

	beforeSimGas := suite.ctx.GasMeter().GasConsumed()

	// run antehandler with simulate=true
	suite.ctx, err = antehandler(suite.ctx, sigTx, true)
	consumedSimGas := suite.ctx.GasMeter().GasConsumed() - beforeSimGas

	// require that antehandler passes and does not underestimate decorator cost
	suite.Require().Nil(err, "ConsumeTxSizeGasDecorator returned error: %v", err)
	suite.Require().True(consumedSimGas >= expectedGas, "Simulate mode underestimates gas on AnteDecorator. Simulated cost: %d, expected cost: %d", consumedSimGas, expectedGas)
}

func TestAnteBasicTestSuite(t *testing.T) {
	suite.Run(t, new(AnteTestSuite))
}
