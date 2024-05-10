package types

import (
	"cosmossdk.io/core/registry"
	coretransaction "cosmossdk.io/core/transaction"
	"cosmossdk.io/x/auth/migrations/legacytx"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterLegacyAminoCodec registers the account interfaces and concrete types on the
// provided LegacyAmino codec. These types are used for Amino JSON serialization
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*sdk.ModuleAccountI)(nil), nil)
	cdc.RegisterInterface((*GenesisAccount)(nil), nil)
	cdc.RegisterInterface((*sdk.AccountI)(nil), nil)
	cdc.RegisterConcrete(&BaseAccount{}, "cosmos-sdk/BaseAccount", nil)
	cdc.RegisterConcrete(&ModuleAccount{}, "cosmos-sdk/ModuleAccount", nil)
	cdc.RegisterConcrete(Params{}, "cosmos-sdk/x/auth/Params", nil)
	cdc.RegisterConcrete(&ModuleCredential{}, "cosmos-sdk/GroupAccountCredential", nil)

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
