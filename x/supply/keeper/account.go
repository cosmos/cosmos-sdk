package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply/types"
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
func (k Keeper) GetModuleAccount(ctx sdk.Context, moduleName string) types.ModuleAccountI {
	addr := k.GetModuleAddress(moduleName)
	if addr == nil {
		return nil
	}

	acc := k.ak.GetAccount(ctx, addr)
	macc, ok := acc.(types.ModuleAccountI)
	if !ok {
		return nil
	}

	return macc
}

// SetModuleAccount sets the module account to the auth account store
func (k Keeper) SetModuleAccount(ctx sdk.Context, macc types.ModuleAccountI) {
	k.ak.SetAccount(ctx, macc)
}

// GetCoins alias for bank keeper
func (k Keeper) GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return k.bk.GetCoins(ctx, addr)
}
