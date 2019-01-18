package auth

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// Register concrete types on codec codec for default AppAccount
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*Account)(nil), nil)
	cdc.RegisterConcrete(&BaseAccount{}, "auth/Account", nil)
	cdc.RegisterInterface((*VestingAccount)(nil), nil)
	cdc.RegisterConcrete(&BaseVestingAccount{}, "auth/BaseVestingAccount", nil)
	cdc.RegisterConcrete(&ContinuousVestingAccount{}, "auth/ContinuousVestingAccount", nil)
	cdc.RegisterConcrete(&DelayedVestingAccount{}, "auth/DelayedVestingAccount", nil)
	cdc.RegisterConcrete(StdTx{}, "auth/StdTx", nil)
}

var msgCdc = codec.New()

func init() {
	RegisterCodec(msgCdc)
	codec.RegisterCrypto(msgCdc)
}
