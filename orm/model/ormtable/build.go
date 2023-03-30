package ormtable

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	ormv1 "cosmossdk.io/api/cosmos/orm/v1"

	"github.com/cosmos/cosmos-sdk/orm/encoding/encodeutil"
	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/internal/fieldnames"
	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"
)

const (
	primaryKeyID uint32 = 0
	indexIDLimit uint32 = 32768
	seqID               = indexIDLimit
)

// Options are options for building a Table.
type Options struct {
	// Prefix is an optional prefix used to build the table's prefix.
	Prefix []byte

	// MessageType is the protobuf message type of the table.
	MessageType protoreflect.MessageType

	// TableDescriptor is an optional table descriptor to be explicitly used
	// with the table. Generally this should be nil and the table descriptor
	// should be pulled from the table message option. TableDescriptor
	// cannot be used together with SingletonDescriptor.
	TableDescriptor *ormv1.TableDescriptor

	// SingletonDescriptor is an optional singleton descriptor to be explicitly used.
	// Generally this should be nil and the table descriptor
	// should be pulled from the singleton message option. SingletonDescriptor
	// cannot be used together with TableDescriptor.
	SingletonDescriptor *ormv1.SingletonDescriptor

	// TypeResolver is an optional type resolver to be used when unmarshaling
	// protobuf messages.
	TypeResolver TypeResolver

	// JSONValidator is an optional validator that can be used for validating
	// messaging when using ValidateJSON. If it is nil, DefaultJSONValidator
	// will be used
	JSONValidator func(proto.Message) error

	// BackendResolver is an optional function which retrieves a Backend from the context.
	// If it is nil, the default behavior will be to attempt to retrieve a
	// backend using the method that WrapContextDefault uses. This method
	// can be used to implement things like "store keys" which would allow a
	// table to only be used with a specific backend and to hide direct
	// access to the backend other than through the table interface.
	// Mutating operations will attempt to cast ReadBackend to Backend and
	// will return an error if that fails.
	BackendResolver BackendResolver
}

// TypeResolver is an interface that can be used for the protoreflect.UnmarshalOptions.Resolver option.
type TypeResolver interface {
	protoregistry.MessageTypeResolver
	protoregistry.ExtensionTypeResolver
}

// Build builds a Table instance from the provided Options.
func Build(options Options) (Table, error) {
	messageDescriptor := options.MessageType.Descriptor()

	backendResolver := options.BackendResolver
	if backendResolver == nil {
		backendResolver = getBackendDefault
	}

	table := &tableImpl{
		primaryKeyIndex: &primaryKeyIndex{
			indexers:   []indexer{},
			getBackend: backendResolver,
		},
		indexes:               []Index{},
		indexesByFields:       map[fieldnames.FieldNames]concreteIndex{},
		uniqueIndexesByFields: map[fieldnames.FieldNames]UniqueIndex{},
		entryCodecsByID:       map[uint32]ormkv.EntryCodec{},
		indexesByID:           map[uint32]Index{},
		typeResolver:          options.TypeResolver,
		customJSONValidator:   options.JSONValidator,
	}

	pkIndex := table.primaryKeyIndex

	tableDesc := options.TableDescriptor
	if tableDesc == nil {
		tableDesc = proto.GetExtension(messageDescriptor.Options(), ormv1.E_Table).(*ormv1.TableDescriptor)
	}

	singletonDesc := options.SingletonDescriptor
	if singletonDesc == nil {
		singletonDesc = proto.GetExtension(messageDescriptor.Options(), ormv1.E_Singleton).(*ormv1.SingletonDescriptor)
	}

	switch {
	case tableDesc != nil:
		if singletonDesc != nil {
			return nil, ormerrors.InvalidTableDefinition.Wrapf("message %s cannot be declared as both a table and a singleton", messageDescriptor.FullName())
		}
	case singletonDesc != nil:
		if singletonDesc.Id == 0 {
			return nil, ormerrors.InvalidTableId.Wrapf("%s", messageDescriptor.FullName())
		}

		prefix := encodeutil.AppendVarUInt32(options.Prefix, singletonDesc.Id)
		pkCodec, err := ormkv.NewPrimaryKeyCodec(
			prefix,
			options.MessageType,
			nil,
			proto.UnmarshalOptions{Resolver: options.TypeResolver},
		)
		if err != nil {
			return nil, err
		}

		pkIndex.PrimaryKeyCodec = pkCodec
		table.tablePrefix = prefix
		table.tableID = singletonDesc.Id

		return &singleton{table}, nil
	default:
		return nil, ormerrors.NoTableDescriptor.Wrapf("missing table descriptor for %s", messageDescriptor.FullName())
	}

	tableID := tableDesc.Id
	if tableID == 0 {
		return nil, ormerrors.InvalidTableId.Wrapf("table %s", messageDescriptor.FullName())
	}

	prefix := options.Prefix
	prefix = encodeutil.AppendVarUInt32(prefix, tableID)
	table.tablePrefix = prefix
	table.tableID = tableID

	if tableDesc.PrimaryKey == nil {
		return nil, ormerrors.MissingPrimaryKey.Wrap(string(messageDescriptor.FullName()))
	}

	pkFields := fieldnames.CommaSeparatedFieldNames(tableDesc.PrimaryKey.Fields)
	table.primaryKeyIndex.fields = pkFields
	pkFieldNames := pkFields.Names()
	if len(pkFieldNames) == 0 {
		return nil, ormerrors.InvalidTableDefinition.Wrapf("empty primary key fields for %s", messageDescriptor.FullName())
	}

	pkPrefix := encodeutil.AppendVarUInt32(prefix, primaryKeyID)
	pkCodec, err := ormkv.NewPrimaryKeyCodec(
		pkPrefix,
		options.MessageType,
		pkFieldNames,
		proto.UnmarshalOptions{Resolver: options.TypeResolver},
	)
	if err != nil {
		return nil, err
	}

	pkIndex.PrimaryKeyCodec = pkCodec
	table.indexesByFields[pkFields] = pkIndex
	table.uniqueIndexesByFields[pkFields] = pkIndex
	table.entryCodecsByID[primaryKeyID] = pkIndex
	table.indexesByID[primaryKeyID] = pkIndex
	table.indexes = append(table.indexes, pkIndex)

	for _, idxDesc := range tableDesc.Index {
		id := idxDesc.Id
		if id == 0 || id >= indexIDLimit {
			return nil, ormerrors.InvalidIndexId.Wrapf("index on table %s with fields %s, invalid id %d", messageDescriptor.FullName(), idxDesc.Fields, id)
		}

		if _, ok := table.entryCodecsByID[id]; ok {
			return nil, ormerrors.DuplicateIndexId.Wrapf("id %d on table %s", id, messageDescriptor.FullName())
		}

		idxFields := fieldnames.CommaSeparatedFieldNames(idxDesc.Fields)
		idxPrefix := encodeutil.AppendVarUInt32(prefix, id)
		var index concreteIndex

		// altNames contains all the alternative "names" of this index
		altNames := map[fieldnames.FieldNames]bool{idxFields: true}

		if idxDesc.Unique && isNonTrivialUniqueKey(idxFields.Names(), pkFieldNames) {
			uniqCdc, err := ormkv.NewUniqueKeyCodec(
				idxPrefix,
				options.MessageType,
				idxFields.Names(),
				pkFieldNames,
			)
			if err != nil {
				return nil, err
			}
			uniqIdx := &uniqueKeyIndex{
				UniqueKeyCodec: uniqCdc,
				fields:         idxFields,
				primaryKey:     pkIndex,
				getReadBackend: backendResolver,
			}
			table.uniqueIndexesByFields[idxFields] = uniqIdx
			index = uniqIdx
		} else {
			idxCdc, err := ormkv.NewIndexKeyCodec(
				idxPrefix,
				options.MessageType,
				idxFields.Names(),
				pkFieldNames,
			)
			if err != nil {
				return nil, err
			}
			index = &indexKeyIndex{
				IndexKeyCodec:  idxCdc,
				fields:         idxFields,
				primaryKey:     pkIndex,
				getReadBackend: backendResolver,
			}

			// non-unique indexes can sometimes be named by several sub-lists of
			// fields and we need to handle all of them. For example consider,
			// a primary key for fields "a,b,c" and an index on field "c". Because the
			// rest of the primary key gets appended to the index key, the index for "c"
			// is actually stored as "c,a,b". So this index can be referred to
			// by the fields "c", "c,a", or "c,a,b".
			allFields := index.GetFieldNames()
			allFieldNames := fieldnames.FieldsFromNames(allFields)
			altNames[allFieldNames] = true
			for i := 1; i < len(allFields); i++ {
				altName := fieldnames.FieldsFromNames(allFields[:i])
				if altNames[altName] {
					continue
				}

				// we check by generating a codec for each sub-list of fields,
				// then we see if the full list of fields matches.
				altIdxCdc, err := ormkv.NewIndexKeyCodec(
					idxPrefix,
					options.MessageType,
					allFields[:i],
					pkFieldNames,
				)
				if err != nil {
					return nil, err
				}

				if fieldnames.FieldsFromNames(altIdxCdc.GetFieldNames()) == allFieldNames {
					altNames[altName] = true
				}
			}
		}

		for name := range altNames {
			if _, ok := table.indexesByFields[name]; ok {
				return nil, fmt.Errorf("duplicate index for fields %s", name)
			}

			table.indexesByFields[name] = index
		}

		table.entryCodecsByID[id] = index
		table.indexesByID[id] = index
		table.indexes = append(table.indexes, index)
		table.indexers = append(table.indexers, index.(indexer))
	}

	if tableDesc.PrimaryKey.AutoIncrement {
		autoIncField := pkCodec.GetFieldDescriptors()[0]
		if len(pkFieldNames) != 1 && autoIncField.Kind() != protoreflect.Uint64Kind {
			return nil, ormerrors.InvalidAutoIncrementKey.Wrapf("field %s", autoIncField.FullName())
		}

		seqPrefix := encodeutil.AppendVarUInt32(prefix, seqID)
		seqCodec := ormkv.NewSeqCodec(options.MessageType, seqPrefix)
		table.entryCodecsByID[seqID] = seqCodec
		return &autoIncrementTable{
			tableImpl:    table,
			autoIncField: autoIncField,
			seqCodec:     seqCodec,
		}, nil
	}

	return table, nil
}
