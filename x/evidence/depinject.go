package evidence

import (
	modulev1 "cosmossdk.io/api/cosmos/evidence/module/v1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"cosmossdk.io/x/evidence/keeper"
	"cosmossdk.io/x/evidence/types"

	"github.com/cosmos/cosmos-sdk/codec"
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

	StoreService store.KVStoreService
	Cdc          codec.Codec

	StakingKeeper  types.StakingKeeper
	SlashingKeeper types.SlashingKeeper
	AddressCodec   address.Codec
}

type ModuleOutputs struct {
	depinject.Out

	EvidenceKeeper keeper.Keeper
	Module         appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	k := keeper.NewKeeper(in.Cdc, in.StoreService, in.StakingKeeper, in.SlashingKeeper, in.AddressCodec)
	m := NewAppModule(*k)

	return ModuleOutputs{EvidenceKeeper: *k, Module: m}
}
