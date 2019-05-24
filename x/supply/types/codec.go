package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers the account types and interface
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*ModuleAccount)(nil), nil)
	cdc.RegisterConcrete(&ModuleHolderAccount{}, "cosmos-sdk/ModuleHolderAccount", nil)
	cdc.RegisterConcrete(&ModuleMinterAccount{}, "cosmos-sdk/ModuleMinterAccount", nil)
}

// ModuleCdc generic sealed codec to be used throughout module
var ModuleCdc *codec.Codec

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	ModuleCdc = cdc.Seal()
}
