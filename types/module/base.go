package module

import "github.com/cosmos/cosmos-sdk/codec/types"

// Module is the base module type that all modules (client and server) must satisfy.
type Module interface {
	Name() string
}

// TypeModule is an interface that modules should implement to register types.
type TypeModule interface {
	Module

	RegisterInterfaces(types.InterfaceRegistry)
}
