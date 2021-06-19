package module

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

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

func (v viewOnlyKeeper) GetParams(ctx sdk.Context) (params types.Params) {
	return v.k.GetParams(ctx)
}

var _ types.ViewKeeper = viewOnlyKeeper{}
