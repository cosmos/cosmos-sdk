package params

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var msgCdc = codec.New()

func init() {
	RegisterCodec(msgCdc)
}

// Registers types to codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(ParamChangeProposal{}, "params/ParamChangeProposal", nil)
}
