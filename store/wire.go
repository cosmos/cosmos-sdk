package store

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

var cdc = wire.NewCodec()

// Register concrete types on wire codec
func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(RootMultistoreWrapper{}, "cosmos-sdk/RootMultistoreWrapper", nil)
}
