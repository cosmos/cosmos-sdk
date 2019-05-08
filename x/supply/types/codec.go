package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on the codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*PoolAccount)(nil), nil)
	cdc.RegisterConcrete(&PoolHolderAccount{}, "auth/PoolHolderAccount", nil)
	cdc.RegisterConcrete(&PoolMinterAccount{}, "auth/PoolMinterAccount", nil)
}
