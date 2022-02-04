package internal

import (
	"embed"

	"google.golang.org/protobuf/proto"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/container"
)

var ModuleRegistry = map[protoreflect.FullName]*ModuleInitializer{}

type ModuleInitializer struct {
	ConfigType         protoreflect.MessageType
	PinnedProtoImageFS embed.FS
	ProviderFactories  []func(proto.Message) container.ProviderDescriptor
}
