package stablejson

import (
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func (opts MarshalOptions) marshalMessage(writer io.Writer, value protoreflect.Message) (continueRange bool, error error) {
	switch value.Descriptor().FullName() {
	case timestampFullName:
		return false, marshalTimestamp(writer, value)
	case durationFullName:
		return false, marshalDuration(writer, value)
	case structFullName:
		return false, marshalStruct(writer, value)
	case listValueFullName:
		return false, marshalListValue(writer, value)
	case valueFullName:
		return false, marshalValue(writer, value)
	case nullValueFullName:
		_, err := writer.Write([]byte("null"))
		return false, err
	case boolValueFullName, int32ValueFullName, int64ValueFullName, uint32ValueFullName, uint64ValueFullName,
		stringValueFullName, bytesValueFullName, floatValueFullName, doubleValueFullName:
		return false, opts.marshalWrapper(writer, value)
	case fieldMaskFullName:
		return false, marshalFieldMask(writer, value)
	}

	_, err := writer.Write([]byte("{"))
	return true, err
}

const (
	timestampFullName   protoreflect.FullName = "google.protobuf.Timestamp"
	durationFullName                          = "google.protobuf.Duration"
	structFullName                            = "google.protobuf.Struct"
	valueFullName                             = "google.protobuf.Value"
	listValueFullName                         = "google.protobuf.ListValue"
	nullValueFullName                         = "google.protobuf.NullValue"
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
