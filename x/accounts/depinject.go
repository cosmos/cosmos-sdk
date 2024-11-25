package accounts

import (
	modulev1 "cosmossdk.io/api/cosmos/accounts/module/v1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"cosmossdk.io/x/accounts/accountstd"
	txdecode "cosmossdk.io/x/tx/decode"

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

	Cdc          codec.Codec
	Environment  appmodule.Environment
	AddressCodec address.Codec
	Registry     cdctypes.InterfaceRegistry

	Accounts []accountstd.DepinjectAccount // at least one account must be provided
}

type ModuleOutputs struct {
	depinject.Out

	AccountsKeeper Keeper
	Module         appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	accCreators := make([]accountstd.AccountCreatorFunc, len(in.Accounts))
	for i, acc := range in.Accounts {
		accCreators[i] = acc.MakeAccount
	}

	txDec, err := txdecode.NewDecoder(txdecode.Options{
		SigningContext: in.Registry.SigningContext(),
		ProtoCodec:     in.Cdc,
	})
	if err != nil {
		panic(err)
	}

	accountsKeeper, err := NewKeeper(
		in.Cdc, in.Environment, in.AddressCodec, in.Registry, txDec,
		accCreators...,
	)
	if err != nil {
		panic(err)
	}
	m := NewAppModule(in.Cdc, accountsKeeper)
	return ModuleOutputs{AccountsKeeper: accountsKeeper, Module: m}
}
