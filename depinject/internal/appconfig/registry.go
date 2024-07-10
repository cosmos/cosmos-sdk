package internal

import (
	"fmt"
	"reflect"

	gogoproto "github.com/cosmos/gogoproto/proto"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/depinject/appconfig/v1alpha1"
)

// ModuleRegistry is the registry of module initializers indexed by their golang
// type to avoid any issues with protobuf descriptor initialization.
var ModuleRegistry = map[reflect.Type]*ModuleInitializer{}

// ModuleInitializer describes how to initialize a module.
type ModuleInitializer struct {
	ConfigGoType       reflect.Type
	ConfigProtoMessage gogoproto.Message
	Error              error
	Providers          []interface{}
	Invokers           []interface{}
}

// ModulesByModuleTypeName should be used to retrieve modules by their module type name.
// This is done lazily after module registration to deal with non-deterministic issues
// that can occur with respect to protobuf descriptor initialization.
func ModulesByModuleTypeName() (map[string]*ModuleInitializer, error) {
	res := map[string]*ModuleInitializer{}

	for _, initializer := range ModuleRegistry {
		// as of gogoproto v1.5.0 this should work with either gogoproto or golang v2 proto
		fullName := gogoproto.MessageName(initializer.ConfigProtoMessage)

		if desc, err := gogoproto.HybridResolver.FindDescriptorByName(protoreflect.FullName(fullName)); err == nil {
			modDesc, err := GetModuleDescriptor(desc)
			if err != nil {
				return nil, err
			}

			if modDesc == nil {
				return nil, fmt.Errorf(
					"protobuf type %s registered as a module should have the option %s",
					desc.FullName(),
					v1alpha1.E_Module.Name)
			}

			if modDesc.GoImport == "" {
				return nil, fmt.Errorf(
					"protobuf type %s registered as a module should have ModuleDescriptor.go_import specified",
					fullName,
				)
			}
		}

		if _, ok := res[fullName]; ok {
			return nil, fmt.Errorf("duplicate module registration for %s", fullName)
		}

		res[fullName] = initializer
	}

	return res, nil
}

// GetModuleDescriptor returns the cosmos.app.v1alpha1.ModuleDescriptor or nil if one isn't found.
// Errors are returned in unexpected cases.
func GetModuleDescriptor(desc protoreflect.Descriptor) (*v1alpha1.ModuleDescriptor, error) {
	var modDesc protov2.Message
	// we range over the extensions to retrieve the module extension dynamically without needing to import the api module
	protov2.RangeExtensions(desc.Options(), func(extensionType protoreflect.ExtensionType, value any) bool {
		if string(extensionType.TypeDescriptor().FullName()) == v1alpha1.E_Module.Name {
			modDesc = value.(protov2.Message)
			return false
		}
		return true
	})

	if modDesc == nil {
		return nil, nil
	}

	// we convert the returned descriptor to the concrete gogo type we have in v1alpha1 to have a concrete type here
	var modDescGogo v1alpha1.ModuleDescriptor
	bz, err := protov2.Marshal(modDesc)
	if err != nil {
		return nil, err
	}
	err = gogoproto.Unmarshal(bz, &modDescGogo)
	if err != nil {
		return nil, err
	}

	return &modDescGogo, nil
}
