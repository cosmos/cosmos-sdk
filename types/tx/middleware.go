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

// Request is the tx request type used in middlewares.
// At least one of Tx or TxBytes must be set. If only TxBytes is set, then
// Tx will be populated by the TxDecoderMiddleware. If only Tx is set, then
// some middlewares (such as signature verification) will fail.
//
// In practice, the middleware stack is called from {Check,Deliver}Tx, which
// only passes the TxBytes. Then, the TxDecoderMiddleware decodes the bytes
// into the Tx field.
type Request struct {
	Tx      sdk.Tx
	TxBytes []byte
}

// Response is the tx response type used in middlewares.
type Response struct {
	GasWanted uint64
	GasUsed   uint64
	// MsgResponses is an array containing each Msg service handler's response
	// type, packed in an Any. This will get proto-serialized into the `Data` field
	// in the ABCI Check/DeliverTx responses.
	MsgResponses []*codectypes.Any
	Log          string
	Events       []abci.Event
}

// RequestCheckTx is the additional request type used in middlewares CheckTx
// method.
type RequestCheckTx struct {
	Type abci.CheckTxType
}

// RequestCheckTx is the additional response type used in middlewares CheckTx
// method.
type ResponseCheckTx struct {
	Priority int64
}

// TxHandler defines the baseapp's CheckTx, DeliverTx and Simulate respective
// handlers. It is designed as a middleware stack.
type Handler interface {
	CheckTx(ctx context.Context, req Request, checkReq RequestCheckTx) (Response, ResponseCheckTx, error)
	DeliverTx(ctx context.Context, req Request) (Response, error)
	SimulateTx(ctx context.Context, req Request) (Response, error)
}

// TxMiddleware defines one layer of the TxHandler middleware stack.
type Middleware func(Handler) Handler
