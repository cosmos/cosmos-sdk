package connection

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var MsgCdc *codec.Codec

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgOpenInit{}, "ibc/client/MsgOpenInit", nil)
	cdc.RegisterConcrete(MsgOpenTry{}, "ibc/client/MsgOpenTry", nil)
	cdc.RegisterConcrete(MsgOpenAck{}, "ibc/client/MsgOpenAck", nil)
	cdc.RegisterConcrete(MsgOpenConfirm{}, "ibc/client/MsgOpenConfirm", nil)
}

func SetMsgCodec(cdc *codec.Codec) {
	// TODO
	/*
		if MsgCdc != nil && MsgCdc != cdc {
			panic("MsgCdc set more than once")
		}
	*/
	MsgCdc = cdc
}
