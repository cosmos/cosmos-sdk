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

// generic sealed codec to be used throughout this module
var moduleCdc *codec.Codec

func init() {
	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	moduleCdc = cdc.Seal()
}
