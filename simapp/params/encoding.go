package params

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// EncodingConfig specifies the concrete encoding types to use for a given app.
// This is provided for compatibility between protobuf and amino implementations.
type EncodingConfig struct {
	InterfaceRegistry types.InterfaceRegistry
	Marshaler         codec.Marshaler
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}

// RegisterCodecsTests registars codecs from the config to the global and
// modules scope (mbm != nil).
func (ec *EncodingConfig) RegisterCodecsTests(mbm module.BasicManager) {
	std.RegisterLegacyAminoCodec(ec.Amino)
	std.RegisterInterfaces(ec.InterfaceRegistry)
	if mbm != nil {
		mbm.RegisterLegacyAminoCodec(ec.Amino)
		mbm.RegisterInterfaces(ec.InterfaceRegistry)
	}
}
