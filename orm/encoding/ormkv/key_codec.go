package ormkv

import (
	"bytes"
	"io"

	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"

	"google.golang.org/protobuf/reflect/protoreflect"

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
	fieldCodecs      []ormfield.Codec
}

func NewKeyCodec(prefix []byte, fieldDescs []protoreflect.FieldDescriptor) (*KeyCodec, error) {
	n := len(fieldDescs)
	var valueCodecs []ormfield.Codec
	var variableSizers []struct {
		cdc ormfield.Codec
		i   int
	}
	fixedSize := 0
	names := make([]protoreflect.Name, len(fieldDescs))
	for i := 0; i < n; i++ {
		nonTerminal := true
		if i == n-1 {
			nonTerminal = false
		}
		field := fieldDescs[i]
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
		valueCodecs = append(valueCodecs, cdc)
		names[i] = field.Name()
	}

	return &KeyCodec{
		fieldCodecs:      valueCodecs,
		fieldDescriptors: fieldDescs,
		prefix:           prefix,
		fixedSize:        fixedSize,
		variableSizers:   variableSizers,
	}, nil
}

// Encode encodes the values assuming that they correspond to the fields
// specified for the key. If the array of values is shorter than the
// number of fields in the key, a partial "prefix" key will be encoded
// which can be used for constructing a prefix iterator.
func (cdc *KeyCodec) Encode(values []protoreflect.Value) ([]byte, error) {
	sz, err := cdc.ComputeBufferSize(values)
	if err != nil {
		return nil, err
	}

	w := bytes.NewBuffer(make([]byte, 0, sz))
	_, err = w.Write(cdc.prefix)
	if err != nil {
		return nil, err
	}

	n := len(values)
	if n > len(cdc.fieldCodecs) {
		return nil, ormerrors.IndexOutOfBounds
	}

	for i := 0; i < n; i++ {
		err = cdc.fieldCodecs[i].Encode(values[i], w)
		if err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

// GetValues extracts the values specified by the key fields from the message.
func (cdc *KeyCodec) GetValues(message protoreflect.Message) []protoreflect.Value {
	var res []protoreflect.Value
	for _, f := range cdc.fieldDescriptors {
		res = append(res, message.Get(f))
	}
	return res
}

// Decode decodes the values in the key specified by the reader. If the
// provided key is a prefix key, the values that could be decoded will
// be returned with io.EOF as the error.
func (cdc *KeyCodec) Decode(r *bytes.Reader) ([]protoreflect.Value, error) {
	err := skipPrefix(r, cdc.prefix)
	if err != nil {
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

// EncodeFromMessage combines GetValues and Encode.
func (cdc *KeyCodec) EncodeFromMessage(message protoreflect.Message) ([]protoreflect.Value, []byte, error) {
	values := cdc.GetValues(message)
	bz, err := cdc.Encode(values)
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

// CompareValues compares the provided values which must correspond to the
// fields in this key. Prefix keys of different lengths are  supported but the
// function will panic if either array is too long.
func (cdc *KeyCodec) CompareValues(values1, values2 []protoreflect.Value) int {
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
	if j == k {
		return 0
	} else if j < k {
		return -1
	} else {
		return 1
	}
}

// ComputeBufferSize computes the required buffer size for the provided values
// which can represent a full or prefix key.
func (cdc KeyCodec) ComputeBufferSize(values []protoreflect.Value) (int, error) {
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

// SetValues sets the provided values on the message which must correspond
// exactly to the field descriptors for this key. Prefix keys aren't
// supported.
func (cdc *KeyCodec) SetValues(message protoreflect.Message, values []protoreflect.Value) {
	for i, f := range cdc.fieldDescriptors {
		message.Set(f, values[i])
	}
}

// CheckValidRangeIterationKeys checks if the start and end key prefixes are valid
// for range iteration meaning that for each non-equal field in the prefixes
// those field types support ordered iteration.
func (cdc KeyCodec) CheckValidRangeIterationKeys(start, end []protoreflect.Value) error {
	lenStart := len(start)
	n := lenStart
	lenEnd := len(end)
	if lenEnd > n {
		n = lenEnd
	}

	if n > len(cdc.fieldCodecs) {
		return ormerrors.IndexOutOfBounds
	}

	var x protoreflect.Value
	var y protoreflect.Value
	var cmp int

	for i := 0; i < n; i++ {
		fieldCdc := cdc.fieldCodecs[i]

		if i < lenStart {
			x = start[i]
		} else {
			// if values are omitted use the default
			x = fieldCdc.DefaultValue()
		}

		if i < lenEnd {
			y = end[i]
		} else {
			// if values are omitted use the default
			y = fieldCdc.DefaultValue()
		}

		cmp = fieldCdc.Compare(x, y)
		if cmp > 0 {
			return ormerrors.InvalidRangeIterationKeys.Wrapf(
				"start must be before end for field %s",
				cdc.fieldDescriptors[i].FullName(),
			)
		} else if !fieldCdc.IsOrdered() && cmp != 0 {
			descriptor := cdc.fieldDescriptors[i]
			return ormerrors.InvalidRangeIterationKeys.Wrapf(

				"field %s of kind %s doesn't support ordered range iteration",
				descriptor.FullName(),
				descriptor.Kind(),
			)
		}
	}

	// the last prefix value must not be equal
	if cmp == 0 {
		return ormerrors.InvalidRangeIterationKeys.Wrapf(
			"start must be before end for field %s",
			cdc.fieldDescriptors[n-1].FullName(),
		)
	}

	return nil
}
