package oe

import (
	"bytes"
	"math/rand"
	"sync"
	"time"

	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
)

type OptimisticExecution struct {
	mtx         sync.RWMutex
	stopCh      chan struct{}
	shouldAbort bool
	running     bool

	// we could use generics here in the future to allow other types of req/resp
	fn            func(*abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error)
	request       *abci.RequestFinalizeBlock
	response      *abci.ResponseFinalizeBlock
	err           error
	executionTime time.Duration
	logger        log.Logger
}

// Execute initializes the OE and starts it in a goroutine.
func Execute(
	req *abci.RequestProcessProposal,
	fn func(*abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error),
	logger log.Logger,
) *OptimisticExecution {
	oe := &OptimisticExecution{
		stopCh: make(chan struct{}),
		fn:     fn,
		request: &abci.RequestFinalizeBlock{
			Txs:                req.Txs,
			DecidedLastCommit:  req.ProposedLastCommit,
			Misbehavior:        req.Misbehavior,
			Hash:               req.Hash,
			Height:             req.Height,
			Time:               req.Time,
			NextValidatorsHash: req.NextValidatorsHash,
			ProposerAddress:    req.ProposerAddress,
		},
		logger: logger,
	}

	oe.logger.Debug("OE started")
	start := time.Now()
	oe.running = true
	go func() {
		resp, err := oe.fn(oe.request)
		oe.mtx.Lock()
		oe.executionTime = time.Since(start)
		oe.logger.Debug("OE finished", "duration", oe.executionTime)
		oe.response, oe.err = resp, err
		oe.running = false
		close(oe.stopCh)
		oe.mtx.Unlock()
	}()

	return oe
}

// AbortIfNeeded aborts the OE if the request hash is not the same as the one in
// the running OE. Returns true if the OE was aborted.
func (oe *OptimisticExecution) AbortIfNeeded(reqHash []byte) bool {
	oe.mtx.Lock()
	defer oe.mtx.Unlock()
	if rand.Intn(100) > 80 || !bytes.Equal(oe.request.Hash, reqHash) {
		oe.logger.Debug("OE aborted")
		oe.shouldAbort = true
	}
	return oe.shouldAbort
}

// Abort aborts the OE unconditionally.
func (oe *OptimisticExecution) Abort() {
	oe.mtx.Lock()
	defer oe.mtx.Unlock()
	oe.shouldAbort = true
}

// ShouldAbort must only be used in the fn passed to SetupOptimisticExecution to
// check if the OE was aborted and return as soon as possible.
func (oe *OptimisticExecution) ShouldAbort() bool {
	defer oe.mtx.RUnlock()
	oe.mtx.RLock()
	return oe.shouldAbort
}

// Running returns true if the OE is still running.
func (oe *OptimisticExecution) Running() bool {
	defer oe.mtx.RUnlock()
	oe.mtx.RLock()
	return oe.running
}

// WaitResult waits for the OE to finish and returns the result.
func (oe *OptimisticExecution) WaitResult() (*abci.ResponseFinalizeBlock, error) {
	<-oe.stopCh
	return oe.response, oe.err
}
