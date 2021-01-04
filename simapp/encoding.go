package simapp

import (
	"github.com/cosmos/cosmos-sdk/codec"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/std"
)

// MakeTestEncodingConfig creates an EncodingConfig for testing.
// This function should be used only internally (in the SDK).
// App user shouldn't create new codecs - use the app.AppCodec instead.
// [DEPRECATED]
func MakeTestEncodingConfig() simappparams.EncodingConfig {
	encodingConfig := simappparams.MakeTestEncodingConfig()
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}

// MakeTestCodecs is a helper function which uses MakeTestEncodingConfig to
// construct new *std.Codec and *codec.LegacyAmino instances.
// It must not be used in non test files, instead app.AppCodec should be used.
// [DEPRECATED]
func MakeTestCodecs() (codec.Marshaler, *codec.LegacyAmino) {
	config := MakeTestEncodingConfig()
	return config.Marshaler, config.Amino
}
