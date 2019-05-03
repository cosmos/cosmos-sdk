package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var cdc = codec.New()

// RegisterCodec concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*NFT)(nil), nil)
	cdc.RegisterConcrete(&BaseNFT{}, "cosmos-sdk/BaseNFT", nil)
	cdc.RegisterConcrete(MsgTransferNFT{}, "cosmos-sdk/MsgTransferNFT", nil)
	cdc.RegisterConcrete(MsgEditNFTMetadata{}, "cosmos-sdk/MsgEditNFTMetadata", nil)
}

func init() {
	RegisterCodec(cdc)
}
