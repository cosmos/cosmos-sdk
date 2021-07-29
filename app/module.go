package app

import "github.com/cosmos/cosmos-sdk/codec/types"

type Module interface {
	RegisterTypes(types.InterfaceRegistry)
}
