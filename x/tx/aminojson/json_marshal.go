package aminojson

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	cosmos_proto "github.com/cosmos/cosmos-proto"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/api/amino"
)

type JSONMarshaller interface {
	MarshalAmino(proto.Message) ([]byte, error)
}

type AminoJSON struct {
	// maps cosmos_proto.scalar -> zero value factory
	zeroValues      map[string]func() protoreflect.Value
	messageEncoders map[string]func(message protoreflect.Message) protoreflect.Value
}

func NewAminoJSON() AminoJSON {
	aj := AminoJSON{
		zeroValues: map[string]func() protoreflect.Value{
			"cosmos.Dec": func() protoreflect.Value {
				return protoreflect.ValueOfString("0")
			},
		},
		messageEncoders: map[string]func(message protoreflect.Message) protoreflect.Value{
			"key_field": func(message protoreflect.Message) protoreflect.Value {
				// todo
				// err handling
				keyField := message.Get(message.Descriptor().Fields().ByName("key"))
				bz := keyField.Bytes()
				return protoreflect.ValueOfBytes(bz)
			},
		},
	}
	return aj
}

func MarshalAmino(message proto.Message) ([]byte, error) {
	buf := &bytes.Buffer{}
	aj := NewAminoJSON()
	vmsg := protoreflect.ValueOfMessage(message.ProtoReflect())
	err := aj.marshal(vmsg, buf)
	return buf.Bytes(), err
}

func (aj AminoJSON) marshal(value protoreflect.Value, writer io.Writer) error {
	switch val := value.Interface().(type) {
	case protoreflect.Message:
		return aj.marshalMessage(val, writer)

	case protoreflect.Map:
		return errors.New("maps are not supported")

	case protoreflect.List:
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

func (aj AminoJSON) marshalMessage(msg protoreflect.Message, writer io.Writer) error {
	if encoded, encodingOption := aj.encodeMessage(msg); encodingOption {
		return aj.marshal(encoded, writer)
	}

	named := false
	opts := msg.Descriptor().Options()
	if proto.HasExtension(opts, amino.E_Name) {
		name := proto.GetExtension(opts, amino.E_Name)
		_, err := writer.Write([]byte(fmt.Sprintf(`{"type":"%s","value":`, name)))
		if err != nil {
			return err
		}
		named = true
	}

	_, err := writer.Write([]byte("{"))
	if err != nil {
		return err
	}

	fields := msg.Descriptor().Fields()
	first := true
	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		v := msg.Get(f)
		name := fieldName(f)

		if !msg.Has(f) {
			if omitEmpty(f) {
				continue
			} else {
				zv, found := aj.fieldZeroValue(f)
				if found {
					v = zv
				}
			}
		}

		if !first {
			_, err = writer.Write([]byte(","))
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

		err = aj.marshal(v, writer)
		if err != nil {
			return err
		}

		first = false
	}

	_, err = writer.Write([]byte("}"))
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

	// replicate https://github.com/tendermint/go-amino/blob/rc0/v0.16.0/json-encode.go#L222
	if n == 0 {
		_, err := writer.Write([]byte("null"))
		return err
	}

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

// omitEmpty returns true if the field should be omitted if empty. Empty field omission is the default behavior.
func omitEmpty(field protoreflect.FieldDescriptor) bool {
	opts := field.Options()
	if proto.HasExtension(opts, amino.E_DontOmitempty) {
		dontOmitEmpty := proto.GetExtension(opts, amino.E_DontOmitempty).(bool)
		return !dontOmitEmpty
	}
	// legacy support for gogoproto would need to look something like below.
	//
	// if gproto.GetBoolExtension(opts, gogoproto.E_Nullable, true) {
	//
	// }
	return true
}

func fieldName(field protoreflect.FieldDescriptor) string {
	opts := field.Options()
	if proto.HasExtension(opts, amino.E_FieldName) {
		return proto.GetExtension(opts, amino.E_FieldName).(string)
	}
	return string(field.Name())
}

func (aj AminoJSON) fieldZeroValue(field protoreflect.FieldDescriptor) (protoreflect.Value, bool) {
	opts := field.Options()
	if proto.HasExtension(opts, cosmos_proto.E_Scalar) {
		scalar := proto.GetExtension(opts, cosmos_proto.E_Scalar).(string)
		if fn, ok := aj.zeroValues[scalar]; ok {
			return fn(), true
		}
	}
	return field.Default(), false
}

func (aj AminoJSON) encodeMessage(message protoreflect.Message) (protoreflect.Value, bool) {
	opts := message.Descriptor().Options()
	if proto.HasExtension(opts, amino.E_MessageEncoding) {
		encoding := proto.GetExtension(opts, amino.E_MessageEncoding).(string)
		if fn, ok := aj.messageEncoders[encoding]; ok {
			return fn(message), true
		}
	}
	return protoreflect.Value{}, false
}
