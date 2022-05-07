package appconfig

import (
	"fmt"
	"reflect"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/anypb"
	"sigs.k8s.io/yaml"

	"github.com/cosmos/cosmos-sdk/container"

	appv1alpha1 "github.com/cosmos/cosmos-sdk/api/cosmos/app/v1alpha1"

	"cosmossdk.io/core/internal"
)

func LoadJSON(bz []byte) container.Option {
	config := &appv1alpha1.Config{}
	err := protojson.Unmarshal(bz, config)
	if err != nil {
		return container.Error(err)
	}

	return Compose(config)
}

func LoadYAML(bz []byte) container.Option {
	j, err := yaml.YAMLToJSON(bz)
	if err != nil {
		return container.Error(err)
	}

	return LoadJSON(j)
}

func Compose(appConfig *appv1alpha1.Config) container.Option {
	opts := []container.Option{
		container.Supply(appConfig),
	}

	for _, module := range appConfig.Modules {
		if module.Name == "" {
			return container.Error(fmt.Errorf("module is missing name"))
		}

		if module.Config == nil {
			return container.Error(fmt.Errorf("module %s is missing a config object", module.Name))
		}

		msgType, err := protoregistry.GlobalTypes.FindMessageByURL(module.Config.TypeUrl)
		if err != nil {
			return container.Error(err)
		}

		init, ok := internal.ModuleRegistry[msgType.Descriptor().FullName()]
		if !ok {
			modDesc := proto.GetExtension(msgType.Descriptor().Options(), appv1alpha1.E_Module).(*appv1alpha1.ModuleDescriptor)
			if modDesc == nil {
				return container.Error(fmt.Errorf("no module registered for type URL %s and that protobuf type does not have the option %s",
					module.Config.TypeUrl, appv1alpha1.E_Module.TypeDescriptor().FullName()))
			}

			return container.Error(fmt.Errorf("no module registered for type URL %s, did you forget to import %s\n\n%s",
				module.Config.TypeUrl, modDesc.GoImport, dumpRegisteredModules()))
		}

		config := init.ConfigProtoType.New().Interface()
		err = anypb.UnmarshalTo(module.Config, config, proto.UnmarshalOptions{})
		if err != nil {

			return container.Error(err)
		}

		opts = append(opts, container.Provide(container.ProviderDescriptor{
			Inputs:  nil,
			Outputs: []container.ProviderOutput{{Type: init.ConfigGoType}},
			Fn: func(values []reflect.Value) ([]reflect.Value, error) {
				return []reflect.Value{reflect.ValueOf(config)}, nil
			},
			Location: container.LocationFromCaller(0),
		}))

		for _, provider := range init.Providers {
			opts = append(opts, container.ProvideInModule(module.Name, provider))
		}
	}

	return container.Options(opts...)
}

func dumpRegisteredModules() string {
	var mods []string
	for name := range internal.ModuleRegistry {
		mods = append(mods, "  "+string(name))
	}
	return fmt.Sprintf("modules are:\n%s", strings.Join(mods, "\n"))
}

func MustWrapAny(message proto.Message) *anypb.Any {
	a, err := anypb.New(message)
	if err != nil {
		panic(err)
	}
	return a
}
