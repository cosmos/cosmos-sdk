package types

import (
	"context"

	"cosmossdk.io/core/address"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// AccountKeeper defines the account contract that must be fulfilled when
// creating a x/bank keeper.
type AccountKeeper interface {
	AddressCodec() address.Codec

	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	HasAccount(ctx context.Context, addr sdk.AccAddress) bool
	SetAccount(ctx context.Context, acc sdk.AccountI)

	GetModuleAddress(moduleName string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
	GetModulePermissions() map[string]types.PermissionsForAddress
}
