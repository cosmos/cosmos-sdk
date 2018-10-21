package slashing

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Amino) {
	cdc.RegisterConcrete(MsgUnjail{}, "cosmos-sdk/MsgUnjail", nil)
}

var cdcEmpty = codec.New()
