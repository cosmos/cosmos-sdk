package crisis

import (
	"github.com/YunSuk-Yeo/cosmos-sdk/codec"
)

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgVerifyInvariant{}, "cosmos-sdk/MsgVerifyInvariant", nil)
}

// generic sealed codec to be used throughout module
var MsgCdc *codec.Codec

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	MsgCdc = cdc.Seal()
}

// SetMsgCodex allows sdk users use custom codex at GetSignBytes
func SetMsgCodec(cdc *codec.Codec) {
	MsgCdc = cdc
}
