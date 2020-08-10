package types

import (
	"github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

// Register the sdk message type
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*Msg)(nil), nil)
	cdc.RegisterInterface((*Tx)(nil), nil)
}

// Register the sdk message type
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterInterface("cosmos.Msg", (*Msg)(nil))
}
