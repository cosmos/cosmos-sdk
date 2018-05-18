package merkle

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

// RegisterCodec registers interfaces to the codec
func RegisterCodec(cdc *wire.Codec) {
	cdc.RegisterInterface((*KeyProof)(nil), nil)
	cdc.RegisterConcrete(ExistsProof{}, "cosmos-sdk/ExistsProof", nil)

	cdc.RegisterInterface((*Wrapper)(nil), nil)
}
