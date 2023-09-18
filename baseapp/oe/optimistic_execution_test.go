package oe

import (
	"context"
	"errors"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/assert"

	"cosmossdk.io/log"
)

func testFinalizeBlock(_ context.Context, _ *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	return nil, errors.New("test error")
}

func TestOptimisticExecution(t *testing.T) {
	oe := NewOptimisticExecution(log.NewNopLogger(), testFinalizeBlock)
	assert.True(t, oe.Enabled())
	oe.Execute(&abci.RequestProcessProposal{
		Hash: []byte("test"),
	})
	assert.True(t, oe.Initialized())

	resp, err := oe.WaitResult()
	assert.Nil(t, resp)
	assert.EqualError(t, err, "test error")

	assert.False(t, oe.AbortIfNeeded([]byte("test")))
	assert.True(t, oe.AbortIfNeeded([]byte("wrong_hash")))

	oe.Reset()
}
