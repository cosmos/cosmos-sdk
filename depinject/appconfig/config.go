package appconfig

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-proto/anyutil"
	"google.golang.org/protobuf/encoding/protojson"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/anypb"
	"sigs.k8s.io/yaml"

	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	appv2 "cosmossdk.io/api/cosmos/app/v2"
	"cosmossdk.io/depinject"
	internal "cosmossdk.io/depinject/internal/appconfig"

	"github.com/cosmos/gogoproto/jsonpb"
	"github.com/cosmos/gogoproto/proto"
	gogotypes "github.com/cosmos/gogoproto/types"
)

// LoadJSON loads an app config in JSON format.
func LoadJSON(bz []byte) depinject.Config {

	config := &appv1alpha1.Config{}
	err := protojson.Unmarshal(bz, config)
	if err == nil {
		return Compose(config)
	}

	gogoConfig := &appv2.Config{}
	err = jsonpb.UnmarshalString(string(bz), gogoConfig)
	if err == nil {
		return Compose(gogoConfig)
	}

	return depinject.Error(err)
}

// LoadYAML loads an app config in YAML format.
func LoadYAML(bz []byte) depinject.Config {
	j, err := yaml.YAMLToJSON(bz)
	if err != nil {
		return depinject.Error(err)
	}

	return LoadJSON(j)
}

// WrapAny marshals a proto message into a proto Any instance
func WrapAny(config protoreflect.ProtoMessage) *anypb.Any {
	cfg, err := anyutil.New(config)
	if err != nil {
		panic(err)
	}

	return cfg
}

type ModuleConfigI interface {
	GetName() string
	GetTypeUrl() (string, error)
	GetGolangBindingsStrings() ([]string, []string)
}

// Compose composes a v1alpha1 app config into a container option by resolving
// the required modules and composing their options.
func Compose(appConfig proto.Message) depinject.Config {

	opts := []depinject.Config{
		depinject.Supply(appConfig),
	}

	modules, err := internal.ModulesByProtoMessageName()
	if err != nil {
		return depinject.Error(err)
	}

	var configModules []ModuleConfigI

	appConfigV2, isProtov2 := appConfig.(*appv1alpha1.Config)
	if isProtov2 {
		for _, m := range appConfigV2.Modules {
			configModules = append(configModules, m)

		}
	} else {
		appConfigV1 := appConfig.(*appv2.Config)
		for _, m := range appConfigV1.Modules {
			configModules = append(configModules, m)
		}
	}

	for _, module := range configModules {
		if module.GetName() == "" {
			return depinject.Error(fmt.Errorf("module is missing name"))
		}

		msgTypeUrl, err := module.GetTypeUrl()
		if err != nil {
			return depinject.Error(err)
		}

		var descriptor protoreflect.MessageDescriptor

		var moduleV2 *appv1alpha1.ModuleConfig
		var moduleV1 *appv2.ModuleConfig

		if isProtov2 {
			moduleV2 = module.(*appv1alpha1.ModuleConfig)
		} else {
			moduleV1 = module.(*appv2.ModuleConfig)
		}

		if isProtov2 {
			msgType, err := protoregistry.GlobalTypes.FindMessageByURL(msgTypeUrl)
			if err != nil {
				return depinject.Error(err)
			}
			descriptor = msgType.Descriptor()
		} else {
			des, err := proto.GogoResolver.FindDescriptorByName(protoreflect.FullName(GetProtoV1MsgNameFromAnyTypeUrl(msgTypeUrl)))
			if err != nil {
				return depinject.Error(err)
			}
			descriptor = des.(protoreflect.MessageDescriptor)
		}

		init, ok := modules[descriptor.FullName()]
		if !ok {
			modDesc := protov2.GetExtension(descriptor.Options(), appv1alpha1.E_Module).(*appv1alpha1.ModuleDescriptor)
			if modDesc == nil {
				return depinject.Error(fmt.Errorf("no module registered for type URL %s and that protobuf type does not have the option %s\n\n%s",
					msgTypeUrl, appv1alpha1.E_Module.TypeDescriptor().FullName(), dumpRegisteredModules(modules)))
			}

			return depinject.Error(fmt.Errorf("no module registered for type URL %s, did you forget to import %s: find more information on how to make a module ready for app wiring: https://docs.cosmos.network/main/building-modules/depinject\n\n%s",
				msgTypeUrl, modDesc.GoImport, dumpRegisteredModules(modules)))
		}

		if isProtov2 {
			config := init.ConfigProtoMessage.(protov2.Message).ProtoReflect().Type().New().Interface()
			err = anypb.UnmarshalTo(moduleV2.Config, config, protov2.UnmarshalOptions{})
			if err != nil {

				return depinject.Error(err)
			}
			opts = append(opts, depinject.Supply(config))
			for _, binding := range moduleV2.GolangBindings {
				opts = append(opts, depinject.BindInterfaceInModule(moduleV2.GetName(), binding.InterfaceType, binding.Implementation))
			}
		} else {
			config := init.ConfigProtoMessage.(proto.Message)

			err = gogotypes.UnmarshalAny(moduleV1.Config, config)
			if err != nil {
				return depinject.Error(err)
			}
			opts = append(opts, depinject.Supply(config))
			for _, binding := range moduleV1.GolangBindings {
				opts = append(opts, depinject.BindInterfaceInModule(moduleV1.GetName(), binding.InterfaceType, binding.Implementation))
			}
		}

		for _, provider := range init.Providers {
			opts = append(opts, depinject.ProvideInModule(module.GetName(), provider))
		}

		for _, invoker := range init.Invokers {
			opts = append(opts, depinject.InvokeInModule(module.GetName(), invoker))
		}

		interfaceTypes, implementations := module.GetGolangBindingsStrings()
		for i, _ := range interfaceTypes {
			opts = append(opts, depinject.BindInterfaceInModule(module.GetName(), interfaceTypes[i], implementations[i]))
		}
	}

	if isProtov2 {
		for _, binding := range appConfigV2.GolangBindings {
			opts = append(opts, depinject.BindInterface(binding.InterfaceType, binding.Implementation))
		}
	} else {
		appConfigV1 := appConfig.(*appv2.Config)
		for _, binding := range appConfigV1.GolangBindings {
			opts = append(opts, depinject.BindInterface(binding.InterfaceType, binding.Implementation))
		}
	}

	return depinject.Configs(opts...)
}

func dumpRegisteredModules(modules map[protoreflect.FullName]*internal.ModuleInitializer) string {
	var mods []string
	for name := range modules {
		mods = append(mods, "  "+string(name))
	}
	return fmt.Sprintf("registered modules are:\n%s", strings.Join(mods, "\n"))
}

func GetProtoV1MsgNameFromAnyTypeUrl(url string) string {
	msgName := strings.Split(url, "/")
	return msgName[1]
}