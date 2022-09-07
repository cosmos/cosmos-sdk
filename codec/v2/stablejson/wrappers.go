package stablejson

import (
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	valueField = "value"
)

func (opts MarshalOptions) marshalWrapper(writer *strings.Builder, message protoreflect.Message) error {
	value := message.Get(message.Descriptor().Fields().ByName(valueField))
	return opts.marshalScalar(writer, value.Interface())
}
