package ormtable

import (
	"encoding/json"
	"io"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/backend/kv"
)

type Table interface {
	UniqueIndex

	MessageType() protoreflect.MessageType

	Save(store kv.IndexCommitmentStore, message proto.Message, mode SaveMode) error
	Delete(store kv.IndexCommitmentStore, primaryKey []protoreflect.Value) error

	GetIndex(fields FieldNames) Index
	GetUniqueIndex(fields FieldNames) UniqueIndex
	Indexes() []Index

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
