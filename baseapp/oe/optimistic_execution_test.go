package oe

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"
)

func testFinalizeBlock(_ context.Context, _ *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	return nil, errors.New("test error")
}

func TestOptimisticExecution(t *testing.T) {
	oe := NewOptimisticExecution(log.NewNopLogger(), testFinalizeBlock)
	assert.True(t, oe.Enabled())
	oe.Execute(&abci.RequestProcessProposal{
		Hash:   []byte("test"),
		Height: 1,
	})
	assert.True(t, oe.Initialized())

	resp, err := oe.WaitResult()
	assert.Nil(t, resp)
	assert.EqualError(t, err, "test error")

	assert.False(t, oe.AbortIfNeeded([]byte("test"), 1))
	assert.True(t, oe.AbortIfNeeded([]byte("wrong_hash"), 1))

	oe.Reset()
}

// TestAbortIfNeededStaleHeight covers the case where ProcessProposal for H+1 is
// skipped and FinalizeBlock(H+1) arrives while OE still holds height H.
// Abort must treat the height mismatch as stale instead of a hash mismatch.
func TestAbortIfNeededStaleHeight(t *testing.T) {
	oe := NewOptimisticExecution(log.NewNopLogger(), func(_ context.Context, _ *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
		return &abci.ResponseFinalizeBlock{}, nil
	})

	oe.Execute(&abci.RequestProcessProposal{
		Hash:   []byte("hash-h"),
		Height: 10,
	})
	_, err := oe.WaitResult()
	require.NoError(t, err)
	require.True(t, oe.Initialized())

	// Same height, different hash → true hash mismatch abort.
	assert.True(t, oe.AbortIfNeeded([]byte("other-hash"), 10))

	// Re-execute for the stale-height path.
	oe.Reset()
	oe.Execute(&abci.RequestProcessProposal{
		Hash:   []byte("hash-h"),
		Height: 10,
	})
	_, err = oe.WaitResult()
	require.NoError(t, err)

	// Height H+1 with a different hash must abort as stale (not keep OE alive).
	assert.True(t, oe.AbortIfNeeded([]byte("hash-h+1"), 11))
}

// TestResetAfterSuccessfulConsume mirrors FinalizeBlock consuming OE then
// receiving FinalizeBlock for the next height without ProcessProposal.
func TestResetAfterSuccessfulConsume(t *testing.T) {
	var calls atomic.Int32
	oe := NewOptimisticExecution(log.NewNopLogger(), func(_ context.Context, req *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
		calls.Add(1)
		return &abci.ResponseFinalizeBlock{}, nil
	})

	oe.Execute(&abci.RequestProcessProposal{
		Hash:   []byte("hash-10"),
		Height: 10,
	})
	res, err := oe.WaitResult()
	require.NoError(t, err)
	require.NotNil(t, res)
	require.False(t, oe.AbortIfNeeded([]byte("hash-10"), 10))

	// Successful FinalizeBlock path must clear OE (as baseapp now does).
	oe.Reset()
	assert.False(t, oe.Initialized())

	// FinalizeBlock(H+1) with no ProcessProposal: OE is cleared, so no abort.
	assert.False(t, oe.Initialized())
	assert.Equal(t, int32(1), calls.Load())
}

func TestAbortIfNeededSameHeightWrongHash(t *testing.T) {
	done := make(chan struct{})
	oe := NewOptimisticExecution(log.NewNopLogger(), func(ctx context.Context, _ *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-done:
			return &abci.ResponseFinalizeBlock{}, nil
		case <-time.After(2 * time.Second):
			return &abci.ResponseFinalizeBlock{}, nil
		}
	})

	oe.Execute(&abci.RequestProcessProposal{
		Hash:   []byte("hash-a"),
		Height: 5,
	})

	assert.True(t, oe.AbortIfNeeded([]byte("hash-b"), 5))
	close(done)
	_, _ = oe.WaitResult()
	oe.Reset()
}
