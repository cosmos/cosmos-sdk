package ormtable

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"io"

	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
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

	Save(store kvstore.IndexCommitmentStore, message proto.Message, mode SaveMode) error
	Delete(store kvstore.IndexCommitmentStore, primaryKey []protoreflect.Value) error

	DefaultJSON() json.RawMessage
	ValidateJSON(io.Reader) error
	ImportJSON(kvstore.IndexCommitmentStore, io.Reader) error
	ExportJSON(kvstore.IndexCommitmentReadStore, io.Writer) error
}

type SaveMode int

const (
	SAVE_MODE_DEFAULT SaveMode = iota
	SAVE_MODE_CREATE
	SAVE_MODE_UPDATE
)
