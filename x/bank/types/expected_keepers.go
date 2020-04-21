package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
)

// AccountKeeper defines the account contract that must be fulfilled when
// creating a x/bank keeper.
type AccountKeeper interface {
	NewAccount(sdk.Context, exported.Account) exported.Account
	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) exported.Account

	GetAccount(ctx sdk.Context, addr sdk.AccAddress) exported.Account
	GetAllAccounts(ctx sdk.Context) []exported.Account
	SetAccount(ctx sdk.Context, acc exported.Account)

	IterateAccounts(ctx sdk.Context, process func(exported.Account) bool)

	ValidatePermissions(macc exported.ModuleAccountI) error

	GetModuleAddress(moduleName string) sdk.AccAddress
	GetModuleAddressAndPermissions(moduleName string) (addr sdk.AccAddress, permissions []string)
	GetModuleAccountAndPermissions(ctx sdk.Context, moduleName string) (exported.ModuleAccountI, []string)
	GetModuleAccount(ctx sdk.Context, moduleName string) exported.ModuleAccountI
	SetModuleAccount(ctx sdk.Context, macc exported.ModuleAccountI)
}
