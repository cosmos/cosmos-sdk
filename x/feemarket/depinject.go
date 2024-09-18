package feemarket

import (
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	modulev1 "cosmossdk.io/api/cosmos/feemarket/module/v1"
	"cosmossdk.io/x/feemarket/keeper"
	"cosmossdk.io/x/feemarket/types"
)

const govModuleName = "gov"

func init() {
	appconfig.Register(
		&modulev1.Module{},
		appconfig.Provide(ProvideModule),
	)
}

type Inputs struct {
	depinject.In

	Config        *modulev1.Module
	Cdc           codec.Codec
	Env           appmodule.Environment
	AccountKeeper types.AccountKeeper
}

type Outputs struct {
	depinject.Out

	Keeper keeper.Keeper
	Module appmodule.AppModule
}

func ProvideModule(in Inputs) Outputs {
	var (
		authority sdk.AccAddress
		err       error
	)
	if in.Config.Authority != "" {
		authority, err = sdk.AccAddressFromBech32(in.Config.Authority)
		if err != nil {
			panic(err)
		}
	} else {
		authority = authtypes.NewModuleAddress(govModuleName)
	}

	Keeper := keeper.NewKeeper(
		in.Cdc,
		in.Env,
		in.AccountKeeper,
		nil,
		authority.String(),
	)

	m := NewAppModule(in.Cdc, *Keeper)

	return Outputs{Keeper: *Keeper, Module: m}
}
