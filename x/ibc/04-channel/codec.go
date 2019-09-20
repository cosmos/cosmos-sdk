package channel

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var msgCdc *codec.Codec

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*Packet)(nil), nil)

	cdc.RegisterConcrete(MsgOpenInit{}, "ibc/channel/MsgOpenInit", nil)
	cdc.RegisterConcrete(MsgOpenTry{}, "ibc/channel/MsgOpenTry", nil)
	cdc.RegisterConcrete(MsgOpenAck{}, "ibc/channel/MsgOpenAck", nil)
	cdc.RegisterConcrete(MsgOpenConfirm{}, "ibc/channel/MsgOpenConfirm", nil)
}

func SetMsgCodec(cdc *codec.Codec) {
	// TODO
	/*
		if msgCdc != nil && msgCdc != cdc {
			panic("MsgCdc set more than once")
		}
	*/
	msgCdc = cdc
}
