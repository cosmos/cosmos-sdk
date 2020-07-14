package simapp

import (
	"github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/std"
)

// MakeEncodingConfig creates an EncodingConfig for an amino based test configuration.
//
// TODO: this file should add a "+build test_amino" flag for #6190 and a proto.go file with a protobuf configuration
func MakeEncodingConfig() params.EncodingConfig {
	encodingConfig := params.MakeEncodingConfig()
	std.RegisterCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ModuleBasics.RegisterCodec(encodingConfig.Amino)
	ModuleBasics.RegisterInterfaceModules(encodingConfig.InterfaceRegistry)
	return encodingConfig
}
