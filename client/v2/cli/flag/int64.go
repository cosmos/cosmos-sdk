package flag

import (
	"context"
	"strconv"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type int64Type struct{}

func (u int64Type) NewValue(context.Context, *Builder) Value {
	v := new(int64)
	return (*int64Value)(v)
}

func (u int64Type) DefaultValue() string {

	return "0"
}

var _ Type = int64Type{}

type int64Value int64

func (i *int64Value) Bind(message protoreflect.Message, field protoreflect.FieldDescriptor) {
	message.Set(field, protoreflect.ValueOfInt64(int64(*i)))
}

func (i *int64Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	*i = int64Value(v)
	return err
}

func (i *int64Value) Type() string {
	return "int64"
}

func (i *int64Value) String() string { return strconv.FormatInt(int64(*i), 10) }
