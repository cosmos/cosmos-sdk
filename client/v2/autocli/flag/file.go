package flag

import (
	"context"
	"encoding/hex"
	"google.golang.org/protobuf/reflect/protoreflect"
	"os"
)

type fileBinaryType struct{}

func (f fileBinaryType) NewValue(_ context.Context, _ *Builder) Value {
	return &fileBinaryValue{}
}

func (f fileBinaryType) DefaultValue() string {
	return ""
}

// fileBinaryValue is a Value that holds a binary file.
type fileBinaryValue struct {
	value []byte
}

func (f fileBinaryValue) Get(protoreflect.Value) (protoreflect.Value, error) {
	return protoreflect.ValueOfBytes(f.value), nil
}

func (f fileBinaryValue) String() string {
	return string(f.value)
}

func (f fileBinaryValue) Set(s string) error {
	var fileBytes []byte
	var err error
	var value []byte
	// check if file exist
	_, err = os.Stat(s)
	if err == nil {
		// open file at path s
		fileBytes, err = os.ReadFile(s)
		if err != nil {
			return err
		}
		value = make([]byte, hex.EncodedLen(len(fileBytes)))
		hex.Encode(value, fileBytes)
	} else {
		// s is not a file path, so it must be hex encoded
		value, err = hex.DecodeString(s)
		if err != nil {
			return err
		}
	}
	f.value = value
	return nil
}

func (f fileBinaryValue) Type() string {
	return "binary file"
}
