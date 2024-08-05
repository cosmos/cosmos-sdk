package types

import (
	corelegacy "cosmossdk.io/core/legacy"
	"cosmossdk.io/core/registry"
	coretransaction "cosmossdk.io/core/transaction"
	authtypes "cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/auth/vesting/exported"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the vesting interfaces and concrete types on the
// provided LegacyAmino codec. These types are used for Amino JSON serialization
func RegisterLegacyAminoCodec(cdc corelegacy.Amino) {
	cdc.RegisterInterface((*exported.VestingAccount)(nil), nil)
	cdc.RegisterConcrete(&BaseVestingAccount{}, "cosmos-sdk/BaseVestingAccount")
	cdc.RegisterConcrete(&ContinuousVestingAccount{}, "cosmos-sdk/ContinuousVestingAccount")
	cdc.RegisterConcrete(&DelayedVestingAccount{}, "cosmos-sdk/DelayedVestingAccount")
	cdc.RegisterConcrete(&PeriodicVestingAccount{}, "cosmos-sdk/PeriodicVestingAccount")
	cdc.RegisterConcrete(&PermanentLockedAccount{}, "cosmos-sdk/PermanentLockedAccount")
	legacy.RegisterAminoMsg(cdc, &MsgCreateVestingAccount{}, "cosmos-sdk/MsgCreateVestingAccount")
	legacy.RegisterAminoMsg(cdc, &MsgCreatePermanentLockedAccount{}, "cosmos-sdk/MsgCreatePermLockedAccount")
	legacy.RegisterAminoMsg(cdc, &MsgCreatePeriodicVestingAccount{}, "cosmos-sdk/MsgCreatePeriodVestAccount")
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

	registrar.RegisterImplementations(
		(*coretransaction.Msg)(nil),
		&MsgCreateVestingAccount{},
		&MsgCreatePermanentLockedAccount{},
	)

	msgservice.RegisterMsgServiceDesc(registrar, &_Msg_serviceDesc)
}
