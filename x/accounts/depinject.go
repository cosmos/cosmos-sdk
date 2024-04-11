package accounts

import (
	"context"

	modulev1 "cosmossdk.io/api/cosmos/accounts/module/v1"
	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	baseaccount "cosmossdk.io/x/accounts/defaults/base"
	"cosmossdk.io/x/tx/signing"

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
	ExecRouter   MsgRouter
	QueryRouter  QueryRouter
	Registry     cdctypes.InterfaceRegistry
}

type ModuleOutputs struct {
	depinject.Out

	AccountsKeeper Keeper
	Module         appmodule.AppModule
}

var _ signing.SignModeHandler = directHandler{}

type directHandler struct{}

func (s directHandler) Mode() signingv1beta1.SignMode {
	return signingv1beta1.SignMode_SIGN_MODE_DIRECT
}

func (s directHandler) GetSignBytes(_ context.Context, _ signing.SignerData, _ signing.TxData) ([]byte, error) {
	panic("not implemented")
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	handler := directHandler{}
	account := baseaccount.NewAccount("base", signing.NewHandlerMap(handler))
	accountskeeper, err := NewKeeper(
		in.Cdc, in.Environment, in.AddressCodec, in.Cdc,
		in.ExecRouter, in.QueryRouter, in.Registry, account,
	)
	if err != nil {
		panic(err)
	}
	m := NewAppModule(in.Cdc, accountskeeper)
	return ModuleOutputs{AccountsKeeper: accountskeeper, Module: m}
}
