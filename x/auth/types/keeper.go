package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ViewKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) AccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
	ValidatePermissions(macc ModuleAccountI) error
	GetModuleAddressAndPermissions(moduleName string) (addr sdk.AccAddress, permissions []string)
	GetModuleAccountAndPermissions(ctx sdk.Context, moduleName string) (ModuleAccountI, []string)
	GetModuleAccount(ctx sdk.Context, moduleName string) ModuleAccountI
	GetParams(ctx sdk.Context) (params Params)
}

type Keeper interface {
	ViewKeeper

	NewAccount(sdk.Context, AccountI) AccountI
	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) AccountI
	SetAccount(ctx sdk.Context, acc AccountI)
	SetModuleAccount(ctx sdk.Context, macc ModuleAccountI)
}
