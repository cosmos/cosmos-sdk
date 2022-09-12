package flag

import (
	"context"
	"strconv"

	"google.golang.org/protobuf/reflect/protoreflect"
)

var boolType = Type{
	NewValue: func(ctx context.Context, builder *Builder) Value {
		v := new(bool)
		return (*boolValue)(v)
	},
	DefaultValue:      "",
	NoOptDefaultValue: "",
}

type boolValue bool

func (b *boolValue) Bind(message protoreflect.Message, field protoreflect.FieldDescriptor) {
	message.Set(field, protoreflect.ValueOfBool(bool(*b)))
}

func (b *boolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	*b = boolValue(v)
	return err
}

func (b *boolValue) Type() string {
	return "bool"
}

func (b *boolValue) String() string { return strconv.FormatBool(bool(*b)) }

func (b *boolValue) IsBoolFlag() bool { return true }
