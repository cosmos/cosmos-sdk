package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var (
	amino = codec.New()
)

func init() {
	codec.RegisterCrypto(amino)
	amino.Seal()
}
