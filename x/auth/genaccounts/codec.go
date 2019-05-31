package genaccounts

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// generic sealed codec to be used throughout this module
var moduleCdc = codec.New()

func init() {
	codec.RegisterCrypto(moduleCdc)
	moduleCdc.Seal()
}
