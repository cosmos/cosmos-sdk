package ormfield

import (
	"encoding/binary"
	io "io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// EnumCodec encodes enum values as varints.
type EnumCodec struct{}

func (EnumCodec) Decode(r Reader) (protoreflect.Value, error) {
	x, err := binary.ReadVarint(r)
	return protoreflect.ValueOfEnum(protoreflect.EnumNumber(x)), err
}

func (EnumCodec) Encode(value protoreflect.Value, w io.Writer) error {
	var x protoreflect.EnumNumber
	if value.IsValid() {
		x = value.Enum()
	}
	buf := make([]byte, binary.MaxVarintLen32)
	n := binary.PutVarint(buf, int64(x))
	_, err := w.Write(buf[:n])
	return err
}

func (EnumCodec) Compare(v1, v2 protoreflect.Value) int {
	var x, y protoreflect.EnumNumber
	if v1.IsValid() {
		x = v1.Enum()
	}
	if v2.IsValid() {
		y = v2.Enum()
	}
	if x == y {
		return 0
	} else if x < y {
		return -1
	} else {
		return 1
	}
}

func (EnumCodec) IsOrdered() bool {
	return false
}

func (EnumCodec) FixedBufferSize() int {
	return binary.MaxVarintLen32
}

func (e EnumCodec) ComputeBufferSize(protoreflect.Value) (int, error) {
	return e.FixedBufferSize(), nil
}
