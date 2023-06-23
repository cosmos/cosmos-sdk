package ormfield

import (
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"cosmossdk.io/orm/types/ormerrors"
)

// Codec defines an interface for decoding and encoding values in ORM index keys.
type Codec interface {
	// Decode decodes a value in a key.
	Decode(r Reader) (protoreflect.Value, error)

	// Encode encodes a value in a key.
	Encode(value protoreflect.Value, w io.Writer) error

	// Compare compares two values of this type and should primarily be used
	// for testing.
	Compare(v1, v2 protoreflect.Value) int

	// IsOrdered returns true if callers can always assume that this ordering
	// is suitable for sorted iteration.
	IsOrdered() bool

	// FixedBufferSize returns a positive value if encoders should assume a
	// fixed size buffer for encoding. Encoders will use at most this much size
	// to encode the value.
	FixedBufferSize() int

	// ComputeBufferSize estimates the buffer size needed to encode the field.
	// Encoders will use at most this much size to encode the value.
	ComputeBufferSize(value protoreflect.Value) (int, error)
}

type Reader interface {
	io.Reader
	io.ByteReader
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
	if field == nil {
		return nil, ormerrors.InvalidKeyField.Wrap("nil field")
	}
	if field.IsList() {
		return nil, ormerrors.InvalidKeyField.Wrapf("repeated field %s", field.FullName())
	}

	if field.ContainingOneof() != nil {
		return nil, ormerrors.InvalidKeyField.Wrapf("oneof field %s", field.FullName())
	}

	if field.HasOptionalKeyword() {
		return nil, ormerrors.InvalidKeyField.Wrapf("optional field %s", field.FullName())
	}

	switch field.Kind() {
	case protoreflect.BytesKind:
		if nonTerminal {
			return NonTerminalBytesCodec{}, nil
		}

		return BytesCodec{}, nil
	case protoreflect.StringKind:
		if nonTerminal {
			return NonTerminalStringCodec{}, nil
		}

		return StringCodec{}, nil

	case protoreflect.Uint32Kind:
		return CompactUint32Codec{}, nil
	case protoreflect.Fixed32Kind:
		return FixedUint32Codec{}, nil
	case protoreflect.Uint64Kind:
		return CompactUint64Codec{}, nil
	case protoreflect.Fixed64Kind:
		return FixedUint64Codec{}, nil
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
			return nil, ormerrors.InvalidKeyField.Wrapf("%s of type %s", field.FullName(), msgName)
		}
	default:
		return nil, ormerrors.InvalidKeyField.Wrapf("%s of kind %s", field.FullName(), field.Kind())
	}
}
