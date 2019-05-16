package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// RegisterCodec registers the account types and interface on the auth codec
func RegisterCodec() {
	auth.RegisterAccountInterface((*PoolAccount)(nil))
	auth.RegisterAccountType(&PoolHolderAccount{}, "auth/PoolHolderAccount")
	auth.RegisterAccountType(&PoolMinterAccount{}, "auth/PoolMinterAccount")
}

// ModuleCdc generic sealed codec to be used throughout module
var ModuleCdc *codec.Codec

func init() {
	cdc := codec.New()
	RegisterCodec()
	codec.RegisterCrypto(cdc)
	ModuleCdc = cdc.Seal()
}
