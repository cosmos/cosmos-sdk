package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// RegisterCodec registers the account types and interface on the auth codec
// FIXME: panic: TypeInfo already exists for types.PoolAccount
// FIXME: panic: types.PoolHolderAccount conflicts with 2 other(s). Add it to the priority list for auth.Account.
// FIXME: panic: unmarshal to auth.Account failed after 4 bytes
func RegisterCodec() {
	// auth.RegisterAccountInterface((*PoolAccount)(nil))
	auth.RegisterAccountType(&PoolHolderAccount{}, "auth/PoolHolderAccount")
	auth.RegisterAccountType(&PoolMinterAccount{}, "auth/PoolMinterAccount")
}

// ModuleCdc generic sealed codec to be used throughout module
var ModuleCdc *codec.Codec

func init() {
	cdc := codec.New()
	// RegisterCodec()
	codec.RegisterCrypto(cdc)
	ModuleCdc = cdc.Seal()
}
