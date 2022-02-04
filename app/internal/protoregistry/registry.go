package protoregistry

import (
	"embed"

	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type Registry interface {
	protoregistry.MessageTypeResolver
	protoregistry.ExtensionTypeResolver
	protodesc.Resolver
}

type RegistryBuilder struct{}

func NewRegistryBuilder() *RegistryBuilder {
	return &RegistryBuilder{}
}

func (b *RegistryBuilder) RegisterModule(
	moduleConfigType protoreflect.FullName,
	pinnedProtoImage embed.FS,
) error {
	// TODO
	return nil
}

func (b *RegistryBuilder) Build() (Registry, error) {
	// This implementation naively uses the global registry and will
	// handle unpacking pinned file descriptors and validation later.
	return &registry{
		Files: protoregistry.GlobalFiles,
		Types: protoregistry.GlobalTypes,
	}, nil
}

type registry struct {
	*protoregistry.Files
	*protoregistry.Types
}
