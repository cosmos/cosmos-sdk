package ormdb

import (
	"bytes"
	"context"
	"encoding/binary"
	"math"

	"google.golang.org/protobuf/reflect/protoregistry"

	ormv1alpha1 "cosmossdk.io/api/cosmos/orm/v1alpha1"

	"github.com/cosmos/cosmos-sdk/orm/types/ormjson"

	"google.golang.org/protobuf/reflect/protodesc"

	"github.com/cosmos/cosmos-sdk/orm/encoding/encodeutil"

	"google.golang.org/protobuf/proto"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/model/ormtable"
	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"
)

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

	// GetBackendResolver returns a backend resolver for the requested storage
	// type or an error if this type of storage isn't supported.
	GetBackendResolver func(ormv1alpha1.StorageType) (ormtable.BackendResolver, error)
}

// NewModuleDB constructs a ModuleDB instance from the provided schema and options.
func NewModuleDB(schema *ormv1alpha1.ModuleSchemaDescriptor, options ModuleDBOptions) (ModuleDB, error) {
	prefix := schema.Prefix
	db := &moduleDB{
		prefix:       prefix,
		filesById:    map[uint32]*fileDescriptorDB{},
		tablesByName: map[protoreflect.FullName]ormtable.Table{},
	}

	fileResolver := options.FileResolver
	if fileResolver == nil {
		fileResolver = protoregistry.GlobalFiles
	}

	for _, entry := range schema.SchemaFile {
		var backendResolver ormtable.BackendResolver
		var err error
		if options.GetBackendResolver != nil {
			backendResolver, err = options.GetBackendResolver(entry.StorageType)
			if err != nil {
				return nil, err
			}
		}

		id := entry.Id
		fileDescriptor, err := fileResolver.FindFileByPath(entry.ProtoFileName)
		if err != nil {
			return nil, err
		}

		if id == 0 {
			return nil, ormerrors.InvalidFileDescriptorID.Wrapf("for %s", fileDescriptor.Path())
		}

		opts := fileDescriptorDBOptions{
			ID:              id,
			Prefix:          prefix,
			TypeResolver:    options.TypeResolver,
			JSONValidator:   options.JSONValidator,
			BackendResolver: backendResolver,
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
