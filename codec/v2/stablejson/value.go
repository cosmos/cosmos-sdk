package stablejson

import (
	"encoding/base64"
	"fmt"
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func (opts MarshalOptions) MarshalFieldValue(fieldDescriptor protoreflect.FieldDescriptor, value protoreflect.Value, writer io.Writer) error {
	if fieldDescriptor.IsList() {
		return opts.marshalList(fieldDescriptor, value.List(), writer)
	} else if fieldDescriptor.IsMap() {
		return opts.marshalMap(fieldDescriptor, value.Map(), writer)
	} else {
		return opts.marshalSingleValue(fieldDescriptor, value, writer)
	}
}

// marshalSingleValue marshals values that are not list/map fields or that are elements of list map fields
func (opts MarshalOptions) marshalSingleValue(fieldDescriptor protoreflect.FieldDescriptor, value protoreflect.Value, writer io.Writer) error {
	switch value := value.Interface().(type) {
	case protoreflect.Message:
		return opts.marshalMessage(value, writer)
	case protoreflect.EnumNumber:
		return opts.marshalEnum(fieldDescriptor, value, writer)
	default:
		return opts.marshalPrimitive(writer, value)
	}
}

func (opts MarshalOptions) marshalPrimitive(writer io.Writer, value interface{}) error {
	var err error
	switch value := value.(type) {
	case string:
		_, err = fmt.Fprintf(writer, "%q", value)
	case []byte:
		_, err = writer.Write([]byte(`"`))
		if err != nil {
			return err
		}

		if opts.HexBytes {
			_, err = fmt.Fprintf(writer, "%X", value)
		} else {
			_, err = writer.Write([]byte(base64.StdEncoding.EncodeToString(value)))
		}
		if err != nil {
			return err
		}

		_, err = writer.Write([]byte(`"`))
	case bool:
		_, err = fmt.Fprintf(writer, "%t", value)
	case int32:
		_, err = fmt.Fprintf(writer, "%d", value)
	case uint32:
		_, err = fmt.Fprintf(writer, "%d", value)
	case int64:
		_, err = fmt.Fprintf(writer, `"%d"`, value) // quoted
	case uint64:
		_, err = fmt.Fprintf(writer, `"%d"`, value) // quoted
	case float32:
		err = marshalFloat(writer, float64(value))
	case float64:
		err = marshalFloat(writer, value)
	default:
		return fmt.Errorf("unexpected type %T", value)
	}
	return err
}
