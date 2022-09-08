package stablejson

import (
	"fmt"
	"io"
	"sort"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func (opts MarshalOptions) marshalMessage(message protoreflect.Message, writer io.Writer) error {
	switch message.Descriptor().FullName() {
	case timestampFullName:
		return marshalTimestamp(writer, message)
	case durationFullName:
		return marshalDuration(writer, message)
	case structFullName:
		return marshalStruct(writer, message)
	case listValueFullName:
		return marshalListValue(writer, message)
	case valueFullName:
		return marshalStructValue(writer, message)
	case nullValueFullName:
		_, err := writer.Write([]byte("null"))
		return err
	case boolValueFullName, int32ValueFullName, int64ValueFullName, uint32ValueFullName, uint64ValueFullName,
		stringValueFullName, bytesValueFullName, floatValueFullName, doubleValueFullName:
		return opts.marshalWrapper(writer, message)
	case fieldMaskFullName:
		return marshalFieldMask(writer, message)
	case anyFullName:
		return opts.marshalAny(message, writer)
	}

	_, err := writer.Write([]byte("{"))
	if err != nil {
		return err
	}

	err = opts.marshalMessageFields(message, writer, true)
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte("}"))
	return err
}

func (opts MarshalOptions) marshalMessageFields(message protoreflect.Message, writer io.Writer, first bool) error {
	fields := message.Descriptor().Fields()
	numFields := fields.Len()
	allFields := make([]protoreflect.FieldDescriptor, numFields)
	for i := 0; i < numFields; i++ {
		allFields[i] = fields.Get(i)
	}
	sort.Slice(allFields, func(i, j int) bool {
		return allFields[i].Number() < allFields[j].Number()
	})

	for _, field := range allFields {
		if !message.Has(field) {
			continue
		}

		if !first {
			_, err := writer.Write([]byte(","))
			if err != nil {
				return err
			}
		}
		first = false

		_, err := fmt.Fprintf(writer, "%q:", field.Name())
		if err != nil {
			return err
		}

		err = opts.MarshalFieldValue(field, message.Get(field), writer)
		if err != nil {
			return err
		}
	}

	return nil
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
	anyFullName                               = "google.protobuf.Any"
)
