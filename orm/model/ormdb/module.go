package ormdb

import (
	"bytes"
	"context"
	"encoding/binary"
	"math"

	"github.com/cosmos/cosmos-sdk/orm/types/ormjson"

	"google.golang.org/protobuf/reflect/protodesc"

	"github.com/cosmos/cosmos-sdk/orm/encoding/encodeutil"

	"google.golang.org/protobuf/proto"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/model/ormtable"
	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"
)

// ModuleSchema describes the ORM schema for a module.
type ModuleSchema struct {
	// FileDescriptors are the file descriptors that contain ORM tables to use in this schema.
	// Each file descriptor must have an unique non-zero uint32 ID associated with it.
	FileDescriptors map[uint32]protoreflect.FileDescriptor

	// Prefix is an optional prefix to prepend to all keys. It is recommended
	// to leave it empty.
	Prefix []byte
}

// ModuleDB defines the ORM database type to be used by modules.
type ModuleDB interface {
	ormtable.Schema

	// DefaultJSON writes default JSON for each table in the module to the target.
	DefaultJSON(ormjson.WriteTarget) error

	// ValidateJSON validates JSON for each table in the module.
	ValidateJSON(ormjson.ReadSource) error

	// ImportJSON imports JSON for each table in the module which has JSON
	// defined in the read source.
	ImportJSON(context.Context, ormjson.ReadSource) error

	// ExportJSON exports JSON for each table in the module.
	ExportJSON(context.Context, ormjson.WriteTarget) error
}

type moduleDB struct {
	prefix       []byte
	filesById    map[uint32]*fileDescriptorDB
	tablesByName map[protoreflect.FullName]ormtable.Table
}

// ModuleDBOptions are options for constructing a ModuleDB.
type ModuleDBOptions struct {

	// TypeResolver is an optional type resolver to be used when unmarshaling
	// protobuf messages. If it is nil, protoregistry.GlobalTypes will be used.
	TypeResolver ormtable.TypeResolver

	// FileResolver is an optional file resolver that can be used to retrieve
	// pinned file descriptors that may be different from those available at
	// runtime. The file descriptor versions returned by this resolver will be
	// used instead of the ones provided at runtime by the ModuleSchema.
	FileResolver protodesc.Resolver

	// JSONValidator is an optional validator that can be used for validating
	// messaging when using ValidateJSON. If it is nil, DefaultJSONValidator
	// will be used
	JSONValidator func(proto.Message) error

	// GetBackend is the function used to retrieve the table backend.
	// See ormtable.Options.GetBackend for more details.
	GetBackend func(context.Context) (ormtable.Backend, error)

	// GetReadBackend is the function used to retrieve a table read backend.
	// See ormtable.Options.GetReadBackend for more details.
	GetReadBackend func(context.Context) (ormtable.ReadBackend, error)
}

// NewModuleDB constructs a ModuleDB instance from the provided schema and options.
func NewModuleDB(schema ModuleSchema, options ModuleDBOptions) (ModuleDB, error) {
	prefix := schema.Prefix
	db := &moduleDB{
		prefix:       prefix,
		filesById:    map[uint32]*fileDescriptorDB{},
		tablesByName: map[protoreflect.FullName]ormtable.Table{},
	}

	for id, fileDescriptor := range schema.FileDescriptors {
		if id == 0 {
			return nil, ormerrors.InvalidFileDescriptorID.Wrapf("for %s", fileDescriptor.Path())
		}

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

		db.filesById[id] = fdSchema
		for name, table := range fdSchema.tablesByName {
			if _, ok := db.tablesByName[name]; ok {
				return nil, ormerrors.UnexpectedError.Wrapf("duplicate table %s", name)
			}

			db.tablesByName[name] = table
		}
	}

	return db, nil
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
