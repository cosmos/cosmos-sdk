package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers the IBC transfer types
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgTransfer{}, "ibc/transfer/MsgTransfer", nil)
	cdc.RegisterConcrete(FungibleTokenPacketData{}, "ibc/transfer/PacketDataTransfer", nil)
}

var (
	amino = codec.New()

	// ModuleCdc references the global x/ibc/20-transfer module codec. Note, the codec
	// should ONLY be used in certain instances of tests and for JSON encoding as Amino
	// is still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/ibc/20-transfer and
	// defined at the application level.
	ModuleCdc = codec.NewHybridCodec(amino)
)

func init() {
	RegisterCodec(amino)
	codec.RegisterCrypto(amino)
	amino.Seal()
}
