package types

import (
	corelegacy "cosmossdk.io/core/legacy"
	"cosmossdk.io/core/registry"
	coretransaction "cosmossdk.io/core/transaction"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

// RegisterLegacyAminoCodec registers the account interfaces and concrete types on the
// provided LegacyAmino codec. These types are used for Amino JSON serialization
func RegisterLegacyAminoCodec(cdc corelegacy.Amino) {
	cdc.RegisterInterface((*sdk.ModuleAccountI)(nil), nil)
	cdc.RegisterInterface((*GenesisAccount)(nil), nil)
	cdc.RegisterInterface((*sdk.AccountI)(nil), nil)
	cdc.RegisterConcrete(&BaseAccount{}, "cosmos-sdk/BaseAccount")
	cdc.RegisterConcrete(&ModuleAccount{}, "cosmos-sdk/ModuleAccount")
	cdc.RegisterConcrete(Params{}, "cosmos-sdk/x/auth/Params")
	cdc.RegisterConcrete(&ModuleCredential{}, "cosmos-sdk/GroupAccountCredential")

	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "cosmos-sdk/x/auth/MsgUpdateParams")

	legacytx.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces associates protoName with AccountI interface
// and creates a registry of it's concrete implementations
func RegisterInterfaces(registrar registry.InterfaceRegistrar) {
	registrar.RegisterInterface(
		"cosmos.auth.v1beta1.AccountI",
		(*AccountI)(nil),
		&BaseAccount{},
		&ModuleAccount{},
	)

	registrar.RegisterInterface(
		"cosmos.auth.v1beta1.AccountI",
		(*sdk.AccountI)(nil),
		&BaseAccount{},
		&ModuleAccount{},
	)

	registrar.RegisterInterface(
		"cosmos.auth.v1beta1.GenesisAccount",
		(*GenesisAccount)(nil),
		&BaseAccount{},
		&ModuleAccount{},
	)

	registrar.RegisterInterface(
		"cosmos.auth.v1.ModuleCredential",
		(*cryptotypes.PubKey)(nil),
		&ModuleCredential{},
	)

	registrar.RegisterImplementations((*coretransaction.Msg)(nil),
		&MsgUpdateParams{},
		&MsgNonAtomicExec{},
	)
}
