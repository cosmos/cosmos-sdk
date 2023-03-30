package flag

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"github.com/cockroachdb/errors"
	"google.golang.org/protobuf/reflect/protoreflect"
	"os"
)

type binaryType struct{}

var _ Value = (*fileBinaryValue)(nil)

func (f binaryType) NewValue(_ context.Context, _ *Builder) Value {
	return &fileBinaryValue{}
}

func (f binaryType) DefaultValue() string {
	return ""
}

// fileBinaryValue is a Value that holds a binary file.
type fileBinaryValue struct {
	value []byte
}

func (f *fileBinaryValue) Get(protoreflect.Value) (protoreflect.Value, error) {
	return protoreflect.ValueOfBytes(f.value), nil
}

func (f *fileBinaryValue) String() string {
	return string(f.value)
}

// Set implements the flag.Value interface for binary files.
// If the input string is a valid file path the file path most have a file extension, it will read the file.
// If the input string is a valid hex or base64 encoded string, it will decode the string.
// If the input string is neither a valid file path, hex, or base64 encoded,
// it will return an error.
func (f *fileBinaryValue) Set(s string) error {
	if data, err := os.ReadFile(s); err == nil {
		f.value = data
		return nil
	}

	if data, err := hex.DecodeString(s); err == nil {
		f.value = data
		return nil
	}

	if data, err := base64.StdEncoding.DecodeString(s); err == nil {
		f.value = data
		return nil
	}

	return errors.New("input string is neither a valid file path, hex, or base64 encoded")

}

func (f *fileBinaryValue) Type() string {
	return "binary"
}
