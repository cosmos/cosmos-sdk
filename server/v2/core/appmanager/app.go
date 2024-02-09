package appmanager

import (
	"time"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/server/v2/core/event"
)

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
