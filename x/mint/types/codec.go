package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy_global"
)

var (
	amino = codec.New()
)

func init() {
	legacy_global.RegisterCrypto(amino)
	amino.Seal()
}
