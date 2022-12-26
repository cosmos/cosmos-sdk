package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// AccountKeeper defines the account contract that must be fulfilled when
// creating a x/bank keeper.
type AccountKeeper interface {
	NewAccount(sdk.Context, types.AccountAliasI) types.AccountAliasI
	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) types.AccountAliasI

	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountAliasI
	GetAllAccounts(ctx sdk.Context) []types.AccountAliasI
	HasAccount(ctx sdk.Context, addr sdk.AccAddress) bool
	SetAccount(ctx sdk.Context, acc types.AccountAliasI)

	IterateAccounts(ctx sdk.Context, process func(types.AccountAliasI) bool)

	ValidatePermissions(macc types.ModuleAccountI) error

	GetModuleAddress(moduleName string) sdk.AccAddress
	GetModuleAddressAndPermissions(moduleName string) (addr sdk.AccAddress, permissions []string)
	GetModuleAccountAndPermissions(ctx sdk.Context, moduleName string) (types.ModuleAccountI, []string)
	GetModuleAccount(ctx sdk.Context, moduleName string) types.ModuleAccountI
	SetModuleAccount(ctx sdk.Context, macc types.ModuleAccountI)
	GetModulePermissions() map[string]types.PermissionsForAddress
}
