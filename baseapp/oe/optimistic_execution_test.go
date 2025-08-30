package oe

import (
	"context"
	"errors"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/assert"

	"cosmossdk.io/log"
)

// mockFinalizeBlock is a mock function that simulates the ABCI FinalizeBlock function
// for testing purposes. It always returns an error to test error handling.
func mockFinalizeBlock(_ context.Context, _ *abci.FinalizeBlockRequest) (*abci.FinalizeBlockResponse, error) {
	return nil, errors.New("test error")
}

// TestOptimisticExecution tests the basic functionality of the OptimisticExecution struct
// including initialization, execution, result waiting, and abort functionality.
func TestOptimisticExecution(t *testing.T) {
	// Create a new OptimisticExecution instance with mock function
	oe := NewOptimisticExecution(log.NewNopLogger(), mockFinalizeBlock)
	assert.True(t, oe.Enabled())

	// Execute optimistic execution with test data
	oe.Execute(&abci.ProcessProposalRequest{
		Hash: []byte("test"),
	})
	assert.True(t, oe.Initialized())

	// Wait for result and verify error handling
	resp, err := oe.WaitResult()
	assert.Nil(t, resp)
	assert.EqualError(t, err, "test error")

	// Test abort functionality with matching and non-matching hashes
	assert.False(t, oe.AbortIfNeeded([]byte("test")))
	assert.True(t, oe.AbortIfNeeded([]byte("wrong_hash")))

	// Reset the optimistic execution context
	oe.Reset()
}
