package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"
)

// reexport
type Codec = codec.Codec

// Register the sdk message type
func RegisterCodec(cdc *codec.Amino) {
	cdc.RegisterInterface((*Msg)(nil), nil)
	cdc.RegisterInterface((*Tx)(nil), nil)
}

// Register the go-crypto to the codec
func RegisterCrypto(cdc *codec.Amino) {
	cryptoAmino.RegisterAmino(cdc.Codec)
}
