package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterLegacyAminoCodec does nothing. IBC does not support amino, so capability.
func RegisterLegacyAminoCodec(*codec.LegacyAmino) {}

var (
	amino = codec.NewLegacyAmino()
)

func init() {
	RegisterLegacyAminoCodec(amino)
	amino.Seal()
}
