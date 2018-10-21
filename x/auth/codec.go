package auth

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Register concrete types on codec codec for default AppAccount
func RegisterCodec(cdc sdk.Codec) {
	cdc.RegisterInterface((*Account)(nil), nil)
	cdc.RegisterConcrete(&BaseAccount{}, "auth/Account", nil)
	cdc.RegisterConcrete(StdTx{}, "auth/StdTx", nil)
}

var msgCdc = codec.New()

func init() {
	RegisterCodec(msgCdc)
	codec.RegisterCrypto(msgCdc)
}
