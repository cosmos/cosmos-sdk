package oe

import (
	"bytes"
	"context"
	"encoding/hex"
	"math/rand"
	"sync"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/log"
)

// FinalizeBlockFunc is the function that is called by the Optimistic Execution (OE)
// to finalize the block. It has the same signature as the ABCI FinalizeBlock method.
type FinalizeBlockFunc func(context.Context, *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error)

// OptimisticExecution manages the optimistic execution of block finalization.
// It runs the FinalizeBlock function in a goroutine and provides mechanisms
// to abort the execution if the block proposal changes.
type OptimisticExecution struct {
	finalizeBlockFunc FinalizeBlockFunc // ABCI FinalizeBlock function to execute
	logger            log.Logger        // Logger for debugging and monitoring

	mtx         sync.Mutex                  // Protects concurrent access to OE state
	stopCh      chan struct{}               // Channel to signal when OE execution completes
	request     *abci.RequestFinalizeBlock  // The block finalization request being processed
	response    *abci.ResponseFinalizeBlock // The result of block finalization
	err         error                       // Any error that occurred during execution
	cancelFunc  func()                      // Function to cancel the execution context
	initialized bool                        // Whether the OE has been initialized with a request

	// debugging/testing options
	abortRate int // Percentage (0-100) of OE executions to randomly abort for testing
}

// NewOptimisticExecution creates a new OptimisticExecution instance.
// The execution is not started until Execute() is called.
func NewOptimisticExecution(logger log.Logger, fn FinalizeBlockFunc, opts ...func(*OptimisticExecution)) *OptimisticExecution {
	logger = logger.With(log.ModuleKey, "oe")
	oe := &OptimisticExecution{logger: logger, finalizeBlockFunc: fn}
	for _, opt := range opts {
		opt(oe)
	}
	return oe
}

// WithAbortRate sets the abort rate for testing purposes.
// The rate is a percentage (0-100) that determines how often OE should be randomly aborted.
// This option is for testing only and must not be used in production.
func WithAbortRate(rate int) func(*OptimisticExecution) {
	return func(oe *OptimisticExecution) {
		oe.abortRate = rate
	}
}

// Reset clears the OE state and invalidates the current execution.
// Must be called before starting a new optimistic execution.
func (oe *OptimisticExecution) Reset() {
	oe.mtx.Lock()
	defer oe.mtx.Unlock()
	oe.request = nil
	oe.response = nil
	oe.err = nil
	oe.initialized = false
}

// Enabled returns true if OptimisticExecution is available (not nil).
func (oe *OptimisticExecution) Enabled() bool {
	return oe != nil
}

// Initialized returns true if the OE has been initialized with a request
// and is either running or has completed execution.
func (oe *OptimisticExecution) Initialized() bool {
	if oe == nil {
		return false
	}
	oe.mtx.Lock()
	defer oe.mtx.Unlock()

	return oe.initialized
}

// Execute initializes the OE with the given proposal and starts block finalization
// in a separate goroutine. The execution can be aborted if the proposal changes.
func (oe *OptimisticExecution) Execute(req *abci.RequestProcessProposal) {
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

	oe.logger.Debug("OE started", "height", req.Height, "hash", hex.EncodeToString(req.Hash), "time", req.Time.String())
	ctx, cancel := context.WithCancel(context.Background())
	oe.cancelFunc = cancel
	oe.initialized = true

	go func() {
		start := time.Now()
		// Execute the block finalization function
		resp, err := oe.finalizeBlockFunc(ctx, oe.request)

		oe.mtx.Lock()

		executionTime := time.Since(start)
		oe.logger.Debug("OE finished", "duration", executionTime.String(), "height", oe.request.Height, "hash", hex.EncodeToString(oe.request.Hash))
		// Store the result and signal completion
		oe.response, oe.err = resp, err

		close(oe.stopCh)
		oe.mtx.Unlock()
	}()
}

// AbortIfNeeded checks if the OE should be aborted based on hash mismatch or test settings.
// Returns true if the OE was aborted, false otherwise.
func (oe *OptimisticExecution) AbortIfNeeded(reqHash []byte) bool {
	if oe == nil {
		return false
	}

	oe.mtx.Lock()
	defer oe.mtx.Unlock()

	if !bytes.Equal(oe.request.Hash, reqHash) {
		// Block proposal changed, abort the current execution
		oe.logger.Error("OE aborted due to hash mismatch", "oe_hash", hex.EncodeToString(oe.request.Hash), "req_hash", hex.EncodeToString(reqHash), "oe_height", oe.request.Height, "req_height", oe.request.Height)
		oe.cancelFunc()
		return true
	} else if oe.abortRate > 0 && rand.Intn(100) < oe.abortRate {
		// Random abort for testing purposes to simulate execution failures
		oe.cancelFunc()
		oe.logger.Error("OE aborted due to test abort rate")
		return true
	}

	return false
}

// Abort immediately cancels the OE execution and waits for the goroutine to finish.
func (oe *OptimisticExecution) Abort() {
	if oe == nil || oe.cancelFunc == nil {
		return
	}

	oe.cancelFunc()
	<-oe.stopCh
}

// WaitResult blocks until the OE execution completes and returns the finalization
// result and any error that occurred during execution.
func (oe *OptimisticExecution) WaitResult() (*abci.ResponseFinalizeBlock, error) {
	<-oe.stopCh
	return oe.response, oe.err
}
