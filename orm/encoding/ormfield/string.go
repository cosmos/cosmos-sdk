package ormfield

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// StringCodec encodes strings as raw bytes.
type StringCodec struct{}

func (s StringCodec) FixedBufferSize() int {
	return -1
}

func (s StringCodec) ComputeBufferSize(value protoreflect.Value) (int, error) {
	if !value.IsValid() {
		return 0, nil
	}

	return len(value.String()), nil
}

func (s StringCodec) IsOrdered() bool {
	return true
}

func (s StringCodec) Compare(v1, v2 protoreflect.Value) int {
	return compareStrings(v1, v2)
}

func (s StringCodec) Decode(r Reader) (protoreflect.Value, error) {
	bz, err := io.ReadAll(r)
	return protoreflect.ValueOfString(string(bz)), err
}

func (s StringCodec) Encode(value protoreflect.Value, w io.Writer) error {
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

func (s NonTerminalStringCodec) FixedBufferSize() int {
	return -1
}

func (s NonTerminalStringCodec) ComputeBufferSize(value protoreflect.Value) (int, error) {
	return len(value.String()) + 1, nil
}

func (s NonTerminalStringCodec) IsOrdered() bool {
	return true
}

func (s NonTerminalStringCodec) Compare(v1, v2 protoreflect.Value) int {
	return compareStrings(v1, v2)
}

func (s NonTerminalStringCodec) Decode(r Reader) (protoreflect.Value, error) {
	var bz []byte
	for {
		b, err := r.ReadByte()
		if b == 0 || errors.Is(err, io.EOF) {
			return protoreflect.ValueOfString(string(bz)), err
		}
		bz = append(bz, b)
	}
}

func (s NonTerminalStringCodec) Encode(value protoreflect.Value, w io.Writer) error {
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
	_, err := w.Write(bz)
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
