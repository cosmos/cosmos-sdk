package flag

import (
	"context"
	"strconv"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type int32Type struct{}

func (u int32Type) NewValue(ctx context.Context, builder *Builder) Value {
	v := new(int32)
	return (*int32Value)(v)
}

func (u int32Type) DefaultValue() string {
	//TODO implement me
	panic("implement me")
}

var _ Type = int32Type{}

type int32Value int32

func (i *int32Value) Bind(message protoreflect.Message, field protoreflect.FieldDescriptor) {
	message.Set(field, protoreflect.ValueOfInt32(int32(*i)))
}

func (i *int32Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 32)
	*i = int32Value(v)
	return err
}

func (i *int32Value) Type() string {
	return "int32"
}

func (i *int32Value) String() string { return strconv.FormatInt(int64(*i), 10) }
