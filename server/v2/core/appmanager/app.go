package appmanager

import (
	"context"
	"time"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/core/event"
	"cosmossdk.io/server/v2/core/store"
)

// PrepareHandler passes in the list of Txs that are being proposed. The app can then do stateful operations
// over the list of proposed transactions. It can return a modified list of txs to include in the proposal.
type PrepareHandler[T transaction.Tx] func(context.Context, store.ReaderMap) ([]T, error)

// ProcessHandler is a function that takes a list of transactions and returns a boolean and an error.
// If the verification of a transaction fails, the boolean is false and the error is non-nil.
type ProcessHandler[T transaction.Tx] func(context.Context, []T, store.ReaderMap) error

type QueryRequest struct {
	Height int64
	Path   string
	Data   []byte
}

type QueryResponse struct {
	Height int64
	Value  []byte
}

type BlockRequest[T any] struct {
	Height            uint64
	Time              time.Time
	Hash              []byte
	Txs               []T
	ConsensusMessages []Type //
}

type BlockResponse struct {
	Apphash          []byte
	ValidatorUpdates []appmodule.ValidatorUpdate
	PreBlockEvents   []event.Event
	BeginBlockEvents []event.Event
	TxResults        []TxResult
	EndBlockEvents   []event.Event
}

type RequestInitChain struct {
	Time          time.Time
	ChainId       string
	Validators    []appmodule.ValidatorUpdate
	AppStateBytes []byte
	InitialHeight int64
}

type ResponseInitChain struct {
	Validators []appmodule.ValidatorUpdate
	AppHash    []byte
}

type TxResult struct {
	Events    []event.Event
	GasUsed   uint64
	GasWanted uint64
	Resp      []Type
	Error     error
}
