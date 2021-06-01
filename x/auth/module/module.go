package module

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type Inputs struct {
	Codec                   codec.Codec
	StoreKey                sdk.StoreKey
	ParamStore              paramtypes.Subspace
	RandomGenesisAccountsFn types.RandomGenesisAccountsFn
}

type Outputs struct {
	ViewKeeper types.ViewKeeper
	Keeper     types.Keeper `security:"admin"`
}

func (m Module) NewAppModule(inputs Inputs) (module.AppModule, Outputs, error) {
	newAccFn := func() types.AccountI {
		return &types.BaseAccount{}
	}
	if m.AccountConstructor != nil {
		var accCtr types.AccountConstructor
		err := inputs.Codec.UnpackAny(m.AccountConstructor, &accCtr)
		if err != nil {
			return nil, Outputs{}, err
		}
		newAccFn = func() types.AccountI {
			return accCtr.NewAccount()
		}
	}

	perms := map[string][]string{}
	for _, perm := range m.Permissions {
		perms[perm.Address] = perm.Permissions
	}

	keeper := authkeeper.NewAccountKeeper(inputs.Codec, inputs.StoreKey, inputs.ParamStore, newAccFn, perms)
	appMod := auth.NewAppModule(inputs.Codec, keeper, inputs.RandomGenesisAccountsFn)

	return appMod, Outputs{
		ViewKeeper: viewOnlyKeeper{keeper},
		Keeper:     keeper,
	}, nil
}

// viewOnlyKeeper wraps the full keeper in a view-only interface which can't be easily type cast to the full keeper interface
type viewOnlyKeeper struct {
	k authkeeper.AccountKeeper
}

func (v viewOnlyKeeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI {
	return v.k.GetAccount(ctx, addr)
}

func (v viewOnlyKeeper) GetModuleAddress(moduleName string) sdk.AccAddress {
	return v.k.GetModuleAddress(moduleName)
}

func (v viewOnlyKeeper) ValidatePermissions(macc types.ModuleAccountI) error {
	return v.k.ValidatePermissions(macc)
}

func (v viewOnlyKeeper) GetModuleAddressAndPermissions(moduleName string) (addr sdk.AccAddress, permissions []string) {
	return v.GetModuleAddressAndPermissions(moduleName)
}

func (v viewOnlyKeeper) GetModuleAccountAndPermissions(ctx sdk.Context, moduleName string) (types.ModuleAccountI, []string) {
	return v.GetModuleAccountAndPermissions(ctx, moduleName)
}

func (v viewOnlyKeeper) GetModuleAccount(ctx sdk.Context, moduleName string) types.ModuleAccountI {
	return v.k.GetModuleAccount(ctx, moduleName)
}

var _ types.ViewKeeper = viewOnlyKeeper{}
