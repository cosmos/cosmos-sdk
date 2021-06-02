package app

import "github.com/cosmos/cosmos-sdk/codec/types"

type TypeProvider interface {
	RegisterTypes(registry types.InterfaceRegistry)
}
