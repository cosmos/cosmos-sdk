package types

import "github.com/cosmos/cosmos-sdk/codec"

// Register the sdk message type
func RegisterCodec(cdc *codec.Amino) {
	cdc.RegisterInterface((*Msg)(nil), nil)
	cdc.RegisterInterface((*Tx)(nil), nil)
}
