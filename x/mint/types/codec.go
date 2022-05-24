package types

import (
	"github.com/Stride-Labs/cosmos-sdk/codec"
	cryptocodec "github.com/Stride-Labs/cosmos-sdk/crypto/codec"
)

var (
	amino = codec.NewLegacyAmino()
)

func init() {
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
