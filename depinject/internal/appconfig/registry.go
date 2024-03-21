package internal

import (
	"fmt"
	"reflect"

	"github.com/cosmos/gogoproto/proto"
	"github.com/jhump/protoreflect/desc"

	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
)

// ModuleRegistry is the registry of module initializers indexed by their golang
// type to avoid any issues with protobuf descriptor initialization.
var ModuleRegistry = map[reflect.Type]*ModuleInitializer{}

// ModuleInitializer describes how to initialize a module.
type ModuleInitializer struct {
	ConfigGoType       reflect.Type
	ConfigProtoMessage proto.Message
	Error              error
	Providers          []interface{}
	Invokers           []interface{}
}

// ModulesByProtoMessageName should be used to retrieve modules by their protobuf name.
// This is done lazily after module registration to deal with non-deterministic issues
// that can occur with respect to protobuf descriptor initialization.
func ModulesByProtoMessageName() (map[string]*ModuleInitializer, error) {
	res := map[string]*ModuleInitializer{}

	for _, initializer := range ModuleRegistry {
		descriptor, err := desc.LoadMessageDescriptorForMessage(initializer.ConfigProtoMessage)
		if err != nil {
			return nil, fmt.Errorf("error loading descriptor for %s: %w", initializer.ConfigProtoMessage, err)
		}

		fullName := descriptor.GetName()
		if _, ok := res[fullName]; ok {
			return nil, fmt.Errorf("duplicate module registration for %s", fullName)
		}

		modDescI, err := proto.GetExtension(descriptor.GetOptions(), appv1alpha1.E_Module)
		if err != nil {
			return nil, fmt.Errorf("error getting module descriptor for %s: %w", fullName, err)
		}

		modDesc, ok := modDescI.(*appv1alpha1.ModuleDescriptor)
		if !ok || modDesc == nil {
			return nil, fmt.Errorf("protobuf type %s registered as a module should have the option cosmos.app.v1alpha1.module", fullName)
		}

		if modDesc.GoImport == "" {
			return nil, fmt.Errorf(
				"protobuf type %s registered as a module should have ModuleDescriptor.go_import specified",
				fullName,
			)
		}

		res[fullName] = initializer
	}

	return res, nil
}
