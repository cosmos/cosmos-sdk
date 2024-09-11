package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AuthKeeper defines the account contract that must be fulfilled when
// creating a x/bank keeper.
type AuthKeeper interface {
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
}
