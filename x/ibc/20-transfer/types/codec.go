package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgTransfer{}, "ibc/transfer/MsgTransfer", nil)
	cdc.RegisterConcrete(TransferPacketData{}, "ibc/transfer/TransferPacketData", nil)
}

var MouduleCdc = codec.New()

func init() {
	RegisterCodec(MouduleCdc)
	channel.RegisterCodec(MouduleCdc)
	commitment.RegisterCodec(MouduleCdc)
}
