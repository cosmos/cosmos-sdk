package oe

import (
	"context"
	"errors"
	"testing"

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	"github.com/stretchr/testify/assert"

	"cosmossdk.io/core/server"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
)

func testFinalizeBlock[T transaction.Tx](context.Context, *abci.FinalizeBlockRequest) (*server.BlockResponse, store.WriterMap, []T, error) {
	return nil, nil, nil, errors.New("test error")
}

func TestOptimisticExecution(t *testing.T) {
	oe := NewOptimisticExecution[transaction.Tx](log.NewNopLogger(), testFinalizeBlock)
	oe.Execute(&abci.ProcessProposalRequest{
		Hash: []byte("test"),
	})
	assert.True(t, oe.Initialized())

	resp, err := oe.WaitResult()
	assert.Equal(t, &FinalizeBlockResponse[transaction.Tx]{}, resp) // empty response
	assert.EqualError(t, err, "test error")

	assert.False(t, oe.AbortIfNeeded([]byte("test")))
	assert.True(t, oe.AbortIfNeeded([]byte("wrong_hash")))

	oe.Reset()
}
