package ormfield

import (
	"fmt"
	"io"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// StringCodec encodes strings as raw bytes.
type StringCodec struct{}

func (StringCodec) FixedBufferSize() int {
	return -1
}

func (StringCodec) ComputeBufferSize(value protoreflect.Value) (int, error) {
	if !value.IsValid() {
		return 0, nil
	}

	return len(value.String()), nil
}

func (StringCodec) IsOrdered() bool {
	return true
}

func (StringCodec) Compare(v1, v2 protoreflect.Value) int {
	return compareStrings(v1, v2)
}

func (StringCodec) Decode(r Reader) (protoreflect.Value, error) {
	bz, err := io.ReadAll(r)
	return protoreflect.ValueOfString(string(bz)), err
}

func (StringCodec) Encode(value protoreflect.Value, w io.Writer) error {
	var x string
	if value.IsValid() {
		x = value.String()
	}
	_, err := w.Write([]byte(x))
	return err
}

// NonTerminalStringCodec encodes strings as null-terminated raw bytes. Null
// values within strings will produce an error.
type NonTerminalStringCodec struct{}

func (NonTerminalStringCodec) FixedBufferSize() int {
	return -1
}

func (NonTerminalStringCodec) ComputeBufferSize(value protoreflect.Value) (int, error) {
	return len(value.String()) + 1, nil
}

func (NonTerminalStringCodec) IsOrdered() bool {
	return true
}

func (NonTerminalStringCodec) Compare(v1, v2 protoreflect.Value) int {
	return compareStrings(v1, v2)
}

func (NonTerminalStringCodec) Decode(r Reader) (protoreflect.Value, error) {
	var bz []byte
	for {
		b, err := r.ReadByte()
		if b == 0 || err == io.EOF {
			return protoreflect.ValueOfString(string(bz)), err
		}
		bz = append(bz, b)
	}
}

func (NonTerminalStringCodec) Encode(value protoreflect.Value, w io.Writer) error {
	var str string
	if value.IsValid() {
		str = value.String()
	}
	bz := []byte(str)
	for _, b := range bz {
		if b == 0 {
			return fmt.Errorf("illegal null terminator found in index string: %s", str)
		}
	}
	_, err := w.Write([]byte(str))
	if err != nil {
		return err
	}
	_, err = w.Write(nullTerminator)
	return err
}

var nullTerminator = []byte{0}

func compareStrings(v1, v2 protoreflect.Value) int {
	var x, y string
	if v1.IsValid() {
		x = v1.String()
	}
	if v2.IsValid() {
		y = v2.String()
	}
	return strings.Compare(x, y)
}
