package configinternal

import (
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/cosmos/cosmos-sdk/app/internal/protoregistry"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/app/internal"
	"github.com/cosmos/cosmos-sdk/container"
)

func LoadJSON(bz []byte) container.Option {
	// we don't use proto JSON here because we don't want to unmarshal Any's
	// until we have resolved the FileDescriptorSet we want to unmarshal
	// them with using pinned file descriptors.
	var cfg Config
	err := json.Unmarshal(bz, &cfg)
	if err != nil {
		return container.Error(err)
	}

	if cfg.Version != "v1alpha" {
		return container.Error(fmt.Errorf("unsupported config version %s", cfg.Version))
	}

	registryBuilder := protoregistry.NewRegistryBuilder()
	moduleInitializers := make([]*internal.ModuleInitializer, len(cfg.Modules))
	for i, mod := range cfg.Modules {
		var anyWrapper AnyWrapper
		err = json.Unmarshal(mod.Config, &anyWrapper)
		if err != nil {
			return container.Error(err)
		}

		url := anyWrapper.TypeURL
		fullName := protoreflect.FullName(url)
		if i := strings.LastIndexByte(url, '/'); i >= 0 {
			fullName = fullName[i+len("/"):]
		}

		moduleInitializer, ok := internal.ModuleRegistry[fullName]
		if !ok {
			return container.Error(fmt.Errorf("can't resolve module for config type URL %s", url))
		}

		moduleInitializers[i] = moduleInitializer

		err = registryBuilder.RegisterModule(fullName, moduleInitializer.PinnedProtoImageFS)
		if err != nil {
			return container.Error(err)
		}
	}

	registry, err := registryBuilder.Build()
	if err != nil {
		return container.Error(err)
	}

	unmarshalJsonOpts := protojson.UnmarshalOptions{
		Resolver: registry,
	}

	var containerOpts []container.Option

	for i, mod := range cfg.Modules {
		moduleInitializer := moduleInitializers[i]

		configObj := moduleInitializer.ConfigType.New().Interface()
		err = unmarshalJsonOpts.Unmarshal(mod.Config, configObj)
		if err != nil {
			return container.Error(err)
		}

		for _, factory := range moduleInitializer.ProviderFactories {
			containerOpts = append(containerOpts, container.Provide(factory(configObj)))
		}
	}

	return container.Options(containerOpts...)
}

// Config is a pure golang version of cosmos.app.v1alpha1.Config that won't unmarshal Any's.
type Config struct {
	Version string   `json:"version"`
	Modules []Module `json:"modules"`
}

// Module is a pure golang version of cosmos.app.v1alpha1.ModuleConfig that won't unmarshal Any's.
type Module struct {
	Name   string          `json:"name"`
	Config json.RawMessage `json:"config"`
}

type AnyWrapper struct {
	TypeURL string `json:"@type"`
}
