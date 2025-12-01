package blockstm

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp/txnrunner"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Mock TxDecoder for testing
func mockTxDecoder(txBytes []byte) (sdk.Tx, error) {
	if len(txBytes) == 0 {
		return nil, errors.New("empty tx")
	}
	// Valid transaction if first byte is not 0xFF
	if txBytes[0] == 0xFF {
		return nil, errors.New("invalid tx")
	}
	return &mockTx{txBytes: txBytes}, nil
}

const TestCoinDenom = "stake"

func testCoinDenomFunc(ms storetypes.MultiStore) string {
	return TestCoinDenom
}

type mockTx struct {
	txBytes []byte
}

func (m *mockTx) GetMsgs() []sdk.Msg {
	return nil
}

func (m *mockTx) GetMsgsV2() ([]protov2.Message, error) {
	return nil, nil
}

func (m *mockTx) ValidateBasic() error {
	return nil
}

type mockFeeTx struct {
	mockTx
	feePayer sdk.AccAddress
}

func (m *mockFeeTx) FeePayer() []byte {
	return m.feePayer
}

func (m *mockFeeTx) GetFee() sdk.Coins {
	return nil
}

func (m *mockFeeTx) GetGas() uint64 {
	return 0
}

func mockTxDecoderWithFeeTx(txBytes []byte) (sdk.Tx, error) {
	if len(txBytes) == 0 {
		return nil, errors.New("empty tx")
	}
	if txBytes[0] == 0xFF {
		return nil, errors.New("invalid tx")
	}
	// Use the tx bytes as the fee payer address for testing
	feePayer := sdk.AccAddress(txBytes[:min(len(txBytes), 20)])
	return &mockFeeTx{
		mockTx:   mockTx{txBytes: txBytes},
		feePayer: feePayer,
	}, nil
}

// TestNewSTMRunner tests the STMRunner constructor
func TestNewSTMRunner(t *testing.T) {
	decoder := mockTxDecoder
	stores := []storetypes.StoreKey{StoreKeyAuth, StoreKeyBank}
	workers := 4
	estimate := true

	runner := NewSTMRunner(decoder, stores, workers, estimate, testCoinDenomFunc)

	require.NotNil(t, runner)
	require.NotNil(t, runner.txDecoder)
	require.Equal(t, stores, runner.stores)
	require.Equal(t, workers, runner.workers)
	require.Equal(t, estimate, runner.estimate)
	require.Equal(t, TestCoinDenom, runner.coinDenom(nil))
}

// TestSTMRunner_Run_EmptyBlock tests STMRunner with empty block
func TestSTMRunner_Run_EmptyBlock(t *testing.T) {
	decoder := mockTxDecoder
	stores := []storetypes.StoreKey{StoreKeyAuth, StoreKeyBank}
	runner := NewSTMRunner(decoder, stores, 4, false, testCoinDenomFunc)

	ctx := context.Background()
	ms := msWrapper{NewMultiMemDB(map[storetypes.StoreKey]int{
		StoreKeyAuth: 0,
		StoreKeyBank: 1,
	})}

	deliverTx := func(tx []byte, ms storetypes.MultiStore, txIndex int, cache map[string]any) *abci.ExecTxResult {
		t.Fatal("deliverTx should not be called for empty block")
		return nil
	}

	results, err := runner.Run(ctx, ms, [][]byte{}, deliverTx)

	require.NoError(t, err)
	require.Nil(t, results)
}

// TestSTMRunner_Run_WithoutEstimation tests STMRunner without pre-estimation
func TestSTMRunner_Run_WithoutEstimation(t *testing.T) {
	decoder := mockTxDecoder
	stores := []storetypes.StoreKey{StoreKeyAuth, StoreKeyBank}
	runner := NewSTMRunner(decoder, stores, 2, false, testCoinDenomFunc)

	ctx := context.Background()
	storeIndex := map[storetypes.StoreKey]int{
		StoreKeyAuth: 0,
		StoreKeyBank: 1,
	}
	ms := msWrapper{NewMultiMemDB(storeIndex)}

	txs := [][]byte{
		{0x01},
		{0x02},
		{0x03},
	}

	executionCount := atomic.Int32{}
	deliverTx := func(tx []byte, ms storetypes.MultiStore, txIndex int, cache map[string]any) *abci.ExecTxResult {
		executionCount.Add(1)
		require.NotNil(t, ms)
		return &abci.ExecTxResult{Code: 0}
	}

	results, err := runner.Run(ctx, ms, txs, deliverTx)

	require.NoError(t, err)
	require.Len(t, results, len(txs))
	// STM may execute transactions multiple times due to conflicts
	require.True(t, executionCount.Load() >= int32(len(txs)))
}

// TestSTMRunner_Run_WithEstimation tests STMRunner with pre-estimation enabled
func TestSTMRunner_Run_WithEstimation(t *testing.T) {
	decoder := mockTxDecoderWithFeeTx
	stores := []storetypes.StoreKey{StoreKeyAuth, StoreKeyBank}
	runner := NewSTMRunner(decoder, stores, 2, true, testCoinDenomFunc)

	ctx := context.Background()
	storeIndex := map[storetypes.StoreKey]int{
		StoreKeyAuth: 0,
		StoreKeyBank: 1,
	}
	ms := msWrapper{NewMultiMemDB(storeIndex)}

	// Create transactions with valid structure for estimation
	addr1 := []byte("addr1")
	addr2 := []byte("addr2")
	txs := [][]byte{
		append(addr1, 0x01),
		append(addr2, 0x02),
	}

	executionCount := atomic.Int32{}
	deliverTx := func(tx []byte, ms storetypes.MultiStore, txIndex int, cache map[string]any) *abci.ExecTxResult {
		executionCount.Add(1)
		require.NotNil(t, ms)
		return &abci.ExecTxResult{Code: 0}
	}

	results, err := runner.Run(ctx, ms, txs, deliverTx)

	require.NoError(t, err)
	require.Len(t, results, len(txs))
}

// TestSTMRunner_Run_IncarnationCache tests that incarnation cache is properly managed
func TestSTMRunner_Run_IncarnationCache(t *testing.T) {
	decoder := mockTxDecoder
	stores := []storetypes.StoreKey{StoreKeyAuth, StoreKeyBank}
	runner := NewSTMRunner(decoder, stores, 2, false, testCoinDenomFunc)

	ctx := context.Background()
	storeIndex := map[storetypes.StoreKey]int{
		StoreKeyAuth: 0,
		StoreKeyBank: 1,
	}
	ms := msWrapper{NewMultiMemDB(storeIndex)}

	txs := [][]byte{
		{0x01},
		{0x02},
	}

	cacheReceived := make([]bool, len(txs))
	deliverTx := func(tx []byte, ms storetypes.MultiStore, txIndex int, cache map[string]any) *abci.ExecTxResult {
		if cache != nil {
			cacheReceived[txIndex] = true
		}
		return &abci.ExecTxResult{Code: 0}
	}

	results, err := runner.Run(ctx, ms, txs, deliverTx)

	require.NoError(t, err)
	require.Len(t, results, len(txs))
	// Each transaction should receive a cache (even if empty)
	for i, received := range cacheReceived {
		require.True(t, received, "transaction %d should receive cache", i)
	}
}

// TestSTMRunner_Run_StoreIndexMapping tests that store keys are correctly mapped
func TestSTMRunner_Run_StoreIndexMapping(t *testing.T) {
	decoder := mockTxDecoder
	stores := []storetypes.StoreKey{StoreKeyAuth, StoreKeyBank}
	runner := NewSTMRunner(decoder, stores, 2, false, testCoinDenomFunc)

	ctx := context.Background()
	storeIndex := map[storetypes.StoreKey]int{
		StoreKeyAuth: 0,
		StoreKeyBank: 1,
	}
	ms := msWrapper{NewMultiMemDB(storeIndex)}

	txs := [][]byte{{0x01}}

	deliverTx := func(tx []byte, ms storetypes.MultiStore, txIndex int, cache map[string]any) *abci.ExecTxResult {
		// Verify we can access both stores
		authStore := ms.GetKVStore(StoreKeyAuth)
		bankStore := ms.GetKVStore(StoreKeyBank)
		require.NotNil(t, authStore)
		require.NotNil(t, bankStore)
		return &abci.ExecTxResult{Code: 0}
	}

	results, err := runner.Run(ctx, ms, txs, deliverTx)

	require.NoError(t, err)
	require.Len(t, results, 1)
}

// TestSTMRunner_Run_ContextCancellation tests context cancellation for STMRunner
func TestSTMRunner_Run_ContextCancellation(t *testing.T) {
	decoder := mockTxDecoder
	stores := []storetypes.StoreKey{StoreKeyAuth, StoreKeyBank}
	runner := NewSTMRunner(decoder, stores, 2, false, testCoinDenomFunc)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	storeIndex := map[storetypes.StoreKey]int{
		StoreKeyAuth: 0,
		StoreKeyBank: 1,
	}
	ms := msWrapper{NewMultiMemDB(storeIndex)}

	// Create a large block to ensure context timeout
	txs := make([][]byte, 1000)
	for i := range txs {
		txs[i] = []byte{byte(i % 256)}
	}

	deliverTx := func(tx []byte, ms storetypes.MultiStore, txIndex int, cache map[string]any) *abci.ExecTxResult {
		time.Sleep(1 * time.Millisecond) // Slow down execution
		return &abci.ExecTxResult{Code: 0}
	}

	results, err := runner.Run(ctx, ms, txs, deliverTx)

	// Should error due to context cancellation
	require.Error(t, err)
	require.Nil(t, results)
}

// TestPreEstimates tests the preEstimates function
func TestPreEstimates(t *testing.T) {
	t.Run("empty transactions", func(t *testing.T) {
		decoder := mockTxDecoderWithFeeTx
		memTxs, estimates := preEstimates([][]byte{}, 2, 0, 1, "stake", decoder)

		require.Empty(t, memTxs)
		require.Empty(t, estimates)
	})

	t.Run("valid transactions with estimation", func(t *testing.T) {
		decoder := mockTxDecoderWithFeeTx

		// Create test addresses
		addr1 := sdk.AccAddress([]byte("address1"))
		addr2 := sdk.AccAddress([]byte("address2"))

		txs := [][]byte{
			append(addr1, 0x01),
			append(addr2, 0x02),
		}

		memTxs, estimates := preEstimates(txs, 2, 0, 1, "stake", decoder)

		require.Len(t, memTxs, len(txs))
		require.Len(t, estimates, len(txs))

		// Check that estimates are generated for valid transactions
		for i, estimate := range estimates {
			if estimate != nil {
				// Should have auth store estimate (index 0)
				require.Contains(t, estimate, 0, "transaction %d should have auth store estimate", i)
				// Should have bank store estimate (index 1)
				require.Contains(t, estimate, 1, "transaction %d should have bank store estimate", i)
			}
		}
	})

	t.Run("invalid transactions", func(t *testing.T) {
		decoder := mockTxDecoderWithFeeTx

		txs := [][]byte{
			{0xFF, 0xFF}, // invalid
			{0x01, 0x02}, // valid
		}

		memTxs, estimates := preEstimates(txs, 2, 0, 1, "stake", decoder)

		require.Len(t, memTxs, len(txs))
		require.Len(t, estimates, len(txs))

		// Invalid transaction should not have memTx or estimates
		require.Nil(t, memTxs[0])
		require.Nil(t, estimates[0])

		// Valid transaction should have memTx
		require.NotNil(t, memTxs[1])
	})

	t.Run("parallel processing with multiple workers", func(t *testing.T) {
		decoder := mockTxDecoderWithFeeTx

		// Create many transactions
		txs := make([][]byte, 100)
		for i := range txs {
			addr := sdk.AccAddress([]byte{byte(i)})
			txs[i] = append(addr, byte(i))
		}

		memTxs, estimates := preEstimates(txs, 4, 0, 1, "stake", decoder)

		require.Len(t, memTxs, len(txs))
		require.Len(t, estimates, len(txs))
	})

	t.Run("non-FeeTx transactions", func(t *testing.T) {
		// Use decoder that doesn't return FeeTx
		decoder := mockTxDecoder

		txs := [][]byte{
			{0x01, 0x02},
			{0x03, 0x04},
		}

		memTxs, estimates := preEstimates(txs, 2, 0, 1, "stake", decoder)

		require.Len(t, memTxs, len(txs))
		require.Len(t, estimates, len(txs))

		// Non-FeeTx should not have estimates
		for _, estimate := range estimates {
			require.Nil(t, estimate)
		}
	})
}

// TestPreEstimates_KeyEncoding tests that account and balance keys are correctly encoded
func TestPreEstimates_KeyEncoding(t *testing.T) {
	decoder := mockTxDecoderWithFeeTx

	addr := sdk.AccAddress([]byte("testaddress12345"))
	tx := append(addr, 0x01)

	memTxs, estimates := preEstimates([][]byte{tx}, 1, 0, 1, "stake", decoder)

	require.Len(t, memTxs, 1)
	require.Len(t, estimates, 1)

	if estimates[0] != nil {
		// Verify account key encoding
		authEstimate := estimates[0][0]
		require.NotEmpty(t, authEstimate)

		// The key should be properly encoded
		expectedAccKey, err := collections.EncodeKeyWithPrefix(
			collections.NewPrefix(1),
			sdk.AccAddressKey,
			addr,
		)
		require.NoError(t, err)
		require.Contains(t, authEstimate, expectedAccKey)

		// Verify balance key encoding
		bankEstimate := estimates[0][1]
		require.NotEmpty(t, bankEstimate)

		expectedBalanceKey, err := collections.EncodeKeyWithPrefix(
			collections.NewPrefix(2),
			collections.PairKeyCodec(sdk.AccAddressKey, collections.StringKey),
			collections.Join(addr, "stake"),
		)
		require.NoError(t, err)
		require.Contains(t, bankEstimate, expectedBalanceKey)
	}
}

// TestTxRunnerInterface tests that both runners implement TxRunner interface
func TestTxRunnerInterface(t *testing.T) {
	decoder := mockTxDecoder

	var _ sdk.TxRunner = NewSTMRunner(decoder, []storetypes.StoreKey{}, 1, false, testCoinDenomFunc)
}

// TestSTMRunner_Integration tests integration between STMRunner and actual block execution
func TestSTMRunner_Integration(t *testing.T) {
	decoder := mockTxDecoder
	stores := []storetypes.StoreKey{StoreKeyAuth, StoreKeyBank}
	runner := NewSTMRunner(decoder, stores, 4, false, testCoinDenomFunc)

	ctx := context.Background()
	storeIndex := map[storetypes.StoreKey]int{
		StoreKeyAuth: 0,
		StoreKeyBank: 1,
	}
	ms := msWrapper{NewMultiMemDB(storeIndex)}

	// Create a mock block with actual transactions
	blk := testBlock(20, 10)

	// Use STMRunner to execute
	var results []*abci.ExecTxResult
	deliverTx := func(tx []byte, mstore storetypes.MultiStore, txIndex int, cache map[string]any) *abci.ExecTxResult {
		// Execute using the mock block's transaction logic
		if txIndex < blk.Size() {
			// Convert multistore wrapper to MultiStore for block execution
			if wrapper, ok := mstore.(msWrapper); ok {
				blk.ExecuteTx(TxnIndex(txIndex), wrapper.MultiStore)
			}
		}
		return &abci.ExecTxResult{Code: 0}
	}

	// Create raw tx bytes for the runner
	txs := make([][]byte, blk.Size())
	for i := range txs {
		txs[i] = []byte{byte(i)}
	}

	results, err := runner.Run(ctx, ms, txs, deliverTx)

	require.NoError(t, err)
	require.Len(t, results, blk.Size())
}

// TestRunnerComparison tests that both DefaultRunner and STMRunner can execute successfully
func TestRunnerComparison(t *testing.T) {
	decoder := mockTxDecoder

	txs := [][]byte{
		{0x01, 0x02},
		{0x03, 0x04},
		{0x05, 0x06},
	}

	executionCount := atomic.Int32{}
	deliverTx := func(tx []byte, _ storetypes.MultiStore, txIndex int, cache map[string]any) *abci.ExecTxResult {
		executionCount.Add(1)
		return &abci.ExecTxResult{Code: 0, Data: tx}
	}

	ctx := context.Background()

	// Test DefaultRunner
	t.Run("DefaultRunner", func(t *testing.T) {
		runner := txnrunner.NewDefaultRunner(decoder)
		executionCount.Store(0)

		results, err := runner.Run(ctx, nil, txs, deliverTx)

		require.NoError(t, err)
		require.Len(t, results, len(txs))
		require.Equal(t, int32(len(txs)), executionCount.Load())
	})

	// Test STMRunner
	t.Run("STMRunner", func(t *testing.T) {
		stores := []storetypes.StoreKey{StoreKeyAuth, StoreKeyBank}
		runner := NewSTMRunner(decoder, stores, 2, false, testCoinDenomFunc)
		storeIndex := map[storetypes.StoreKey]int{
			StoreKeyAuth: 0,
			StoreKeyBank: 1,
		}
		ms := msWrapper{NewMultiMemDB(storeIndex)}
		executionCount.Store(0)

		results, err := runner.Run(ctx, ms, txs, deliverTx)

		require.NoError(t, err)
		require.Len(t, results, len(txs))
		// STM may execute more times due to conflicts
		require.True(t, executionCount.Load() >= int32(len(txs)))
	})
}
