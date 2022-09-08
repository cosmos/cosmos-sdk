package stablejson

import (
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func (opts MarshalOptions) marshalMap(descriptor protoreflect.FieldDescriptor, value protoreflect.Map, writer io.Writer) error {
	return nil
}
