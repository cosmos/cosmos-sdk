package stablejson

import (
	"fmt"
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func (opts MarshalOptions) marshalEnum(fieldDescriptor protoreflect.FieldDescriptor, value protoreflect.EnumNumber, writer io.Writer) error {
	enumDescriptor := fieldDescriptor.Enum()
	if enumDescriptor == nil {
		return fmt.Errorf("expected enum descriptor for %s", fieldDescriptor.FullName())
	}

	enumValueDescriptor := enumDescriptor.Values().ByNumber(value)
	var err error
	if enumValueDescriptor != nil {
		_, err = fmt.Fprintf(writer, "%q", enumValueDescriptor.Name())
	} else {
		_, err = fmt.Fprintf(writer, "%d", value)
	}
	return err
}
