package ormdb

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"math"

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

	// DefaultJSON returns default JSON that can be used as a template for
	// genesis files.
	//
	// For regular tables this an empty JSON array, but for singletons an
	// empty instance of the singleton is marshaled.
	DefaultJSON(sink JSONSink) error

	// ValidateJSON validates JSON streamed from the reader.
	ValidateJSON(source JSONSource) error

	// ImportJSON imports JSON into the store, streaming one entry at a time.
	// Each table should be import from a separate JSON file to enable proper
	// streaming.
	//
	// Regular tables should be stored as an array of objects with each object
	// corresponding to a single record in the table.
	//
	// Auto-incrementing tables
	// can optionally have the last sequence value as the first element in the
	// array. If the last sequence value is provided, then each value of the
	// primary key in the file must be <= this last sequence value or omitted
	// entirely. If no last sequence value is provided, no entries should
	// contain the primary key as this will be auto-assigned.
	//
	// Singletons should define a single object and not an array.
	//
	// ImportJSON is not atomic with respect to the underlying store, meaning
	// that in the case of an error, some records may already have been
	// imported. It is assumed that ImportJSON is called in the context of some
	// larger transaction isolation.
	ImportJSON(context.Context, JSONSource) error

	// ExportJSON exports JSON in the format accepted by ImportJSON.
	// Auto-incrementing tables will export the last sequence number as the
	// first element in the JSON array.
	ExportJSON(context.Context, JSONSink) error
}

type JSONSource interface {
	JSONReader(tableName protoreflect.FullName) (io.Reader, error)
}

type JSONSink interface {
	JSONWriter(tableName protoreflect.FullName) (io.Writer, error)
}

type moduleDB struct {
	prefix       []byte
	filesById    map[uint32]*fileDescriptorDB
	tablesByName map[protoreflect.FullName]ormtable.Table
}

func (m moduleDB) DefaultJSON(sink JSONSink) error {
	for name, table := range m.tablesByName {
		w, err := sink.JSONWriter(name)
		if err != nil {
			return err
		}

		_, err = w.Write(table.DefaultJSON())
		if err != nil {
			return err
		}
	}
	return nil
}

func (m moduleDB) ValidateJSON(source JSONSource) error {
	var errors map[protoreflect.FullName]error
	for name, table := range m.tablesByName {
		r, err := source.JSONReader(name)
		if err != nil {
			return err
		}

		err = table.ValidateJSON(r)
		if err != nil {
			errors[name] = err
		}
	}

	if len(errors) != 0 {
		panic("TODO")
	}
	return nil
}

func (m moduleDB) ImportJSON(ctx context.Context, source JSONSource) error {
	//TODO need sorted map iteration
	panic("implement me")
}

func (m moduleDB) ExportJSON(ctx context.Context, sink JSONSink) error {
	for name, table := range m.tablesByName {
		w, err := sink.JSONWriter(name)
		if err != nil {
			return err
		}

		err = table.ExportJSON(ctx, w)
		if err != nil {
			return err
		}
	}
	return nil
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
