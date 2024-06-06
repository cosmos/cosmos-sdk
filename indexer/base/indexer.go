package indexerbase

import (
	"encoding/json"
)

type Indexer interface {
	EnsureSetup(*SetupData) error
	StartBlock(uint64) error
	IndexBlockHeader(*BlockHeaderData) error
	IndexTx(*TxData) error
	IndexEvent(*EventData) error
	IndexEntityUpdate(EntityUpdate) error
	IndexEntityDelete(EntityDelete) error
	Commit() error
}

type SetupData struct {
	Schema Schema
}

type BlockHeaderData struct {
	Height uint64
	Header ToJSON
}

type TxData struct {
	TxIndex uint64
	Bytes   []byte
	JSON    ToJSON
}

type EventData struct {
	TxIndex    uint64
	MsgIndex   uint64
	EventIndex uint64
	Type       string
	Data       ToJSON
}

type ToJSON interface {
	ToJSON() json.RawMessage
}
