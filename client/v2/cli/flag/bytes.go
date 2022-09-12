package flag

import (
	"context"
	"encoding/base64"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

var bytesBase64Type = Type{
	NewValue: func(ctx context.Context, builder *Builder) Value {
		x := new([]byte)
		return (*bytesBase64Value)(x)
	},
}

// BytesBase64 adapts []byte for use as a flag. Value of flag is Base64 encoded
type bytesBase64Value []byte

func (bytesBase64 *bytesBase64Value) Bind(message protoreflect.Message, field protoreflect.FieldDescriptor) {
	message.Set(field, protoreflect.ValueOfBytes(*bytesBase64))
}

// String implements pflag.Value.String.
func (bytesBase64 *bytesBase64Value) String() string {
	return base64.StdEncoding.EncodeToString(*bytesBase64)
}

// Set implements pflag.Value.Set.
func (bytesBase64 *bytesBase64Value) Set(value string) error {
	bin, err := base64.StdEncoding.DecodeString(strings.TrimSpace(value))

	if err != nil {
		return err
	}

	*bytesBase64 = bin

	return nil
}

// Type implements pflag.Value.Type.
func (*bytesBase64Value) Type() string {
	return "bytesBase64"
}
