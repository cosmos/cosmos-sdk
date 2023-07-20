package baseapp

import (
	"bytes"
	"log"
	"math/rand"
	"sync"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
)

type OptimisticExecutionInfo struct {
	mtx         sync.RWMutex
	stopCh      chan struct{}
	shouldAbort bool
	running     bool

	// we could use generics here in the future to allow other types of req/resp
	fn            func(*abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error)
	Request       *abci.RequestFinalizeBlock
	Response      *abci.ResponseFinalizeBlock
	Error         error
	executionTime time.Duration
}

func SetupOptimisticExecution(
	req *abci.RequestProcessProposal,
	fn func(*abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error),
) *OptimisticExecutionInfo {
	return &OptimisticExecutionInfo{
		stopCh: make(chan struct{}),
		fn:     fn,
		Request: &abci.RequestFinalizeBlock{
			Txs:                req.Txs,
			DecidedLastCommit:  req.ProposedLastCommit,
			Misbehavior:        req.Misbehavior,
			Hash:               req.Hash,
			Height:             req.Height,
			Time:               req.Time,
			NextValidatorsHash: req.NextValidatorsHash,
			ProposerAddress:    req.ProposerAddress,
		},
	}
}

func (oe *OptimisticExecutionInfo) Execute() {
	log.Println("Start OE ✅")
	start := time.Now()
	oe.running = true
	go func() {
		resp, err := oe.fn(oe.Request)
		oe.mtx.Lock()
		oe.executionTime = time.Since(start)
		oe.Response, oe.Error = resp, err
		oe.running = false
		close(oe.stopCh)
		oe.mtx.Unlock()
	}()
}

// AbortIfNeeded
// If the request hash is not the same as the one in the OE, then abort the OE
// and wait for the abort to happen. Returns true if the OE was aborted.
func (oe *OptimisticExecutionInfo) AbortIfNeeded(reqHash []byte) bool {
	oe.mtx.Lock()
	defer oe.mtx.Unlock()
	if rand.Intn(100) > 80 || !bytes.Equal(oe.Request.Hash, reqHash) {
		log.Println("OE aborted ❌")
		oe.shouldAbort = true
	}
	return oe.shouldAbort
}

func (oe *OptimisticExecutionInfo) Abort() {
	oe.mtx.Lock()
	defer oe.mtx.Unlock()
	oe.shouldAbort = true
}

// ShouldAbort must only be used in the fn passed to SetupOptimisticExecution to
// check if the OE was aborted and return as soon as possible.
// TODO: figure out a better name, maybe ReturnEarly?
func (oe *OptimisticExecutionInfo) ShouldAbort() bool {
	defer oe.mtx.RUnlock()
	oe.mtx.RLock()
	return oe.shouldAbort
}

func (oe *OptimisticExecutionInfo) Running() bool {
	defer oe.mtx.RUnlock()
	oe.mtx.RLock()
	return oe.running
}

func (oe *OptimisticExecutionInfo) WaitResult() (*abci.ResponseFinalizeBlock, error) {
	<-oe.stopCh
	log.Println("OE took ⏱", oe.executionTime)
	return oe.Response, oe.Error
}
