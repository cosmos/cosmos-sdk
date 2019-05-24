package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers the account types and interface
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*PoolAccount)(nil), nil)
	cdc.RegisterConcrete(&PoolHolderAccount{}, "cosmos-sdk/PoolHolderAccount", nil)
	cdc.RegisterConcrete(&PoolMinterAccount{}, "cosmos-sdk/PoolMinterAccount", nil)
}

// ModuleCdc generic sealed codec to be used throughout module
var ModuleCdc *codec.Codec

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	ModuleCdc = cdc.Seal()
}
