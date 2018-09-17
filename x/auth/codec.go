package auth

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var msgCdc = codec.New()

// Register concrete types on codec codec for default AppAccount
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*Account)(nil), nil)
	cdc.RegisterConcrete(&BaseAccount{}, "auth/Account", nil)
	cdc.RegisterConcrete(StdTx{}, "auth/StdTx", nil)
	cdc.RegisterInterface((*VestingAccount)(nil), nil)
	cdc.RegisterConcrete(&ContinuousVestingAccount{}, "auth/ContinuousVestingAccount", nil)
	cdc.RegisterConcrete(&DelayTransferAccount{}, "auth/DelayTransferAccount", nil)
}

func init() {
	RegisterCodec(msgCdc)
	codec.RegisterCrypto(msgCdc)
}
