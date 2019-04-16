package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*TokenHolder)(nil), nil)
	cdc.RegisterConcrete(&BaseTokenHolder{}, "bank/BaseTokenHolder", nil)
}

var msgCdc = codec.New()

func init() {
	RegisterCodec(msgCdc)
}
