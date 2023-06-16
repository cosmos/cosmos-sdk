package baseapp

import abci "github.com/cometbft/cometbft/abci/types"

type OptimisticExecutionInfo struct {
	Aborted    bool
	Completion chan struct{}
	Request    *abci.RequestFinalizeBlock
	Response   *abci.ResponseFinalizeBlock
	Error      error
}
