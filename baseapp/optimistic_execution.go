package baseapp

import (
	"bytes"
	"log"

	abci "github.com/cometbft/cometbft/abci/types"
)

type OptimisticExecutionInfo struct {
	// we could use generics here in the future to allow other types of req/resp
	completeSignal chan struct{}
	abortSignal    chan struct{}
	fn             func(*abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error)
	Aborted        bool
	Request        *abci.RequestFinalizeBlock
	Response       *abci.ResponseFinalizeBlock
	Error          error
}

func SetupOptimisticExecution(
	req *abci.RequestProcessProposal,
	fn func(*abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error),
) *OptimisticExecutionInfo {
	return &OptimisticExecutionInfo{
		completeSignal: make(chan struct{}),
		abortSignal:    make(chan struct{}),
		fn:             fn,
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
	go func() {
		log.Println("Running OE ✅")
		oe.Response, oe.Error = oe.fn(oe.Request)
		oe.completeSignal <- struct{}{}
	}()
}

// AbortIfNeeded
// If the request hash is not the same as the one in the OE, then abort the OE
// and wait for the abort to happen. Returns true if the OE was aborted.
func (oe *OptimisticExecutionInfo) AbortIfNeeded(reqHash []byte) bool {
	if !bytes.Equal(oe.Request.Hash, reqHash) {
		oe.abortSignal <- struct{}{}
		// wait for the abort to happen
		<-oe.completeSignal
		oe.Aborted = true
		return true
	}
	return false
}

// ShouldAbort must only be used in the fn passed to SetupOptimisticExecution to
// check if the OE was aborted and return as soon as possible.
// TODO: figure out a better name, maybe ReturnEarly?
func (oe *OptimisticExecutionInfo) ShouldAbort() bool {
	select {
	case <-oe.abortSignal:
		return true
	default:
		return false
	}
}

func (oe *OptimisticExecutionInfo) WaitResult() (*abci.ResponseFinalizeBlock, error) {
	<-oe.completeSignal
	return oe.Response, oe.Error
}
