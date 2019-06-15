package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on the codec.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSwapOrder{}, "cosmos-sdk/MsgSwapOrder", nil)
	cdc.RegisterConcrete(MsgAddLiquidity{}, "cosmos-sdk/MsgAddLiquidity", nil)
	cdc.RegisterConcrete(MsgRemoveLiquidity{}, "cosmos-sdk/MsgRemoveLiquidity", nil)
}

// module codec
var ModuleCdc = codec.New()

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	ModuleCdc.Seal()
}
