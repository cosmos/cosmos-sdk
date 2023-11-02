package ormkv

import (
	"bytes"
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/orm/types/ormerrors"
)

// IndexKeyCodec is the codec for (non-unique) index keys.
type IndexKeyCodec struct {
	*KeyCodec
	pkFieldOrder []int
}

var _ IndexCodec = &IndexKeyCodec{}

// NewIndexKeyCodec creates a new IndexKeyCodec with an optional prefix for the
// provided message descriptor, index and primary key fields.
func NewIndexKeyCodec(prefix []byte, messageType protoreflect.MessageType, indexFields, primaryKeyFields []protoreflect.Name) (*IndexKeyCodec, error) {
	if len(indexFields) == 0 {
		return nil, ormerrors.InvalidTableDefinition.Wrapf("index fields are empty")
	}

	if len(primaryKeyFields) == 0 {
		return nil, ormerrors.InvalidTableDefinition.Wrapf("primary key fields are empty")
	}

	indexFieldMap := map[protoreflect.Name]int{}

	keyFields := make([]protoreflect.Name, 0, len(indexFields)+len(primaryKeyFields))
	for i, f := range indexFields {
		indexFieldMap[f] = i
		keyFields = append(keyFields, f)
	}

	numIndexFields := len(indexFields)
	numPrimaryKeyFields := len(primaryKeyFields)
	pkFieldOrder := make([]int, numPrimaryKeyFields)
	k := 0
	for j, f := range primaryKeyFields {
		if i, ok := indexFieldMap[f]; ok {
			pkFieldOrder[j] = i
			continue
		}
		keyFields = append(keyFields, f)
		pkFieldOrder[j] = numIndexFields + k
		k++
	}

	cdc, err := NewKeyCodec(prefix, messageType, keyFields)
	if err != nil {
		return nil, err
	}

	return &IndexKeyCodec{
		KeyCodec:     cdc,
		pkFieldOrder: pkFieldOrder,
	}, nil
}

func (cdc IndexKeyCodec) DecodeIndexKey(k, _ []byte) (indexFields, primaryKey []protoreflect.Value, err error) {
	values, err := cdc.DecodeKey(bytes.NewReader(k))
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

func (cdc IndexKeyCodec) DecodeEntry(k, v []byte) (Entry, error) {
	idxValues, pk, err := cdc.DecodeIndexKey(k, v)
	if err != nil {
		return nil, err
	}

	return &IndexKeyEntry{
		TableName:   cdc.messageType.Descriptor().FullName(),
		Fields:      cdc.fieldNames,
		IndexValues: idxValues,
		PrimaryKey:  pk,
	}, nil
}

func (cdc IndexKeyCodec) EncodeEntry(entry Entry) (k, v []byte, err error) {
	indexEntry, ok := entry.(*IndexKeyEntry)
	if !ok {
		return nil, nil, ormerrors.BadDecodeEntry
	}

	if indexEntry.TableName != cdc.messageType.Descriptor().FullName() {
		return nil, nil, ormerrors.BadDecodeEntry
	}

	bz, err := cdc.KeyCodec.EncodeKey(indexEntry.IndexValues)
	if err != nil {
		return nil, nil, err
	}

	return bz, []byte{}, nil
}

func (cdc IndexKeyCodec) EncodeKVFromMessage(message protoreflect.Message) (k, v []byte, err error) {
	_, k, err = cdc.EncodeKeyFromMessage(message)
	return k, []byte{}, err
}
