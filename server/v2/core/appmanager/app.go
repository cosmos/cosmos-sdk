package appmanager

import (
	"context"
	"time"

	"cosmossdk.io/server/v2/core/mempool"
	"google.golang.org/protobuf/proto"

	"cosmossdk.io/server/v2/core/event"
	"cosmossdk.io/server/v2/core/store"
	"cosmossdk.io/server/v2/core/transaction"
)

// PrepareHandler is a function that takes a list of transactions and returns a list of transactions and a list of changesets
// In the block building phase the transactions from the mempool can be verified, executed right away or lazily.
type PrepareHandler[T transaction.Tx] func(context.Context, uint32, mempool.Mempool[T], store.ReadonlyState) ([]T, error)

// ProcessHandler is a function that takes a list of transactions and returns a boolean and an error. If the verification of a transaction fails, the boolean is false and the error is non-nil.
type ProcessHandler[T transaction.Tx] func(context.Context, []T, store.ReadonlyState) error

type Type = proto.Message

type App[T transaction.Tx] interface {
	ChainID() string
	AppVersion() (uint64, error)

	InitChain(context.Context, RequestInitChain) (ResponseInitChain, error)
	DeliverBlock(context.Context, BlockRequest) (BlockResponse, error)

	Query(context.Context, *QueryRequest) (*QueryResponse, error)
}

type QueryRequest struct {
	Height int64
	Path   string
	Data   []byte
}

type QueryResponse struct {
	Height int64
	Value  []byte
}

type BlockRequest struct {
	Height            uint64
	Time              time.Time
	Hash              []byte
	Txs               [][]byte
	ConsensusMessages []Type //
}

type BlockResponse struct {
	Apphash          []byte
	ValidatorUpdates []ValidatorUpdate
	PreBlockEvents   []event.Event
	BeginBlockEvents []event.Event
	TxResults        []TxResult
	EndBlockEvents   []event.Event
}

type RequestInitChain struct {
	Time          time.Time
	ChainId       string
	Validators    []ValidatorUpdate
	AppStateBytes []byte
	InitialHeight int64
}

type ResponseInitChain struct {
	Validators []ValidatorUpdate
	AppHash    []byte
}

type TxResult struct {
	Events  []event.Event
	GasUsed uint64
	Resp    []Type
	Error   error
}
