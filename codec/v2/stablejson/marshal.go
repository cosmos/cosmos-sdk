package stablejson

import (
	"bytes"
	"io"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func Marshal(message proto.Message) ([]byte, error) {
	return MarshalOptions{}.Marshal(message)
}

type MarshalOptions struct {
	// HexBytes specifies whether bytes fields should be marshaled as upper-case
	// hex strings. If set to false, bytes fields will be encoded as standard
	// base64 strings as specified by the official proto3 JSON mapping.
	HexBytes bool

	// Resolver is used for looking up types when expanding google.protobuf.Any
	// messages. If nil, this defaults to using protoregistry.GlobalTypes.
	Resolver interface {
		protoregistry.ExtensionTypeResolver
		protoregistry.MessageTypeResolver
	}
}

func (opts MarshalOptions) Marshal(message proto.Message) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := opts.MarshalTo(message, buf)
	return buf.Bytes(), err
}

func (opts MarshalOptions) MarshalTo(message proto.Message, writer io.Writer) error {
	return opts.marshalMessage(message.ProtoReflect(), writer)
}
