package ormfield

import (
	"encoding/binary"
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// Int32Codec encodes 32-bit integers as big-endian unsigned 32-bit integers
// by adding the maximum value of int32 (2147583647) + 1 before encoding so
// that these values can be used for ordered iteration.
type Int32Codec struct{}

var int32Codec = Int32Codec{}

const (
	int32Max    = 2147483647
	int32Offset = int32Max + 1
)

func (i Int32Codec) Decode(r Reader) (protoreflect.Value, error) {
	var x uint32
	err := binary.Read(r, binary.BigEndian, &x)
	y := int64(x) - int32Offset
	return protoreflect.ValueOfInt32(int32(y)), err
}

func (i Int32Codec) Encode(value protoreflect.Value, w io.Writer) error {
	var x int64
	if value.IsValid() {
		x = value.Int()
	}
	x += int32Offset
	return binary.Write(w, binary.BigEndian, uint32(x))
}

func (i Int32Codec) Compare(v1, v2 protoreflect.Value) int {
	return compareInt(v1, v2)
}

func (i Int32Codec) IsOrdered() bool {
	return true
}

func (i Int32Codec) FixedBufferSize() int {
	return 4
}

func (i Int32Codec) ComputeBufferSize(protoreflect.Value) (int, error) {
	return i.FixedBufferSize(), nil
}
