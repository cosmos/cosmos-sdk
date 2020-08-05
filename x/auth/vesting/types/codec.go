package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
)

// RegisterCodec registers the vesting interfaces and concrete types on the
// provided Amino codec.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.VestingAccount)(nil), nil)
	cdc.RegisterConcrete(&BaseVestingAccount{}, "cosmos-sdk/BaseVestingAccount", nil)
	cdc.RegisterConcrete(&ContinuousVestingAccount{}, "cosmos-sdk/ContinuousVestingAccount", nil)
	cdc.RegisterConcrete(&DelayedVestingAccount{}, "cosmos-sdk/DelayedVestingAccount", nil)
	cdc.RegisterConcrete(&PeriodicVestingAccount{}, "cosmos-sdk/PeriodicVestingAccount", nil)
}

// RegisterInterface associates protoName with AccountI and VestingAccount
// Interfaces and creates a registry of it's concrete implementations
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterInterface(
		"cosmos.vesting.v1beta1.VestingAccount",
		(*exported.VestingAccount)(nil),
		&ContinuousVestingAccount{},
		&DelayedVestingAccount{},
		&PeriodicVestingAccount{},
	)
	registry.RegisterImplementations(
		(*authtypes.AccountI)(nil),
		&DelayedVestingAccount{},
		&ContinuousVestingAccount{},
		&PeriodicVestingAccount{},
	)
}

var amino = codec.New()

func init() {
	RegisterCodec(amino)
	amino.Seal()
}
