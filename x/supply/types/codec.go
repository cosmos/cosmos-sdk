package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers the account types and interface on the auth codec
// FIXME: panic: TypeInfo already exists for types.PoolAccount
// FIXME: panic: types.PoolHolderAccount conflicts with 2 other(s). Add it to the priority list for auth.Account.
// FIXME: panic: unmarshal to auth.Account failed after 4 bytes
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*PoolAccount)(nil), nil)
	cdc.RegisterConcrete(&PoolHolderAccount{}, "auth/PoolHolderAccount", nil)
	cdc.RegisterConcrete(&PoolMinterAccount{}, "auth/PoolMinterAccount", nil)
}

// ModuleCdc generic sealed codec to be used throughout module
var ModuleCdc *codec.Codec

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	ModuleCdc = cdc.Seal()
}
