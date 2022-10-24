package params

import (
	"github.com/pointnetwork/cosmos-point-sdk/client"
	"github.com/pointnetwork/cosmos-point-sdk/codec"
	"github.com/pointnetwork/cosmos-point-sdk/codec/types"
)

// EncodingConfig specifies the concrete encoding types to use for a given app.
// This is provided for compatibility between protobuf and amino implementations.
type EncodingConfig struct {
	InterfaceRegistry types.InterfaceRegistry
	Codec             codec.Codec
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}
