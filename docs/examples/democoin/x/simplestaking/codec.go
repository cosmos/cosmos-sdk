package simplestaking

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgBond{}, "simplestaking/BondMsg", nil)
	cdc.RegisterConcrete(MsgUnbond{}, "simplestaking/UnbondMsg", nil)
}
