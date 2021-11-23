package ormvalue

import (
	"bytes"
	"fmt"
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type Codec interface {
	Decode(r *bytes.Reader) (protoreflect.Value, error)
	Encode(value protoreflect.Value, w io.Writer) error
	Compare(v1, v2 protoreflect.Value) int
	IsOrdered() bool

	// FixedSize returns a positive value if this codec encodes values with
	// a fixed number of bytes.
	FixedSize() int

	// Size computes the size of the field for the given value.
	Size(value protoreflect.Value) (int, error)
}

func MakeCodec(field protoreflect.FieldDescriptor, nonTerminal bool) (Codec, error) {
	if field.IsList() {
		return nil, fmt.Errorf("repeated fields aren't supported in keys")
	}

	if field.ContainingOneof() != nil {
		return nil, fmt.Errorf("oneof fields aren't supported in keys")
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
	case protoreflect.Uint32Kind:
		return Uint32Codec{}, nil
	case protoreflect.Uint64Kind:
		return Uint64Codec{}, nil
	default:
		return nil, fmt.Errorf("unsupported index key kind %s", field.Kind())
	}
}
