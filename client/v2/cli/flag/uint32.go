package flag

import (
	"context"
	"strconv"

	"google.golang.org/protobuf/reflect/protoreflect"
)

var uint32Type = Type{
	NewValue: func(ctx context.Context, builder *Builder) Value {
		v := new(uint32)
		return (*uint32Value)(v)
	},
}

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
