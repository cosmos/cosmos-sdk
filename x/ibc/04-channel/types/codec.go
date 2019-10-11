package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
)

var SubModuleCdc *codec.Codec

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.PacketI)(nil), nil)

	cdc.RegisterConcrete(MsgChanOpenInit{}, "ibc/channel/MsgChanOpenInit", nil)
	cdc.RegisterConcrete(MsgChanOpenTry{}, "ibc/channel/MsgChanOpenTry", nil)
	cdc.RegisterConcrete(MsgChanOpenAck{}, "ibc/channel/MsgChanOpenAck", nil)
	cdc.RegisterConcrete(MsgChanOpenConfirm{}, "ibc/channel/MsgChanOpenConfirm", nil)
}

func SetMsgChanCodec(cdc *codec.Codec) {
	SubModuleCdc = cdc
}
