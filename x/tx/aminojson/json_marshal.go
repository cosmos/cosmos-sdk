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
	authapi "cosmossdk.io/api/cosmos/auth/v1beta1"
)

type MessageEncoder func(protoreflect.Message, io.Writer) error
type FieldEncoder func(protoreflect.Value, io.Writer) error

type AminoJSON struct {
	// maps cosmos_proto.scalar -> zero value factory
	zeroValues      map[string]func() protoreflect.Value
	messageEncoders map[string]MessageEncoder
	encodings       map[string]FieldEncoder
}

func NewAminoJSON() AminoJSON {
	aj := AminoJSON{
		zeroValues: map[string]func() protoreflect.Value{
			"cosmos.Dec": func() protoreflect.Value {
				return protoreflect.ValueOfString("0")
			},
		},
		messageEncoders: map[string]MessageEncoder{
			"key_field": func(msg protoreflect.Message, w io.Writer) error {
				keyField := msg.Descriptor().Fields().ByName("key")
				if keyField == nil {
					return errors.New(`message encoder for key_field: no field named "key" found`)
				}
				bz := msg.Get(keyField).Bytes()
				_, err := fmt.Fprintf(w, `"%s"`, base64.StdEncoding.EncodeToString(bz))
				if err != nil {
					return err
				}
				return nil
			},
			"module_account": moduleAccountEncoder,
		},
		encodings: map[string]FieldEncoder{
			"empty_string": func(v protoreflect.Value, writer io.Writer) error {
				_, err := writer.Write([]byte(`""`))
				return err
			},
			"json_default": func(v protoreflect.Value, writer io.Writer) error {
				switch val := v.Interface().(type) {
				case string, bool, int32, uint32, uint64, int64, protoreflect.EnumNumber:
					return jsonMarshal(writer, val)
				default:
					return fmt.Errorf("unsupported type %T", val)
				}
			},
		},
	}
	return aj
}

func (aj AminoJSON) DefineMessageEncoding(name string, encoder MessageEncoder) {
	aj.messageEncoders[name] = encoder
}

func (aj AminoJSON) MarshalAmino(message proto.Message) ([]byte, error) {
	buf := &bytes.Buffer{}
	//vmsg := protoreflect.ValueOfMessage(message.ProtoReflect())
	err := aj.beginMarshal(message.ProtoReflect(), buf)
	return buf.Bytes(), err
}

// TODO
// move into marshalMessage
func (aj AminoJSON) beginMarshal(msg protoreflect.Message, writer io.Writer) error {
	if encoder := aj.getMessageEncoder(msg); encoder != nil {
		err := encoder(msg, writer)
		return err
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

	//if encoder := aj.getMessageEncoder(msg); encoder != nil {
	//	err := encoder(msg, writer)
	//	if err != nil {
	//		return err
	//	}
	//} else {
	err := aj.marshal(protoreflect.ValueOfMessage(msg), writer)
	if err != nil {
		return err
	}
	//}

	if named {
		_, err := writer.Write([]byte("}"))
		if err != nil {
			return err
		}
	}

	return nil
}

func (aj AminoJSON) marshal(value protoreflect.Value, writer io.Writer) error {
	switch val := value.Interface().(type) {
	case protoreflect.Message:
		_, err := writer.Write([]byte("{"))
		if err != nil {
			return err
		}
		err = aj.marshalMessage(val, writer)
		if err != nil {
			return err
		}
		_, err = writer.Write([]byte("}"))
		return err

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

		// TODO timestamp
		// this is what is breaking MsgGrant

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
			err = encoder(v, writer)
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
	if encoder := aj.getMessageEncoder(msg); encoder != nil {
		err := encoder(msg, writer)
		return err
	}

	//named := false

	//opts := msg.Descriptor().Options()
	//if proto.HasExtension(opts, amino.E_Name) {
	//	name := proto.GetExtension(opts, amino.E_Name)
	//	_, err := writer.Write([]byte(fmt.Sprintf(`{"type":"%s","value":`, name)))
	//	if err != nil {
	//		return err
	//	}
	//	named = true
	//}

	//_, err := writer.Write([]byte("{"))
	//if err != nil {
	//	return err
	//}
	var err error

	fields := msg.Descriptor().Fields()
	first := true
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
			err = encoder(v, writer)
			if err != nil {
				return err
			}
		} else {
			err = aj.marshal(v, writer)
			if err != nil {
				return err
			}
		}

		first = false
	}

	//_, err = writer.Write([]byte("}"))
	//if err != nil {
	//	return err
	//}
	//
	//if named {
	//	_, err = writer.Write([]byte("}"))
	//	if err != nil {
	//		return err
	//	}
	//}

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

func getFieldName(field protoreflect.FieldDescriptor) string {
	opts := field.Options()
	if proto.HasExtension(opts, amino.E_FieldName) {
		return proto.GetExtension(opts, amino.E_FieldName).(string)
	}
	return string(field.Name())
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
		if fn, ok := aj.encodings[enc]; ok {
			return fn
		}
	}
	return nil
}

type moduleAccountPretty struct {
	Address       string   `json:"address"`
	PubKey        string   `json:"public_key"`
	AccountNumber uint64   `json:"account_number"`
	Sequence      uint64   `json:"sequence"`
	Name          string   `json:"name"`
	Permissions   []string `json:"permissions"`
}

type typeWrapper struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

// moduleAccountEncoder replicates the behavior in
// https://github.com/cosmos/cosmos-sdk/blob/41a3dfeced2953beba3a7d11ec798d17ee19f506/x/auth/types/account.go#L230-L2549
func moduleAccountEncoder(msg protoreflect.Message, w io.Writer) error {
	ma := msg.Interface().(*authapi.ModuleAccount)
	pretty := moduleAccountPretty{
		PubKey:      "",
		Name:        ma.Name,
		Permissions: ma.Permissions,
	}
	if ma.BaseAccount != nil {
		pretty.Address = ma.BaseAccount.Address
		pretty.AccountNumber = ma.BaseAccount.AccountNumber
		pretty.Sequence = ma.BaseAccount.Sequence
	} else {
		pretty.Address = ""
		pretty.AccountNumber = 0
		pretty.Sequence = 0
	}

	bz, err := json.Marshal(typeWrapper{Type: "cosmos-sdk/ModuleAccount", Value: pretty})
	//bz, err := json.Marshal(pretty)
	if err != nil {
		return err
	}
	_, err = w.Write(bz)
	return err
}
