package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(Packet{}, "ibcmockbank/Packet", nil)
	cdc.RegisterConcrete(TransferPacketData{}, "ibcmockbank/TransferPacketData", nil)
	cdc.RegisterConcrete(MsgTransfer{}, "ibcmockbank/MsgTransfer", nil)
	cdc.RegisterConcrete(MsgRecvTransferPacket{}, "ibcmockbank/MsgRecvTransferPacket", nil)
}

var MouduleCdc = codec.New()

func init() {
	RegisterCodec(MouduleCdc)
	channel.RegisterCodec(MouduleCdc)
	commitment.RegisterCodec(MouduleCdc)
	merkle.RegisterCodec(MouduleCdc)
}
