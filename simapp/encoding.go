package simapp

import (
	simapparams "github.com/cosmos/cosmos-sdk/simapp/params"
)

// MakeEncodingConfigTests creates an EncodingConfig for testing
func MakeEncodingConfigTests() simapparams.EncodingConfig {
	encodingConfig := simapparams.MakeEncodingConfig()
	encodingConfig.RegisterCodecsTests(ModuleBasics)
	return encodingConfig
}
