package ante_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestValidateBasic(t *testing.T) {
	// setup
	_, ctx := createTestApp(true)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()

	// msg and signatures
	msg1 := types.NewTestMsg(addr1)
	fee := types.NewTestStdFee()

	msgs := []sdk.Msg{msg1}

	privs, accNums, seqs := []crypto.PrivKey{}, []uint64{}, []uint64{}
	invalidTx := types.NewTestTx(ctx, msgs, privs, accNums, seqs, fee)

	vbd := ante.NewValidateBasicDecorator()
	antehandler := sdk.ChainAnteDecorators(vbd)
	_, err := antehandler(ctx, invalidTx, false)

	require.NotNil(t, err, "Did not error on invalid tx")

	privs, accNums, seqs = []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	validTx := types.NewTestTx(ctx, msgs, privs, accNums, seqs, fee)

	_, err = antehandler(ctx, validTx, false)
	require.Nil(t, err, "ValidateBasicDecorator returned error on valid tx. err: %v", err)

	// test decorator skips on recheck
	ctx = ctx.WithIsReCheckTx(true)

	// decorator should skip processing invalidTx on recheck and thus return nil-error
	_, err = antehandler(ctx, invalidTx, false)

	require.Nil(t, err, "ValidateBasicDecorator ran on ReCheck")
}

func TestValidateMemo(t *testing.T) {
	// setup
	app, ctx := createTestApp(true)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()

	// msg and signatures
	msg1 := types.NewTestMsg(addr1)
	fee := types.NewTestStdFee()

	msgs := []sdk.Msg{msg1}

	privs, accNums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	invalidTx := types.NewTestTxWithMemo(ctx, msgs, privs, accNums, seqs, fee, strings.Repeat("01234567890", 500))

	// require that long memos get rejected
	vmd := ante.NewValidateMemoDecorator(app.AccountKeeper)
	antehandler := sdk.ChainAnteDecorators(vmd)
	_, err := antehandler(ctx, invalidTx, false)

	require.NotNil(t, err, "Did not error on tx with high memo")

	validTx := types.NewTestTxWithMemo(ctx, msgs, privs, accNums, seqs, fee, strings.Repeat("01234567890", 10))

	// require small memos pass ValidateMemo Decorator
	_, err = antehandler(ctx, validTx, false)
	require.Nil(t, err, "ValidateBasicDecorator returned error on valid tx. err: %v", err)
}

func TestConsumeGasForTxSize(t *testing.T) {
	// setup
	app, ctx := createTestApp(true)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()

	// msg and signatures
	msg1 := types.NewTestMsg(addr1)
	fee := types.NewTestStdFee()

	msgs := []sdk.Msg{msg1}

	privs, accNums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx := types.NewTestTxWithMemo(ctx, msgs, privs, accNums, seqs, fee, strings.Repeat("01234567890", 10))
	txBytes, err := json.Marshal(tx)
	require.Nil(t, err, "Cannot marshal tx: %v", err)

	cgtsd := ante.NewConsumeGasForTxSizeDecorator(app.AccountKeeper)
	antehandler := sdk.ChainAnteDecorators(cgtsd)

	params := app.AccountKeeper.GetParams(ctx)
	expectedGas := sdk.Gas(len(txBytes)) * params.TxSizeCostPerByte

	// Set ctx with TxBytes manually
	ctx = ctx.WithTxBytes(txBytes)

	// track how much gas is necessary to retrieve parameters
	beforeGas := ctx.GasMeter().GasConsumed()
	app.AccountKeeper.GetParams(ctx)
	afterGas := ctx.GasMeter().GasConsumed()
	expectedGas += afterGas - beforeGas

	beforeGas = ctx.GasMeter().GasConsumed()
	ctx, err = antehandler(ctx, tx, false)
	require.Nil(t, err, "ConsumeTxSizeGasDecorator returned error: %v", err)

	// require that decorator consumes expected amount of gas
	consumedGas := ctx.GasMeter().GasConsumed() - beforeGas
	require.Equal(t, expectedGas, consumedGas, "Decorator did not consume the correct amount of gas")

	// simulation must not underestimate gas of this decorator even with nil signatures
	sigTx := tx.(types.StdTx)
	sigTx.Signatures = []types.StdSignature{{}}

	simTxBytes, err := json.Marshal(sigTx)
	require.Nil(t, err)
	// require that simulated tx is smaller than tx with signatures
	require.True(t, len(simTxBytes) < len(txBytes), "simulated tx still has signatures")

	// Set ctx with smaller simulated TxBytes manually
	ctx = ctx.WithTxBytes(txBytes)

	beforeSimGas := ctx.GasMeter().GasConsumed()

	// run antehandler with simulate=true
	ctx, err = antehandler(ctx, sigTx, true)
	consumedSimGas := ctx.GasMeter().GasConsumed() - beforeSimGas

	// require that antehandler passes and does not underestimate decorator cost
	require.Nil(t, err, "ConsumeTxSizeGasDecorator returned error: %v", err)
	require.True(t, consumedSimGas >= expectedGas, "Simulate mode underestimates gas on AnteDecorator. Simulated cost: %d, expected cost: %d", consumedSimGas, expectedGas)
}
