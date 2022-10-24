package types

import (
	"github.com/pointnetwork/cosmos-point-sdk/codec"
	cryptocodec "github.com/pointnetwork/cosmos-point-sdk/crypto/codec"
)

var amino = codec.NewLegacyAmino()

func init() {
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
