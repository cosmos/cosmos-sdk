package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSwapOrder{}, "cosmos-sdk/MsgSwapOrder", nil)
	cdc.RegisterConcrete(MsgAddLiquidity{}, "cosmos-sdk/MsgAddLiquidity", nil)
	cdc.RegisterConcrete(MsgRemoveLiquidity{}, "cosmos-sdk/MsgRemoveLiquidity", nil)
}

var moduleCdc = codec.New()

func init() {
	RegisterCodec(moduleCdc)
}
