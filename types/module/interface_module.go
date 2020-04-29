package module

import "github.com/cosmos/cosmos-sdk/codec/types"

type InterfaceModule interface {
	RegisterInterfaceTypes(registry types.InterfaceRegistry)
}

func RegisterInterfaceModules(manager BasicManager, registry types.InterfaceRegistry) {
	for _, m := range manager {
		im, ok := m.(InterfaceModule)
		if !ok {
			continue
		}
		im.RegisterInterfaceTypes(registry)
	}
}
