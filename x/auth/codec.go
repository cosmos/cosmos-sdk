package auth

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on the codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*Account)(nil), nil)
	cdc.RegisterConcrete(&BaseAccount{}, "auth/Account", nil)
	cdc.RegisterInterface((*VestingAccount)(nil), nil)
	cdc.RegisterConcrete(&BaseVestingAccount{}, "auth/BaseVestingAccount", nil)
	cdc.RegisterConcrete(&ContinuousVestingAccount{}, "auth/ContinuousVestingAccount", nil)
	cdc.RegisterConcrete(&DelayedVestingAccount{}, "auth/DelayedVestingAccount", nil)
	cdc.RegisterConcrete(StdTx{}, "auth/StdTx", nil)
}

// RegisterBaseAccount most users shouldn't use this, but this comes in handy for tests.
func RegisterBaseAccount(cdc *codec.Codec) {
	cdc.RegisterInterface((*Account)(nil), nil)
	cdc.RegisterInterface((*VestingAccount)(nil), nil)
	cdc.RegisterConcrete(&BaseAccount{}, "cosmos-sdk/BaseAccount", nil)
	cdc.RegisterConcrete(&BaseVestingAccount{}, "cosmos-sdk/BaseVestingAccount", nil)
	cdc.RegisterConcrete(&ContinuousVestingAccount{}, "cosmos-sdk/ContinuousVestingAccount", nil)
	cdc.RegisterConcrete(&DelayedVestingAccount{}, "cosmos-sdk/DelayedVestingAccount", nil)
	codec.RegisterCrypto(cdc)
}

// RegisterAccountType registers an external account type defined
// in another module for the internal ModuleCdc. This allows the account type
// to be correctly Amino encoded and decoded.
func RegisterAccountType(o interface{}, name string) {
	moduleCdc.RegisterConcrete(o, name, nil)
}

// RegisterAccountInterface registers an external account interface defined
// in another module for the internal ModuleCdc. This allows the account interface
// to be correctly Amino encoded and decoded.
func RegisterAccountInterface(o interface{}) {
	moduleCdc.RegisterInterface(o, nil)
}

var moduleCdc = codec.New()

func init() {
	RegisterCodec(moduleCdc)
	codec.RegisterCrypto(moduleCdc)
}
