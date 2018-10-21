package types

import "github.com/cosmos/cosmos-sdk/codec"

// reexport
type Codec = codec.Codec

// Register the sdk message type
func RegisterCodec(cdc *codec.Amino) {
	cdc.RegisterInterface((*Msg)(nil), nil)
	cdc.RegisterInterface((*Tx)(nil), nil)
}
