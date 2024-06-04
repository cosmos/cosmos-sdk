package internal

import (
	"fmt"
	"reflect"

	gogoproto "github.com/cosmos/gogoproto/proto"
	protov2 "google.golang.org/protobuf/proto"
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
		if msgv2, ok := initializer.ConfigProtoMessage.(protov2.Message); ok {
			fullName = string(msgv2.ProtoReflect().Descriptor().FullName())
		} else {
			fullName = gogoproto.MessageName(initializer.ConfigProtoMessage)
		}

		if _, ok := res[fullName]; ok {
			return nil, fmt.Errorf("duplicate module registration for %s", fullName)
		}

		res[fullName] = initializer
	}

	return res, nil
}
