package ormtable

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	ormv1alpha1 "github.com/cosmos/cosmos-sdk/api/cosmos/orm/v1alpha1"
	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/model/ormindex"
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

	pkIndex := ormindex.NewPrimaryKeyIndex(pkCodec)

	table := &TableImpl{
		PrimaryKeyIndex:       pkIndex,
		indexers:              []ormindex.Indexer{},
		indexes:               []ormindex.Index{},
		indexesByFields:       map[FieldNames]ormindex.Index{},
		uniqueIndexesByFields: map[FieldNames]ormindex.UniqueIndex{},
		indexesById:           map[uint32]ormindex.Index{},
		tablePrefix:           options.Prefix,
		typeResolver:          options.TypeResolver,
	}

	table.indexesByFields[pkFields] = pkIndex
	table.uniqueIndexesByFields[pkFields] = pkIndex
	table.indexesById[PrimaryKeyId] = pkIndex

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
		var cdc ormkv.IndexCodec
		var index ormindex.Index
		if idxDesc.Unique {
			uniqCdc, err := ormkv.NewUniqueKeyCodec(idxPrefix, messageDescriptor, idxFields.Names(), pkFieldNames)
			if err != nil {
				return nil, err
			}
			uniqIdx := ormindex.NewUniqueKeyIndex(uniqCdc, pkIndex)
			table.uniqueIndexesByFields[idxFields] = uniqIdx
			index = ormindex.NewUniqueKeyIndex(uniqCdc, pkIndex)
			cdc = uniqCdc
		} else {
			idxCdc, err := ormkv.NewIndexKeyCodec(idxPrefix, messageDescriptor, idxFields.Names(), pkFieldNames)
			if err != nil {
				return nil, err
			}
			panic("TODO")
			//idx := ormindex.NewUniqueKeyIndex(uniqCdc, pkIndex)
			//table.uniqueIndexesByFields[idxFields] = uniqIdx
			//index = ormindex.NewUniqueKeyIndex(uniqCdc, pkIndex)
			//cdc = uniqCdc
		}
	}

	return table, nil
}

type TableOptions struct {
	Prefix           []byte
	MessageType      protoreflect.MessageType
	TableDescriptor  *ormv1alpha1.TableDescriptor
	UnmarshalOptions proto.UnmarshalOptions
	TypeResolver     TypeResolver
}
