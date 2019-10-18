package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(Packet{}, "ibcmockbank/Packet", nil)
	cdc.RegisterConcrete(TransferPacketData{}, "ibcmockbank/TransferPacketData", nil)
	cdc.RegisterConcrete(MsgTransfer{}, "ibcmockbank/MsgTransfer", nil)
	cdc.RegisterConcrete(MsgSendTransferPacket{}, "ibcmockbank/MsgSendTransferPacket", nil)
}

var MouduleCdc = codec.New()

func init() {
	RegisterCodec(MouduleCdc)
}
