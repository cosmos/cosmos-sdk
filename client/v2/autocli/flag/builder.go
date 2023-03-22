package flag

import (
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// Builder manages options for building pflag flags for protobuf messages.
type Builder struct {
	// TypeResolver specifies how protobuf types will be resolved. If it is
	// nil protoregistry.GlobalTypes will be used.
	TypeResolver interface {
		protoregistry.MessageTypeResolver
		protoregistry.ExtensionTypeResolver
	}

	// FileResolver specifies how protobuf file descriptors will be resolved. If it is
	// nil protoregistry.GlobalFiles will be used.
	FileResolver protodesc.Resolver

	messageFlagTypes map[protoreflect.FullName]Type
	scalarFlagTypes  map[string]Type

	// AddressPrefix is the prefix for the address flag
	AddressPrefix string
	// reflectionClient is the reflection client for the address flag
	GetClientConn func() (grpc.ClientConnInterface, error)
}

func (b *Builder) init() {
	if b.messageFlagTypes == nil {
		b.messageFlagTypes = map[protoreflect.FullName]Type{}
		b.messageFlagTypes["google.protobuf.Timestamp"] = timestampType{}
		b.messageFlagTypes["google.protobuf.Duration"] = durationType{}
	}

	if b.scalarFlagTypes == nil {
		b.scalarFlagTypes = map[string]Type{}
		b.scalarFlagTypes["cosmos.AddressString"] = addressStringType{}
	}
}

func (b *Builder) DefineMessageFlagType(messageName protoreflect.FullName, flagType Type) {
	b.init()
	b.messageFlagTypes[messageName] = flagType
}

func (b *Builder) DefineScalarFlagType(scalarName string, flagType Type) {
	b.init()
	b.scalarFlagTypes[scalarName] = flagType
}
