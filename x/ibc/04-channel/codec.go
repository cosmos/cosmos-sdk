package channel

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var MsgCdc = codec.New()

func init() {
	RegisterCodec(MsgCdc)
}

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*Packet)(nil), nil)

	cdc.RegisterConcrete(MsgOpenInit{}, "cosmos-sdk/ibc/channel/MsgOpenInit", nil)
	cdc.RegisterConcrete(MsgOpenTry{}, "cosmos-sdk/ibc/channel/MsgOpenTry", nil)
	cdc.RegisterConcrete(MsgOpenAck{}, "cosmos-sdk/ibc/channel/MsgOpenAck", nil)
	cdc.RegisterConcrete(MsgOpenConfirm{}, "cosmos-sdk/ibc/channel/MsgOpenConfirm", nil)
	cdc.RegisterConcrete(MsgReceive{}, "cosmos-sdk/ibc/channel/MsgReceive", nil)
}
