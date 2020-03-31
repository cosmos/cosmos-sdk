package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

// ModuleCdc defines the IBC transfer codec.
var ModuleCdc = codec.New()

// PacketDataI is our abstract packet data type for this module.
type PacketDataI interface {
	GetBytes() []byte
}

// RegisterCodec registers the IBC transfer types
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*PacketDataI)(nil), nil)
	cdc.RegisterConcrete(MsgTransfer{}, "ibc/transfer/MsgTransfer", nil)
	cdc.RegisterConcrete(FungibleTokenPacketData{}, "ibc/transfer/PacketDataTransfer", nil)
}

func init() {
	RegisterCodec(ModuleCdc)
	channel.RegisterCodec(ModuleCdc)
	commitmenttypes.RegisterCodec(ModuleCdc)
}
