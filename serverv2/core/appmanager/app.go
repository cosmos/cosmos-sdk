package appmanager

import (
	"context"
	"time"

	"github.com/cosmos/cosmos-sdk/serverv2/core/event"
	"github.com/cosmos/cosmos-sdk/serverv2/core/transaction"
	"github.com/cosmos/cosmos-sdk/serverv2/core/validator"
)

type App[T transaction.Tx] interface {
	ChainID() string
	AppVersion() (uint64, error)

	InitChain(context.Context, RequestInitChain) (ResponseInitChain, error)
	DeliverBlock(context.Context, RequestDeliverBlock[T]) (ResponseDeliverBlock, error)
}

type RequestDeliverBlock[T transaction.Tx] struct {
	Height int64
	Time   time.Time
	Hash   []byte

	Txs []T
}

type ResponseDeliverBlock struct {
	Apphash          []byte
	ValidatorUpdates []validator.Update
	TxResults        []TxResult
	Events           []event.Event
}

type RequestInitChain struct {
	Time          time.Time
	ChainId       string
	Validators    []validator.Update
	AppStateBytes []byte
	InitialHeight int64
}

type ResponseInitChain struct {
	Validators []validator.Update
	AppHash    []byte
}

type TxResult struct {
	GasWanted int64
	GasUsed   int64
	Log       string
	Data      string
	Events    event.Event
}
