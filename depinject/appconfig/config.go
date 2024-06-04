package appconfig

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cosmos/cosmos-proto/anyutil"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/protobuf/encoding/protojson"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/anypb"
	"sigs.k8s.io/yaml"

	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"

	"cosmossdk.io/depinject"
	internal "cosmossdk.io/depinject/internal/appconfig"
)

// LoadJSON loads an app config in JSON format.
func LoadJSON(bz []byte) depinject.Config {
	config := &appv1alpha1.Config{}
	err := protojson.UnmarshalOptions{
		Resolver: dynamicResolver{resolver: gogoproto.HybridResolver},
	}.Unmarshal(bz, config)
	if err != nil {
		return depinject.Error(err)
	}

	return Compose(config)
}

// dynamic resolver allows marshaling gogo proto messages from the gogoproto.HybridResolver as long as those
// files have been imported before calling LoadJSON.
type dynamicResolver struct {
	resolver protodesc.Resolver
}

func (r dynamicResolver) FindExtensionByName(field protoreflect.FullName) (protoreflect.ExtensionType, error) {
	ext, err := protoregistry.GlobalTypes.FindExtensionByName(field)
	if err == nil {
		return ext, nil
	}

	desc, err := r.resolver.FindDescriptorByName(field)
	if err != nil {
		return nil, err
	}

	return dynamicpb.NewExtensionType(desc.(protoreflect.ExtensionTypeDescriptor)), nil
}

func (r dynamicResolver) FindExtensionByNumber(message protoreflect.FullName, field protoreflect.FieldNumber) (protoreflect.ExtensionType, error) {
	return protoregistry.GlobalTypes.FindExtensionByNumber(message, field)
}

func (r dynamicResolver) FindMessageByName(message protoreflect.FullName) (protoreflect.MessageType, error) {
	typ, err := protoregistry.GlobalTypes.FindMessageByName(message)
	if err == nil {
		return typ, nil
	}

	desc, err := r.resolver.FindDescriptorByName(message)
	if err != nil {
		return nil, err
	}

	return dynamicpb.NewMessageType(desc.(protoreflect.MessageDescriptor)), nil
}

func (r dynamicResolver) FindMessageByURL(url string) (protoreflect.MessageType, error) {
	if i := strings.LastIndexByte(url, '/'); i >= 0 {
		url = url[i+1:]
	}

	return r.FindMessageByName(protoreflect.FullName(url))
}

var _ protoregistry.MessageTypeResolver = dynamicResolver{}
var _ protoregistry.ExtensionTypeResolver = dynamicResolver{}

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

// Compose composes a v1alpha1 app config into a container option by resolving
// the required modules and composing their options.
func Compose(appConfig *appv1alpha1.Config) depinject.Config {
	opts := []depinject.Config{
		depinject.Supply(appConfig),
	}

	modules, err := internal.ModulesByModuleTypeName()
	if err != nil {
		return depinject.Error(err)
	}

	for _, module := range appConfig.Modules {
		if module.Name == "" {
			return depinject.Error(fmt.Errorf("module is missing name"))
		}

		if module.Config == nil {
			return depinject.Error(fmt.Errorf("module %q is missing a config object", module.Name))
		}

		msgName := module.Config.TypeUrl
		// strip type URL prefix
		if slashIdx := strings.LastIndex(msgName, "/"); slashIdx >= 0 {
			msgName = msgName[slashIdx+1:]
		}
		if msgName == "" {
			return depinject.Error(fmt.Errorf("module %q is missing a type URL", module.Name))
		}

		init, ok := modules[msgName]
		if !ok {
			if msgDesc, err := gogoproto.HybridResolver.FindDescriptorByName(protoreflect.FullName(msgName)); err == nil {
				modDesc := protov2.GetExtension(msgDesc.Options(), appv1alpha1.E_Module).(*appv1alpha1.ModuleDescriptor)
				if modDesc == nil {
					return depinject.Error(fmt.Errorf("no module registered for type URL %s and that protobuf type does not have the option %s\n\n%s",
						module.Config.TypeUrl, appv1alpha1.E_Module.TypeDescriptor().FullName(), dumpRegisteredModules(modules)))
				}

				return depinject.Error(fmt.Errorf("no module registered for type URL %s, did you forget to import %s: find more information on how to make a module ready for app wiring: https://docs.cosmos.network/main/building-modules/depinject\n\n%s",
					module.Config.TypeUrl, modDesc.GoImport, dumpRegisteredModules(modules)))
			}

		}

		var config any
		if configInit, ok := init.ConfigProtoMessage.(protov2.Message); ok {
			configProto := configInit.ProtoReflect().Type().New().Interface()
			err = anypb.UnmarshalTo(module.Config, configProto, protov2.UnmarshalOptions{})
			if err != nil {
				return depinject.Error(err)
			}
			config = configProto
		} else {
			configProto := reflect.New(init.ConfigGoType.Elem()).Interface().(gogoproto.Message)
			err = gogoproto.Unmarshal(module.Config.Value, configProto)
			if err != nil {
				return depinject.Error(err)
			}
			config = configProto
		}

		opts = append(opts, depinject.Supply(config))

		for _, provider := range init.Providers {
			opts = append(opts, depinject.ProvideInModule(module.Name, provider))
		}

		for _, invoker := range init.Invokers {
			opts = append(opts, depinject.InvokeInModule(module.Name, invoker))
		}

		for _, binding := range module.GolangBindings {
			opts = append(opts, depinject.BindInterfaceInModule(module.Name, binding.InterfaceType, binding.Implementation))
		}
	}

	for _, binding := range appConfig.GolangBindings {
		opts = append(opts, depinject.BindInterface(binding.InterfaceType, binding.Implementation))
	}

	return depinject.Configs(opts...)
}

func dumpRegisteredModules(modules map[string]*internal.ModuleInitializer) string {
	var mods []string
	for name := range modules {
		mods = append(mods, "  "+name)
	}
	return fmt.Sprintf("registered modules are:\n%s", strings.Join(mods, "\n"))
}
