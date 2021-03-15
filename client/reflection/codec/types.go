package codec

import (
	"context"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// ReadonlyTypeRegistry defines an immutable protobuf type registry
type ReadonlyTypeRegistry interface {
	protoregistry.ExtensionTypeResolver
	protoregistry.MessageTypeResolver
}

// ReadonlyProtoFileRegistry defines an immutable protobuf file registry
type ReadonlyProtoFileRegistry interface {
	// FindFileByPath returns a protoreflect.FileDescriptor given the file
	// path, ex: proto/cosmos/base/v1beta1/coin.proto
	FindFileByPath(path string) (protoreflect.FileDescriptor, error)
	// FindDescriptorByName returns protoreflect.FileDescriptor given
	// its protoreflect.FullName
	FindDescriptorByName(name protoreflect.FullName) (protoreflect.Descriptor, error)
}

// ProtoImportsDownloader defines the behaviour of an object that can
// download the raw descriptors of protobuf files. It is used to resolve
// protobuf dependencies in a dynamic way.
type ProtoImportsDownloader interface {
	// DownloadDescriptorByPath returns the protobuf descriptor raw bytes given the file path
	DownloadDescriptorByPath(ctx context.Context, path string) (rawDescriptor []byte, err error)
}
