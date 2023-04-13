package ormkv

import (
	"bytes"
	"io"

	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/encoding/encodeutil"
	"github.com/cosmos/cosmos-sdk/orm/encoding/ormfield"
)

type KeyCodec struct {
	fixedSize      int
	variableSizers []struct {
		cdc ormfield.Codec
		i   int
	}

	prefix           []byte
	fieldDescriptors []protoreflect.FieldDescriptor
	fieldNames       []protoreflect.Name
	fieldCodecs      []ormfield.Codec
	messageType      protoreflect.MessageType
}

// NewKeyCodec returns a new KeyCodec with an optional prefix for the provided
// message descriptor and fields.
func NewKeyCodec(prefix []byte, messageType protoreflect.MessageType, fieldNames []protoreflect.Name) (*KeyCodec, error) {
	n := len(fieldNames)
	fieldCodecs := make([]ormfield.Codec, n)
	fieldDescriptors := make([]protoreflect.FieldDescriptor, n)
	var variableSizers []struct {
		cdc ormfield.Codec
		i   int
	}
	fixedSize := 0
	messageFields := messageType.Descriptor().Fields()

	for i := 0; i < n; i++ {
		nonTerminal := i != n-1
		field := messageFields.ByName(fieldNames[i])
		if field == nil {
			return nil, ormerrors.FieldNotFound.Wrapf("field %s on %s", fieldNames[i], messageType.Descriptor().FullName())
		}
		cdc, err := ormfield.GetCodec(field, nonTerminal)
		if err != nil {
			return nil, err
		}
		if x := cdc.FixedBufferSize(); x > 0 {
			fixedSize += x
		} else {
			variableSizers = append(variableSizers, struct {
				cdc ormfield.Codec
				i   int
			}{cdc, i})
		}
		fieldCodecs[i] = cdc
		fieldDescriptors[i] = field
	}

	return &KeyCodec{
		fieldCodecs:      fieldCodecs,
		fieldDescriptors: fieldDescriptors,
		fieldNames:       fieldNames,
		prefix:           prefix,
		fixedSize:        fixedSize,
		variableSizers:   variableSizers,
		messageType:      messageType,
	}, nil
}

// EncodeKey encodes the values assuming that they correspond to the fields
// specified for the key. If the array of values is shorter than the
// number of fields in the key, a partial "prefix" key will be encoded
// which can be used for constructing a prefix iterator.
func (cdc *KeyCodec) EncodeKey(values []protoreflect.Value) ([]byte, error) {
	sz, err := cdc.ComputeKeyBufferSize(values)
	if err != nil {
		return nil, err
	}

	w := bytes.NewBuffer(make([]byte, 0, sz+len(cdc.prefix)))
	if _, err = w.Write(cdc.prefix); err != nil {
		return nil, err
	}

	n := len(values)
	if n > len(cdc.fieldCodecs) {
		return nil, ormerrors.IndexOutOfBounds.Wrapf("cannot encode %d values into %d fields", n, len(cdc.fieldCodecs))
	}

	for i := 0; i < n; i++ {
		if err = cdc.fieldCodecs[i].Encode(values[i], w); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

// GetKeyValues extracts the values specified by the key fields from the message.
func (cdc *KeyCodec) GetKeyValues(message protoreflect.Message) []protoreflect.Value {
	res := make([]protoreflect.Value, len(cdc.fieldDescriptors))
	for i, f := range cdc.fieldDescriptors {
		if message.Has(f) {
			res[i] = message.Get(f)
		}
	}
	return res
}

// DecodeKey decodes the values in the key specified by the reader. If the
// provided key is a prefix key, the values that could be decoded will
// be returned with io.EOF as the error.
func (cdc *KeyCodec) DecodeKey(r *bytes.Reader) ([]protoreflect.Value, error) {
	if err := encodeutil.SkipPrefix(r, cdc.prefix); err != nil {
		return nil, err
	}

	n := len(cdc.fieldCodecs)
	values := make([]protoreflect.Value, 0, n)
	for i := 0; i < n; i++ {
		value, err := cdc.fieldCodecs[i].Decode(r)
		if err == io.EOF {
			return values, err
		} else if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}

// EncodeKeyFromMessage combines GetKeyValues and EncodeKey.
func (cdc *KeyCodec) EncodeKeyFromMessage(message protoreflect.Message) ([]protoreflect.Value, []byte, error) {
	values := cdc.GetKeyValues(message)
	bz, err := cdc.EncodeKey(values)
	return values, bz, err
}

// IsFullyOrdered returns true if all fields are also ordered.
func (cdc *KeyCodec) IsFullyOrdered() bool {
	for _, p := range cdc.fieldCodecs {
		if !p.IsOrdered() {
			return false
		}
	}
	return true
}

// CompareKeys compares the provided values which must correspond to the
// fields in this key. Prefix keys of different lengths are supported but the
// function will panic if either array is too long. A negative value is returned
// if values1 is less than values2, 0 is returned if the two arrays are equal,
// and a positive value is returned if values2 is greater.
func (cdc *KeyCodec) CompareKeys(values1, values2 []protoreflect.Value) int {
	j := len(values1)
	k := len(values2)
	n := j
	if k < j {
		n = k
	}

	if n > len(cdc.fieldCodecs) {
		panic("array is too long")
	}

	var cmp int
	for i := 0; i < n; i++ {
		cmp = cdc.fieldCodecs[i].Compare(values1[i], values2[i])
		// any non-equal parts determine our ordering
		if cmp != 0 {
			return cmp
		}
	}

	// values are equal but arrays of different length
	switch {
	case j == k:
		return 0
	case j < k:
		return -1
	default:
		return 1
	}
}

// ComputeKeyBufferSize computes the required buffer size for the provided values
// which can represent a full or prefix key.
func (cdc KeyCodec) ComputeKeyBufferSize(values []protoreflect.Value) (int, error) {
	size := cdc.fixedSize
	n := len(values)
	for _, sz := range cdc.variableSizers {
		// handle prefix key encoding case where don't need all the sizers
		if sz.i >= n {
			return size, nil
		}

		x, err := sz.cdc.ComputeBufferSize(values[sz.i])
		if err != nil {
			return 0, err
		}
		size += x
	}
	return size, nil
}

// SetKeyValues sets the provided values on the message which must correspond
// exactly to the field descriptors for this key. Prefix keys aren't
// supported.
func (cdc *KeyCodec) SetKeyValues(message protoreflect.Message, values []protoreflect.Value) {
	for i, f := range cdc.fieldDescriptors {
		value := values[i]
		if value.IsValid() {
			message.Set(f, value)
		}
	}
}

// CheckValidRangeIterationKeys checks if the start and end key prefixes are valid
// for range iteration meaning that for each non-equal field in the prefixes
// those field types support ordered iteration. If start or end is longer than
// the other, the omitted values will function as the minimum and maximum
// values of that type respectively.
func (cdc KeyCodec) CheckValidRangeIterationKeys(start, end []protoreflect.Value) error {
	lenStart := len(start)
	shortest := lenStart
	longest := lenStart
	lenEnd := len(end)
	if lenEnd < shortest {
		shortest = lenEnd
	} else {
		longest = lenEnd
	}

	if longest > len(cdc.fieldCodecs) {
		return ormerrors.IndexOutOfBounds
	}

	i := 0
	var cmp int

	for ; i < shortest; i++ {
		fieldCdc := cdc.fieldCodecs[i]
		x := start[i]
		y := end[i]

		cmp = fieldCdc.Compare(x, y)
		switch {
		case cmp > 0:
			return ormerrors.InvalidRangeIterationKeys.Wrapf(
				"start must be before end for field %s",
				cdc.fieldDescriptors[i].FullName(),
			)
		case !fieldCdc.IsOrdered() && cmp != 0:
			descriptor := cdc.fieldDescriptors[i]
			return ormerrors.InvalidRangeIterationKeys.Wrapf(
				"field %s of kind %s doesn't support ordered range iteration",
				descriptor.FullName(),
				descriptor.Kind(),
			)
		case cmp < 0:
			break
		}
	}

	// the last prefix value must not be equal if the key lengths are the same
	if lenStart == lenEnd {
		if cmp == 0 {
			return ormerrors.InvalidRangeIterationKeys
		}
	} else {
		// check any remaining values in start or end
		for j := i; j < longest; j++ {
			if !cdc.fieldCodecs[j].IsOrdered() {
				return ormerrors.InvalidRangeIterationKeys.Wrapf(
					"field %s of kind %s doesn't support ordered range iteration",
					cdc.fieldDescriptors[j].FullName(),
					cdc.fieldDescriptors[j].Kind(),
				)
			}
		}
	}

	return nil
}

// GetFieldDescriptors returns the field descriptors for this codec.
func (cdc *KeyCodec) GetFieldDescriptors() []protoreflect.FieldDescriptor {
	return cdc.fieldDescriptors
}

// GetFieldNames returns the field names for this codec.
func (cdc *KeyCodec) GetFieldNames() []protoreflect.Name {
	return cdc.fieldNames
}

// Prefix returns the prefix applied to keys in this codec before any field
// values are encoded.
func (cdc *KeyCodec) Prefix() []byte {
	return cdc.prefix
}

// MessageType returns the message type of fields in this key.
func (cdc *KeyCodec) MessageType() protoreflect.MessageType {
	return cdc.messageType
}
