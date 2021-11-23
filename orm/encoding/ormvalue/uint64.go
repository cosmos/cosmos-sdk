package ormvalue

import (
	"bytes"
	"encoding/binary"
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type Uint64Codec struct{}

func (u Uint64Codec) FixedSize() int {
	return 8
}

func (u Uint64Codec) Size(protoreflect.Value) (int, error) {
	return u.FixedSize(), nil
}

func (u Uint64Codec) IsOrdered() bool {
	return true
}

func (u Uint64Codec) Compare(v1, v2 protoreflect.Value) int {
	return compareUint(v1, v2)
}

func (u Uint64Codec) Decode(r *bytes.Reader) (protoreflect.Value, error) {
	var x uint64
	err := binary.Read(r, binary.BigEndian, &x)
	return protoreflect.ValueOfUint64(x), err
}

func (u Uint64Codec) Encode(value protoreflect.Value, w io.Writer) error {
	return binary.Write(w, binary.BigEndian, value.Uint())
}

func compareUint(v1, v2 protoreflect.Value) int {
	x := v1.Uint()
	y := v2.Uint()
	if x == y {
		return 0
	} else if x < y {
		return -1
	} else {
		return 1
	}
}
