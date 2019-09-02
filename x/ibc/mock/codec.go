package mock

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/mock/types"
)

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(types.PacketSequence{}, "ibcmock/types.PacketSequence", nil)
}
