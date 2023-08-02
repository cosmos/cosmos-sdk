package oe

import (
	"bytes"
	"math/rand"
	"sync"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/log"
)

type OptimisticExecution struct {
	mtx         sync.RWMutex
	stopCh      chan struct{}
	shouldAbort bool
	running     bool
	initialized bool

	// we could use generics here in the future to allow other types of req/resp
	fn            func(*abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error)
	request       *abci.RequestFinalizeBlock
	response      *abci.ResponseFinalizeBlock
	err           error
	executionTime time.Duration
	logger        log.Logger

	// debugging options
	abortRate int // number from 0 to 100
}

func NewOptimisticExecution(logger log.Logger, fn func(*abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error), opts ...func(*OptimisticExecution)) *OptimisticExecution {
	oe := &OptimisticExecution{logger: logger, fn: fn}
	for _, opt := range opts {
		opt(oe)
	}
	return oe
}

func WithAbortRate(rate int) func(*OptimisticExecution) {
	return func(oe *OptimisticExecution) {
		oe.abortRate = rate
	}
}

// Reset resets the OE context. Must be called whenever we want to invalidate
// the current OE. For example when on FinalizeBlock we want to process the
// block async, we run Reset() to make sure ShouldAbort() returns always false.
func (oe *OptimisticExecution) Reset() {
	oe.mtx.Lock()
	defer oe.mtx.Unlock()
	oe.request = nil
	oe.response = nil
	oe.err = nil
	oe.executionTime = 0
	oe.shouldAbort = false
	oe.running = false
	oe.initialized = false
}

func (oe *OptimisticExecution) Enabled() bool {
	return oe != nil
}

// Initialized returns true if the OE was initialized, meaning that it contains
// a request and it was run or it is running.
func (oe *OptimisticExecution) Initialized() bool {
	if oe == nil {
		return false
	}
	oe.mtx.RLock()
	defer oe.mtx.RUnlock()

	return oe.initialized
}

// Execute initializes the OE and starts it in a goroutine.
func (oe *OptimisticExecution) Execute(
	req *abci.RequestProcessProposal,
) {
	oe.mtx.Lock()
	defer oe.mtx.Unlock()

	oe.stopCh = make(chan struct{})
	oe.request = &abci.RequestFinalizeBlock{
		Txs:                req.Txs,
		DecidedLastCommit:  req.ProposedLastCommit,
		Misbehavior:        req.Misbehavior,
		Hash:               req.Hash,
		Height:             req.Height,
		Time:               req.Time,
		NextValidatorsHash: req.NextValidatorsHash,
		ProposerAddress:    req.ProposerAddress,
	}

	oe.logger.Debug("OE started")
	start := time.Now()
	oe.running = true
	oe.initialized = true

	go func() {
		resp, err := oe.fn(oe.request)
		oe.mtx.Lock()
		oe.executionTime = time.Since(start)
		oe.logger.Debug("OE finished", "duration", oe.executionTime.String())
		oe.response, oe.err = resp, err
		oe.running = false
		close(oe.stopCh)
		oe.mtx.Unlock()
	}()
}

// AbortIfNeeded aborts the OE if the request hash is not the same as the one in
// the running OE. Returns true if the OE was aborted.
func (oe *OptimisticExecution) AbortIfNeeded(reqHash []byte) bool {
	if oe == nil {
		return false
	}

	oe.mtx.Lock()
	defer oe.mtx.Unlock()

	if !bytes.Equal(oe.request.Hash, reqHash) {
		oe.logger.Debug("OE aborted due to hash mismatch", "oe_hash", oe.request.Hash, "req_hash", reqHash)
		oe.shouldAbort = true
	}

	// test abort rate
	if oe.abortRate > 0 && !oe.shouldAbort {
		oe.shouldAbort = rand.Intn(100) < oe.abortRate
		if oe.shouldAbort {
			oe.logger.Debug("OE aborted due to test abort rate")
		}
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
	if oe == nil {
		return false
	}

	oe.mtx.RLock()
	defer oe.mtx.RUnlock()
	return oe.shouldAbort
}

// Running returns true if the OE is still running.
func (oe *OptimisticExecution) Running() bool {
	if oe == nil {
		return false
	}

	oe.mtx.RLock()
	defer oe.mtx.RUnlock()
	return oe.running
}

// WaitResult waits for the OE to finish and returns the result.
func (oe *OptimisticExecution) WaitResult() (*abci.ResponseFinalizeBlock, error) {
	<-oe.stopCh
	return oe.response, oe.err
}
