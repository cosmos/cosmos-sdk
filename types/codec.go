package types

import (
	"github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

// Register the sdk message type
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*Msg)(nil), nil)
	cdc.RegisterInterface((*Tx)(nil), nil)
}

// Register the sdk message type
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterInterface("cosmos_sdk.v1.Msg", (*Msg)(nil))
}
