package ormdb

import (
	"bytes"
	"encoding/binary"
	"math"

	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/cosmos/cosmos-sdk/orm/encoding/encodeutil"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/model/ormtable"
)

type fileDescriptorDBOptions struct {
	Prefix          []byte
	ID              uint32
	TypeResolver    ormtable.TypeResolver
	JSONValidator   func(proto.Message) error
	BackendResolver ormtable.BackendResolver
}

type fileDescriptorDB struct {
	id             uint32
	prefix         []byte
	tablesById     map[uint32]ormtable.Table
	tablesByName   map[protoreflect.FullName]ormtable.Table
	fileDescriptor protoreflect.FileDescriptor
}

func newFileDescriptorDB(fileDescriptor protoreflect.FileDescriptor, options fileDescriptorDBOptions) (*fileDescriptorDB, error) {
	prefix := encodeutil.AppendVarUInt32(options.Prefix, options.ID)

	schema := &fileDescriptorDB{
		id:             options.ID,
		prefix:         prefix,
		tablesById:     map[uint32]ormtable.Table{},
		tablesByName:   map[protoreflect.FullName]ormtable.Table{},
		fileDescriptor: fileDescriptor,
	}

	resolver := options.TypeResolver
	if resolver == nil {
		resolver = protoregistry.GlobalTypes
	}

	messages := fileDescriptor.Messages()
	n := messages.Len()
	for i := 0; i < n; i++ {
		messageDescriptor := messages.Get(i)
		tableName := messageDescriptor.FullName()
		messageType, err := resolver.FindMessageByName(tableName)
		if err != nil {
			return nil, err
		}

		table, err := ormtable.Build(ormtable.Options{
			Prefix:          prefix,
			MessageType:     messageType,
			TypeResolver:    resolver,
			JSONValidator:   options.JSONValidator,
			BackendResolver: options.BackendResolver,
		})
		if err != nil {
			return nil, err
		}

		id := table.ID()
		if _, ok := schema.tablesById[id]; ok {
			return nil, ormerrors.InvalidTableId.Wrapf("duplicate ID %d for %s", id, tableName)
		}
		schema.tablesById[id] = table

		if _, ok := schema.tablesByName[tableName]; ok {
			return nil, ormerrors.InvalidTableDefinition.Wrapf("duplicate table %s", tableName)
		}
		schema.tablesByName[tableName] = table
	}

	return schema, nil
}

func (f fileDescriptorDB) DecodeEntry(k, v []byte) (ormkv.Entry, error) {
	r := bytes.NewReader(k)
	err := encodeutil.SkipPrefix(r, f.prefix)
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

	table, ok := f.tablesById[uint32(id)]
	if !ok {
		return nil, ormerrors.UnexpectedDecodePrefix.Wrapf("can't find table with id %d", id)
	}

	return table.DecodeEntry(k, v)
}

func (f fileDescriptorDB) EncodeEntry(entry ormkv.Entry) (k, v []byte, err error) {
	table, ok := f.tablesByName[entry.GetTableName()]
	if !ok {
		return nil, nil, ormerrors.BadDecodeEntry.Wrapf("can't find table %s", entry.GetTableName())
	}

	return table.EncodeEntry(entry)
}

var _ ormkv.EntryCodec = fileDescriptorDB{}
