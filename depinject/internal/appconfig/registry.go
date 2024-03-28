package internal

import (
	"fmt"
	"reflect"

	gogoproto "github.com/cosmos/gogoproto/proto"
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
		var fullName string
		if initializer.ConfigProtoMessage != nil {
			fullName = gogoproto.MessageName(initializer.ConfigProtoMessage)
			//modDesc := proto.GetExtension(descriptor.Options(), appv1alpha1.E_Module).(*appv1alpha1.ModuleDescriptor)
			//if modDesc == nil {
			//	return nil, fmt.Errorf(
			//		"protobuf type %s registered as a module should have the option %s",
			//		fullName,
			//		appv1alpha1.E_Module.TypeDescriptor().FullName())
			//}
			//
			//if modDesc.GoImport == "" {
			//	return nil, fmt.Errorf(
			//		"protobuf type %s registered as a module should have ModuleDescriptor.go_import specified",
			//		fullName,
			//	)
			//}
		} else {
			fullName = fmt.Sprintf("%s.%s", initializer.ConfigGoType.PkgPath(), initializer.ConfigGoType.Name())
		}

		if _, ok := res[fullName]; ok {
			return nil, fmt.Errorf("duplicate module registration for %s", fullName)
		}

		res[fullName] = initializer
	}

	return res, nil
}
