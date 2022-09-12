package flag

import (
	"context"

	"google.golang.org/protobuf/reflect/protoreflect"
)

var addressStringType = Type{
	NewValue: func(ctx context.Context, builder *Builder) Value {
		return &addressValue{}
	},
}

type addressValue struct {
	value string
}

func (a addressValue) Bind(message protoreflect.Message, field protoreflect.FieldDescriptor) {
	message.Set(field, protoreflect.ValueOfString(a.value))
}

func (a addressValue) Get() (protoreflect.Value, error) {
	return protoreflect.ValueOfString(a.value), nil
}

func (a addressValue) String() string {
	return a.value
}

func (a *addressValue) Set(s string) error {
	a.value = s
	// TODO handle bech32 validation
	return nil
}

func (a addressValue) Type() string {
	return "bech32 account address key name"
}
