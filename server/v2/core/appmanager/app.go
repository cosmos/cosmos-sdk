package appmanager

import (
	"context"
	"time"

	"cosmossdk.io/core/appmodule"
	corecontext "cosmossdk.io/core/context"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/core/event"
	"cosmossdk.io/server/v2/core/store"

	"github.com/cosmos/gogoproto/proto"
)

// TODO: is this proto.Message the correct one?
// PrepareHandler passes in the list of Txs that are being proposed. The app can then do stateful operations
// over the list of proposed transactions. It can return a modified list of txs to include in the proposal.
type PrepareHandler[T transaction.Tx] func(context.Context, AppManager[T], []T, proto.Message) ([]T, error)

// ProcessHandler is a function that takes a list of transactions and returns a boolean and an error.
// If the verification of a transaction fails, the boolean is false and the error is non-nil.
type ProcessHandler[T transaction.Tx] func(context.Context, AppManager[T], []T, proto.Message) error

type AppManager[T transaction.Tx] interface {
	BuildBlock(ctx context.Context, height, maxBlockBytes uint64) ([]T, error)
	VerifyBlock(ctx context.Context, height uint64, txs []T) error
	DeliverBlock(ctx context.Context, block *BlockRequest[T]) (*BlockResponse, store.WriterMap, error)
	ValidateTx(ctx context.Context, tx T, execMode corecontext.ExecMode) (TxResult, error)
	Simulate(ctx context.Context, tx T) (TxResult, store.WriterMap, error)
	Query(ctx context.Context, version uint64, request Type) (response Type, err error)
	QueryWithState(ctx context.Context, state store.ReaderMap, request Type) (Type, error)
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
