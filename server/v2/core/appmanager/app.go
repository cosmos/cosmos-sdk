package appmanager

import (
	"context"
	"time"

	"google.golang.org/protobuf/proto"

	"cosmossdk.io/server/v2/core/event"
	"cosmossdk.io/server/v2/core/transaction"
)

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

	Resp  Type
	Error error
}
||||||| 39865d852f
=======
package appmanager

import (
	"context"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/serverv2/core/event"
	"github.com/cosmos/cosmos-sdk/serverv2/core/transaction"
)

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

	Resp  Type
	Error error
}
>>>>>>> marko/app_manager
=======
package appmanager

import (
	"context"
	"time"

	"google.golang.org/protobuf/proto"

	"cosmossdk.io/server/v2/core/event"
	"cosmossdk.io/server/v2/core/transaction"
)

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

	Resp  Type
	Error error
}
