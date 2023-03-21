package flag

import (
	"context"
	"encoding/hex"
	"google.golang.org/protobuf/reflect/protoreflect"
	"os"
	"regexp"
)

// regex for file path
var isFilePathRegex = regexp.MustCompile(`^([a-zA-Z0-9_\-\.\/\\])+$`)

type fileBinaryType struct{}

func (f fileBinaryType) NewValue(_ context.Context, _ *Builder) Value {
	return &fileBinaryValue{}
}

func (f fileBinaryType) DefaultValue() []byte {
	return []byte{}
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

func (f *fileBinaryValue) Set(s string) error {
	var fileBytes []byte
	var err error

	if isFilePathRegex.MatchString(s) {
		// open file at path s
		fileBytes, err = os.ReadFile(s)
		if err != nil {
			return err
		}
	} else {
		fileBytes = []byte(s)
	}

	value := make([]byte, hex.EncodedLen(len(fileBytes)))
	hex.Encode(value, fileBytes)
	f.value = value
	return nil
}

func (f fileBinaryValue) Type() string {
	return "binary file"
}
