package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"
)

// RegisterLegacyAminoCodec registers the account interfaces and concrete types on the
// provided LegacyAmino codec. These types are used for Amino JSON serialization
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*ModuleAccountI)(nil), nil)
	cdc.RegisterInterface((*GenesisAccount)(nil), nil)
	cdc.RegisterInterface((*AccountI)(nil), nil)
	cdc.RegisterConcrete(&BaseAccount{}, "cosmos-sdk/BaseAccount", nil)
	cdc.RegisterConcrete(&ModuleAccount{}, "cosmos-sdk/ModuleAccount", nil)

	legacytx.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces associates protoName with AccountI interface
// and creates a registry of it's concrete implementations
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterInterface(
		"cosmos.auth.v1beta1.AccountI",
		(*AccountI)(nil),
		&BaseAccount{},
		&ModuleAccount{},
	)

	registry.RegisterInterface(
		"cosmos.auth.v1beta1.GenesisAccount",
		(*GenesisAccount)(nil),
		&BaseAccount{},
		&ModuleAccount{},
	)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	sdk.RegisterLegacyAminoCodec(amino)

	// Register all Amino interfaces and concrete types on the authz Amino codec so that this can later be
	// used to properly serialize MsgGrant and MsgExec instances
	RegisterLegacyAminoCodec(authzcodec.Amino)
}
