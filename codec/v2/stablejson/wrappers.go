package stablejson

import (
	io "io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	valueField = "value"
)

func (opts MarshalOptions) marshalWrapper(writer io.Writer, message protoreflect.Message) error {
	value := message.Get(message.Descriptor().Fields().ByName(valueField))
	return opts.marshalPrimitive(writer, value.Interface())
}
