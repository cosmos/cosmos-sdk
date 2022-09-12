package flag

import (
	"context"
	"strconv"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type boolType struct{}

func (b boolType) NewValue(context.Context, *Builder) Value {
	v := new(bool)
	return (*boolValue)(v)
}

func (b boolType) NoOptDefaultValue() string {
	return "true"
}

var _ Type = boolType{}
var _ HasNoOptDefaultValue = boolType{}

type boolValue bool

func (b *boolValue) Get(protoreflect.Value) (protoreflect.Value, error) {
	return protoreflect.ValueOfBool(bool(*b)), nil
}

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
