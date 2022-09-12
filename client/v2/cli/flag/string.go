package flag

import (
	"context"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type stringType struct{}

func (s stringType) NewValue(context.Context, *Builder) Value {
	v := new(string)
	return (*stringValue)(v)
}

var _ Type = stringType{}

type stringValue string

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}
func (s *stringValue) Type() string {
	return "string"
}

func (s *stringValue) String() string { return string(*s) }

func (s *stringValue) Bind(message protoreflect.Message, field protoreflect.FieldDescriptor) {
	message.Set(field, protoreflect.ValueOfString(string(*s)))
}
