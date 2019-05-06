package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on the codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*ModuleAccount)(nil), nil)
	cdc.RegisterConcrete(&ModuleHolderAccount{}, "auth/ModuleHolderAccount", nil)
	cdc.RegisterConcrete(&ModuleMinterAccount{}, "auth/ModuleMinterAccount", nil)
}