package ante_test

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/header"
	"cosmossdk.io/x/auth/ante"
	"cosmossdk.io/x/auth/ante/unorderedtx"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

func TestUnorderedTxDecorator_OrderedTx(t *testing.T) {
	txm := unorderedtx.NewManager(t.TempDir())
	defer func() {
		require.NoError(t, txm.Close())
	}()

	txm.Start()

	suite := SetupTestSuite(t, false)

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(unorderedtx.DefaultmaxTimeoutDuration, txm, suite.accountKeeper.GetEnvironment()))

	tx, txBz := genUnorderedTx(t, false, time.Time{})
	ctx := sdk.Context{}.WithTxBytes(txBz)

	_, err := chain(ctx, tx, false)
	require.NoError(t, err)
}

func TestUnorderedTxDecorator_UnorderedTx_NoTTL(t *testing.T) {
	txm := unorderedtx.NewManager(t.TempDir())
	defer func() {
		require.NoError(t, txm.Close())
	}()

	txm.Start()

	suite := SetupTestSuite(t, false)

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(unorderedtx.DefaultmaxTimeoutDuration, txm, suite.accountKeeper.GetEnvironment()))

	tx, txBz := genUnorderedTx(t, true, time.Time{})
	ctx := sdk.Context{}.WithTxBytes(txBz)

	_, err := chain(ctx, tx, false)
	require.Error(t, err)
}

func TestUnorderedTxDecorator_UnorderedTx_InvalidTTL(t *testing.T) {
	txm := unorderedtx.NewManager(t.TempDir())
	defer func() {
		require.NoError(t, txm.Close())
	}()

	txm.Start()

	suite := SetupTestSuite(t, false)

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(unorderedtx.DefaultmaxTimeoutDuration, txm, suite.accountKeeper.GetEnvironment()))

	tx, txBz := genUnorderedTx(t, true, time.Now().Add(unorderedtx.DefaultmaxTimeoutDuration+time.Second))
	ctx := sdk.Context{}.WithTxBytes(txBz).WithHeaderInfo(header.Info{Time: time.Now()})
	_, err := chain(ctx, tx, false)
	require.Error(t, err)
}

func TestUnorderedTxDecorator_UnorderedTx_AlreadyExists(t *testing.T) {
	txm := unorderedtx.NewManager(t.TempDir())
	defer func() {
		require.NoError(t, txm.Close())
	}()

	txm.Start()

	suite := SetupTestSuite(t, false)

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(unorderedtx.DefaultmaxTimeoutDuration, txm, suite.accountKeeper.GetEnvironment()))

	tx, txBz := genUnorderedTx(t, true, time.Now().Add(time.Minute))
	ctx := sdk.Context{}.WithTxBytes(txBz).WithHeaderInfo(header.Info{Time: time.Now()})

	txHash := sha256.Sum256(txBz)
	txm.Add(txHash, time.Now().Add(time.Minute))

	_, err := chain(ctx, tx, false)
	require.Error(t, err)
}

func TestUnorderedTxDecorator_UnorderedTx_ValidCheckTx(t *testing.T) {
	txm := unorderedtx.NewManager(t.TempDir())
	defer func() {
		require.NoError(t, txm.Close())
	}()

	txm.Start()

	suite := SetupTestSuite(t, false)

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(unorderedtx.DefaultmaxTimeoutDuration, txm, suite.accountKeeper.GetEnvironment()))

	tx, txBz := genUnorderedTx(t, true, time.Now().Add(time.Minute))
	ctx := sdk.Context{}.WithTxBytes(txBz).WithHeaderInfo(header.Info{Time: time.Now()}).WithExecMode(sdk.ExecModeCheck)

	_, err := chain(ctx, tx, false)
	require.NoError(t, err)
}

func TestUnorderedTxDecorator_UnorderedTx_ValidDeliverTx(t *testing.T) {
	txm := unorderedtx.NewManager(t.TempDir())
	defer func() {
		require.NoError(t, txm.Close())
	}()

	txm.Start()

	suite := SetupTestSuite(t, false)

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(unorderedtx.DefaultmaxTimeoutDuration, txm, suite.accountKeeper.GetEnvironment()))

	tx, txBz := genUnorderedTx(t, true, time.Now().Add(time.Minute))
	ctx := sdk.Context{}.WithTxBytes(txBz).WithHeaderInfo(header.Info{Time: time.Now()}).WithExecMode(sdk.ExecModeFinalize)

	_, err := chain(ctx, tx, false)
	require.NoError(t, err)

	txHash := sha256.Sum256(txBz)
	require.True(t, txm.Contains(txHash))
}

func genUnorderedTx(t *testing.T, unordered bool, timestamp time.Time) (sdk.Tx, []byte) {
	t.Helper()

	s := SetupTestSuite(t, true)
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	require.NoError(t, s.txBuilder.SetMsgs(msg))

	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)
	s.txBuilder.SetUnordered(unordered)
	s.txBuilder.SetTimeoutTimestamp(timestamp)

	privKeys, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(s.ctx, privKeys, accNums, accSeqs, s.ctx.ChainID(), signing.SignMode_SIGN_MODE_DIRECT)
	require.NoError(t, err)

	txBz, err := s.encCfg.TxConfig.TxEncoder()(tx)
	require.NoError(t, err)

	return tx, txBz
}
