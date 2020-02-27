package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
)

var SubModuleCdc *codec.Codec

// RegisterCodec registers types declared in this package
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.Root)(nil), nil)
	cdc.RegisterInterface((*exported.Prefix)(nil), nil)
	cdc.RegisterInterface((*exported.Path)(nil), nil)
	cdc.RegisterInterface((*exported.Proof)(nil), nil)

	cdc.RegisterConcrete(MerkleRoot{}, "ibc/commitment/MerkleRoot", nil)
	cdc.RegisterConcrete(MerklePrefix{}, "ibc/commitment/MerklePrefix", nil)
	cdc.RegisterConcrete(MerklePath{}, "ibc/commitment/MerklePath", nil)
	cdc.RegisterConcrete(MerkleProof{}, "ibc/commitment/MerkleProof", nil)

	SetSubModuleCodec(cdc)
}

func SetSubModuleCodec(cdc *codec.Codec) {
	SubModuleCdc = cdc
}
