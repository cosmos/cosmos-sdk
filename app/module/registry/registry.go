package registry

import (
	"github.com/cosmos/cosmos-sdk/app/internal"
	"google.golang.org/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/container"
)

func Resolve(moduleConfig proto.Message, moduleName string) container.Option {
	config, ok := internal.ModuleRegistry[moduleConfig.ProtoReflect().Descriptor().FullName()]
	if !ok {
		return container.Options()
	}

	var opts []container.Option
	for _, provider := range config.ProviderFactories {
		opts = append(opts, container.ProvideInModule(moduleName, provider))
	}

	return container.Options(opts...)
}
