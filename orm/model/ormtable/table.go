package ormtable

import (
	"encoding/json"
	"io"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/backend/kv"
)

type View interface {
	UniqueIndex
	GetIndex(fields FieldNames) Index
	GetUniqueIndex(fields FieldNames) UniqueIndex
	Indexes() []Index
}

type Table interface {
	View

	ormkv.EntryCodec

	Save(store kv.IndexCommitmentStore, message proto.Message, mode SaveMode) error
	Delete(store kv.IndexCommitmentStore, primaryKey []protoreflect.Value) error

	DefaultJSON() json.RawMessage
	ValidateJSON(io.Reader) error
	ImportJSON(kv.IndexCommitmentStore, io.Reader) error
	ExportJSON(kv.IndexCommitmentReadStore, io.Writer) error
}

type SaveMode int

const (
	SAVE_MODE_DEFAULT SaveMode = iota
	SAVE_MODE_CREATE
	SAVE_MODE_UPDATE
)
