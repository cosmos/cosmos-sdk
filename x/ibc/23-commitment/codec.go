package ics23

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers types declared in this package
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*Root)(nil), nil)
	cdc.RegisterInterface((*Prefix)(nil), nil)
	cdc.RegisterInterface((*Proof)(nil), nil)
}
