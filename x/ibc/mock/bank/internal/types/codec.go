package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(TransferPacketData{}, "ibcmockbank/TransferPacketData", nil)
	cdc.RegisterConcrete(MsgTransfer{}, "ibcmockbank/MsgTransfer", nil)
}

var MouduleCdc = codec.New()

func init() {
	RegisterCodec(MouduleCdc)
}
