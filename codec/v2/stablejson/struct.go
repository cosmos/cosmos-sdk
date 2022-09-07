package stablejson

import (
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	structFieldsFullName    protoreflect.FullName = "google.protobuf.Struct.fields"
	listValueValuesFullName protoreflect.FullName = "google.protobuf.ListValue.values"
	kindOneOf               protoreflect.Name     = "kind"
	nullValueField          protoreflect.Name     = "null_value"
	numberValueField        protoreflect.Name     = "number_value"
	stringValueField        protoreflect.Name     = "string_value"
	boolValueField          protoreflect.Name     = "bool_value"
	structValueField        protoreflect.Name     = "struct_value"
	listValueField          protoreflect.Name     = "list_value"
)

func marshalValue(writer *strings.Builder, value protoreflect.Message) error {
	field := value.WhichOneof(value.Descriptor().Oneofs().ByName(kindOneOf))
	switch field.Name() {
	case nullValueField:
		writer.WriteString("null")
	case numberValueField:
		marshalFloat(writer, value.Get(field).Float())
	case stringValueField:
		_, _ = fmt.Fprintf(writer, "%q", value.Get(field).String())
	case boolValueField:
		writer.WriteString(strconv.FormatBool(value.Get(field).Bool()))
	case structValueField:
		return nil
	case listValueField:
		return nil
	default:
		return fmt.Errorf("unexpected field in google.protobuf.Value: %v", field)
	}
	return nil
}
