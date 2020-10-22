package simapp

import (
	simapparams "github.com/cosmos/cosmos-sdk/simapp/params"
)

// MakeEncodingConfig creates an EncodingConfig for testing
func MakeEncodingConfig() simapparams.EncodingConfig {
	encodingConfig := simapparams.MakeEncodingConfig()
	encodingConfig.RegisterCodecsTests(ModuleBasics)
	return encodingConfig
}
