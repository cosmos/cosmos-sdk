package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
)

// ModuleCdc auth module wide codec
var ModuleCdc = codec.New()

// RegisterCodec registers concrete types on the codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.GenesisAccount)(nil), nil)
	cdc.RegisterInterface((*exported.Account)(nil), nil)
	cdc.RegisterConcrete(&BaseAccount{}, "cosmos-sdk/Account", nil)
	cdc.RegisterInterface((*exported.VestingAccount)(nil), nil)
	cdc.RegisterConcrete(&BaseVestingAccount{}, "cosmos-sdk/BaseVestingAccount", nil)
	cdc.RegisterConcrete(&ContinuousVestingAccount{}, "cosmos-sdk/ContinuousVestingAccount", nil)
	cdc.RegisterConcrete(&DelayedVestingAccount{}, "cosmos-sdk/DelayedVestingAccount", nil)
	cdc.RegisterConcrete(StdTx{}, "cosmos-sdk/StdTx", nil)
}

// RegisterAccountTypeCodec registers an external account type defined in
// another module for the internal ModuleCdc.
func RegisterAccountTypeCodec(o interface{}, name string) {
	ModuleCdc.RegisterConcrete(o, name, nil)
}

var validAccountTypes = map[string]struct{}{
	"BaseAccount":              {},
	"BaseVestingAccount":       {},
	"ContinuousVestingAccount": {},
	"DelayedVestingAccount":    {},
}

// RegisterAccountType registers a account type. It will panic if the type is
// already registered.
func RegisterAccountType(ty string) {
	if _, ok := validAccountTypes[ty]; ok {
		panic(fmt.Sprintf("already registered proposal type: %s", ty))
	}

	validAccountTypes[ty] = struct{}{}
}

func init() {
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
}
