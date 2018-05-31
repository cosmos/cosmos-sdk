package slashing

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

// Register concrete types on wire codec
func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(MsgUnrevoke{}, "cosmos-sdk/MsgUnrevoke", nil)
}

var cdcEmpty = wire.NewCodec()
