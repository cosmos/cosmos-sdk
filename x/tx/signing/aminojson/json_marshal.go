package aminojson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"cosmossdk.io/x/tx/signing"
)

// MessageEncoder is a function that can encode a protobuf protoreflect.Message to JSON.
type MessageEncoder func(*Encoder, protoreflect.Message, io.Writer) error

// FieldEncoder is a function that can encode a protobuf protoreflect.Value to JSON.
type FieldEncoder func(*Encoder, protoreflect.Value, io.Writer) error

// EncoderOptions are options for creating a new Encoder.
type EncoderOptions struct {
	// DonotSortFields when set turns off sorting of field names.
	DoNotSortFields bool
	// TypeResolver is used to resolve protobuf message types by TypeURL when marshaling any packed messages.
	TypeResolver protoregistry.MessageTypeResolver
	// FileResolver is used to resolve protobuf file descriptors TypeURL when TypeResolver fails.
	FileResolver signing.ProtoFileResolver
}

// Encoder is a JSON encoder that uses the Amino JSON encoding rules for protobuf messages.
type Encoder struct {
	// maps cosmos_proto.scalar -> field encoder
	scalarEncoders  map[string]FieldEncoder
	messageEncoders map[string]MessageEncoder
	fieldEncoders   map[string]FieldEncoder
	fileResolver    signing.ProtoFileResolver
	typeResolver    protoregistry.MessageTypeResolver
	doNotSortFields bool
}

// NewEncoder returns a new Encoder capable of serializing protobuf messages to JSON using the Amino JSON encoding
// rules.
func NewEncoder(options EncoderOptions) Encoder {
	if options.FileResolver == nil {
		options.FileResolver = protoregistry.GlobalFiles
	}
	if options.TypeResolver == nil {
		options.TypeResolver = protoregistry.GlobalTypes
	}
	enc := Encoder{
		scalarEncoders: map[string]FieldEncoder{
			"cosmos.Dec": cosmosDecEncoder,
			"cosmos.Int": cosmosIntEncoder,
		},
		messageEncoders: map[string]MessageEncoder{
			"key_field":        keyFieldEncoder,
			"module_account":   moduleAccountEncoder,
			"threshold_string": thresholdStringEncoder,
		},
		fieldEncoders: map[string]FieldEncoder{
			"legacy_coins":     nullSliceAsEmptyEncoder,
			"cosmos_dec_bytes": cosmosDecEncoder,
		},
		fileResolver:    options.FileResolver,
		typeResolver:    options.TypeResolver,
		doNotSortFields: options.DoNotSortFields,
	}
	return enc
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
func (enc Encoder) DefineMessageEncoding(name string, encoder MessageEncoder) Encoder {
	if enc.messageEncoders == nil {
		enc.messageEncoders = map[string]MessageEncoder{}
	}
	enc.messageEncoders[name] = encoder
	return enc
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
func (enc Encoder) DefineFieldEncoding(name string, encoder FieldEncoder) Encoder {
	if enc.fieldEncoders == nil {
		enc.fieldEncoders = map[string]FieldEncoder{}
	}
	enc.fieldEncoders[name] = encoder
	return enc
}

// Marshal serializes a protobuf message to JSON.
func (enc Encoder) Marshal(message proto.Message) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := enc.beginMarshal(message.ProtoReflect(), buf)
	return buf.Bytes(), err
}

func (enc Encoder) beginMarshal(msg protoreflect.Message, writer io.Writer) error {
	name, named := getMessageAminoName(msg.Descriptor().Options())
	if named {
		_, err := writer.Write([]byte(fmt.Sprintf(`{"type":"%s","value":`, name)))
		if err != nil {
			return err
		}
	}

	err := enc.marshal(protoreflect.ValueOfMessage(msg), writer)
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

func (enc Encoder) marshal(value protoreflect.Value, writer io.Writer) error {
	switch val := value.Interface().(type) {
	case protoreflect.Message:
		err := enc.marshalMessage(val, writer)
		return err

	case protoreflect.Map:
		return errors.New("maps are not supported")

	case protoreflect.List:
		if !val.IsValid() {
			_, err := writer.Write([]byte("null"))
			return err
		}
		return enc.marshalList(val, writer)

	case string, bool, int32, uint32, []byte, protoreflect.EnumNumber:
		return jsonMarshal(writer, val)

	case uint64, int64:
		_, err := fmt.Fprintf(writer, `"%d"`, val) // quoted
		return err

	default:
		return errors.Errorf("unknown type %T", val)
	}
}

type nameAndIndex struct {
	i    int
	name string
}

func (enc Encoder) marshalMessage(msg protoreflect.Message, writer io.Writer) error {
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
		return enc.marshalAny(msg, writer)
	}

	if encoder := enc.getMessageEncoder(msg); encoder != nil {
		err := encoder(&enc, msg, writer)
		return err
	}

	_, err := writer.Write([]byte("{"))
	if err != nil {
		return err
	}

	fields := msg.Descriptor().Fields()
	first := true
	emptyOneOfWritten := map[string]bool{}

	// 1. If permitted, ensure the names are sorted.
	indices := make([]*nameAndIndex, 0, fields.Len())
	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		name := getAminoFieldName(f)
		indices = append(indices, &nameAndIndex{i: i, name: name})
	}

	if shouldSortFields := !enc.doNotSortFields; shouldSortFields {
		sort.Slice(indices, func(i, j int) bool {
			ni, nj := indices[i], indices[j]
			return ni.name < nj.name
		})
	}

	for _, ni := range indices {
		i := ni.i
		name := ni.name
		f := fields.Get(i)
		v := msg.Get(f)
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
			switch {
			case isOneOf && msg.WhichOneof(oneof) == nil && !emptyOneOfWritten[oneofFieldName]:
				name = oneofFieldName
				writeNil = true
				emptyOneOfWritten[oneofFieldName] = true
			case omitEmpty(f):
				continue
			case f.Kind() == protoreflect.MessageKind &&
				f.Cardinality() != protoreflect.Repeated &&
				!v.Message().IsValid():
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
		if encoder := enc.getFieldEncoding(f); encoder != nil {
			err = encoder(&enc, v, writer)
			if err != nil {
				return err
			}
		} else if writeNil {
			_, err = writer.Write([]byte("null"))
			if err != nil {
				return err
			}
		} else {
			err = enc.marshal(v, writer)
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

func (enc Encoder) marshalList(list protoreflect.List, writer io.Writer) error {
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

		err = enc.marshal(list.Get(i), writer)
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
