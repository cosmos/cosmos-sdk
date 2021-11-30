package ormkv

import (
	"bytes"
	"io"

	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type IndexKeyCodec struct {
	*KeyCodec
	tableName    protoreflect.FullName
	fieldNames   Fields
	pkFieldOrder []int
}

var _ IndexCodec = &IndexKeyCodec{}

func NewIndexKeyCodec(prefix []byte, tableName protoreflect.FullName, indexFields []protoreflect.FieldDescriptor, primaryKeyFields []protoreflect.FieldDescriptor) (*IndexKeyCodec, error) {
	indexFieldMap := map[protoreflect.Name]int{}

	var keyFields []protoreflect.FieldDescriptor
	for i, f := range indexFields {
		indexFieldMap[f.Name()] = i
		keyFields = append(keyFields, f)
	}

	numIndexFields := len(indexFields)
	numPrimaryKeyFields := len(primaryKeyFields)
	pkFieldOrder := make([]int, numPrimaryKeyFields)
	k := 0
	for j, f := range primaryKeyFields {
		if i, ok := indexFieldMap[f.Name()]; ok {
			pkFieldOrder[j] = i
			continue
		}
		keyFields = append(keyFields, f)
		pkFieldOrder[j] = numIndexFields + k
		k++
	}

	cdc, err := NewKeyCodec(prefix, keyFields)
	if err != nil {
		return nil, err
	}

	fields := FieldsFromDescriptors(cdc.fieldDescriptors)

	return &IndexKeyCodec{
		KeyCodec:     cdc,
		pkFieldOrder: pkFieldOrder,
		fieldNames:   fields,
		tableName:    tableName,
	}, nil
}

func (cdc IndexKeyCodec) DecodeIndexKey(k, _ []byte) (indexFields, primaryKey []protoreflect.Value, err error) {

	values, err := cdc.Decode(bytes.NewReader(k))
	// got prefix key
	if err == io.EOF {
		return values, nil, nil
	} else if err != nil {
		return nil, nil, err
	}

	// got prefix key
	if len(values) < len(cdc.fieldCodecs) {
		return values, nil, nil
	}

	numPkFields := len(cdc.pkFieldOrder)
	pkValues := make([]protoreflect.Value, numPkFields)

	for i := 0; i < numPkFields; i++ {
		pkValues[i] = values[cdc.pkFieldOrder[i]]
	}

	return values, pkValues, nil
}

var _ IndexCodec = &IndexKeyCodec{}

func (cdc IndexKeyCodec) DecodeEntry(k, v []byte) (Entry, error) {
	idxValues, pk, err := cdc.DecodeIndexKey(k, v)
	if err != nil {
		return nil, err
	}

	return &IndexKeyEntry{
		TableName:   cdc.tableName,
		Fields:      cdc.fieldNames,
		IndexValues: idxValues,
		PrimaryKey:  pk,
	}, nil
}

func (i IndexKeyCodec) EncodeEntry(entry Entry) (k, v []byte, err error) {
	indexEntry, ok := entry.(*IndexKeyEntry)
	if !ok {
		return nil, nil, ormerrors.BadDecodeEntry
	}

	if indexEntry.TableName != i.tableName {
		return nil, nil, ormerrors.BadDecodeEntry
	}

	bz, err := i.KeyCodec.Encode(indexEntry.IndexValues)
	if err != nil {
		return nil, nil, err
	}

	return bz, sentinel, nil
}

var sentinel = []byte{0}

func (cdc IndexKeyCodec) EncodeKVFromMessage(message protoreflect.Message) (k, v []byte, err error) {
	_, k, err = cdc.EncodeFromMessage(message)
	return k, sentinel, err
}
