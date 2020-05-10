package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
)

// RegisterCodec registers the account interfaces and concrete types on the
// provided Amino codec.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*ModuleAccountI)(nil), nil)
	cdc.RegisterInterface((*GenesisAccount)(nil), nil)
	cdc.RegisterInterface((*AccountI)(nil), nil)
	cdc.RegisterConcrete(&BaseAccount{}, "cosmos-sdk/AccountI", nil)
	cdc.RegisterConcrete(&ModuleAccount{}, "cosmos-sdk/ModuleAccount", nil)
	cdc.RegisterConcrete(StdTx{}, "cosmos-sdk/StdTx", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterInterface(
		"cosmos_sdk.auth.v1.auth",
		(*AccountI)(nil),
		&BaseAccount{},
		&ModuleAccount{},
		//&vesting.DelayedVestingAccount{},
	)
}

// RegisterKeyTypeCodec registers an external concrete type defined in
// another module for the internal ModuleCdc.
func RegisterKeyTypeCodec(o interface{}, name string) {
	amino.RegisterConcrete(o, name, nil)
}

var (
	amino = codec.New()

	ModuleCdc = codec.NewHybridCodec(amino, types.NewInterfaceRegistry())
)

func init() {
	RegisterCodec(amino)
	codec.RegisterCrypto(amino)
}
