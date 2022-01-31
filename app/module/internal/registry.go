package internal

import (
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/container"
)

var ModuleRegistry = map[protoreflect.FullName]*ModuleConfig{}

type ModuleConfig struct {
	Providers []container.ProviderDescriptor
}
