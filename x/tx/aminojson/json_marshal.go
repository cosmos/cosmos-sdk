package aminojson

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	cosmos_proto "github.com/cosmos/cosmos-proto"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"io"

	"cosmossdk.io/api/amino"
)

type MessageEncoder func(protoreflect.Message, io.Writer) error
type FieldEncoder func(AminoJSON, protoreflect.Value, io.Writer) error

type AminoJSON struct {
	// maps cosmos_proto.scalar -> zero value factory
	scalarEncoders  map[string]FieldEncoder
	zeroValues      map[string]func() protoreflect.Value
	messageEncoders map[string]MessageEncoder
	fieldEncoders   map[string]FieldEncoder
}

func NewAminoJSON() AminoJSON {
	aj := AminoJSON{
		zeroValues: map[string]func() protoreflect.Value{
			"cosmos.Dec": func() protoreflect.Value {
				return protoreflect.ValueOfString("0")
			},
		},
		scalarEncoders: map[string]FieldEncoder{
			"cosmos.Dec": cosmosDecEncoder,
			"cosmos.Int": cosmosDecEncoder,
		},
		messageEncoders: map[string]MessageEncoder{
			"key_field":      keyFieldEncoder,
			"module_account": moduleAccountEncoder,
		},
		fieldEncoders: map[string]FieldEncoder{
			"empty_string":        emptyStringEncoder,
			"json_default":        jsonDefaultEncoder,
			"null_slice_as_empty": nullSliceAsEmptyEncoder,
			"cosmos_dec_bytes":    cosmosDecBytesEncoder,
		},
	}
	return aj
}

func (aj AminoJSON) DefineMessageEncoding(name string, encoder MessageEncoder) {
	aj.messageEncoders[name] = encoder
}

func (aj AminoJSON) DefineFieldEncoding(name string, encoder FieldEncoder) {
	aj.fieldEncoders[name] = encoder
}

func (aj AminoJSON) MarshalAmino(message proto.Message) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := aj.beginMarshal(message.ProtoReflect(), buf)
	return buf.Bytes(), err
}

// TODO
// move into marshalMessage
func (aj AminoJSON) beginMarshal(msg protoreflect.Message, writer io.Writer) error {
	name, named := getMessageName(msg)
	if named {
		_, err := writer.Write([]byte(fmt.Sprintf(`{"type":"%s","value":`, name)))
		if err != nil {
			return err
		}
	}

	err := aj.marshal(protoreflect.ValueOfMessage(msg), writer)
	if err != nil {
		return err
	}

	if named {
		_, err = writer.Write([]byte("}"))
		if err != nil {
			return err
		}
	}

	return nil
}

func (aj AminoJSON) marshal(value protoreflect.Value, writer io.Writer) error {
	// TODO timestamp
	// this is what is breaking MsgGrant

	switch val := value.Interface().(type) {
	case protoreflect.Message:
		err := aj.marshalMessage(val, writer)
		return err

	case protoreflect.Map:
		return errors.New("maps are not supported")

	case protoreflect.List:
		if !val.IsValid() {
			_, err := writer.Write([]byte("null"))
			return err
		}
		return aj.marshalList(val, writer)

	case string, bool, int32, uint32, protoreflect.EnumNumber:
		return jsonMarshal(writer, val)

	case uint64, int64:
		_, err := fmt.Fprintf(writer, `"%d"`, val) // quoted
		return err

	case []byte:
		_, err := fmt.Fprintf(writer, `"%s"`, base64.StdEncoding.EncodeToString(val))
		return err

	default:
		return errors.Errorf("unknown type %T", val)
	}
}

// TODO
// merge with marshalMessage or if embed ends up not being needed delete it.
func (aj AminoJSON) marshalEmbedded(msg protoreflect.Message, writer io.Writer) (bool, error) {
	fields := msg.Descriptor().Fields()
	first := true
	wrote := false
	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		v := msg.Get(f)
		name := getFieldName(f)

		if !msg.Has(f) {
			if omitEmpty(f) {
				continue
			} else {
				zv, found := aj.getZeroValue(f)
				if found {
					v = zv
				}
			}
		}

		if !first {
			_, err := writer.Write([]byte(","))
			if err != nil {
				return wrote, err
			}
		}

		err := jsonMarshal(writer, name)
		wrote = true
		if err != nil {
			return wrote, err
		}

		_, err = writer.Write([]byte(":"))
		if err != nil {
			return wrote, err
		}

		// encode value
		if encoder := aj.getFieldEncoding(f); encoder != nil {
			err = encoder(aj, v, writer)
			if err != nil {
				return wrote, err
			}
		} else {
			err = aj.marshal(v, writer)
			if err != nil {
				return wrote, err
			}
		}

		first = false
	}

	return wrote, nil
}

func (aj AminoJSON) marshalMessage(msg protoreflect.Message, writer io.Writer) error {
	if msg == nil {
		return errors.New("nil message")
	}

	switch msg.Descriptor().FullName() {
	case timestampFullName:
		// replicate https://github.com/tendermint/go-amino/blob/8e779b71f40d175cd1302d3cd41a75b005225a7a/json-encode.go#L45-L51
		return marshalTimestamp(msg, writer)
	case durationFullName:
		return marshalDuration(msg, writer)
	case anyFullName:
		return aj.marshalAny(msg, writer)
	}

	if encoder := aj.getMessageEncoder(msg); encoder != nil {
		err := encoder(msg, writer)
		return err
	}

	_, err := writer.Write([]byte("{"))
	if err != nil {
		return err
	}

	fields := msg.Descriptor().Fields()
	first := true
	emptyOneOfWritten := map[string]bool{}
	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		v := msg.Get(f)
		name := getFieldName(f)
		oneof := f.ContainingOneof()
		isOneOf := oneof != nil
		oneofFieldName, oneofTypeName, err := getOneOfNames(f)
		if err != nil && isOneOf {
			return err
		}
		writeNil := false

		if !msg.Has(f) {
			// msg.WhichOneof(oneof) == nil: no field of the oneof has been set
			// !emptyOneOfWritten: we haven't written a null for this oneof yet (only write one null per empty oneof)
			if isOneOf && msg.WhichOneof(oneof) == nil && emptyOneOfWritten[oneofFieldName] != true {
				name = oneofFieldName
				writeNil = true
				emptyOneOfWritten[oneofFieldName] = true
			} else if omitEmpty(f) {
				continue
			} else if f.Kind() == protoreflect.MessageKind &&
				f.Cardinality() != protoreflect.Repeated &&
				!v.Message().IsValid() {
				return errors.Errorf("not supported: dont_omit_empty=true on invalid (nil?) message field: %s", name)
			}
		}

		if isFieldEmbedded(v, f) {
			wrote, err := aj.marshalEmbedded(v.Message(), writer)
			if err != nil {
				return err
			}
			if wrote {
				first = false
			}
			continue
		}

		if !first {
			_, err = writer.Write([]byte(","))
			if err != nil {
				return err
			}
		}

		if isOneOf && !writeNil {
			_, err = writer.Write([]byte(fmt.Sprintf(`"%s":{"type":"%s","value":{`,
				oneofFieldName, oneofTypeName)))
			if err != nil {
				return err
			}
		}

		err = jsonMarshal(writer, name)
		if err != nil {
			return err
		}

		_, err = writer.Write([]byte(":"))
		if err != nil {
			return err
		}

		// encode value
		if encoder := aj.getFieldEncoding(f); encoder != nil {
			err = encoder(aj, v, writer)
			if err != nil {
				return err
			}
		} else if writeNil {
			_, err = writer.Write([]byte("null"))
			if err != nil {
				return err
			}
		} else {
			err = aj.marshal(v, writer)
			if err != nil {
				return err
			}
		}

		if isOneOf && !writeNil {
			_, err = writer.Write([]byte("}}"))
			if err != nil {
				return err
			}
		}

		first = false
	}

	_, err = writer.Write([]byte("}"))
	return err
}

func jsonMarshal(w io.Writer, v interface{}) error {
	blob, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(blob)
	return err
}

func (aj AminoJSON) marshalList(list protoreflect.List, writer io.Writer) error {
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

		err = aj.marshal(list.Get(i), writer)
		if err != nil {
			return err
		}
	}

	_, err = writer.Write([]byte("]"))
	return err
}

func getMessageName(msg protoreflect.Message) (string, bool) {
	opts := msg.Descriptor().Options()
	if proto.HasExtension(opts, amino.E_Name) {
		name := proto.GetExtension(opts, amino.E_Name)
		return name.(string), true
	}
	return "", false
}

// omitEmpty returns true if the field should be omitted if empty. Empty field omission is the default behavior.
func omitEmpty(field protoreflect.FieldDescriptor) bool {
	opts := field.Options()
	if proto.HasExtension(opts, amino.E_DontOmitempty) {
		dontOmitEmpty := proto.GetExtension(opts, amino.E_DontOmitempty).(bool)
		return !dontOmitEmpty
	}
	//if field.ContainingOneof() != nil {
	//	return false
	//}

	// legacy support for gogoproto would need to look something like below.
	//
	// if gproto.GetBoolExtension(opts, gogoproto.E_Nullable, true) {
	//
	// }
	return true
}

func getFieldName(field protoreflect.FieldDescriptor) string {
	opts := field.Options()
	if proto.HasExtension(opts, amino.E_FieldName) {
		return proto.GetExtension(opts, amino.E_FieldName).(string)
	}
	return string(field.Name())
}

func getOneOfNames(field protoreflect.FieldDescriptor) (string, string, error) {
	opts := field.Options()
	oneOf := field.ContainingOneof()
	if oneOf == nil {
		return "", "", errors.Errorf("field %s must be within a oneof", field.Name())
	}
	oneOfOpts := oneOf.Options()

	var fieldName, typeName string
	if proto.HasExtension(oneOfOpts, amino.E_OneofFieldName) {
		fieldName = proto.GetExtension(oneOfOpts, amino.E_OneofFieldName).(string)
	} else {
		return "", "", errors.Errorf("oneof %s must have the amino.oneof_field_name option set", oneOf.Name())
	}
	if proto.HasExtension(opts, amino.E_OneofTypeName) {
		typeName = proto.GetExtension(opts, amino.E_OneofTypeName).(string)
	} else {
		return "", "", errors.Errorf("field %s within a oneof must have the amino.oneof_type_name option set",
			field.Name())
	}

	return fieldName, typeName, nil
}

func (aj AminoJSON) getZeroValue(field protoreflect.FieldDescriptor) (protoreflect.Value, bool) {
	opts := field.Options()
	if proto.HasExtension(opts, cosmos_proto.E_Scalar) {
		scalar := proto.GetExtension(opts, cosmos_proto.E_Scalar).(string)
		if fn, ok := aj.zeroValues[scalar]; ok {
			return fn(), true
		}
	}
	return field.Default(), false
}

func (aj AminoJSON) getMessageEncoder(message protoreflect.Message) MessageEncoder {
	opts := message.Descriptor().Options()
	if proto.HasExtension(opts, amino.E_MessageEncoding) {
		encoding := proto.GetExtension(opts, amino.E_MessageEncoding).(string)
		if fn, ok := aj.messageEncoders[encoding]; ok {
			return fn
		}
	}
	return nil
}

func isFieldEmbedded(fieldValue protoreflect.Value, field protoreflect.FieldDescriptor) bool {
	opts := field.Options()
	if proto.HasExtension(opts, amino.E_Embed) {
		embedded := proto.GetExtension(opts, amino.E_Embed).(bool)
		switch fieldValue.Interface().(type) {
		case protoreflect.Message:
			return embedded
		default:
			fmt.Printf("WARN: field %s is not a message, but has the embedded option set to true. Ignoring option.\n", field.Name())
			return false
		}
	}
	return false
}

func (aj AminoJSON) getFieldEncoding(field protoreflect.FieldDescriptor) FieldEncoder {
	opts := field.Options()
	if proto.HasExtension(opts, amino.E_Encoding) {
		enc := proto.GetExtension(opts, amino.E_Encoding).(string)
		if fn, ok := aj.fieldEncoders[enc]; ok {
			return fn
		}
	}
	if proto.HasExtension(opts, cosmos_proto.E_Scalar) {
		scalar := proto.GetExtension(opts, cosmos_proto.E_Scalar).(string)
		if fn, ok := aj.scalarEncoders[scalar]; ok {
			return fn
		}
	}
	return nil
}

const (
	timestampFullName protoreflect.FullName = "google.protobuf.Timestamp"
	durationFullName  protoreflect.FullName = "google.protobuf.Duration"
	anyFullName       protoreflect.FullName = "google.protobuf.Any"
)
