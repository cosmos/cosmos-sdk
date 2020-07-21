package module

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
)

// InterfaceModule is an interface that modules can implement in order to
// register their interfaces and implementations in an InterfaceRegistry
type InterfaceModule interface {
	RegisterInterfaceTypes(registry types.InterfaceRegistry)
}

// RegisterInterfaceModules calls RegisterInterfaceTypes with the registry
// parameter on all of the modules which implement InterfaceModule in the manager
func (bm BasicManager) RegisterInterfaceModules(registry types.InterfaceRegistry) {
	for _, m := range bm {
		im, ok := m.(InterfaceModule)
		if !ok {
			continue
		}

		im.RegisterInterfaceTypes(registry)
	}
}
