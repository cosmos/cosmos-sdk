package ormdb

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	ormv1alpha1 "cosmossdk.io/api/cosmos/orm/v1alpha1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/orm/encoding/encodeutil"
	"cosmossdk.io/orm/encoding/ormkv"
	"cosmossdk.io/orm/model/ormtable"
	"cosmossdk.io/orm/types/ormerrors"
)

// ModuleDB defines the ORM database type to be used by modules.
type ModuleDB interface {
	ormtable.Schema

	// GenesisHandler returns an implementation of appmodule.HasGenesis
	// to be embedded in or called from app module implementations.
	// Ex:
	//   type AppModule struct {
	//     appmodule.HasGenesis
	//   }
	//
	//   func NewKeeper(db ModuleDB) *Keeper {
	//     return &Keeper{genesisHandler: db.GenesisHandler()}
	//   }
	//
	//  func NewAppModule(keeper keeper.Keeper) AppModule {
	//    return AppModule{HasGenesis: keeper.GenesisHandler()}
	//  }
	GenesisHandler() appmodule.HasGenesis

	private()
}

type moduleDB struct {
	prefix       []byte
	filesByID    map[uint32]*fileDescriptorDB
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

	// KVStoreService is the storage service to use for the DB if default KV-store storage is used.
	KVStoreService store.KVStoreService

	// KVStoreService is the storage service to use for the DB if memory storage is used.
	MemoryStoreService store.MemoryStoreService

	// KVStoreService is the storage service to use for the DB if transient storage is used.
	TransientStoreService store.TransientStoreService
}

// NewModuleDB constructs a ModuleDB instance from the provided schema and options.
func NewModuleDB(schema *ormv1alpha1.ModuleSchemaDescriptor, options ModuleDBOptions) (ModuleDB, error) {
	prefix := schema.Prefix
	db := &moduleDB{
		prefix:       prefix,
		filesByID:    map[uint32]*fileDescriptorDB{},
		tablesByName: map[protoreflect.FullName]ormtable.Table{},
	}

	fileResolver := options.FileResolver
	if fileResolver == nil {
		fileResolver = protoregistry.GlobalFiles
	}

	for _, entry := range schema.SchemaFile {
		var backendResolver ormtable.BackendResolver

		switch entry.StorageType {
		case ormv1alpha1.StorageType_STORAGE_TYPE_DEFAULT_UNSPECIFIED:
			service := options.KVStoreService
			if service != nil {
				// for testing purposes, the ORM allows KVStoreService to be omitted
				// and a default test backend can be used
				backendResolver = func(ctx context.Context) (ormtable.ReadBackend, error) {
					kvStore := service.OpenKVStore(ctx)
					return ormtable.NewBackend(ormtable.BackendOptions{
						CommitmentStore: kvStore,
						IndexStore:      kvStore,
					}), nil
				}
			}
		case ormv1alpha1.StorageType_STORAGE_TYPE_MEMORY:
			service := options.MemoryStoreService
			if service == nil {
				return nil, fmt.Errorf("missing MemoryStoreService")
			}

			backendResolver = func(ctx context.Context) (ormtable.ReadBackend, error) {
				kvStore := service.OpenMemoryStore(ctx)
				return ormtable.NewBackend(ormtable.BackendOptions{
					CommitmentStore: kvStore,
					IndexStore:      kvStore,
				}), nil
			}
		case ormv1alpha1.StorageType_STORAGE_TYPE_TRANSIENT:
			service := options.TransientStoreService
			if service == nil {
				return nil, fmt.Errorf("missing TransientStoreService")
			}

			backendResolver = func(ctx context.Context) (ormtable.ReadBackend, error) {
				kvStore := service.OpenTransientStore(ctx)
				return ormtable.NewBackend(ormtable.BackendOptions{
					CommitmentStore: kvStore,
					IndexStore:      kvStore,
				}), nil
			}
		default:
			return nil, fmt.Errorf("unsupported storage type %s", entry.StorageType)
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

		db.filesByID[id] = fdSchema
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

	fileSchema, ok := m.filesByID[uint32(id)]
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

func (m moduleDB) GenesisHandler() appmodule.HasGenesis {
	return appModuleGenesisWrapper{m}
}

func (moduleDB) private() {}
