package pow

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Amino) {
	cdc.RegisterConcrete(MsgMine{}, "pow/Mine", nil)
}
