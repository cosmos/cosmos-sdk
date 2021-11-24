package ormfield

import (
	"encoding/binary"
	io "io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// EnumCodec encodes enum values as varints.
type EnumCodec struct{}

func (e EnumCodec) Decode(r Reader) (protoreflect.Value, error) {
	x, err := binary.ReadVarint(r)
	return protoreflect.ValueOfEnum(protoreflect.EnumNumber(x)), err
}

func (e EnumCodec) Encode(value protoreflect.Value, w io.Writer) error {
	x := value.Enum()
	buf := make([]byte, binary.MaxVarintLen32)
	n := binary.PutVarint(buf, int64(x))
	_, err := w.Write(buf[:n])
	return err
}

func (e EnumCodec) Compare(v1, v2 protoreflect.Value) int {
	x := v1.Enum()
	y := v2.Enum()
	if x == y {
		return 0
	} else if x < y {
		return -1
	} else {
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
