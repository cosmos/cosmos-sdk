package ormdb

import (
	"bytes"
	"context"
	"encoding/binary"
	"math"

	"google.golang.org/protobuf/reflect/protodesc"

	"github.com/cosmos/cosmos-sdk/orm/encoding/encodeutil"

	"google.golang.org/protobuf/proto"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/model/ormtable"
	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"
)

type ModuleSchema struct {
	FileDescriptors map[uint32]protoreflect.FileDescriptor
	Prefix          []byte
}

type ModuleDB interface {
	ormkv.EntryCodec
	GetTable(message proto.Message) ormtable.Table
}

type moduleDB struct {
	prefix       []byte
	filesById    map[uint32]*fileDescriptorDB
	tablesByName map[protoreflect.FullName]ormtable.Table
}

type ModuleDBOptions struct {
	// TypeResolver is an optional type resolver to be used when unmarshaling
	// protobuf messages.
	TypeResolver ormtable.TypeResolver

	FileResolver protodesc.Resolver

	// JSONValidator is an optional validator that can be used for validating
	// messaging when using ValidateJSON. If it is nil, DefaultJSONValidator
	// will be used
	JSONValidator func(proto.Message) error

	GetBackend func(context.Context) (ormtable.Backend, error)

	GetReadBackend func(context.Context) (ormtable.ReadBackend, error)
}

func NewModuleDB(desc ModuleSchema, options ModuleDBOptions) (ModuleDB, error) {
	prefix := desc.Prefix
	schema := &moduleDB{
		prefix:       prefix,
		filesById:    map[uint32]*fileDescriptorDB{},
		tablesByName: map[protoreflect.FullName]ormtable.Table{},
	}

	for id, fileDescriptor := range desc.FileDescriptors {
		opts := fileDescriptorDBOptions{
			ID:             id,
			Prefix:         prefix,
			TypeResolver:   options.TypeResolver,
			JSONValidator:  options.JSONValidator,
			GetBackend:     options.GetBackend,
			GetReadBackend: options.GetReadBackend,
		}

		if options.FileResolver != nil {
			// if a FileResolver is provided, we use that to resolve the file
			// and not the one provided as a different pinned file descriptor
			// may have been provided
			var err error
			fileDescriptor, err = options.FileResolver.FindFileByPath(fileDescriptor.Path())
			if err != nil {
				return nil, err
			}
		}

		fdSchema, err := newFileDescriptorDB(fileDescriptor, opts)
		if err != nil {
			return nil, err
		}

		schema.filesById[id] = fdSchema
		for name, table := range fdSchema.tablesByName {
			if _, ok := schema.tablesByName[name]; ok {
				return nil, ormerrors.UnexpectedError.Wrapf("duplicate table %s", name)
			}

			schema.tablesByName[name] = table
		}
	}

	return schema, nil
}

func (m moduleDB) DecodeEntry(k, v []byte) (ormkv.Entry, error) {
	r := bytes.NewReader(k)
	err := encodeutil.SkipPrefix(r, m.prefix)
	if err != nil {
		return nil, err
	}

	id, err := binary.ReadUvarint(r)
	if err != nil {
		return nil, err
	}

	if id > math.MaxUint32 {
		return nil, ormerrors.UnexpectedDecodePrefix.Wrapf("uint32 varint id out of range %d", id)
	}

	fileSchema, ok := m.filesById[uint32(id)]
	if !ok {
		return nil, ormerrors.UnexpectedDecodePrefix.Wrapf("can't find FileDescriptor schema with id %d", id)
	}

	return fileSchema.DecodeEntry(k, v)
}

func (m moduleDB) EncodeEntry(entry ormkv.Entry) (k, v []byte, err error) {
	tableName := entry.GetTableName()
	table, ok := m.tablesByName[tableName]
	if !ok {
		return nil, nil, ormerrors.BadDecodeEntry.Wrapf("can't find table %s", tableName)
	}

	return table.EncodeEntry(entry)
}

func (m moduleDB) GetTable(message proto.Message) ormtable.Table {
	return m.tablesByName[message.ProtoReflect().Descriptor().FullName()]
}
