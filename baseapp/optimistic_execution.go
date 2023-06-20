package baseapp

import abci "github.com/cometbft/cometbft/abci/types"

type OptimisticExecutionInfo struct {
	Completion chan struct{}
	Abort      chan struct{}
	Aborted    bool
	Request    *abci.RequestFinalizeBlock
	Response   *abci.ResponseFinalizeBlock
	Error      error
}
