package internal

import (
	"github.com/cosmos/cosmos-sdk/types/module"
)

type Named interface {
	Name() string
}

type Handler interface {
	RegisterServices(configurator module.Configurator)
}
