package ante_test

import (
	"crypto/sha256"
	"testing"
	"time"

	"cosmossdk.io/x/auth/ante"
	"github.com/stretchr/testify/require"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

func TestUnorderedTxManager_Close(t *testing.T) {
	txm := ante.NewUnorderedTxManager()
	txm.Start()

	require.NoError(t, txm.Close())
	require.Panics(t, func() { txm.Close() })
}

func TestUnorderedTxManager_SimpleSize(t *testing.T) {
	txm := ante.NewUnorderedTxManager()
	defer txm.Close()

	txm.Start()

	txm.Add([32]byte{0xFF}, 100)
	txm.Add([32]byte{0xAA}, 100)
	txm.Add([32]byte{0xCC}, 100)

	require.Equal(t, 3, txm.Size())
}

func TestUnorderedTxManager_SimpleContains(t *testing.T) {
	txm := ante.NewUnorderedTxManager()
	defer txm.Close()

	txm.Start()

	for i := 0; i < 10; i++ {
		txHash := [32]byte{byte(i)}
		txm.Add(txHash, 100)
		require.True(t, txm.Contains(txHash))
	}

	for i := 10; i < 20; i++ {
		txHash := [32]byte{byte(i)}
		require.False(t, txm.Contains(txHash))
	}
}

func TestUnorderedTxManager_Flow(t *testing.T) {
	txm := ante.NewUnorderedTxManager()
	defer txm.Close()

	txm.Start()

	// Seed the manager with a txs, some of which should eventually be purged and
	// the others will remain. Txs with TTL less than or equal to 50 should be purged.
	for i := 1; i <= 100; i++ {
		txHash := [32]byte{byte(i)}

		if i <= 50 {
			txm.Add(txHash, uint64(i))
		} else {
			txm.Add(txHash, 100)
		}
	}

	// start a goroutine that mimics new blocks being made every 500ms
	doneBlockCh := make(chan bool)
	go func() {
		ticker := time.NewTicker(time.Millisecond * 500)
		defer ticker.Stop()

		var (
			height uint64 = 1
			i             = 101
		)
		for range ticker.C {
			txm.OnNewBlock(height)
			height++

			if height > 51 {
				doneBlockCh <- true
				return
			} else {
				txm.Add([32]byte{byte(i)}, 50)
			}
		}
	}()

	// Eventually all the txs that should be expired by block 50 should be purged.
	// The remaining txs should remain.
	require.Eventually(
		t,
		func() bool {
			return txm.Size() == 50
		},
		2*time.Minute,
		5*time.Second,
	)

	<-doneBlockCh
}

func TestUnorderedTxDecorator_OrderedTx(t *testing.T) {
	txm := ante.NewUnorderedTxManager()
	defer txm.Close()

	txm.Start()

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(ante.DefaultMaxUnOrderedTTL, txm))

	tx, txBz := genUnorderedTx(t, false, 0)
	ctx := sdk.Context{}.WithTxBytes(txBz).WithBlockHeight(100)

	_, err := chain(ctx, tx, false)
	require.NoError(t, err)
}

func TestUnorderedTxDecorator_UnorderedTx_NoTTL(t *testing.T) {
	txm := ante.NewUnorderedTxManager()
	defer txm.Close()

	txm.Start()

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(ante.DefaultMaxUnOrderedTTL, txm))

	tx, txBz := genUnorderedTx(t, true, 0)
	ctx := sdk.Context{}.WithTxBytes(txBz).WithBlockHeight(100)

	_, err := chain(ctx, tx, false)
	require.Error(t, err)
}

func TestUnorderedTxDecorator_UnorderedTx_InvalidTTL(t *testing.T) {
	txm := ante.NewUnorderedTxManager()
	defer txm.Close()

	txm.Start()

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(ante.DefaultMaxUnOrderedTTL, txm))

	tx, txBz := genUnorderedTx(t, true, 100+ante.DefaultMaxUnOrderedTTL+1)
	ctx := sdk.Context{}.WithTxBytes(txBz).WithBlockHeight(100)

	_, err := chain(ctx, tx, false)
	require.Error(t, err)
}

func TestUnorderedTxDecorator_UnorderedTx_AlreadyExists(t *testing.T) {
	txm := ante.NewUnorderedTxManager()
	defer txm.Close()

	txm.Start()

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(ante.DefaultMaxUnOrderedTTL, txm))

	tx, txBz := genUnorderedTx(t, true, 150)
	ctx := sdk.Context{}.WithTxBytes(txBz).WithBlockHeight(100)

	txHash := sha256.Sum256(txBz)
	txm.Add(txHash, 150)

	_, err := chain(ctx, tx, false)
	require.Error(t, err)
}

func TestUnorderedTxDecorator_UnorderedTx_ValidCheckTx(t *testing.T) {
	txm := ante.NewUnorderedTxManager()
	defer txm.Close()

	txm.Start()

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(ante.DefaultMaxUnOrderedTTL, txm))

	tx, txBz := genUnorderedTx(t, true, 150)
	ctx := sdk.Context{}.WithTxBytes(txBz).WithBlockHeight(100).WithExecMode(sdk.ExecModeCheck)

	_, err := chain(ctx, tx, false)
	require.NoError(t, err)
}

func TestUnorderedTxDecorator_UnorderedTx_ValidDeliverTx(t *testing.T) {
	txm := ante.NewUnorderedTxManager()
	defer txm.Close()

	txm.Start()

	chain := sdk.ChainAnteDecorators(ante.NewUnorderedTxDecorator(ante.DefaultMaxUnOrderedTTL, txm))

	tx, txBz := genUnorderedTx(t, true, 150)
	ctx := sdk.Context{}.WithTxBytes(txBz).WithBlockHeight(100).WithExecMode(sdk.ExecModeFinalize)

	_, err := chain(ctx, tx, false)
	require.NoError(t, err)

	txHash := sha256.Sum256(txBz)
	require.True(t, txm.Contains(txHash))
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

	txBz, err := s.encCfg.TxConfig.TxEncoder()(tx)
	require.NoError(t, err)

	return tx, txBz
}
