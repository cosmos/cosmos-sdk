package ante_test

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/header"
	gastestutil "cosmossdk.io/core/testing/gas"
	"cosmossdk.io/x/auth/ante"
	"cosmossdk.io/x/auth/ante/unorderedtx"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
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

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(unorderedtx.DefaultMaxTimeoutDuration, txm, suite.accountKeeper.GetEnvironment(), ante.DefaultSha256Cost))

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

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(unorderedtx.DefaultMaxTimeoutDuration, txm, suite.accountKeeper.GetEnvironment(), ante.DefaultSha256Cost))

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

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(unorderedtx.DefaultMaxTimeoutDuration, txm, suite.accountKeeper.GetEnvironment(), ante.DefaultSha256Cost))

	tx, txBz := genUnorderedTx(t, true, time.Now().Add(unorderedtx.DefaultMaxTimeoutDuration+time.Second))
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

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(unorderedtx.DefaultMaxTimeoutDuration, txm, suite.accountKeeper.GetEnvironment(), ante.DefaultSha256Cost))

	tx, txBz := genUnorderedTx(t, true, time.Now().Add(time.Minute))

	ctrl := gomock.NewController(t)
	mockMeter := gastestutil.NewMockMeter(ctrl)
	mockMeter.EXPECT().Consume(gasConsumed, "consume gas for calculating tx hash").Times(1)

	ctx := sdk.Context{}.WithTxBytes(txBz).WithHeaderInfo(header.Info{Time: time.Now()}).WithGasMeter(runtime.NewSDKGasMeter(
		mockMeter,
	))

	bz := [32]byte{}
	copy(bz[:], txBz[:32])
	txm.Add(bz, time.Now().Add(time.Minute))

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

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(unorderedtx.DefaultMaxTimeoutDuration, txm, suite.accountKeeper.GetEnvironment(), ante.DefaultSha256Cost))

	tx, txBz := genUnorderedTx(t, true, time.Now().Add(time.Minute))

	ctrl := gomock.NewController(t)
	mockMeter := gastestutil.NewMockMeter(ctrl)
	mockMeter.EXPECT().Consume(gasConsumed, "consume gas for calculating tx hash").Times(1)

	ctx := sdk.Context{}.WithTxBytes(txBz).WithHeaderInfo(header.Info{Time: time.Now()}).WithExecMode(sdk.ExecModeCheck).WithGasMeter(runtime.NewSDKGasMeter(
		mockMeter,
	))

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

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(unorderedtx.DefaultMaxTimeoutDuration, txm, suite.accountKeeper.GetEnvironment(), ante.DefaultSha256Cost))

	tx, txBz := genUnorderedTx(t, true, time.Now().Add(time.Minute))

	ctrl := gomock.NewController(t)
	mockMeter := gastestutil.NewMockMeter(ctrl)
	mockMeter.EXPECT().Consume(gasConsumed, "consume gas for calculating tx hash").Times(1)

	ctx := sdk.Context{}.WithTxBytes(txBz).WithHeaderInfo(header.Info{Time: time.Now()}).WithExecMode(sdk.ExecModeFinalize).WithGasMeter(runtime.NewSDKGasMeter(
		mockMeter,
	))

	_, err := chain(ctx, tx, false)
	require.NoError(t, err)

	bz := [32]byte{}
	copy(bz[:], txBz[:32])

	require.True(t, txm.Contains(bz))
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

	txBz, err := ante.TxIdentifier(uint64(timestamp.Unix()), tx)

	require.NoError(t, err)

	return tx, txBz[:]
}
