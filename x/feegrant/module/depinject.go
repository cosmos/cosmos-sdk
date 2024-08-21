package module

import (
	modulev1 "cosmossdk.io/api/cosmos/feegrant/module/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"cosmossdk.io/x/auth/ante"
	"cosmossdk.io/x/feegrant"
	"cosmossdk.io/x/feegrant/keeper"

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

type FeegrantInputs struct {
	depinject.In

	Environment   appmodule.Environment
	Cdc           codec.Codec
	AccountKeeper feegrant.AccountKeeper
	BankKeeper    feegrant.BankKeeper
	Registry      cdctypes.InterfaceRegistry

	FeeTxValidator ante.FeeTxValidator `optional:"true"` // server v2
}

func ProvideModule(in FeegrantInputs) (keeper.Keeper, appmodule.AppModule) {
	k := keeper.NewKeeper(in.Environment, in.Cdc, in.AccountKeeper)
	m := NewAppModule(in.Cdc, in.AccountKeeper, in.BankKeeper, k, in.Registry)

	if in.FeeTxValidator != nil {
		feeTxValidator := in.FeeTxValidator.SetFeegrantKeeper(k)
		m.SetFeeTxValidator(feeTxValidator)
	}

	return k, m
}
