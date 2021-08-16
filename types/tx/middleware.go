package tx

import (
	context "context"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RequestSimulateTx is the request type for the TxHandler.RequestSimulateTx
// method.
type RequestSimulateTx struct {
	TxBytes []byte
}

// ResponseSimulateTx is the response type for the TxHandler.RequestSimulateTx
// method.
type ResponseSimulateTx struct {
	GasInfo sdk.GasInfo
	Result  *sdk.Result
}

// TxHandler defines the baseapp's CheckTx, DeliverTx and Simulate respective
// handlers. It is designed as a middleware stack.
type TxHandler interface {
	CheckTx(ctx context.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error)
	DeliverTx(ctx context.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error)
	SimulateTx(ctx context.Context, tx sdk.Tx, req RequestSimulateTx) (ResponseSimulateTx, error)
}

// TxMiddleware defines one layer of the TxHandler middleware stack.
type TxMiddleware func(TxHandler) TxHandler
