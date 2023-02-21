package aminojson

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// MessageEncoder is a function that can encode a protobuf protoreflect.Message to JSON.
type MessageEncoder func(protoreflect.Message, io.Writer) error

// FieldEncoder is a function that can encode a protobuf protoreflect.Value to JSON.
type FieldEncoder func(AminoJSON, protoreflect.Value, io.Writer) error

// AminoJSON is a JSON encoder that uses the Amino JSON encoding rules.
type AminoJSON struct {
	// maps cosmos_proto.scalar -> zero value factory
	scalarEncoders  map[string]FieldEncoder
	messageEncoders map[string]MessageEncoder
	fieldEncoders   map[string]FieldEncoder
}

// NewAminoJSON returns a new AminoJSON encoder.
func NewAminoJSON() AminoJSON {
	aj := AminoJSON{
		scalarEncoders: map[string]FieldEncoder{
			"cosmos.Dec": cosmosDecEncoder,
			"cosmos.Int": cosmosIntEncoder,
		},
		messageEncoders: map[string]MessageEncoder{
			"key_field":      keyFieldEncoder,
			"module_account": moduleAccountEncoder,
		},
		fieldEncoders: map[string]FieldEncoder{
			"legacy_coins":     nullSliceAsEmptyEncoder,
			"cosmos_dec_bytes": cosmosDecEncoder,
		},
	}
	return aj
}

// DefineMessageEncoding defines a custom encoding for a protobuf message.  The `name` field must match a usage of
// an (amino.message_encoding) option in the protobuf message as in the following example.  This encoding will be
// used instead of the default encoding for all usages of the tagged message.
//
//	message ModuleAccount {
//	  option (amino.name)                        = "cosmos-sdk/ModuleAccount";
//	  option (amino.message_encoding)            = "module_account";
//	  ...
//	}
func (aj AminoJSON) DefineMessageEncoding(name string, encoder MessageEncoder) {
	aj.messageEncoders[name] = encoder
}

// DefineFieldEncoding defines a custom encoding for a protobuf field.  The `name` field must match a usage of
// an (amino.encoding) option in the protobuf message as in the following example. This encoding will be used
// instead of the default encoding for all usages of the tagged field.
//
//	message Balance {
//	  repeated cosmos.base.v1beta1.Coin coins = 2 [
//	    (amino.encoding)         = "legacy_coins",
//	    (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins",
//	    (gogoproto.nullable)     = false,
//	    (amino.dont_omitempty)   = true
//	  ];
//	  ...
//	}
func (aj AminoJSON) DefineFieldEncoding(name string, encoder FieldEncoder) {
	aj.fieldEncoders[name] = encoder
}

// MarshalAmino serializes a protobuf message to JSON.
func (aj AminoJSON) MarshalAmino(message proto.Message) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := aj.beginMarshal(message.ProtoReflect(), buf)
	return buf.Bytes(), err
}

func (aj AminoJSON) beginMarshal(msg protoreflect.Message, writer io.Writer) error {
	name, named := getMessageAminoName(msg)
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
		name := getAminoFieldName(f)
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
			if isOneOf && msg.WhichOneof(oneof) == nil && !emptyOneOfWritten[oneofFieldName] {
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

const (
	timestampFullName protoreflect.FullName = "google.protobuf.Timestamp"
	durationFullName  protoreflect.FullName = "google.protobuf.Duration"
	anyFullName       protoreflect.FullName = "google.protobuf.Any"
)
