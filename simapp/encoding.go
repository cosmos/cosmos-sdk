package simapp

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/std"
)

func MakeCodecs() (codec.Marshaler, codectypes.InterfaceRegistry, *codec.Codec) {
	cfg := MakeEncodingConfig()
	return cfg.Marshaler, cfg.InterfaceRegistry, cfg.Amino
}

func MakeEncodingConfig() params.EncodingConfig {
	encodingConfig := params.MakeEncodingConfig()
	std.RegisterCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ModuleBasics.RegisterCodec(encodingConfig.Amino)
	ModuleBasics.RegisterInterfaceModules(encodingConfig.InterfaceRegistry)
	return encodingConfig
}
