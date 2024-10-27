package ormfield

import (
	io "io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// BoolCodec encodes a bool value as a single byte 0 or 1.
type BoolCodec struct{}

func (b BoolCodec) Decode(r Reader) (protoreflect.Value, error) {
	x, err := r.ReadByte()
	return protoreflect.ValueOfBool(x != 0), err
}

var (
	zeroBz = []byte{0}
	oneBz  = []byte{1}
)

func (b BoolCodec) Encode(value protoreflect.Value, w io.Writer) error {
	var err error
	if value.Bool() {
		_, err = w.Write(oneBz)
	} else {
		_, err = w.Write(zeroBz)
	}
	return err
}

func (b BoolCodec) Compare(v1, v2 protoreflect.Value) int {
	b1 := v1.Bool()
	b2 := v2.Bool()
	if b1 == b2 {
		return 0
	} else if b1 {
		return -1
	} else {
		return 1
	}
}

func (b BoolCodec) IsOrdered() bool {
	return false
}

func (b BoolCodec) FixedBufferSize() int {
	return 1
}

func (b BoolCodec) ComputeBufferSize(protoreflect.Value) (int, error) {
	return b.FixedBufferSize(), nil
}
