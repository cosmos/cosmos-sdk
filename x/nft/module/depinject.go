package module

import (
	modulev1 "cosmossdk.io/api/cosmos/nft/module/v1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"cosmossdk.io/x/nft"
	"cosmossdk.io/x/nft/keeper"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
)

var _ depinject.OnePerModuleType = AppModule{}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

func init() {
	appconfig.RegisterModule(&modulev1.Module{},
		appconfig.Provide(ProvideModule),
	)
}

type ModuleInputs struct {
	depinject.In

	Environment appmodule.Environment
	Cdc         codec.Codec
	Registry    cdctypes.InterfaceRegistry
	AddressCdc  address.Codec

	BankKeeper nft.BankKeeper
}

type ModuleOutputs struct {
	depinject.Out

	NFTKeeper      keeper.Keeper
	Module         appmodule.AppModule
	ModuleAccounts []runtime.ModuleAccount
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	k := keeper.NewKeeper(in.Environment, in.Cdc, in.BankKeeper, in.AddressCdc)
	m := NewAppModule(in.Cdc, k, in.BankKeeper, in.Registry, in.AddressCdc)

	return ModuleOutputs{NFTKeeper: k, Module: m, ModuleAccounts: []runtime.ModuleAccount{runtime.NewModuleAccount(nft.ModuleName)}}
}
