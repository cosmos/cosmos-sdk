package genutil

import (
	modulev1 "cosmossdk.io/api/cosmos/genutil/module/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/genesis"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

var _ depinject.OnePerModuleType = AppModule{}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

func init() {
	appconfig.RegisterModule(&modulev1.Module{},
		appconfig.Provide(ProvideModule),
	)
}

// ModuleInputs defines the inputs needed for the genutil module.
type ModuleInputs struct {
	depinject.In

	AccountKeeper  types.AccountKeeper
	StakingKeeper  types.StakingKeeper
	DeliverTx      genesis.TxHandler
	Config         client.TxConfig
	Cdc            codec.Codec
	GenTxValidator types.MessageValidator `optional:"true"`
}

func ProvideModule(in ModuleInputs) appmodule.AppModule {
	if in.GenTxValidator == nil {
		in.GenTxValidator = types.DefaultMessageValidator
	}

	m := NewAppModule(in.Cdc, in.AccountKeeper, in.StakingKeeper, in.DeliverTx, in.Config, in.GenTxValidator)
	return m
}
