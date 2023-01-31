package aminojson

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type JSONMarshaller interface {
	MarshalAmino(proto.Message) ([]byte, error)
}

type User struct {
	Name string `json:"name_field"`
	Age  int    `json:"age_field"`
}

type AminoJson struct {
	nilMapFields map[string]bool
}

func MarshalAmino(message proto.Message) ([]byte, error) {
	buf := &bytes.Buffer{}
	aj := AminoJson{
		nilMapFields: make(map[string]bool),
	}
	rt := reflect.TypeOf(message).Elem()
	rv := reflect.ValueOf(message).Elem()

	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.Type.Kind() == reflect.Map {
			protoFieldNo := strings.Split(f.Tag.Get("protobuf"), ",")[1]
			aj.nilMapFields[protoFieldNo] = rv.Field(i).IsNil()
		}
	}

	vmsg := protoreflect.ValueOfMessage(message.ProtoReflect())
	err := aj.marshal(vmsg, nil, buf)
	return buf.Bytes(), err
}

func (aj AminoJson) marshal(
	value protoreflect.Value,
	field protoreflect.FieldDescriptor,
	writer io.Writer) error {

	switch typedValue := value.Interface().(type) {
	case protoreflect.Message:
		return aj.marshalMessage(typedValue, writer)

	case protoreflect.Map:
		return aj.marshalMap(field, typedValue, writer)

	case protoreflect.List:
		return aj.marshalList(field, typedValue, writer)

	case string, bool, int32, uint32, protoreflect.EnumNumber:
		return invokeStdlibJSONMarshal(writer, typedValue)

	case uint64, int64:
		_, err := fmt.Fprintf(writer, `"%d"`, typedValue) // quoted
		return err

	case []byte:
		_, err := fmt.Fprintf(writer, `"%s"`,
			base64.StdEncoding.EncodeToString([]byte(typedValue)))
		return err
	}

	return nil
}

func (aj AminoJson) marshalMessage(msg protoreflect.Message, writer io.Writer) error {
	_, err := writer.Write([]byte("{"))
	if err != nil {
		return err
	}

	fields := msg.Descriptor().Fields()
	first := true
	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		v := msg.Get(f)

		if f.IsMap() {
			// legacy behavior for maps:
			// - if nil, omit
			// - if empty, include
			// - if non-empty, include
			if isNil := aj.nilMapFields[fmt.Sprintf("%v", f.Number())]; isNil {
				continue
			}
		} else if !msg.Has(f) {
			continue
		}

		// legacy behavior omits maps with no entries, but .Has only checks for absence

		if !first {
			_, err = writer.Write([]byte(","))
			if err != nil {
				return err
			}
		}

		err = invokeStdlibJSONMarshal(writer, f.Name())
		if err != nil {
			return err
		}

		_, err = writer.Write([]byte(":"))
		if err != nil {
			return err
		}

		err = aj.marshal(v, f, writer)
		if err != nil {
			return err
		}

		first = false
	}

	_, err = writer.Write([]byte("}"))
	if err != nil {
		return err
	}
	return nil
}

func invokeStdlibJSONMarshal(w io.Writer, v interface{}) error {
	blob, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(blob)
	return err
}

func (aj AminoJson) marshalAny(message protoreflect.Message, writer io.Writer) error {
	fields := message.Descriptor().Fields()
	typeUrlField := fields.ByName(typeUrlName)
	if typeUrlField == nil {
		return fmt.Errorf("expected type_url field")
	}

	//_, err := writer.Write([]byte("{"))
	//if err != nil {
	//	return err
	//}

	typeUrl := message.Get(typeUrlField).String()
	// TODO
	// do we need an resolver other than protoregistry.GlobalTypes?
	resolver := protoregistry.GlobalTypes

	typ, err := resolver.FindMessageByURL(typeUrl)
	if err != nil {
		return errors.Wrapf(err, "can't resolve type URL %s", typeUrl)
	}

	_, err = fmt.Fprintf(writer, `"@type_url":%q`, typeUrl)
	if err != nil {
		return err
	}

	valueField := fields.ByName(valueName)
	if valueField == nil {
		return fmt.Errorf("expected value field")
	}

	valueBz := message.Get(valueField).Bytes()

	valueMsg := typ.New()
	err = proto.Unmarshal(valueBz, valueMsg.Interface())
	if err != nil {
		return err
	}

	return aj.marshal(protoreflect.ValueOfMessage(valueMsg), nil, writer)

	//err = aj.marshalMessageFields(valueMsg, writer, false)
	//if err != nil {
	//	return err
	//}
	//
	//_, err = writer.Write([]byte("}"))
	//return err
}

func (aj AminoJson) marshalMap(
	fieldDescriptor protoreflect.FieldDescriptor,
	value protoreflect.Map,
	writer io.Writer) error {
	_, err := writer.Write([]byte("{"))
	if err != nil {
		return err
	}

	allKeys := make([]protoreflect.MapKey, 0, value.Len())
	//sortedKeys := make([]protoreflect.MapKey, 0, value.Len())
	value.Range(func(key protoreflect.MapKey, _ protoreflect.Value) bool {
		allKeys = append(allKeys, key)
		return true
	})

	// TODO fail if key type is not string

	valueField := fieldDescriptor.MapValue()
	first := true
	for _, key := range allKeys {
		if !first {
			_, err = writer.Write([]byte(","))
			if err != nil {
				return err
			}
		}
		first = false

		err = invokeStdlibJSONMarshal(writer, key.String())
		if err != nil {
			return err
		}

		_, err = writer.Write([]byte(":"))
		if err != nil {
			return err
		}

		err = aj.marshal(value.Get(key), valueField, writer)
		if err != nil {
			return err
		}
	}

	_, err = writer.Write([]byte("}"))
	return err
}

func (aj AminoJson) marshalList(
	fieldDescriptor protoreflect.FieldDescriptor,
	list protoreflect.List,
	writer io.Writer) error {
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

		err = aj.marshal(list.Get(i), fieldDescriptor, writer)
		if err != nil {
			return err
		}
	}

	_, err = writer.Write([]byte("]"))
	return err
}

const (
	// type names
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

	// field names
	typeUrlName protoreflect.Name = "type_url"
	valueName   protoreflect.Name = "value"
)
