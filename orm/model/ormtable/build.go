package ormtable

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	ormv1alpha1 "github.com/cosmos/cosmos-sdk/api/cosmos/orm/v1alpha1"
	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"
)

const (
	PrimaryKeyId uint32 = 0
	IndexIdLimit uint32 = 32768
	SeqId               = IndexIdLimit
)

func BuildTable(options TableOptions) (Table, error) {
	messageDescriptor := options.MessageType.Descriptor()

	table := &TableImpl{
		indexers:              []Indexer{},
		indexes:               []Index{},
		indexesByFields:       map[FieldNames]concreteIndex{},
		uniqueIndexesByFields: map[FieldNames]UniqueIndex{},
		entryCodecsById:       map[uint32]ormkv.EntryCodec{},
		tablePrefix:           options.Prefix,
		typeResolver:          options.TypeResolver,
		customJSONValidator:   options.JSONValidator,
		hooks:                 options.Hooks,
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
		pkPrefix := AppendVarUInt32(options.Prefix, PrimaryKeyId)
		pkCodec, err := ormkv.NewPrimaryKeyCodec(pkPrefix, options.MessageType, nil, options.UnmarshalOptions)
		if err != nil {
			return nil, err
		}

		table.PrimaryKeyIndex = NewPrimaryKeyIndex(pkCodec)

		return &Singleton{
			table,
		}, nil
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

	pkPrefix := AppendVarUInt32(options.Prefix, PrimaryKeyId)
	pkCodec, err := ormkv.NewPrimaryKeyCodec(pkPrefix, options.MessageType, pkFieldNames, options.UnmarshalOptions)
	if err != nil {
		return nil, err
	}

	pkIndex := NewPrimaryKeyIndex(pkCodec)

	table.PrimaryKeyIndex = pkIndex
	table.indexesByFields[pkFields] = pkIndex
	table.uniqueIndexesByFields[pkFields] = pkIndex
	table.entryCodecsById[PrimaryKeyId] = pkIndex
	table.indexes = append(table.indexes, pkIndex)

	for _, idxDesc := range tableDesc.Index {
		id := idxDesc.Id
		if id == 0 || id >= IndexIdLimit {
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
		table.indexers = append(table.indexers, index.(Indexer))
	}

	if tableDesc.PrimaryKey.AutoIncrement {
		autoIncField := pkCodec.GetFieldDescriptors()[0]
		if len(pkFieldNames) != 1 && autoIncField.Kind() != protoreflect.Uint64Kind {
			return nil, ormerrors.InvalidAutoIncrementKey.Wrapf("field %s", autoIncField.FullName())
		}

		seqPrefix := AppendVarUInt32(options.Prefix, SeqId)
		seqCodec := ormkv.NewSeqCodec(options.MessageType, seqPrefix)
		table.entryCodecsById[SeqId] = seqCodec
		return &AutoIncrementTable{
			TableImpl:    table,
			autoIncField: autoIncField,
			seqCodec:     seqCodec,
		}, nil
	}

	return table, nil
}

type TableOptions struct {
	Prefix              []byte
	MessageType         protoreflect.MessageType
	TableDescriptor     *ormv1alpha1.TableDescriptor
	SingletonDescriptor *ormv1alpha1.SingletonDescriptor
	UnmarshalOptions    proto.UnmarshalOptions
	TypeResolver        TypeResolver
	JSONValidator       func(proto.Message) error
	Hooks               Hooks
}
