package stablejson

import (
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func (opts MarshalOptions) marshalList(fieldDescriptor protoreflect.FieldDescriptor, list protoreflect.List, writer io.Writer) error {
	n := list.Len()
	_, err := writer.Write([]byte("["))
	if err != nil {
		return err
	}

	first := true
	for i := 0; i < n; i++ {
		if !first {
			_, err := writer.Write([]byte(","))
			if err != nil {
				return err
			}
		}
		first = false

		err = opts.marshalSingleValue(fieldDescriptor, list.Get(i), writer)
	}

	_, err = writer.Write([]byte("]"))
	return err
}
