package flag

import (
	"context"
	"strconv"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type uint64Type struct{}

func (u uint64Type) NewValue(ctx context.Context, builder *Builder) Value {
	v := new(uint64)
	return (*uint64Value)(v)
}

func (u uint64Type) DefaultValue() string {
	//TODO implement me
	panic("implement me")
}

var _ Type = uint64Type{}

type uint64Value uint64

func (i *uint64Value) Bind(message protoreflect.Message, field protoreflect.FieldDescriptor) {
	message.Set(field, protoreflect.ValueOfUint64(uint64(*i)))
}

func (i *uint64Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	*i = uint64Value(v)
	return err
}

func (i *uint64Value) Type() string {
	return "uint64"
}

func (i *uint64Value) String() string { return strconv.FormatUint(uint64(*i), 10) }
