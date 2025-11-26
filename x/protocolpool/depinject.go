package protocolpool

import (
	modulev1 "cosmossdk.io/api/cosmos/protocolpool/module/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/keeper"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

var _ depinject.OnePerModuleType = AppModule{}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

func init() {
	appconfig.RegisterModule(
		&modulev1.Module{},
		appconfig.Provide(ProvideModule),
	)
}

type ModuleInputs struct {
	depinject.In

	Config       *modulev1.Module
	Codec        codec.Codec
	StoreService store.KVStoreService

	AccountKeeper types.AccountKeeper
	BankKeeper    types.BankKeeper
}

type ModuleOutputs struct {
	depinject.Out

	Keeper keeper.Keeper
	Module appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	k := keeper.NewKeeper(in.Codec, in.StoreService, in.AccountKeeper, in.BankKeeper)
	m := NewAppModule(k, in.AccountKeeper, in.BankKeeper)

	return ModuleOutputs{
		Keeper: k,
		Module: m,
	}
}
