package ormsql

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type schema struct {
	jsonMarshalOptions   protojson.MarshalOptions
	jsonUnmarshalOptions protojson.UnmarshalOptions
	resolver             protoregistry.MessageTypeResolver
	messageCodecs        map[protoreflect.FullName]*messageCodec
	Error                error
}
