package codec

import (
	"github.com/pointnetwork/cosmos-point-sdk/codec"
	cryptocodec "github.com/pointnetwork/cosmos-point-sdk/crypto/codec"
	sdk "github.com/pointnetwork/cosmos-point-sdk/types"
)

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(Amino)
)

func init() {
	cryptocodec.RegisterCrypto(Amino)
	codec.RegisterEvidences(Amino)
	sdk.RegisterLegacyAminoCodec(Amino)
}
