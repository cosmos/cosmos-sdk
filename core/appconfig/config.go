package appconfig

import (
	"fmt"
	"reflect"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/anypb"
	"sigs.k8s.io/yaml"

	"github.com/cosmos/cosmos-sdk/container"

	appv1alpha1 "github.com/cosmos/cosmos-sdk/api/cosmos/app/v1alpha1"

	"cosmossdk.io/core/internal"
)

// LoadJSON loads an app config in JSON format.
func LoadJSON(bz []byte) container.Option {
	config := &appv1alpha1.Config{}
	err := protojson.Unmarshal(bz, config)
	if err != nil {
		return container.Error(err)
	}

	return Compose(config)
}

// LoadYAML loads an app config in YAML format.
func LoadYAML(bz []byte) container.Option {
	j, err := yaml.YAMLToJSON(bz)
	if err != nil {
		return container.Error(err)
	}

	return LoadJSON(j)
}

// Compose composes a v1alpha1 app config into a container option by resolving
// the required modules and composing their options.
func Compose(appConfig *appv1alpha1.Config) container.Option {
	opts := []container.Option{
		container.Supply(appConfig),
	}

	for _, module := range appConfig.Modules {
		if module.Name == "" {
			return container.Error(fmt.Errorf("module is missing name"))
		}

		if module.Config == nil {
			return container.Error(fmt.Errorf("module %q is missing a config object", module.Name))
		}

		msgType, err := protoregistry.GlobalTypes.FindMessageByURL(module.Config.TypeUrl)
		if err != nil {
			return container.Error(err)
		}

		modules, err := internal.ModulesByProtoMessageName()
		if err != nil {
			return container.Error(err)
		}

		init, ok := modules[msgType.Descriptor().FullName()]
		if !ok {
			modDesc := proto.GetExtension(msgType.Descriptor().Options(), appv1alpha1.E_Module).(*appv1alpha1.ModuleDescriptor)
			if modDesc == nil {
				return container.Error(fmt.Errorf("no module registered for type URL %s and that protobuf type does not have the option %s\n\n%s",
					module.Config.TypeUrl, appv1alpha1.E_Module.TypeDescriptor().FullName(), dumpRegisteredModules(modules)))
			}

			return container.Error(fmt.Errorf("no module registered for type URL %s, did you forget to import %s\n\n%s",
				module.Config.TypeUrl, modDesc.GoImport, dumpRegisteredModules(modules)))
		}

		config := init.ConfigProtoMessage.ProtoReflect().Type().New().Interface()
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

func dumpRegisteredModules(modules map[protoreflect.FullName]*internal.ModuleInitializer) string {
	var mods []string
	for name := range modules {
		mods = append(mods, "  "+string(name))
	}
	return fmt.Sprintf("registered modules are:\n%s", strings.Join(mods, "\n"))
}
