package aminojson

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"time"

	cosmos_proto "github.com/cosmos/cosmos-proto"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/api/amino"
	authapi "cosmossdk.io/api/cosmos/auth/v1beta1"
)

type MessageEncoder func(protoreflect.Message, io.Writer) error
type FieldEncoder func(AminoJSON, protoreflect.Value, io.Writer) error

type AminoJSON struct {
	// maps cosmos_proto.scalar -> zero value factory
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
		messageEncoders: map[string]MessageEncoder{
			"key_field": func(msg protoreflect.Message, w io.Writer) error {
				keyField := msg.Descriptor().Fields().ByName("key")
				if keyField == nil {
					return errors.New(`message encoder for key_field: no field named "key" found`)
				}

				bz := msg.Get(keyField).Bytes()

				if len(bz) == 0 {
					_, err := fmt.Fprint(w, "null")
					return err
				}

				_, err := fmt.Fprintf(w, `"%s"`, base64.StdEncoding.EncodeToString(bz))
				return err
			},
			"module_account": moduleAccountEncoder,
		},
		fieldEncoders: map[string]FieldEncoder{
			"empty_string": func(_ AminoJSON, v protoreflect.Value, writer io.Writer) error {
				_, err := writer.Write([]byte(`""`))
				return err
			},
			"json_default": func(_ AminoJSON, v protoreflect.Value, writer io.Writer) error {
				switch val := v.Interface().(type) {
				case string, bool, int32, uint32, uint64, int64, protoreflect.EnumNumber:
					return jsonMarshal(writer, val)
				default:
					return fmt.Errorf("unsupported type %T", val)
				}
			},
			// created to replicate: https://github.com/cosmos/cosmos-sdk/blob/be9bd7a8c1b41b115d58f4e76ee358e18a52c0af/types/coin.go#L199-L205
			"null_slice_as_empty": func(aj AminoJSON, v protoreflect.Value, writer io.Writer) error {
				switch list := v.Interface().(type) {
				case protoreflect.List:
					if list.Len() == 0 {
						_, err := writer.Write([]byte("[]"))
						return err
					}
					return aj.marshalList(list, writer)
				default:
					return fmt.Errorf("unsupported type %T", list)
				}
			},
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
	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		v := msg.Get(f)
		name := getFieldName(f)
		writeNil := false

		if !msg.Has(f) {
			if omitEmpty(f) {
				continue
			} else {
				zv, found := aj.getZeroValue(f)
				if found {
					v = zv
				} else if f.Cardinality() == protoreflect.Repeated {
					fmt.Printf("WARN: not supported: dont_omit_empty=true on empty repeated field: %s\n", name)
					//writeNil = true
				} else if f.Kind() == protoreflect.MessageKind && !v.Message().IsValid() {
					return errors.Errorf("not supported: dont_omit_empty=true on invalid (nil?) message field: %s", name)
					//fmt.Printf("WARN: not supported: dont_omit_empty=true on invalid (nil?) message field: %s\n", name)
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
		if fn, ok := aj.fieldEncoders[enc]; ok {
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
// https://github.com/cosmos/cosmos-sdk/blob/41a3dfeced2953beba3a7d11ec798d17ee19f506/x/auth/types/account.go#L230-L254
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

	bz, err := json.Marshal(pretty)
	//bz, err := json.Marshal(pretty)
	if err != nil {
		return err
	}
	_, err = w.Write(bz)
	return err
}

const (
	timestampFullName protoreflect.FullName = "google.protobuf.Timestamp"
	durationFullName  protoreflect.FullName = "google.protobuf.Duration"
	anyFullName       protoreflect.FullName = "google.protobuf.Any"
)

const (
	secondsName protoreflect.Name = "seconds"
	nanosName   protoreflect.Name = "nanos"
)

func marshalTimestamp(message protoreflect.Message, writer io.Writer) error {
	// PROTO3 SPEC:
	// Uses RFC 3339, where generated output will always be Z-normalized and uses 0, 3, 6 or 9 fractional digits.
	// Offsets other than "Z" are also accepted.

	fields := message.Descriptor().Fields()
	secondsField := fields.ByName(secondsName)
	if secondsField == nil {
		return fmt.Errorf("expected seconds field")
	}

	nanosField := fields.ByName(nanosName)
	if nanosField == nil {
		return fmt.Errorf("expected nanos field")
	}

	seconds := message.Get(secondsField).Int()
	nanos := message.Get(nanosField).Int()
	if nanos < 0 {
		return fmt.Errorf("nanos must be non-negative on timestamp %v", message)
	}

	t := time.Unix(seconds, nanos).UTC()
	var str string
	if nanos == 0 {
		str = t.Format(time.RFC3339)
	} else {
		str = t.Format(time.RFC3339Nano)
	}

	_, err := fmt.Fprintf(writer, `"%s"`, str)
	return err
}

// MaxDurationSeconds the maximum number of seconds (when expressed as nanoseconds) which can fit in an int64.
// gogoproto encodes google.protobuf.Duration as a time.Duration, which is 64-bit signed integer.
const MaxDurationSeconds = int64(math.MaxInt64/int(1e9)) - 1

func marshalDuration(message protoreflect.Message, writer io.Writer) error {
	fields := message.Descriptor().Fields()
	secondsField := fields.ByName(secondsName)
	if secondsField == nil {
		return fmt.Errorf("expected seconds field")
	}

	// todo
	// check signs are consistent
	seconds := message.Get(secondsField).Int()
	if seconds > MaxDurationSeconds {
		return fmt.Errorf("%d seconds would overflow an int64 when represented as nanoseconds", seconds)
	}

	nanosField := fields.ByName(nanosName)
	if nanosField == nil {
		return fmt.Errorf("expected nanos field")
	}

	nanos := message.Get(nanosField).Int()
	totalNanos := nanos + (seconds * 1e9)
	_, err := writer.Write([]byte(fmt.Sprintf(`"%d"`, totalNanos)))
	return err
}
