package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// AccountKeeper defines the account contract that must be fulfilled when
// creating a x/bank keeper.
type AccountKeeper interface {
	NewAccount(sdk.Context, types.IAccount) types.IAccount
	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) types.IAccount

	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.IAccount
	GetAllAccounts(ctx sdk.Context) []types.IAccount
	HasAccount(ctx sdk.Context, addr sdk.AccAddress) bool
	SetAccount(ctx sdk.Context, acc types.IAccount)

	IterateAccounts(ctx sdk.Context, process func(types.IAccount) bool)

	ValidatePermissions(macc types.IModuleAccount) error

	GetModuleAddress(moduleName string) sdk.AccAddress
	GetModuleAddressAndPermissions(moduleName string) (addr sdk.AccAddress, permissions []string)
	GetModuleAccountAndPermissions(ctx sdk.Context, moduleName string) (types.IModuleAccount, []string)
	GetModuleAccount(ctx sdk.Context, moduleName string) types.IModuleAccount
	SetModuleAccount(ctx sdk.Context, macc types.IModuleAccount)
	GetModulePermissions() map[string]types.PermissionsForAddress
}
