package types

import (
	"cosmossdk.io/core/registry"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
)

// RegisterLegacyAminoCodec registers the vesting interfaces and concrete types on the
// provided LegacyAmino codec. These types are used for Amino JSON serialization
func RegisterLegacyAminoCodec(cdc registry.AminoRegistrar) {
	cdc.RegisterInterface((*exported.VestingAccount)(nil), nil)
	cdc.RegisterConcrete(&BaseVestingAccount{}, "cosmos-sdk/BaseVestingAccount")
	cdc.RegisterConcrete(&ContinuousVestingAccount{}, "cosmos-sdk/ContinuousVestingAccount")
	cdc.RegisterConcrete(&DelayedVestingAccount{}, "cosmos-sdk/DelayedVestingAccount")
	cdc.RegisterConcrete(&PeriodicVestingAccount{}, "cosmos-sdk/PeriodicVestingAccount")
	cdc.RegisterConcrete(&PermanentLockedAccount{}, "cosmos-sdk/PermanentLockedAccount")
}

// RegisterInterfaces associates protoName with AccountI and VestingAccount
// Interfaces and creates a registry of it's concrete implementations
func RegisterInterfaces(registrar registry.InterfaceRegistrar) {
	registrar.RegisterInterface(
		"cosmos.vesting.v1beta1.VestingAccount",
		(*exported.VestingAccount)(nil),
		&ContinuousVestingAccount{},
		&DelayedVestingAccount{},
		&PeriodicVestingAccount{},
		&PermanentLockedAccount{},
	)

	registrar.RegisterImplementations(
		(*sdk.AccountI)(nil),
		&BaseVestingAccount{},
		&DelayedVestingAccount{},
		&ContinuousVestingAccount{},
		&PeriodicVestingAccount{},
		&PermanentLockedAccount{},
	)

	registrar.RegisterImplementations(
		(*authtypes.GenesisAccount)(nil),
		&BaseVestingAccount{},
		&DelayedVestingAccount{},
		&ContinuousVestingAccount{},
		&PeriodicVestingAccount{},
		&PermanentLockedAccount{},
	)
}
