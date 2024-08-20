package app

import (
	"context"
	"time"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/transaction"
)

// BlockRequest defines the request structure for a block coming from consensus server to the state transition function.
type BlockRequest[T transaction.Tx] struct {
	Height  uint64
	Time    time.Time
	Hash    []byte
	ChainId string
	AppHash []byte
	Txs     []T

	// IsGenesis indicates if this block is the first block of the chain.
	IsGenesis bool
}

// BlockResponse defines the response structure for a block coming from the state transition function to consensus server.
type BlockResponse struct {
	Apphash          []byte
	ValidatorUpdates []appmodulev2.ValidatorUpdate
	PreBlockEvents   []event.Event
	BeginBlockEvents []event.Event
	TxResults        []TxResult
	EndBlockEvents   []event.Event
}

// TxResult defines the result of a transaction execution.
type TxResult struct {
	Events    []event.Event
	Resp      []transaction.Msg
	Error     error
	Code      uint32
	Data      []byte
	Log       string
	Info      string
	GasWanted uint64
	GasUsed   uint64
	Codespace string
}

// VersionModifier defines the interface fulfilled by BaseApp
// which allows getting and setting it's appVersion field. This
// in turn updates the consensus params that are sent to the
// consensus engine in EndBlock
type VersionModifier interface {
	SetAppVersion(context.Context, uint64) error
	AppVersion(context.Context) (uint64, error)
}
