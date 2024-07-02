package ante_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/auth/ante"
	"cosmossdk.io/x/auth/ante/unorderedtx"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

const gasConsumed = uint64(25)

func TestUnorderedTxDecorator_OrderedTx(t *testing.T) {
	txm := unorderedtx.NewManager(t.TempDir())
	defer func() {
		require.NoError(t, txm.Close())
	}()

	txm.Start()

	suite := SetupTestSuite(t, false)

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(unorderedtx.DefaultMaxUnOrderedTTL, txm, suite.accountKeeper.GetEnvironment(), ante.DefaultSha256Cost))

	tx, txBz := genUnorderedTx(t, false, 0)
	ctx := sdk.Context{}.WithTxBytes(txBz).WithBlockHeight(100)

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

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(unorderedtx.DefaultMaxUnOrderedTTL, txm, suite.accountKeeper.GetEnvironment(), ante.DefaultSha256Cost))

	tx, txBz := genUnorderedTx(t, true, 0)
	ctx := sdk.Context{}.WithTxBytes(txBz).WithBlockHeight(100)

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

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(unorderedtx.DefaultMaxUnOrderedTTL, txm, suite.accountKeeper.GetEnvironment(), ante.DefaultSha256Cost))

	tx, txBz := genUnorderedTx(t, true, 100+unorderedtx.DefaultMaxUnOrderedTTL+1)
	ctx := sdk.Context{}.WithTxBytes(txBz).WithBlockHeight(100)

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

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(unorderedtx.DefaultMaxUnOrderedTTL, txm, suite.accountKeeper.GetEnvironment(), ante.DefaultSha256Cost))

	tx, txBz := genUnorderedTx(t, true, 150)
	ctx := sdk.Context{}.WithTxBytes(txBz).WithBlockHeight(100).WithGasMeter(storetypes.NewGasMeter(gasConsumed))

	bz := [32]byte{}
	copy(bz[:], txBz[:32])
	txm.Add(bz, 150)

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

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(unorderedtx.DefaultMaxUnOrderedTTL, txm, suite.accountKeeper.GetEnvironment(), ante.DefaultSha256Cost))

	tx, txBz := genUnorderedTx(t, true, 150)
	ctx := sdk.Context{}.WithTxBytes(txBz).WithBlockHeight(100).WithExecMode(sdk.ExecModeCheck).WithGasMeter(storetypes.NewGasMeter(gasConsumed))

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

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(unorderedtx.DefaultMaxUnOrderedTTL, txm, suite.accountKeeper.GetEnvironment(), ante.DefaultSha256Cost))

	tx, txBz := genUnorderedTx(t, true, 150)
	ctx := sdk.Context{}.WithTxBytes(txBz).WithBlockHeight(100).WithExecMode(sdk.ExecModeFinalize).WithGasMeter(storetypes.NewGasMeter(gasConsumed))

	_, err := chain(ctx, tx, false)
	require.NoError(t, err)

	bz := [32]byte{}
	copy(bz[:], txBz[:32])

	require.True(t, txm.Contains(bz))
}

func genUnorderedTx(t *testing.T, unordered bool, ttl uint64) (sdk.Tx, []byte) {
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
	s.txBuilder.SetTimeoutHeight(ttl)

	privKeys, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(s.ctx, privKeys, accNums, accSeqs, s.ctx.ChainID(), signing.SignMode_SIGN_MODE_DIRECT)
	require.NoError(t, err)

	txBz, err := ante.TxIdentifier(ttl, tx)

	require.NoError(t, err)

	return tx, txBz[:]
}
