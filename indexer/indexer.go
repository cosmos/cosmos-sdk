package indexer

import (
	"encoding/json"
)

type Indexer interface {
	StartBlock(uint64) error
	MigrateSchema(*MigrationData) error
	IndexBlockHeader(*BlockHeaderData) error
	IndexTx(*TxData) error
	IndexEvent(*EventData) error
	IndexEntityUpdate(EntityUpdate) error
	IndexEntityDelete(EntityDelete) error
	CommitBlock() error
}

type MigrationData struct {
	Schema Schema
}

type BlockHeaderData struct {
	Height uint64
	Header json.RawMessage
}

type TxData struct {
	TxIndex uint64
	Bytes   []byte
	JSON    json.RawMessage
}

type EventData struct {
	TxIndex    uint64
	MsgIndex   uint64
	EventIndex uint64
	Type       string
	Data       json.RawMessage
}
