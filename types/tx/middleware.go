package tx

import (
	context "context"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RequestSimulateTx is the request type for the tx.Handler.RequestSimulateTx
// method.
type RequestSimulateTx struct {
	TxBytes []byte
}

// ResponseSimulateTx is the response type for the tx.Handler.RequestSimulateTx
// method.
type ResponseSimulateTx struct {
	GasInfo sdk.GasInfo
	Result  *sdk.Result
}

type Response struct {
	GasWanted    uint64
	GasUsed      uint64
	MsgResponses []*codectypes.Any // Represents each Msg service handler's response type. Will get proto-serialized into the `Data` field in ABCI, see note #2
	Log          string
	Events       []abci.Event
}

type Request struct {
	Tx      sdk.Tx
	TxBytes []byte
}

// TxHandler defines the baseapp's CheckTx, DeliverTx and Simulate respective
// handlers. It is designed as a middleware stack.
type Handler interface {
	CheckTx(ctx context.Context, req Request, checkReq abci.RequestCheckTx) (Response, error)
	DeliverTx(ctx context.Context, req Request) (Response, error)
	SimulateTx(ctx context.Context, req Request) (Response, error)
}

// TxMiddleware defines one layer of the TxHandler middleware stack.
type Middleware func(Handler) Handler
