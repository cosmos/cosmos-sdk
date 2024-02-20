package module

import (
	modulev1 "cosmossdk.io/api/cosmos/nft/module/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"cosmossdk.io/x/nft"
	"cosmossdk.io/x/nft/keeper"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
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

	AccountKeeper nft.AccountKeeper
	BankKeeper    nft.BankKeeper
}

type ModuleOutputs struct {
	depinject.Out

	NFTKeeper keeper.Keeper
	Module    appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	k := keeper.NewKeeper(in.Environment, in.Cdc, in.AccountKeeper, in.BankKeeper)
	m := NewAppModule(in.Cdc, k, in.AccountKeeper, in.BankKeeper, in.Registry)

	return ModuleOutputs{NFTKeeper: k, Module: m}
}
