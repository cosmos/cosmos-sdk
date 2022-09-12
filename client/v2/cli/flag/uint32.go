package flag

import (
	"context"
	"strconv"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type uint32Type struct{}

func (t uint32Type) NewValue(context.Context, *Builder) Value {
	v := new(uint32)
	return (*uint32Value)(v)
}

var _ Type = uint32Type{}

type uint32Value uint32

func (i *uint32Value) Bind(message protoreflect.Message, field protoreflect.FieldDescriptor) {
	message.Set(field, protoreflect.ValueOfUint32(uint32(*i)))
}

func (i *uint32Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 32)
	*i = uint32Value(v)
	return err
}

func (i *uint32Value) Type() string {
	return "uint32"
}

func (i *uint32Value) String() string { return strconv.FormatUint(uint64(*i), 10) }
