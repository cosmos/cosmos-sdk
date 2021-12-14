package ormtable

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoregistry"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	ormv1alpha1 "github.com/cosmos/cosmos-sdk/api/cosmos/orm/v1alpha1"
	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"
)

const (
	primaryKeyId uint32 = 0
	indexIdLimit uint32 = 32768
	seqId               = indexIdLimit
)

// Options are options for building a Table.
type Options struct {
	// Prefix is an optional prefix for the table within the store.
	Prefix []byte

	// MessageType is the protobuf message type of the table.
	MessageType protoreflect.MessageType

	// TableDescriptor is an optional table descriptor to be explicitly used
	// with the table. Generally this should be nil and the table descriptor
	// should be pulled from the table message option. TableDescriptor
	// cannot be used together with SingletonDescriptor.
	TableDescriptor *ormv1alpha1.TableDescriptor

	// SingletonDescriptor is an optional singleton descriptor to be explicitly used.
	// Generally this should be nil and the table descriptor
	// should be pulled from the singleton message option. SingletonDescriptor
	// cannot be used together with TableDescriptor.
	SingletonDescriptor *ormv1alpha1.SingletonDescriptor

	// TypeResolver is an optional type resolver to be used when unmarshaling
	// protobuf messages.
	TypeResolver TypeResolver

	// JSONValidator is an optional validator that can be used for validating
	// messaging when using ValidateJSON. If it is nil, DefaultJSONValidator
	// will be used
	JSONValidator func(proto.Message) error
}

// TypeResolver is an interface that can be used for the protoreflect.UnmarshalOptions.Resolver option.
type TypeResolver interface {
	protoregistry.MessageTypeResolver
	protoregistry.ExtensionTypeResolver
}

// Build builds a Table instance from the provided Options.
func Build(options Options) (Table, error) {
	messageDescriptor := options.MessageType.Descriptor()

	table := &tableImpl{
		indexers:              []indexer{},
		indexes:               []Index{},
		indexesByFields:       map[FieldNames]concreteIndex{},
		uniqueIndexesByFields: map[FieldNames]UniqueIndex{},
		entryCodecsById:       map[uint32]ormkv.EntryCodec{},
		tablePrefix:           options.Prefix,
		typeResolver:          options.TypeResolver,
		customJSONValidator:   options.JSONValidator,
	}

	tableDesc := options.TableDescriptor
	if tableDesc == nil {
		tableDesc = proto.GetExtension(messageDescriptor.Options(), ormv1alpha1.E_Table).(*ormv1alpha1.TableDescriptor)
	}

	singletonDesc := options.SingletonDescriptor
	if singletonDesc == nil {
		singletonDesc = proto.GetExtension(messageDescriptor.Options(), ormv1alpha1.E_Singleton).(*ormv1alpha1.SingletonDescriptor)
	}

	if tableDesc != nil {
		if singletonDesc != nil {
			return nil, ormerrors.InvalidTableDefinition.Wrapf("message %s cannot be declared as both a table and a singleton", messageDescriptor.FullName())
		}
	} else if singletonDesc != nil {
		pkPrefix := AppendVarUInt32(options.Prefix, primaryKeyId)
		pkCodec, err := ormkv.NewPrimaryKeyCodec(
			pkPrefix,
			options.MessageType,
			nil,
			proto.UnmarshalOptions{Resolver: options.TypeResolver},
		)
		if err != nil {
			return nil, err
		}

		table.PrimaryKeyIndex = NewPrimaryKeyIndex(pkCodec)

		return &singleton{table}, nil
	} else {
		return nil, ormerrors.InvalidTableDefinition.Wrapf("missing table descriptor for %s", messageDescriptor.FullName())
	}

	tableId := tableDesc.Id
	if tableId == 0 {
		return nil, ormerrors.InvalidTableId.Wrapf("table %s", messageDescriptor.FullName())
	}

	if tableDesc.PrimaryKey == nil {
		return nil, ormerrors.MissingPrimaryKey.Wrap(string(messageDescriptor.FullName()))
	}

	pkFields, err := CommaSeparatedFieldNames(tableDesc.PrimaryKey.Fields)
	if err != nil {
		return nil, err
	}

	pkFieldNames := pkFields.Names()
	if len(pkFieldNames) == 0 {
		return nil, ormerrors.InvalidTableDefinition.Wrapf("empty primary key fields for %s", messageDescriptor.FullName())
	}

	pkPrefix := AppendVarUInt32(options.Prefix, primaryKeyId)
	pkCodec, err := ormkv.NewPrimaryKeyCodec(
		pkPrefix,
		options.MessageType,
		pkFieldNames,
		proto.UnmarshalOptions{Resolver: options.TypeResolver},
	)
	if err != nil {
		return nil, err
	}

	pkIndex := NewPrimaryKeyIndex(pkCodec)

	table.PrimaryKeyIndex = pkIndex
	table.indexesByFields[pkFields] = pkIndex
	table.uniqueIndexesByFields[pkFields] = pkIndex
	table.entryCodecsById[primaryKeyId] = pkIndex
	table.indexes = append(table.indexes, pkIndex)

	for _, idxDesc := range tableDesc.Index {
		id := idxDesc.Id
		if id == 0 || id >= indexIdLimit {
			return nil, ormerrors.InvalidIndexId.Wrapf("index on table %s with fields %s, invalid id %d", messageDescriptor.FullName(), idxDesc.Fields, id)
		}

		if _, ok := table.entryCodecsById[id]; ok {
			return nil, ormerrors.DuplicateIndexId.Wrapf("id %d on table %s", id, messageDescriptor.FullName())
		}

		idxFields, err := CommaSeparatedFieldNames(idxDesc.Fields)
		if err != nil {
			return nil, err
		}

		idxPrefix := AppendVarUInt32(options.Prefix, id)
		var index concreteIndex

		// altNames contains all the alternative "names" of this index
		altNames := map[FieldNames]bool{idxFields: true}

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
			uniqIdx := NewUniqueKeyIndex(uniqCdc, pkIndex)
			table.uniqueIndexesByFields[idxFields] = uniqIdx
			index = NewUniqueKeyIndex(uniqCdc, pkIndex)
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
			index = NewIndexKeyIndex(idxCdc, pkIndex)

			// non-unique indexes can sometimes be named by several sub-lists of
			// fields and we need to handle all of them. For example consider,
			// a primary key for fields "a,b,c" and an index on field "c". Because the
			// rest of the primary key gets appended to the index key, the index for "c"
			// is actually stored as "c,a,b". So this index can be referred to
			// by the fields "c", "c,a", or "c,a,b".
			allFields := index.GetFieldNames()
			allFieldNames := FieldsFromNames(allFields)
			altNames[allFieldNames] = true
			for i := 1; i < len(allFields); i++ {
				altName := FieldsFromNames(allFields[:i])
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

				if FieldsFromNames(altIdxCdc.GetFieldNames()) == allFieldNames {
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

		table.entryCodecsById[id] = index
		table.indexes = append(table.indexes, index)
		table.indexers = append(table.indexers, index.(indexer))
	}

	if tableDesc.PrimaryKey.AutoIncrement {
		autoIncField := pkCodec.GetFieldDescriptors()[0]
		if len(pkFieldNames) != 1 && autoIncField.Kind() != protoreflect.Uint64Kind {
			return nil, ormerrors.InvalidAutoIncrementKey.Wrapf("field %s", autoIncField.FullName())
		}

		seqPrefix := AppendVarUInt32(options.Prefix, seqId)
		seqCodec := ormkv.NewSeqCodec(options.MessageType, seqPrefix)
		table.entryCodecsById[seqId] = seqCodec
		return &autoIncrementTable{
			tableImpl:    table,
			autoIncField: autoIncField,
			seqCodec:     seqCodec,
		}, nil
	}

	return table, nil
}
