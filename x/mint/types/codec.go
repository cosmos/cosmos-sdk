package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codec2 "github.com/cosmos/cosmos-sdk/crypto/codec"
)

var (
	amino = codec.New()
)

func init() {
	codec2.RegisterCrypto(amino)
	amino.Seal()
}
