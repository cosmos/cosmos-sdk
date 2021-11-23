package ormvalue

import (
	"bytes"
	"io"

	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// Codec defines an interface for decoding and encoding values in ORM index keys.
type Codec interface {

	// Decode decodes a value in a key.
	Decode(r *bytes.Reader) (protoreflect.Value, error)

	// Encode encodes a value in a key.
	Encode(value protoreflect.Value, w io.Writer) error

	// Compare compares two values of this type and should primarily be used
	// for testing.
	Compare(v1, v2 protoreflect.Value) int

	// IsOrdered returns true if callers can always assume that this ordering
	// is suitable for sorted iteration.
	IsOrdered() bool

	// FixedSize returns a positive value if encoders should assume a fixed size
	// buffer for encoding. A codec may choose to estimate this value,
	// in the case of a varint for example, so that computation with Size isn't
	// needed. Encoders should only use the bytes actually written
	// by Encode.
	FixedSize() int

	// Size estimates the size needed to encode the field. It may over-estimate
	// by a small amount if this is more efficient, for instance in the case of
	// a varint. Encoders should only use the bytes actually written
	// by Encode.
	Size(value protoreflect.Value) (int, error)
}

var (
	timestampMsgType  = (&timestamppb.Timestamp{}).ProtoReflect().Type()
	timestampFullName = timestampMsgType.Descriptor().FullName()
	durationMsgType   = (&durationpb.Duration{}).ProtoReflect().Type()
	durationFullName  = durationMsgType.Descriptor().FullName()
)

// GetCodec returns the Codec for the provided field if one is defined.
// nonTerminal should be set to true if this value is being encoded as a
// non-terminal segment of a multi-part key.
func GetCodec(field protoreflect.FieldDescriptor, nonTerminal bool) (Codec, error) {
	if field.IsList() {
		return nil, ormerrors.UnsupportedKeyField.Wrapf("repeated field %s", field.FullName())
	}

	if field.ContainingOneof() != nil {
		return nil, ormerrors.UnsupportedKeyField.Wrapf("oneof field %s", field.FullName())
	}

	switch field.Kind() {
	case protoreflect.BytesKind:
		if nonTerminal {
			return NonTerminalBytesCodec{}, nil
		} else {
			return BytesCodec{}, nil
		}
	case protoreflect.StringKind:
		if nonTerminal {
			return NonTerminalStringCodec{}, nil
		} else {
			return StringCodec{}, nil
		}
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return Uint32Codec{}, nil
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return Uint64Codec{}, nil
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return Int32Codec{}, nil
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return Int64Codec{}, nil
	case protoreflect.BoolKind:
		return BoolCodec{}, nil
	case protoreflect.EnumKind:
		return EnumCodec{}, nil
	case protoreflect.MessageKind:
		msgName := field.Message().FullName()
		switch msgName {
		case timestampFullName:
			return TimestampCodec{}, nil
		case durationFullName:
			return DurationCodec{}, nil
		default:
			return nil, ormerrors.UnsupportedKeyField.Wrapf("%s of type %s", field.FullName(), msgName)
		}
	default:
		return nil, ormerrors.UnsupportedKeyField.Wrapf("%s of kind %s", field.FullName(), field.Kind())
	}
}
