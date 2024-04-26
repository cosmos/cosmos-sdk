package ormfield

import (
	"encoding/binary"
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// Int64Codec encodes 64-bit integers as big-endian unsigned 64-bit integers
// by adding the maximum value of int32 (9223372036854775807) + 1 before encoding so
// that these values can be used for ordered iteration.
type Int64Codec struct{}

var int64Codec = Int64Codec{}

const int64Max = 9223372036854775807

func (i Int64Codec) Decode(r Reader) (protoreflect.Value, error) {
	var x uint64
	err := binary.Read(r, binary.BigEndian, &x)
	if x >= int64Max {
		x = x - int64Max - 1
		return protoreflect.ValueOfInt64(int64(x)), err
	}

	y := int64(x) - int64Max - 1
	return protoreflect.ValueOfInt64(y), err
}

func (i Int64Codec) Encode(value protoreflect.Value, w io.Writer) error {
	var x int64
	if value.IsValid() {
		x = value.Int()
	}
	if x >= -1 {
		y := uint64(x) + int64Max + 1
		return binary.Write(w, binary.BigEndian, y)
	}

	x += int64Max
	x++
	return binary.Write(w, binary.BigEndian, uint64(x))
}

func (i Int64Codec) Compare(v1, v2 protoreflect.Value) int {
	return compareInt(v1, v2)
}

func (i Int64Codec) IsOrdered() bool {
	return true
}

func (i Int64Codec) FixedBufferSize() int {
	return 8
}

func (i Int64Codec) ComputeBufferSize(protoreflect.Value) (int, error) {
	return i.FixedBufferSize(), nil
}

func compareInt(v1, v2 protoreflect.Value) int {
	var x, y int64
	if v1.IsValid() {
		x = v1.Int()
	}
	if v2.IsValid() {
		y = v2.Int()
	}
	switch {
	case x == y:
		return 0
	case x < y:
		return -1
	default:
		return 1
	}
}
