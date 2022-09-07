package stablejson

import (
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func marshalMessage(writer *strings.Builder, value protoreflect.Message) (closingBrace bool, error error) {
	switch value.Descriptor().FullName() {
	case timestampFullName:
		return false, marshalTimestamp(writer, value)
	case durationFullName:
		return false, marshalDuration(writer, value)
	case structFullName:
		// here we let protorange marshal the fields, but just omit braces
		return false, nil
	case listValueFullName:
		// here we let protorange marshal the values, but just omit braces
		return false, nil
	case valueFullName:
		return false, marshalValue(writer, value)
	}

	writer.WriteString("{")
	return true, nil
}

const (
	timestampFullName   protoreflect.FullName = "google.protobuf.Timestamp"
	durationFullName                          = "google.protobuf.Duration"
	structFullName                            = "google.protobuf.Struct"
	valueFullName                             = "google.protobuf.Value"
	listValueFullName                         = "google.protobuf.ListValue"
	nullValueFullName                         = "google.protobuf.NullValue"
	emptyFullName                             = "google.protobuf.Empty"
	boolValueFullName                         = "google.protobuf.BoolValue"
	stringValueFullName                       = "google.protobuf.StringValue"
	bytesValueFullName                        = "google.protobuf.BytesValue"
	int32ValueFullName                        = "google.protobuf.Int32Value"
	int64ValueFullName                        = "google.protobuf.Int64Value"
	uint32ValueFullName                       = "google.protobuf.UInt32Value"
	uint64ValueFullName                       = "google.protobuf.UInt64Value"
	floatValueFullName                        = "google.protobuf.FloatValue"
	doubleValueFullName                       = "google.protobuf.DoubleValue"
	fieldMaskFullName                         = "google.protobuf.FieldMask"
)
