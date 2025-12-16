package txnrunner

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
	return nil, nil
}

// TestNewDefaultRunner tests the constructor
func TestNewDefaultRunner(t *testing.T) {
	runner := NewDefaultRunner(nil)

	require.NotNil(t, runner)
	require.Nil(t, runner.txDecoder)
}

// TestDefaultRunner_Run_Success tests successful execution of transactions
func TestDefaultRunner_Run_Success(t *testing.T) {
	decoder := mockTxDecoder
	runner := NewDefaultRunner(decoder)

	txs := [][]byte{
		{0x01, 0x02, 0x03},
		{0x04, 0x05, 0x06},
		{0x07, 0x08, 0x09},
	}

	executionCount := atomic.Int32{}
	deliverTx := func(tx []byte, ms storetypes.MultiStore, txIndex int, cache map[string]any) *abci.ExecTxResult {
		executionCount.Add(1)
		return &abci.ExecTxResult{
			Code: 0,
			Data: tx,
		}
	}

	ctx := context.Background()
	results, err := runner.Run(ctx, nil, txs, deliverTx)

	require.NoError(t, err)
	require.Len(t, results, len(txs))
	require.Equal(t, int32(len(txs)), executionCount.Load())

	for i, result := range results {
		require.Equal(t, uint32(0), result.Code)
		require.Equal(t, txs[i], result.Data)
	}
}

// TestDefaultRunner_Run_EmptyTxs tests execution with no transactions
func TestDefaultRunner_Run_EmptyTxs(t *testing.T) {
	decoder := mockTxDecoder
	runner := NewDefaultRunner(decoder)

	deliverTx := func(tx []byte, ms storetypes.MultiStore, txIndex int, cache map[string]any) *abci.ExecTxResult {
		t.Fatal("deliverTx should not be called for empty txs")
		return nil
	}

	ctx := context.Background()
	results, err := runner.Run(ctx, nil, [][]byte{}, deliverTx)

	require.NoError(t, err)
	require.Empty(t, results)
}

// TestDefaultRunner_Run_InvalidTx tests handling of invalid transactions
func TestDefaultRunner_Run_InvalidTx(t *testing.T) {
	decoder := mockTxDecoder
	runner := NewDefaultRunner(decoder)

	txs := [][]byte{
		{0x01, 0x02, 0x03}, // valid
		{0xFF, 0xFF, 0xFF}, // invalid (0xFF marker)
		{0x07, 0x08, 0x09}, // valid
	}

	validTxCount := atomic.Int32{}
	deliverTx := func(tx []byte, ms storetypes.MultiStore, txIndex int, cache map[string]any) *abci.ExecTxResult {
		validTxCount.Add(1)
		return &abci.ExecTxResult{Code: 0}
	}

	ctx := context.Background()
	results, err := runner.Run(ctx, nil, txs, deliverTx)

	require.NoError(t, err)
	require.Len(t, results, len(txs))
	// Only 2 valid transactions should be executed
	require.Equal(t, int32(2), validTxCount.Load())

	// The invalid tx should get an error response
	require.Equal(t, sdkerrors.ErrTxDecode.ABCICode(), results[1].Code)
}

// TestDefaultRunner_Run_ContextCancellation tests that execution stops on context cancellation
func TestDefaultRunner_Run_ContextCancellation(t *testing.T) {
	decoder := mockTxDecoder
	runner := NewDefaultRunner(decoder)

	txs := [][]byte{
		{0x01, 0x02, 0x03},
		{0x04, 0x05, 0x06},
		{0x07, 0x08, 0x09},
		{0x0A, 0x0B, 0x0C},
	}

	ctx, cancel := context.WithCancel(context.Background())

	executionCount := atomic.Int32{}
	deliverTx := func(tx []byte, ms storetypes.MultiStore, txIndex int, cache map[string]any) *abci.ExecTxResult {
		count := executionCount.Add(1)
		// Cancel after second transaction
		if count == 2 {
			cancel()
		}
		return &abci.ExecTxResult{Code: 0}
	}

	_, err := runner.Run(ctx, nil, txs, deliverTx)

	require.Error(t, err)
	require.Equal(t, context.Canceled, err)
	// Results may be nil or partial depending on when cancellation occurs
	// The key assertion is that execution was stopped
	require.LessOrEqual(t, executionCount.Load(), int32(len(txs)))
}

// TestDefaultRunner_Run_MultiStoreIsNil tests that nil multistore is handled correctly
func TestDefaultRunner_Run_MultiStoreIsNil(t *testing.T) {
	decoder := mockTxDecoder
	runner := NewDefaultRunner(decoder)

	txs := [][]byte{{0x01}}

	deliverTx := func(tx []byte, ms storetypes.MultiStore, txIndex int, cache map[string]any) *abci.ExecTxResult {
		require.Nil(t, ms, "multistore should be nil for DefaultRunner")
		require.Nil(t, cache, "cache should be nil for DefaultRunner")
		return &abci.ExecTxResult{Code: 0}
	}

	ctx := context.Background()
	results, err := runner.Run(ctx, nil, txs, deliverTx)

	require.NoError(t, err)
	require.Len(t, results, 1)
}
