package ibc

import (
	"github.com/YunSuk-Yeo/cosmos-sdk/codec"
)

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgIBCTransfer{}, "cosmos-sdk/MsgIBCTransfer", nil)
	cdc.RegisterConcrete(MsgIBCReceive{}, "cosmos-sdk/MsgIBCReceive", nil)
}
