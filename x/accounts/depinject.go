package accounts

import (
	modulev1 "cosmossdk.io/api/cosmos/accounts/module/v1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
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

	Cdc            codec.Codec
	Environment    appmodule.Environment
	SignerProvider SignerProvider
	AddressCodec   address.Codec
	ExecRouter     MsgRouter
	QueryRouter    QueryRouter
	Registry       cdctypes.InterfaceRegistry
}

type ModuleOutputs struct {
	depinject.Out

	AccountsKeeper Keeper
	Module         appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	accountskeeper, err := NewKeeper(
		in.Cdc, in.Environment, in.AddressCodec,
		in.SignerProvider, in.ExecRouter, in.QueryRouter, in.Registry, nil,
	)
	if err != nil {
		panic(err)
	}
	m := NewAppModule(in.Cdc, accountskeeper)
	return ModuleOutputs{AccountsKeeper: accountskeeper, Module: m}
}
