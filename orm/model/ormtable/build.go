package ormtable

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	ormv1alpha1 "github.com/cosmos/cosmos-sdk/api/cosmos/orm/v1alpha1"
	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"
)

const (
	PrimaryKeyId uint32 = 0
)

func Build(options TableOptions) (Table, error) {
	messageDescriptor := options.MessageType.Descriptor()
	tableDesc := options.TableDescriptor
	if tableDesc == nil {
		tableDesc = proto.GetExtension(messageDescriptor.Options(), ormv1alpha1.E_Table).(*ormv1alpha1.TableDescriptor)
		if tableDesc == nil {
			return nil, ormerrors.InvalidTableDefinition.Wrapf("missing table descriptor for %s", messageDescriptor.FullName())
		}
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

	table := &TableImpl{
		PrimaryKeyIndex:       pkIndex,
		indexers:              []Indexer{},
		indexes:               []Index{},
		indexesByFields:       map[FieldNames]concreteIndex{},
		uniqueIndexesByFields: map[FieldNames]UniqueIndex{},
		indexesById:           map[uint32]concreteIndex{},
		tablePrefix:           options.Prefix,
		typeResolver:          options.TypeResolver,
		customImportValidator: options.ImportValidator,
	}

	table.indexesByFields[pkFields] = pkIndex
	table.uniqueIndexesByFields[pkFields] = pkIndex
	table.indexesById[PrimaryKeyId] = pkIndex
	table.indexes = append(table.indexes, pkIndex)

	for _, idxDesc := range tableDesc.Index {
		id := idxDesc.Id
		if id == 0 {
			return nil, ormerrors.InvalidIndexId.Wrapf("index on table %s with fields %s", messageDescriptor.FullName(), idxDesc.Fields)
		}

		if _, ok := table.indexesById[id]; ok {
			return nil, ormerrors.DuplicateIndexId.Wrapf("id %d on table %s", id, messageDescriptor.FullName())
		}

		idxFields, err := CommaSeparatedFieldNames(idxDesc.Fields)
		if err != nil {
			return nil, err
		}

		idxPrefix := AppendVarUInt32(options.Prefix, id)
		var index concreteIndex
		if idxDesc.Unique {
			uniqCdc, err := ormkv.NewUniqueKeyCodec(idxPrefix, options.MessageType, idxFields.Names(), pkFieldNames)
			if err != nil {
				return nil, err
			}
			uniqIdx := NewUniqueKeyIndex(uniqCdc, pkIndex)
			table.uniqueIndexesByFields[idxFields] = uniqIdx
			index = NewUniqueKeyIndex(uniqCdc, pkIndex)
		} else {
			idxCdc, err := ormkv.NewIndexKeyCodec(idxPrefix, options.MessageType, idxFields.Names(), pkFieldNames)
			if err != nil {
				return nil, err
			}
			index = NewIndexKeyIndex(idxCdc, pkIndex)
		}
		table.indexesByFields[idxFields] = index
		table.indexesById[idxDesc.Id] = index
		table.indexes = append(table.indexes, index)
		table.indexers = append(table.indexers, index.(Indexer))
	}

	return table, nil
}

type TableOptions struct {
	Prefix           []byte
	MessageType      protoreflect.MessageType
	TableDescriptor  *ormv1alpha1.TableDescriptor
	UnmarshalOptions proto.UnmarshalOptions
	TypeResolver     TypeResolver
	ImportValidator  func(proto.Message) error
}
