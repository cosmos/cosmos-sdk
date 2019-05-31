package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// generic sealed codec to be used throughout this module
var ModuleCdc = codec.New()

func init() {
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
