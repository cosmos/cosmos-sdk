package slashing

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgUnjail{}, "cosmos-sdk/MsgUnjail", nil)
}

var msgCdc *codec.Codec

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	msgCdc = cdc.Seal()
}

// SetMsgCodex allows sdk users use custom codex at GetSignBytes
func SetMsgCodec(cdc *codec.Codec) {
	msgCdc = cdc
}
