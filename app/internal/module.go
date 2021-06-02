package internal

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

type Named interface {
	Name() string
}

type TypeProvider interface {
	RegisterInterfaces(registry types.InterfaceRegistry)
}

type Handler interface {
	RegisterServices(configurator module.Configurator)
}
