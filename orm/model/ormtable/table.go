package ormtable

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/backend/kv"
	"github.com/cosmos/cosmos-sdk/orm/model/ormindex"
)

type Table interface {
	ormindex.UniqueIndex

	MessageType() protoreflect.MessageType

	Save(store kv.IndexCommitmentStore, message proto.Message, mode SaveMode) error
	Delete(store kv.IndexCommitmentStore, primaryKey []protoreflect.Value) error

	GetIndex(fields FieldNames) ormindex.Index
	GetUniqueIndex(fields FieldNames) ormindex.UniqueIndex
	Indexes() []ormindex.Index
}

type SaveMode int

const (
	SAVE_MODE_DEFAULT SaveMode = iota
	SAVE_MODE_CREATE
	SAVE_MODE_UPDATE
)
