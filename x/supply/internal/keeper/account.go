package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply/exported"
	"github.com/cosmos/cosmos-sdk/x/supply/internal/types"
)

// GetModuleAddress returns a an address  based on the name
func (k Keeper) GetModuleAddress(moduleName string) sdk.AccAddress {
	permAddr, ok := k.permAddrs[moduleName]
	if !ok {
		return nil
	}
	return permAddr.address
}

// GetModuleAddressAndPermission returns an address and permission  based on the name
func (k Keeper) GetModuleAddressAndPermission(moduleName string) (addr sdk.AccAddress, permission string) {
	permAddr, ok := k.permAddrs[moduleName]
	if !ok {
		return nil, ""
	}
	return permAddr.address, permAddr.permission
}

// GetModuleAccount gets the module account to the auth account store
func (k Keeper) GetModuleAccountAndPermission(ctx sdk.Context, moduleName string) (exported.ModuleAccountI, string) {
	addr, perm := k.GetModuleAddressAndPermission(moduleName)
	if addr == nil {
		return nil, ""
	}

	acc := k.ak.GetAccount(ctx, addr)
	if acc != nil {
		macc, ok := acc.(exported.ModuleAccountI)
		if !ok {
			panic("account is not a module account")
		}
		return macc, perm
	}

	// create a new module account
	macc := types.NewEmptyModuleAccount(moduleName, perm)
	maccI := (k.ak.NewAccount(ctx, macc)).(exported.ModuleAccountI) // set the account number
	k.SetModuleAccount(ctx, maccI)

	return maccI, perm
}

// GetModuleAccount gets the module account to the auth account store
func (k Keeper) GetModuleAccount(ctx sdk.Context, moduleName string) exported.ModuleAccountI {
	acc, _ := k.GetModuleAccountAndPermission(ctx, moduleName)
	return acc
}

// SetModuleAccount sets the module account to the auth account store
func (k Keeper) SetModuleAccount(ctx sdk.Context, macc exported.ModuleAccountI) {
	k.ak.SetAccount(ctx, macc)
}
