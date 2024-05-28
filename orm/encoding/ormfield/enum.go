package ormfield

import (
	"encoding/binary"
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// EnumCodec encodes enum values as varints.
type EnumCodec struct{}

func (e EnumCodec) Decode(r Reader) (protoreflect.Value, error) {
	x, err := binary.ReadVarint(r)
	return protoreflect.ValueOfEnum(protoreflect.EnumNumber(x)), err
}

func (e EnumCodec) Encode(value protoreflect.Value, w io.Writer) error {
	var x protoreflect.EnumNumber
	if value.IsValid() {
		x = value.Enum()
	}
	buf := make([]byte, binary.MaxVarintLen32)
	n := binary.PutVarint(buf, int64(x))
	_, err := w.Write(buf[:n])
	return err
}

func (e EnumCodec) Compare(v1, v2 protoreflect.Value) int {
	var x, y protoreflect.EnumNumber
	if v1.IsValid() {
		x = v1.Enum()
	}
	if v2.IsValid() {
		y = v2.Enum()
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

func (e EnumCodec) IsOrdered() bool {
	return false
}

func (e EnumCodec) FixedBufferSize() int {
	return binary.MaxVarintLen32
}

func (e EnumCodec) ComputeBufferSize(protoreflect.Value) (int, error) {
	return e.FixedBufferSize(), nil
}
