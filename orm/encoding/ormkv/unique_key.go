package ormkv

import (
	"bytes"
	"io"

	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// UniqueKeyCodec is the codec for unique indexes.
type UniqueKeyCodec struct {
	tableName    protoreflect.FullName
	pkFieldOrder []struct {
		inKey bool
		i     int
	}
	keyCodec   *KeyCodec
	valueCodec *KeyCodec
}

// NewUniqueKeyCodec creates a new UniqueKeyCodec with an optional prefix for the
// provided message descriptor, index and primary key fields.
func NewUniqueKeyCodec(prefix []byte, messageDescriptor protoreflect.MessageDescriptor, indexFields, primaryKeyFields []protoreflect.Name) (*UniqueKeyCodec, error) {
	keyCodec, err := NewKeyCodec(prefix, messageDescriptor, indexFields)
	if err != nil {
		return nil, err
	}

	haveFields := map[protoreflect.Name]int{}
	for i, descriptor := range keyCodec.fieldDescriptors {
		haveFields[descriptor.Name()] = i
	}

	var valueFields []protoreflect.Name
	var pkFieldOrder []struct {
		inKey bool
		i     int
	}
	k := 0
	for _, field := range primaryKeyFields {
		if j, ok := haveFields[field]; ok {
			pkFieldOrder = append(pkFieldOrder, struct {
				inKey bool
				i     int
			}{inKey: true, i: j})
		} else {
			valueFields = append(valueFields, field)
			pkFieldOrder = append(pkFieldOrder, struct {
				inKey bool
				i     int
			}{inKey: false, i: k})
			k++
		}
	}

	valueCodec, err := NewKeyCodec(nil, messageDescriptor, valueFields)
	if err != nil {
		return nil, err
	}

	return &UniqueKeyCodec{
		tableName:    messageDescriptor.FullName(),
		pkFieldOrder: pkFieldOrder,
		keyCodec:     keyCodec,
		valueCodec:   valueCodec,
	}, nil
}

var _ IndexCodec = &UniqueKeyCodec{}

func (u UniqueKeyCodec) DecodeIndexKey(k, v []byte) (indexFields, primaryKey []protoreflect.Value, err error) {
	ks, err := u.keyCodec.Decode(bytes.NewReader(k))

	// got prefix key
	if err == io.EOF {
		return ks, nil, err
	} else if err != nil {
		return nil, nil, err
	}

	// got prefix key
	if len(ks) < len(u.keyCodec.fieldCodecs) {
		return ks, nil, err
	}

	vs, err := u.valueCodec.Decode(bytes.NewReader(v))
	if err != nil {
		return nil, nil, err
	}

	pk := u.extractPrimaryKey(ks, vs)
	return ks, pk, nil
}

func (cdc UniqueKeyCodec) extractPrimaryKey(keyValues, valueValues []protoreflect.Value) []protoreflect.Value {
	numPkFields := len(cdc.pkFieldOrder)
	pkValues := make([]protoreflect.Value, numPkFields)

	for i := 0; i < numPkFields; i++ {
		fo := cdc.pkFieldOrder[i]
		if fo.inKey {
			pkValues[i] = keyValues[fo.i]
		} else {
			pkValues[i] = valueValues[fo.i]
		}
	}

	return pkValues
}

func (u UniqueKeyCodec) DecodeEntry(k, v []byte) (Entry, error) {
	idxVals, pk, err := u.DecodeIndexKey(k, v)
	if err != nil {
		return nil, err
	}

	return &IndexKeyEntry{
		TableName:   u.tableName,
		Fields:      u.keyCodec.fieldNames,
		IsUnique:    true,
		IndexValues: idxVals,
		PrimaryKey:  pk,
	}, err
}

func (u UniqueKeyCodec) EncodeEntry(entry Entry) (k, v []byte, err error) {
	indexEntry, ok := entry.(*IndexKeyEntry)
	if !ok {
		return nil, nil, ormerrors.BadDecodeEntry
	}
	k, err = u.keyCodec.Encode(indexEntry.IndexValues)
	if err != nil {
		return nil, nil, err
	}

	n := len(indexEntry.PrimaryKey)
	if n != len(u.pkFieldOrder) {
		return nil, nil, ormerrors.BadDecodeEntry.Wrapf("wrong primary key length")
	}

	var values []protoreflect.Value
	for i := 0; i < n; i++ {
		value := indexEntry.PrimaryKey[i]
		fieldOrder := u.pkFieldOrder[i]
		if !fieldOrder.inKey {
			// goes in values because it is not present in the index key otherwise
			values = append(values, value)
		} else {
			// does not go in values, but we need to verify that the value in index values matches the primary key value
			if u.keyCodec.fieldCodecs[fieldOrder.i].Compare(value, indexEntry.IndexValues[fieldOrder.i]) != 0 {
				return nil, nil, ormerrors.BadDecodeEntry.Wrapf("value in primary key does not match corresponding value in index key")
			}
		}
	}

	v, err = u.valueCodec.Encode(values)
	return k, v, err
}

func (u UniqueKeyCodec) EncodeKVFromMessage(message protoreflect.Message) (k, v []byte, err error) {
	_, k, err = u.keyCodec.EncodeFromMessage(message)
	if err != nil {
		return nil, nil, err
	}

	_, v, err = u.valueCodec.EncodeFromMessage(message)
	return k, v, err
}
