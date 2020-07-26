package types

import (
	"github.com/KiraCore/cosmos-sdk/codec"
	cryptocodec "github.com/KiraCore/cosmos-sdk/crypto/codec"
)

var (
	amino = codec.New()
)

func init() {
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
