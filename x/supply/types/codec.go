package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// TODO: register in auth codec

// RegisterCodec registers concrete types on the codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*PoolAccount)(nil), nil)
	cdc.RegisterConcrete(&PoolHolderAccount{}, "auth/PoolHolderAccount", nil)
	cdc.RegisterConcrete(&PoolMinterAccount{}, "auth/PoolMinterAccount", nil)
}

// generic sealed codec to be used throughout module
var ModuleCdc *codec.Codec

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	ModuleCdc = cdc.Seal()
}
