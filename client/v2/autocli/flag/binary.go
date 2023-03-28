package flag

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"google.golang.org/protobuf/reflect/protoreflect"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
)

type binaryType struct{}

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
//If the input string is a valid file path the file path most have a file extension, it will read the file.
//If the input string is a valid hex or base64 encoded string, it will decode the string.
//If the input string is neither a valid file path, hex, or base64 encoded,
// it will return an error.
func (f *fileBinaryValue) Set(s string) error {
	var value []byte
	var err error

	if _, err = os.Stat(s); err == nil {
		if filepath.Ext(s) == "" {
			return errors.New("file path must have an extension try renaming the file to include an extension")
		}
		value, err = os.ReadFile(s)
		if err != nil {
			return err
		}
	} else if value, err = hex.DecodeString(s); err == nil {
	} else if value, err = base64.StdEncoding.DecodeString(s); err == nil {
	} else {
		return errors.New("input string is neither a valid file path, hex, or base64 encoded")
	}
	f.value = value
	return nil
}

func (f *fileBinaryValue) Type() string {
	return "binary file"
}
